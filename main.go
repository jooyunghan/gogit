package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
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

func main() {
	dir := gitdir()
	branches := gitBranches(dir)
	for _, branch := range branches {
		fmt.Println(branch + " | " + readCommitId(dir, branch))
	}
}
