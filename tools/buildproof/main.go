// Command buildproof computes deterministic SHA-256 receipts using only the Go
// standard library. With two inputs, it also fails if their bytes differ.
package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: go run ./tools/buildproof FILE [FILE ...]")
		os.Exit(2)
	}
	hashes := make([][sha256.Size]byte, 0, len(os.Args)-1)
	for _, path := range os.Args[1:] {
		hash, err := hashFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "buildproof: %s: %v\n", path, err)
			os.Exit(1)
		}
		hashes = append(hashes, hash)
		fmt.Printf("%x  %s\n", hash, path)
	}
	if len(hashes) == 2 && hashes[0] != hashes[1] {
		fmt.Fprintln(os.Stderr, "buildproof: files are not byte-identical")
		os.Exit(1)
	}
}

func hashFile(path string) ([sha256.Size]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return [sha256.Size]byte{}, err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return [sha256.Size]byte{}, err
	}
	var result [sha256.Size]byte
	copy(result[:], hash.Sum(nil))
	return result, nil
}
