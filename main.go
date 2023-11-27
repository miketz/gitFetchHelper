package main

import (
	"fmt"
	"log"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
)

type SubModule struct {
	Folder        string
	Url           string
	UpstreamAlias string
}

var DB = []SubModule{
	{Folder: "~/.emacs.d/notElpa/paredit", Url: "https://mumble.net/~campbell/git/paredit.git", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/combobulate", Url: "https://github.com/mickeynp/combobulate", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/emacs-buttercup", Url: "https://github.com/jorgenschaefer/emacs-buttercup", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/swiper", Url: "https://github.com/abo-abo/swiper", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/ivy-explorer", Url: "https://github.com/clemera/ivy-explorer", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/iedit", Url: "https://github.com/victorhge/iedit", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/lispy", Url: "https://github.com/abo-abo/lispy", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/evil", Url: "https://github.com/emacs-evil/evil", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/evil-leader", Url: "https://github.com/cofi/evil-leader", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/expand-region.el", Url: "https://github.com/magnars/expand-region.el", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/s.el", Url: "https://github.com/magnars/s.el", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/dash.el", Url: "https://github.com/magnars/dash.el", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/transient", Url: "https://github.com/magit/transient", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/with-editor", Url: "https://github.com/magit/with-editor", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/magit", Url: "https://github.com/magit/magit", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/csharp-mode", Url: "https://github.com/josteink/csharp-mode", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/ctrlf", Url: "https://github.com/radian-software/ctrlf", UpstreamAlias: "origin"},
	{Folder: "~/.emacs.d/notElpa/spinner.el", Url: "https://github.com/Malabarba/spinner.el", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/ggtags", Url: "https://github.com/leoliu/ggtags", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/ido-grid.el", Url: "https://github.com/larkery/ido-grid.el", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/ov", Url: "https://github.com/emacsorphanage/ov", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/emacs-deferred", Url: "https://github.com/kiwanami/emacs-deferred", UpstreamAlias: "upstream"},
}

func main() {
	// get all the submodules
	// git config --file .gitmodules --get-regexp path | awk '{ print $2 }'
	//cmd := exec.Command("git", "config", "--file", ".gitmodules", "--get-regexp", "path", "|", "awk", "'{ print $2 }'")

	fmt.Printf("len: %d\n", len(DB))
	var wg sync.WaitGroup
	for i := 0; i < len(DB); i++ {
		wg.Add(1)
		go fetch(i, &wg)
	}
	wg.Wait()
	fmt.Printf("Done\n")
}

func fetch(i int, wg *sync.WaitGroup) {
	defer wg.Done()
	subMod := DB[i]
	cmd := exec.Command("git", "fetch", subMod.UpstreamAlias)
	cmd.Dir = expandPath(subMod.Folder)

	stdout, err := cmd.Output()
	if err != nil {
		log.Fatalf("error: %v", err.Error())
	}
	if len(stdout) == 0 {
		fmt.Printf("%d: no output for %v\n", i, subMod.Folder)
		return
	}
	fmt.Printf("%d: %s\n", i, string(stdout))
}

func expandPath(path string) string {
	usr, err := user.Current()
	if err != nil {
		log.Fatalf("failed to get user info for translating ~. error: %v", err.Error())
	}
	dir := usr.HomeDir

	if path == "~" {
		// In case of "~", which won't be caught by the "else if"
		path = dir
	} else if strings.HasPrefix(path, "~/") {
		// Use strings.HasPrefix so we don't match paths like
		// "/something/~/something/"
		path = filepath.Join(dir, path[2:])
	}
	return path
}
