package main

import (
	"compress/zlib"
	"fmt"
	"io"
	"os"
)

// Usage: your_program.sh <command> <arg1> <arg2> ...
func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Fprintf(os.Stderr, "Logs from your program will appear here!\n")

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: mygit <command> [<args>...]\n")
		os.Exit(1)
	}

	switch command := os.Args[1]; command {
	case "init":
		// Uncomment this block to pass the first stage!
		//
		for _, dir := range []string{".git", ".git/objects", ".git/refs"} {
			if err := os.MkdirAll(dir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating directory: %s\n", err)
				os.Exit(1)
			}
		}

		headFileContents := []byte("ref: refs/heads/main\n")
		if err := os.WriteFile(".git/HEAD", headFileContents, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %s\n", err)
			os.Exit(1)
		}

		fmt.Println("Initialized git directory")
	case "cat-file":
		if len(os.Args) < 4 {
			fmt.Fprintf(os.Stderr, "usage: mygit <command> [<args>...]\n")
			os.Exit(1)
		}
		catFile(os.Args[2], os.Args[3])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}

type GitObject []byte

func (g GitObject) Type() string {
	return string(g[0:4])
}

func (g GitObject) Size() string {
	return string(g[5:7])
}

func (g GitObject) Content() string {
	return string(g[8:])
}

func catFile(t, obj string) {
	dir := obj[0:2]
	file := obj[2:]

	f, err := os.Open(fmt.Sprintf(".git/objects/%s/%s", dir, file))
	if err != nil {
		fmt.Printf("failed to open file %v\n", err)
		os.Exit(1)
	}

	r, err := zlib.NewReader(f)
	if err != nil {
		fmt.Printf("failed to open file %v\n", err)
		os.Exit(1)
	}

	b, err := io.ReadAll(r)
	if err != nil {
		fmt.Printf("failed to open file %v\n", err)
		os.Exit(1)
	}

	var g GitObject = GitObject(b)

	switch t {
	case "-p":
		fmt.Printf("%s", g.Content())
	case "-t":
		fmt.Printf("%s", g.Type())
	case "-s":
		fmt.Printf("%s", g.Size())
	default:
		fmt.Printf("Unknown argument %s", t)
		os.Exit(1)
	}
}
