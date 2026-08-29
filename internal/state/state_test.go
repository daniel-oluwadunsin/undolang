package state

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestTransactionStateAndLockLifecycle(t *testing.T) {
	root := t.TempDir()
	store, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	tx, err := store.Begin("test", "test.undo", "hash")
	if err != nil {
		t.Fatal(err)
	}
	if tx.Meta.ID == "" || tx.Meta.Status != Planned {
		t.Fatalf("meta=%#v", tx.Meta)
	}
	if _, err := os.Stat(filepath.Join(root, ".undo", "active.lock")); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Begin("second", "test.undo", "hash"); err == nil {
		t.Fatal("second active transaction was allowed")
	}
	for _, status := range []Status{Prepared, Running, Verifying, Committing, Committed} {
		if err := tx.Transition(status); err != nil {
			t.Fatalf("transition %s: %v", status, err)
		}
	}
	if err := tx.Release(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, ".undo", "active.lock")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("lock remains: %v", err)
	}
	meta, err := store.Inspect(tx.Meta.ID)
	if err != nil {
		t.Fatal(err)
	}
	if meta.Status != Committed {
		t.Fatalf("meta=%#v", meta)
	}
	history, err := store.History()
	if err != nil {
		t.Fatal(err)
	}
	if len(history) != 1 || history[0].ID != tx.Meta.ID {
		t.Fatalf("history=%#v", history)
	}
}

func TestInvalidStateTransitionsAndEarlyRelease(t *testing.T) {
	if err := ValidateTransition(Planned, Running); err == nil {
		t.Fatal("invalid transition accepted")
	}
	store, _ := Open(t.TempDir())
	tx, err := store.Begin("test", "test.undo", "hash")
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Close()
	if err := tx.Release(); err == nil {
		t.Fatal("released non-terminal transaction")
	}
}
