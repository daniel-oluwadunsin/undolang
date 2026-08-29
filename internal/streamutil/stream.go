package streamutil

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
)

const BufferSize = 64 * 1024

var ErrEmptyNeedle = errors.New("empty search needle")

func Contains(r io.Reader, needle []byte) (bool, error) {
	if len(needle) == 0 {
		return false, ErrEmptyNeedle
	}
	prefix := make([]int, len(needle))
	for i, j := 1, 0; i < len(needle); i++ {
		for j > 0 && needle[i] != needle[j] {
			j = prefix[j-1]
		}
		if needle[i] == needle[j] {
			j++
		}
		prefix[i] = j
	}
	buf := make([]byte, BufferSize)
	matched := 0
	for {
		n, err := r.Read(buf)
		for _, b := range buf[:n] {
			for matched > 0 && b != needle[matched] {
				matched = prefix[matched-1]
			}
			if b == needle[matched] {
				matched++
				if matched == len(needle) {
					return true, nil
				}
			}
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				return false, nil
			}
			return false, err
		}
		if n == 0 {
			return false, io.ErrNoProgress
		}
	}
}

func SHA256(r io.Reader) (string, int64, error) {
	h := sha256.New()
	n, err := io.CopyBuffer(h, r, make([]byte, BufferSize))
	if err != nil {
		return "", n, err
	}
	return hex.EncodeToString(h.Sum(nil)), n, nil
}
