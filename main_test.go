package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	// setup code here
	homeDir, _ = getHomeDir() // ~ path handling setup

	exitCode := m.Run()

	// teardown code here

	os.Exit(exitCode)
}

// functional test, not unit test.
// may only work on my machine with folders setup
func TestIsGitRepo(t *testing.T) {
	// repos, but not submods
	path1 := expandPath("~/.emacs.d")
	path2 := expandPath("~/.emacs.d/notElpa")
	// repo and is submod
	path3 := expandPath("~/.emacs.d/notElpa/magit")
	// not a repo
	path4 := expandPath("~/.vscode")

	want := true
	got := isInGitRepo(path1)
	if got != want {
		t.Fatalf("got: %t. wanted %t", got, want)
	}

	want = true
	got = isInGitRepo(path2)
	if got != want {
		t.Fatalf("got: %t. wanted %t", got, want)
	}

	want = true
	got = isInGitRepo(path3)
	if got != want {
		t.Fatalf("got: %t. wanted %t", got, want)
	}

	want = false
	got = isInGitRepo(path4)
	if got != want {
		t.Fatalf("got: %t. wanted %t", got, want)
	}
}

// functional test, not unit test.
// may only work on my machine with folders setup
func TestIsInGitSubmodule(t *testing.T) {
	// repos, but not submods
	path1 := expandPath("~/.emacs.d")
	path2 := expandPath("~/.emacs.d/notElpa")
	// repo and is submod
	path3 := expandPath("~/.emacs.d/notElpa/magit")
	// not a repo
	path4 := expandPath("~/.vscode")

	want := false
	got := isInGitSubmodule(path1)
	if got != want {
		t.Fatalf("got: %t. wanted %t", got, want)
	}

	want = false
	got = isInGitSubmodule(path2)
	if got != want {
		t.Fatalf("got: %t. wanted %t", got, want)
	}

	want = true
	got = isInGitSubmodule(path3)
	if got != want {
		t.Fatalf("got: %t. wanted %t", got, want)
	}

	want = false
	got = isInGitSubmodule(path4)
	if got != want {
		t.Fatalf("got: %t. wanted %t", got, want)
	}
}

// functional test, not unit test
// may only work on my machine with folders setup
func TestExists(t *testing.T) {
	// repos, but not submods
	path1 := expandPath("~/.emacs.d")
	path2 := expandPath("~/.emacs.d/notElpa")
	// repo and is submod
	// path3 := expandPath("~/.emacs.d/notElpa/magit")
	// not a repo
	path4 := expandPath("~/.vscode")
	// is a "yolo" repo
	path5 := expandPath("~/.emacs.d/notElpaYolo/binky.el")

	want := true
	got, _ := exists(path1)
	if got != want {
		t.Fatalf("got: %t. wanted %t", got, want)
	}

	want = true
	got, _ = exists(path2)
	if got != want {
		t.Fatalf("got: %t. wanted %t", got, want)
	}

	// want = true
	// got, _ = exists(path3)
	// if got != want {
	// 	t.Fatalf("got: %t. wanted %t", got, want)
	// }

	want = true
	got, _ = exists(path4)
	if got != want {
		t.Fatalf("got: %t. wanted %t", got, want)
	}

	want = true
	got, _ = exists(path5)
	if got != want {
		t.Fatalf("got: %t. wanted %t", got, want)
	}

	want = false
	got, _ = exists("~/fake/path")
	if got != want {
		t.Fatalf("got: %t. wanted %t", got, want)
	}
}

func TestParentDir(t *testing.T) {
	want := expandPath("~/.emacs.d/notElpaYolo")
	got := parentDir("~/.emacs.d/notElpaYolo/binky.el")
	if got != want {
		t.Fatalf("got: %s. wanted %s", got, want)
	}
}

// functional test, not unit test
// may only work on my machine with folders and git repos setup
func TestTrackingBranches(t *testing.T) {
	want := make([]string, 0, 2)
	want = append(want, "origin/master")
	want = append(want, "origin/mine")

	got, err := TrackingBranches(expandPath("~/.emacs.d/notElpaYolo/nov.el"), "origin")
	if err != nil {
		t.Fatalf("err during test: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("got: %v. wanted %v", got, want)
	}
	if got[0] != want[0] {
		t.Fatalf("got: %v. wanted %v", got, want)
	}
	if got[1] != want[1] {
		t.Fatalf("got: %v. wanted %v", got, want)
	}
}

func TestRemoveRemoteFromBranchName(t *testing.T) {
	// name with extra slashes in it
	fullBranchName := "origin/km/reshelve-rewrite"
	got := removeRemoteFromBranchName(fullBranchName)
	want := "km/reshelve-rewrite"
	if got != want {
		t.Fatalf("got: %s. wanted %s", got, want)
	}

	// common typical name. no extra slashes
	fullBranchName2 := "origin/master"
	got = removeRemoteFromBranchName(fullBranchName2)
	want = "master"
	if got != want {
		t.Fatalf("got: %s. wanted %s", got, want)
	}

	// normal master branch. no remote!
	fullBranchName3 := "master"
	got = removeRemoteFromBranchName(fullBranchName3)
	want = "master"
	if got != want {
		t.Fatalf("got: %s. wanted %s", got, want)
	}
}

func BenchmarkMergeAlreadyLatest(b *testing.B) {
	var hasLatest bool
	dir := expandPath("~/.emacs.d/notElpaYolo/mor")
	for i := 0; i < b.N; i++ {
		cmd := exec.Command("git", "merge", "origin/master") // #nosec G204
		cmd.Dir = dir
		stdout, err := cmd.CombinedOutput()
		if err != nil {
			b.Fatalf("merged errored out! %v", err)
		}
		output := string(stdout)
		hasLatest = output == "Already up to date.\n"
	}
	fmt.Printf("merge hasLatest: %v\n", hasLatest)
	b.ReportAllocs() // include alloc info in report
}

func BenchmarkCheckAlreadyLatest(b *testing.B) {
	var hasLatest bool
	dir := expandPath("~/.emacs.d/notElpaYolo/mor")
	for i := 0; i < b.N; i++ {
		// git rev-parse HEAD
		hashLocal, err := GetHash(dir, "HEAD")
		if err != nil {
			b.Fatalf("GetHash errored out! %v", err)
		}
		hashRemote, err := GetHash(dir, "origin/master")
		if err != nil {
			b.Fatalf("GetHash errored out! %v", err)
		}
		hasLatest = hashLocal == hashRemote
	}
	fmt.Printf("check hasLatest: %v\n", hasLatest)
	b.ReportAllocs() // include alloc info in report
}

func BenchmarkSubstringOld(b *testing.B) {
	fullBranchName := "origin/km/reshelve-rewrite"
	for i := 0; i < b.N; i++ {
		_removeRemoteFromBranchName_OLD(fullBranchName)
	}
	b.ReportAllocs() // include alloc info in report
}
func BenchmarkSubstringNew(b *testing.B) {
	fullBranchName := "origin/km/reshelve-rewrite"
	for i := 0; i < b.N; i++ {
		_removeRemoteFromBranchName_NEW(fullBranchName)
	}
	b.ReportAllocs() // include alloc info in report
}

// tmp fn to compare substirng techniques
func _removeRemoteFromBranchName_OLD(remoteBranch string) string {
	parts := strings.Split(remoteBranch, "/")

	// we cannot simply use parts[1] becuase the remainder of the name may have
	// contained slashes "/".
	branchName := "" // := parts[1]
	// trim off the "origin" prefix, but also add back the "/" in the remainder of the name
	for i := 1; i < len(parts); i++ {
		if i == 1 {
			branchName += parts[i]
		} else {
			branchName += "/" + parts[i]
		}
	}
	return branchName
}

// tmp fn to compare substirng techniques
func _removeRemoteFromBranchName_NEW(remoteBranch string) string {
	i := strings.Index(remoteBranch, "/")
	return remoteBranch[i+1:]
}
