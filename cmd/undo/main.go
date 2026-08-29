package main

import (
	"fmt"
	"os"

	"github.com/daniel-oluwadunsin/undolang/internal/buildinfo"
	"github.com/daniel-oluwadunsin/undolang/internal/lang/frontend"
	"github.com/daniel-oluwadunsin/undolang/internal/lang/validate"
)

func main() { os.Exit(run(os.Args[1:])) }

func run(args []string) int {
	if len(args) == 1 && args[0] == "version" {
		fmt.Printf("UndoLang %s\ndsl %s\napi %s\n", buildinfo.Version, buildinfo.DSLVersion, buildinfo.APIVersion)
		return 0
	}
	if len(args) != 2 || args[0] != "check" {
		fmt.Fprintln(os.Stderr, "usage: undo check FILE | undo version")
		return 1
	}
	program, err := frontend.ParseFile(args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Println("valid UndoLang program")
	for _, name := range validate.Names(program) {
		fmt.Printf("transaction: %s\n", name)
	}
	return 0
}
