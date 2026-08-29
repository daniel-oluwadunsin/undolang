package journal

import (
	"encoding/binary"
	json "encoding/json/v2"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"os"
)

const (
	Version    uint16 = 1
	MaxPayload        = 1 << 20
	headerSize        = 20
)

var (
	magic      = [4]byte{'U', 'N', 'D', 'O'}
	castagnoli = crc32.MakeTable(crc32.Castagnoli)
	ErrCorrupt = errors.New("E_JOURNAL_CORRUPT")
)

type Type uint16

const (
	TXBegin Type = iota + 1
	TXState
	OPPrepared
	OPApplied
	AssertResult
	RollbackPrepared
	RollbackApplied
	TXCommit
	TXRollbackComplete
)

func validType(t Type) bool { return t >= TXBegin && t <= TXRollbackComplete }

type Payload struct {
	TransactionID string `json:"transaction_id"`
	OperationID   string `json:"operation_id,omitempty"`
	State         string `json:"state,omitempty"`
	Data          any    `json:"data,omitempty"`
}

type Record struct {
	Type     Type
	Sequence uint64
	Payload  []byte
}

type Appender struct {
	file *os.File
	next uint64
}

func Create(path string) (*Appender, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, err
	}
	return &Appender{file: file, next: 1}, nil
}

func OpenAppend(path string) (*Appender, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND, 0o600)
	if err != nil {
		return nil, err
	}
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		file.Close()
		return nil, err
	}
	replay, err := Decode(file)
	if err != nil {
		file.Close()
		return nil, err
	}
	if replay.TornTail {
		file.Close()
		return nil, fmt.Errorf("%w: journal has torn tail", ErrCorrupt)
	}
	return &Appender{file: file, next: uint64(len(replay.Records) + 1)}, nil
}

func (a *Appender) Append(kind Type, payload Payload) (Record, error) {
	if a == nil || a.file == nil {
		return Record{}, errors.New("journal is closed")
	}
	if !validType(kind) {
		return Record{}, fmt.Errorf("%w: unknown record type %d", ErrCorrupt, kind)
	}
	data, err := json.Marshal(&payload)
	if err != nil {
		return Record{}, err
	}
	if len(data) > MaxPayload {
		return Record{}, fmt.Errorf("%w: payload exceeds %d bytes", ErrCorrupt, MaxPayload)
	}
	header := make([]byte, headerSize)
	copy(header[:4], magic[:])
	binary.BigEndian.PutUint16(header[4:6], Version)
	binary.BigEndian.PutUint16(header[6:8], uint16(kind))
	binary.BigEndian.PutUint64(header[8:16], a.next)
	binary.BigEndian.PutUint32(header[16:20], uint32(len(data)))
	checksum := crc32.New(castagnoli)
	_, _ = checksum.Write(header[4:])
	_, _ = checksum.Write(data)
	crc := make([]byte, 4)
	binary.BigEndian.PutUint32(crc, checksum.Sum32())
	if err = writeFull(a.file, header); err == nil {
		err = writeFull(a.file, data)
	}
	if err == nil {
		err = writeFull(a.file, crc)
	}
	if err == nil {
		err = a.file.Sync()
	}
	if err != nil {
		return Record{}, err
	}
	record := Record{Type: kind, Sequence: a.next, Payload: data}
	a.next++
	return record, nil
}
func writeFull(w io.Writer, p []byte) error {
	for len(p) > 0 {
		n, err := w.Write(p)
		if err != nil {
			return err
		}
		if n == 0 {
			return io.ErrShortWrite
		}
		p = p[n:]
	}
	return nil
}
func (a *Appender) Close() error {
	if a == nil || a.file == nil {
		return nil
	}
	err := a.file.Close()
	a.file = nil
	return err
}

type Replay struct {
	Records       []Record
	TornTail      bool
	TransactionID string
}

func Decode(r io.Reader) (Replay, error) {
	var result Replay
	expected := uint64(1)
	prepared := map[string]bool{}
	applied := map[string]bool{}
	rollbackPrepared := map[string]bool{}
	for {
		header := make([]byte, headerSize)
		n, err := io.ReadFull(r, header)
		if errors.Is(err, io.EOF) && n == 0 {
			if len(result.Records) == 0 {
				return Replay{}, fmt.Errorf("%w: zero-byte journal", ErrCorrupt)
			}
			return result, nil
		}
		if errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, io.EOF) {
			if len(result.Records) == 0 {
				return Replay{}, fmt.Errorf("%w: torn first frame", ErrCorrupt)
			}
			result.TornTail = true
			return result, nil
		}
		if err != nil {
			return Replay{}, err
		}
		if string(header[:4]) != string(magic[:]) {
			return Replay{}, fmt.Errorf("%w: bad magic", ErrCorrupt)
		}
		if binary.BigEndian.Uint16(header[4:6]) != Version {
			return Replay{}, fmt.Errorf("%w: unsupported version", ErrCorrupt)
		}
		kind := Type(binary.BigEndian.Uint16(header[6:8]))
		if !validType(kind) {
			return Replay{}, fmt.Errorf("%w: unknown record type", ErrCorrupt)
		}
		sequence := binary.BigEndian.Uint64(header[8:16])
		if sequence != expected {
			return Replay{}, fmt.Errorf("%w: sequence %d, expected %d", ErrCorrupt, sequence, expected)
		}
		length := binary.BigEndian.Uint32(header[16:20])
		if length > MaxPayload {
			return Replay{}, fmt.Errorf("%w: payload length exceeds cap", ErrCorrupt)
		}
		body := make([]byte, int(length)+4)
		n, err = io.ReadFull(r, body)
		if errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, io.EOF) {
			if len(result.Records) == 0 {
				return Replay{}, fmt.Errorf("%w: torn first payload", ErrCorrupt)
			}
			result.TornTail = true
			return result, nil
		}
		if err != nil {
			return Replay{}, err
		}
		if n != len(body) {
			return Replay{}, io.ErrUnexpectedEOF
		}
		checksum := crc32.New(castagnoli)
		_, _ = checksum.Write(header[4:])
		_, _ = checksum.Write(body[:length])
		if binary.BigEndian.Uint32(body[length:]) != checksum.Sum32() {
			return Replay{}, fmt.Errorf("%w: checksum mismatch", ErrCorrupt)
		}
		var payload Payload
		if err = json.Unmarshal(body[:length], &payload); err != nil {
			return Replay{}, fmt.Errorf("%w: invalid payload: %v", ErrCorrupt, err)
		}
		if expected == 1 && kind != TXBegin {
			return Replay{}, fmt.Errorf("%w: first record is not TX_BEGIN", ErrCorrupt)
		}
		if result.TransactionID == "" {
			result.TransactionID = payload.TransactionID
		} else if payload.TransactionID != result.TransactionID {
			return Replay{}, fmt.Errorf("%w: transaction id mismatch", ErrCorrupt)
		}
		if err = validateReference(kind, payload.OperationID, prepared, applied, rollbackPrepared); err != nil {
			return Replay{}, err
		}
		result.Records = append(result.Records, Record{Type: kind, Sequence: sequence, Payload: append([]byte(nil), body[:length]...)})
		expected++
	}
}

func validateReference(kind Type, id string, prepared, applied, rollbackPrepared map[string]bool) error {
	switch kind {
	case OPPrepared:
		if id == "" || prepared[id] {
			return fmt.Errorf("%w: invalid or duplicate prepared operation", ErrCorrupt)
		}
		prepared[id] = true
	case OPApplied:
		if !prepared[id] || applied[id] {
			return fmt.Errorf("%w: applied operation lacks preparation", ErrCorrupt)
		}
		applied[id] = true
	case RollbackPrepared:
		if !applied[id] || rollbackPrepared[id] {
			return fmt.Errorf("%w: rollback lacks applied operation", ErrCorrupt)
		}
		rollbackPrepared[id] = true
	case RollbackApplied:
		if !rollbackPrepared[id] {
			return fmt.Errorf("%w: rollback applied lacks preparation", ErrCorrupt)
		}
	}
	return nil
}
