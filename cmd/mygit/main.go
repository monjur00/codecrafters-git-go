package main

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"strconv"
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
	case "hash-object":
		if len(os.Args) < 4 {
			fmt.Fprintf(os.Stderr, "usage: mygit <command> [<args>...]\n")
			os.Exit(1)
		}
		hashObject(os.Args[3])
	case "ls-tree":
		if len(os.Args) < 4 {
			fmt.Fprintf(os.Stderr, "usage: mygit <command> [<args>...]\n")
			os.Exit(1)
		}
		lsTree(os.Args[3])
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

func (g GitObject) Hash() string {
	h := sha1.New()
	h.Write(g)
	// return base64.StdEncoding.EncodeToString(h.Sum(nil))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func catFile(t, obj string) {
	dir := obj[0:2]
	file := obj[2:]

	f, err := os.Open(fmt.Sprintf(".git/objects/%s/%s", dir, file))
	if err != nil {
		fmt.Printf("failed to open file %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	r, err := zlib.NewReader(f)
	if err != nil {
		fmt.Printf("failed to open file %v\n", err)
		os.Exit(1)
	}
	defer r.Close()

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

func hashObject(path string) {
	b, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("Failed to read file", err)
		os.Exit(1)
	}

	var content GitObject
	content = append(content, []byte("blob ")...)
	content = append(content, []byte(strconv.Itoa(len(b)))...)
	content = append(content, 0)
	content = append(content, b...)

	h := content.Hash()
	dir := h[0:2]
	file := h[2:]

	os.MkdirAll(fmt.Sprintf(".git/objects/%s/", dir), os.ModePerm)
	var buffer bytes.Buffer
	z := zlib.NewWriter(&buffer)
	z.Write(content)
	z.Close()

	if err := os.WriteFile(fmt.Sprintf(".git/objects/%s/%s", dir, file), buffer.Bytes(), 0644); err != nil {
		fmt.Println("failed to write file", err)
		os.Exit(1)
	}
	fmt.Printf("%s", h)
}

func object(sha string) GitObject {
	f, err := os.Open(fmt.Sprintf(".git/objects/%s/%s", sha[0:2], sha[2:]))
	if err != nil {
		fmt.Printf("failed to open file %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	r, err := zlib.NewReader(f)
	if err != nil {
		fmt.Printf("failed to open file %v\n", err)
		os.Exit(1)
	}
	defer r.Close()

	b, err := io.ReadAll(r)
	if err != nil {
		fmt.Printf("failed to open file %v\n", err)
		os.Exit(1)
	}

	return GitObject(b)
}

func lsTree(sha string) {
	g := object(sha)
	split := bytes.Split([]byte(g.Content()), []byte("\x00"))
	con := split[0 : len(split)-1]

	for _, s := range con {
		d := bytes.Split(s, []byte(" "))[1]
		fmt.Printf("%s\n", d)
	}
}
