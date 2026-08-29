package state

import (
	json "encoding/json/v2"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
	"uuid"

	"github.com/daniel-oluwadunsin/undolang/internal/journal"
)

type Status string

const (
	Planned          Status = "PLANNED"
	Prepared         Status = "PREPARED"
	Running          Status = "RUNNING"
	Verifying        Status = "VERIFYING"
	Committing       Status = "COMMITTING"
	Committed        Status = "COMMITTED"
	RollingBack      Status = "ROLLING_BACK"
	RolledBack       Status = "ROLLED_BACK"
	RecoveryRequired Status = "RECOVERY_REQUIRED"
	RecoveryFailed   Status = "RECOVERY_FAILED"
)

var transitions = map[Status]map[Status]bool{
	Planned: {Prepared: true}, Prepared: {Running: true, RollingBack: true, RecoveryRequired: true}, Running: {Verifying: true, RollingBack: true, RecoveryRequired: true}, Verifying: {Committing: true, RollingBack: true, RecoveryRequired: true}, Committing: {Committed: true, RollingBack: true, RecoveryRequired: true}, RollingBack: {RolledBack: true, RecoveryFailed: true}, RecoveryRequired: {RollingBack: true, RecoveryFailed: true}, RecoveryFailed: {RollingBack: true},
}

func ValidateTransition(from, to Status) error {
	if transitions[from][to] {
		return nil
	}
	return fmt.Errorf("invalid transaction state transition %s -> %s", from, to)
}
func Terminal(s Status) bool { return s == Committed || s == RolledBack }

type Metadata struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	ScriptPath   string    `json:"script_path"`
	ScriptSHA256 string    `json:"script_sha256"`
	Root         string    `json:"root"`
	StartedAt    time.Time `json:"started_at"`
	Status       Status    `json:"status"`
}
type Lock struct {
	TransactionID string    `json:"transaction_id"`
	PID           int       `json:"pid"`
	StartedAt     time.Time `json:"started_at"`
	ScriptSHA256  string    `json:"script_sha256"`
}
type Store struct{ root, stateDir, transactionsDir string }

func Open(root string) (*Store, error) {
	absolute, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	return &Store{root: absolute, stateDir: filepath.Join(absolute, ".undo"), transactionsDir: filepath.Join(absolute, ".undo", "transactions")}, nil
}
func (s *Store) Ensure() error {
	for _, dir := range []string{s.stateDir, s.transactionsDir, filepath.Join(s.stateDir, "history")} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return err
		}
		if err := os.Chmod(dir, 0o700); err != nil {
			return err
		}
	}
	return syncDir(s.root)
}

type Transaction struct {
	Meta                        Metadata
	Dir, BackupDir, JournalPath string
	Journal                     *journal.Appender
	store                       *Store
}

func (s *Store) Begin(name, scriptPath, scriptHash string) (*Transaction, error) {
	if err := s.Ensure(); err != nil {
		return nil, err
	}
	id := uuid.NewV7().String()
	now := time.Now().UTC()
	meta := Metadata{ID: id, Name: name, ScriptPath: scriptPath, ScriptSHA256: scriptHash, Root: s.root, StartedAt: now, Status: Planned}
	lockPath := filepath.Join(s.stateDir, "active.lock")
	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, err
	}
	lock := Lock{TransactionID: id, PID: os.Getpid(), StartedAt: now, ScriptSHA256: scriptHash}
	data, err := json.Marshal(&lock)
	if err == nil {
		err = writeAndSync(lockFile, data)
	}
	closeErr := lockFile.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		_ = os.Remove(lockPath)
		return nil, err
	}
	dir := filepath.Join(s.transactionsDir, id)
	backup := filepath.Join(dir, "backup")
	if err = os.MkdirAll(backup, 0o700); err != nil {
		return nil, err
	}
	if err = writeAtomic(filepath.Join(dir, "meta.json"), meta); err != nil {
		return nil, err
	}
	if err = writeAtomic(filepath.Join(dir, "status"), []byte(meta.Status)); err != nil {
		return nil, err
	}
	journalPath := filepath.Join(dir, "journal.bin")
	appender, err := journal.Create(journalPath)
	if err != nil {
		return nil, err
	}
	if _, err = appender.Append(journal.TXBegin, journal.Payload{TransactionID: id, State: string(Planned)}); err != nil {
		appender.Close()
		return nil, err
	}
	return &Transaction{Meta: meta, Dir: dir, BackupDir: backup, JournalPath: journalPath, Journal: appender, store: s}, nil
}

func (t *Transaction) Transition(to Status) error {
	if err := ValidateTransition(t.Meta.Status, to); err != nil {
		return err
	}
	if _, err := t.Journal.Append(journal.TXState, journal.Payload{TransactionID: t.Meta.ID, State: string(to)}); err != nil {
		return err
	}
	t.Meta.Status = to
	if err := writeAtomic(filepath.Join(t.Dir, "status"), []byte(to)); err != nil {
		return err
	}
	return writeAtomic(filepath.Join(t.Dir, "meta.json"), t.Meta)
}
func (t *Transaction) Close() error {
	if t.Journal == nil {
		return nil
	}
	err := t.Journal.Close()
	t.Journal = nil
	return err
}
func (t *Transaction) Release() error {
	if !Terminal(t.Meta.Status) {
		return errors.New("cannot release active lock before terminal status")
	}
	if err := t.Close(); err != nil {
		return err
	}
	if err := os.Remove(filepath.Join(t.store.stateDir, "active.lock")); err != nil {
		return err
	}
	return syncDir(t.store.stateDir)
}

func (s *Store) Inspect(id string) (Metadata, error) {
	var meta Metadata
	data, err := os.ReadFile(filepath.Join(s.transactionsDir, id, "meta.json"))
	if err != nil {
		return meta, err
	}
	if err = json.Unmarshal(data, &meta); err != nil {
		return meta, err
	}
	if meta.ID != id {
		return Metadata{}, fmt.Errorf("transaction metadata id mismatch")
	}
	return meta, nil
}
func (s *Store) History() ([]Metadata, error) {
	entries, err := os.ReadDir(s.transactionsDir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []Metadata
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		meta, err := s.Inspect(entry.Name())
		if err != nil {
			return nil, err
		}
		out = append(out, meta)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].StartedAt.Before(out[j].StartedAt) })
	return out, nil
}

func writeAtomic(path string, value any) error {
	var data []byte
	var err error
	switch v := value.(type) {
	case []byte:
		data = v
	default:
		data, err = json.Marshal(&v)
		if err != nil {
			return err
		}
	}
	dir := filepath.Dir(path)
	temp, err := os.CreateTemp(dir, ".undo-state-")
	if err != nil {
		return err
	}
	name := temp.Name()
	defer os.Remove(name)
	if err = temp.Chmod(0o600); err == nil {
		err = writeAndSync(temp, data)
	}
	closeErr := temp.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	if err = os.Rename(name, path); err != nil {
		return err
	}
	return syncDir(dir)
}
func writeAndSync(file *os.File, data []byte) error {
	for len(data) > 0 {
		n, err := file.Write(data)
		if err != nil {
			return err
		}
		if n == 0 {
			return errors.New("short write")
		}
		data = data[n:]
	}
	return file.Sync()
}
func syncDir(path string) error {
	dir, err := os.Open(path)
	if err != nil {
		return err
	}
	defer dir.Close()
	return dir.Sync()
}
