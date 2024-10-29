package main

import (
	"os"
	"os/user"
	"runtime"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	// setup code here

	exitCode := m.Run()

	// teardown code here

	os.Exit(exitCode)
}

// functional test, not unit test
// may only work on my machine with folders setup
func TestIsGitRepo(t *testing.T) {
	// ~ path handling setup
	usr, _ := user.Current()
	homeDir = usr.HomeDir
	isMsWindows := strings.HasPrefix(runtime.GOOS, "windows")
	if isMsWindows {
		homeDir += "/AppData/Local"
	}
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

// functional test, not unit test
// may only work on my machine with folders setup
func TestIsInGitSubmodule(t *testing.T) {
	// ~ path handling setup
	usr, _ := user.Current()
	homeDir = usr.HomeDir
	isMsWindows := strings.HasPrefix(runtime.GOOS, "windows")
	if isMsWindows {
		homeDir += "/AppData/Local"
	}
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
