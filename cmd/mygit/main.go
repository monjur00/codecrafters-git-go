package main

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
)

var nullByte = []byte("\u0000")
var whiteSpaceByte = []byte("\u0020")

type GitObject []byte
type TreeContent []byte
type TreeObject []byte

func (g GitObject) Type() string {
	return string(g[0:4])
}

func (g GitObject) Size() string {
	return string(g[5:7])
}

func (g GitObject) Content() string {
	return string(g[8:])
}

func (g GitObject) Hash() []byte {
	h := sha1.New()
	h.Write(g)
	return h.Sum(nil)
}

func (t TreeObject) Contents() []TreeContent {
	// first null byte index
	i := bytes.Index(t, nullByte)
	c := t[i+1:]
	var contents []TreeContent

	for len(c) > 0 {
		k := bytes.Index(c, nullByte)
		newTreeCont := TreeContent(c[:k+21])
		contents = append(contents, newTreeCont)
		c = c[k+21:]
	}

	return contents
}

func (t TreeContent) TypeCode() string {
	k := bytes.Index(t, nullByte)
	typeCode := bytes.Split(t[:k], whiteSpaceByte)[0]
	return string(typeCode)
}
func (t TreeContent) Type() string {
	switch t.TypeCode() {
	case "100644":
		return "blob"
	case "40000":
		return "tree"
	default:
		return "unknown"
	}
}

func (t TreeContent) Name() string {
	k := bytes.Index(t, nullByte)
	name := bytes.Split(t[:k], whiteSpaceByte)[1]
	return string(name)
}

func (t TreeContent) Hash() []byte {
	k := bytes.Index(t, nullByte)
	return t[k+1:]
}

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
		lsTree(os.Args[2], os.Args[3])
	case "write-tree":
		writeTree()
	case "commit-tree":
		//git commit-tree 5b825dc642cb6eb9a060e54bf8d69288fbee4904 -p 3b18e512dba79e4c8300dd08aeb37f8e728b8dad -m "Second commit"
		commitTree(os.Args[2], os.Args[4], os.Args[6])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
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

	h := fmt.Sprintf("%x", content.Hash())
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

func lsTree(arg, sha string) {
	tree := TreeObject(object(sha))
	var output []string

	for _, c := range tree.Contents() {
		if arg == "--name-only" {
			output = append(output, c.Name())
		} else {
			output = append(output, fmt.Sprintf("%s %s %x\t%s", c.TypeCode(), c.Type(), c.Hash(), c.Name()))
		}
	}

	fmt.Println(strings.Join(output, "\n"))
}

func writeTree() {
	_, h := writeTreeRec(".", true)
	fmt.Printf("%x", h)
}

// writeTreeRec recursively create git object
// returns object type and hash
func writeTreeRec(path string, isDir bool) (string, []byte) {
	if isDir {
		// this is a dir, we need to generate
		dirEntries, err := os.ReadDir(path)
		if err != nil {
			fmt.Println("Failed to read dir", err)
			os.Exit(1)
		}
		var contents []string

		slices.SortFunc(dirEntries, func(i, j os.DirEntry) int { return strings.Compare(i.Name(), j.Name()) })
		for _, d := range dirEntries {
			if strings.Contains(d.Name(), ".git") {
				// ignore .git dir
				continue
			}

			t, h := writeTreeRec(path+"/"+d.Name(), d.IsDir())
			var content string
			if t == "tree" {
				content = fmt.Sprintf("40000 %s\u0000", d.Name())
			} else {
				content = fmt.Sprintf("100644 %s\u0000", d.Name())
			}
			cHash := append([]byte(content), h...)
			contents = append(contents, string(cHash))
		}

		cc := strings.Join(contents, "")
		g := newGitObj("tree", []byte(cc))
		h := g.HashObj()
		return "tree", h
	}

	b, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("Failed to read file", err)
		os.Exit(1)
	}

	// generate hash object of the file and return the hash
	g := newGitObj("blob", b)
	h := g.HashObj()
	return "blob", h
}

// newGitObj creates a new git object
// t is type of object like blob, tree
func newGitObj(t string, body []byte) GitObject {
	// s := fmt.Sprintf("%s %d\u0000%s", t, len(body), string(body))
	// return GitObject(s)

	var content GitObject
	switch t {
	case "tree":
		content = append(content, []byte("tree ")...)
		content = append(content, []byte(strconv.Itoa(len(body)))...)
		content = append(content, []byte("\u0000")...)
		// content = append(content, []byte("\n")...)
		content = append(content, body...)
	case "blob":
		content = append(content, []byte("blob ")...)
		content = append(content, []byte(strconv.Itoa(len(body)))...)
		content = append(content, []byte("\u0000")...)
		content = append(content, body...)
	}
	return content
}

// HashObj writes the obj and return hash
func (g GitObject) HashObj() []byte {
	h := g.Hash()
	hHex := fmt.Sprintf("%x", h)
	dir := hHex[0:2]
	file := hHex[2:]

	var buffer bytes.Buffer
	z := zlib.NewWriter(&buffer)
	z.Write([]byte(g))
	z.Close()

	err := os.MkdirAll(fmt.Sprintf(".git/objects/%s/", dir), os.ModePerm)
	if err != nil {
		fmt.Println("failed to write file", err)
		os.Exit(1)
	}

	if err := os.WriteFile(fmt.Sprintf(".git/objects/%s/%s", dir, file), buffer.Bytes(), 0644); err != nil {
		fmt.Println("failed to write file", err)
		os.Exit(1)
	}

	return h
}

type CommitObject []byte

func newCommitObject(treeSha, partentSha, msg string) CommitObject {
	//commit {size}\0{content}
	//tree {tree_sha}
	//parent {parent1_sha}
	// author {author_name} <{author_email}> {author_date_seconds} {author_date_timezone}
	// committer {committer_name} <{committer_email}> {committer_date_seconds} {committer_date_timezone}
	//
	//{commit message}
	content := fmt.Sprintf("tree %s\n", treeSha)
	if len(partentSha) > 0 {
		content += fmt.Sprintf("parent %s\n", partentSha)
	}
	content += fmt.Sprintf("author Scott Chacon <schacon@gmail.com> %d +0530\n", time.Now().Unix())
	content += fmt.Sprintf("committer Scott Chacon <schacon@gmail.com> %d +0530\n", time.Now().Unix())
	content += fmt.Sprintf("\n%s", msg)

	commitObj := fmt.Sprintf("commit %d\u0000%s", len(content), content)
	return CommitObject(commitObj)
}

func commitTree(treeSha, partentSha, msg string) {
	c := newCommitObject(treeSha, partentSha, msg)
	gitObj := GitObject(c)
	h := gitObj.HashObj()
	fmt.Printf("%x", h)
}
