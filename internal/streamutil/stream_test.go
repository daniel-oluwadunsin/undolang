package streamutil

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestContainsAcrossBufferBoundary(t *testing.T) {
	data := append(bytes.Repeat([]byte{'a'}, BufferSize-2), []byte("needle")...)
	found, err := Contains(bytes.NewReader(data), []byte("needle"))
	if err != nil || !found {
		t.Fatalf("found=%v err=%v", found, err)
	}
	found, err = Contains(bytes.NewReader(data), []byte("absent"))
	if err != nil || found {
		t.Fatalf("found=%v err=%v", found, err)
	}
}

func TestSHA256(t *testing.T) {
	data := bytes.Repeat([]byte("0123456789"), 10000)
	want := sha256.Sum256(data)
	got, n, err := SHA256(bytes.NewReader(data))
	if err != nil || n != int64(len(data)) || got != hex.EncodeToString(want[:]) {
		t.Fatalf("digest=%s n=%d err=%v", got, n, err)
	}
}
