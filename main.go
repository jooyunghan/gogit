package main

import (
	"bytes"
	"compress/zlib"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
)


func gitdir() string {
	pwd, err := os.Getwd()
	if err != nil {
		return ""
	}

	for {
		candidate := path.Join(pwd, ".git")
		fileInfo, err := os.Stat(candidate)
		if err == nil && fileInfo.IsDir() {
			return candidate
		}
		if pwd == "/" {
			break
		}
		pwd = path.Dir(pwd) 
	}
	panic("no git dir")
}

func gitBranches(gitdir string) []string {
	refsHeads := path.Join(gitdir, "refs/heads")
	if _, err := os.Open(path.Join(refsHeads)); err != nil {
		panic("Open git dir failed")
	}
	return dirFiles(refsHeads)
}

func dirFiles(baseDir string) (files []string) {
	base, _ := os.Open(baseDir)
	fi, _ := base.Readdir(-1)
	for _, f := range fi {
		if f.IsDir() {
			names := dirFiles(path.Join(baseDir, f.Name()))
			for _, name := range names {
				files = append(files, path.Join(f.Name(), name))
			}
		} else {
			files = append(files, f.Name())
		}
	}
	return
}

func readCommitId(gitdir, branch string) string {
	file, _ := os.Open(path.Join(gitdir, "refs/heads", branch))
	bytes, _ := ioutil.ReadAll(file)
	return strings.TrimSpace(string(bytes))
}

func readObject(gitdir, objId string) []byte {
	objPath := path.Join(gitdir, "objects", objId[:2], objId[2:])
	file, err := os.Open(objPath)
	if err != nil {
		panic("failed to open Object: " + objId)
	}

	r, _ := zlib.NewReader(file)
	bytes, _ := ioutil.ReadAll(r)
	r.Close()
	return bytes
}

func readInt(b []byte) (value int, byteLength int) {
	for _, d := range b {
		if d >= '0' && d <= '9' {
			byteLength++
			value = value * 10 + int(d) - '0'
		} else {
			return
		}
	}
	return
}

type commit struct {
	props map[string]string
	message string
}

type tag struct {
	objectType string
	size int
}

func atoi(s string) int {
	n, err := strconv.ParseInt(s, 10, 0)
	if err != nil {
		panic(err)
	}
	return int(n)
}

func readTag(b []byte) (tag, []byte) {
	index := bytes.IndexByte(b, byte(0))
	if index == -1 {
		panic("no tag")
	}
	elements := strings.Split(string(b[:index]), " ")
	return tag{elements[0], atoi(elements[1])}, b[index+1:]
}

func parseCommit(b []byte) commit {
	_, rest := readTag(b)
	buf := bytes.NewBuffer(rest)
	props := make(map[string]string)
	message := ""

	for {
		line,_ := buf.ReadBytes('\n')
		index := bytes.IndexByte(line, ' ')
		if index != -1 {
			field := line[:index]
			value := line[index+1:]
			props[string(field)] = strings.TrimSpace(string(value))
		} else {
			message = buf.String()
			break
		}
	}
	return commit{props, message}
}


type tree struct {
}

func parseTree(b []byte) tree {
	_, rest := readTag(b)
	for rest != nil {
		index := bytes.IndexByte(rest, byte(0))
		if index == -1 {
			break
		}
		fmt.Println(string(rest[:index]))
		objId := hex.EncodeToString(rest[index+1:][:20])
		fmt.Println(objId)
		rest = rest[index+21:]
	}
	return tree{}
}

func lsTree(gitdir, branch string) {
	commitId := readCommitId(gitdir, branch)
	commitObject := readObject(gitdir, commitId)
//	fmt.Println(string(commitObject))
	c := parseCommit(commitObject)
//	fmt.Println(c.props["tree"])
	treeObject := readObject(gitdir, c.props["tree"])
	fmt.Println(string(treeObject))
	t := parseTree(treeObject)
	fmt.Println(t)
}

func main() {
	dir := gitdir()
	branches := gitBranches(dir)
	firstBranch := branches[0]
	lsTree(dir, firstBranch)
	fmt.Print("ok")
	io.Copy(os.Stdout, strings.NewReader("\n"))
}
