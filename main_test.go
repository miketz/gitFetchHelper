package main

import (
	"fmt"
	"testing"
)

func TestUpstreamAlias(t *testing.T) {
	t.Parallel()
	// set up a dummy DB
	numRepos := 2
	db := RepoTable{
		Name:          make([]string, numRepos, numRepos),
		Comment:       make([]string, numRepos, numRepos),
		Folder:        make([]string, numRepos, numRepos),
		Remotes:       make([][]Remote, numRepos, numRepos),
		RemoteDefault: make([]RemoteSym, numRepos, numRepos),
		MainBranch:    make([]string, numRepos, numRepos),
		UseBranch:     make([]string, numRepos, numRepos),
		UseCommit:     make([]string, numRepos, numRepos),
		DependHard:    make([][]Dependency, numRepos, numRepos),
		DependSoft:    make([][]Dependency, numRepos, numRepos),
		DependBundled: make([][]Dependency, numRepos, numRepos),
	}
	// repo 1, has upstream
	db.Name[0] = "paredit"
	db.Comment[0] = ""
	db.Folder[0] = "~/.emacs.d/notElpa/paredit"
	db.Remotes[0] = append(db.Remotes[0], Remote{Sym: mine, Alias: "origin", URL: "https://github.com/miketz/paredit"})
	db.Remotes[0] = append(db.Remotes[0], Remote{Sym: upstream, Alias: "upstream", URL: "https://mumble.net/~campbell/git/paredit.git"})
	db.RemoteDefault[0] = mine
	db.MainBranch[0] = "master"
	db.UseBranch[0] = "master"
	db.DependHard[0] = nil
	db.DependSoft[0] = nil
	db.DependBundled[0] = nil
	// repo 2, no upstream
	db.Name[1] = "fake"
	db.Comment[1] = ""
	db.Folder[1] = "~/.emacs.d/notElpa/fake"
	db.Remotes[1] = append(db.Remotes[1], Remote{Sym: mine, Alias: "origin", URL: "fakeUrl"})
	db.RemoteDefault[1] = mine
	db.MainBranch[1] = "master"
	db.UseBranch[1] = "master"
	db.DependHard[1] = nil
	db.DependSoft[1] = nil
	db.DependBundled[1] = nil

	// test 1. repo 0 should have an upstream. err should be nil
	upstreamAlias, err := db.UpstreamAlias(0)
	if err != nil {
		t.Errorf("upstream not configured")
	}
	// test 2. repo 1 should not have an upstream. err is expected.
	upstreamAlias, err = db.UpstreamAlias(1)
	if err == nil {
		t.Errorf("failed to supply an error when accessing non-existant upstream")
	}

	fmt.Printf("%s\n", upstreamAlias)
}
