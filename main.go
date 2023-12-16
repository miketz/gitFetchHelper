package main

import (
	"fmt"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type SubModule struct {
	Folder        string
	UpstreamAlias string
}

// DUMMY repo project for testing.
// var DB = []SubModule{
// 	{Folder: "~/proj/dummyProj2", UpstreamAlias: "origin"},
// }

var DB = []SubModule{
	{Folder: "~/.emacs.d/notElpa/paredit", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/combobulate", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/emacs-buttercup", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/swiper", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/ivy-explorer", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/iedit", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/lispy", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/evil", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/evil-leader", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/expand-region.el", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/s.el", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/dash.el", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/transient", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/with-editor", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/magit", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/libegit2", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/csharp-mode", UpstreamAlias: "upstream"},
	// {Folder: "~/.emacs.d/notElpa/ctrlf", UpstreamAlias: "origin"},
	{Folder: "~/.emacs.d/notElpa/spinner.el", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/ggtags", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/goto-chg", UpstreamAlias: "upstream"},
	// {Folder: "~/.emacs.d/notElpa/mine/mor", UpstreamAlias: "nil"},
	{Folder: "~/.emacs.d/notElpa/ido-grid.el", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/ov", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/emacs-deferred", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/flx", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/sallet", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/eros", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/elisp-slime-nav", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/emacs-async", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/lua-mode", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/slime", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/slime-company", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/sly", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/eglot", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/lsp-mode", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/f.el", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/ht.el", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/markdown-mode", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/avy", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/rust-mode", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/emacs-racer", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/helm", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/rg.el", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/rainbow-delimiters", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/js2-mode", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/js2-highlight-vars.el", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/json-mode", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/json-snatcher", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/leerzeichen.el", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/citre", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/haskell-mode", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/Emacs-wgrep", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/projectile", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/swift-mode", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/dank-mode", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/darkroom", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/smex", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/pkg-info", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/epl", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/erc-hl-nicks", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/esxml", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/flycheck", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/smarttabs", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/web-mode", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/puni", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/ace-link", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/ace-window", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/adoc-mode", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/markup-faces", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/hydra", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/nov.el", UpstreamAlias: "upstream"},
	// {Folder: "~/.emacs.d/notElpa/num3-mode", UpstreamAlias: "nil"},
	{Folder: "~/.emacs.d/notElpa/nyan-mode", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/php-mode", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/emacs-reformatter", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/zig-mode", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/go-mode.el", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/zoutline", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/yasnippet", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/yaml-mode", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/vimrc-mode", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/unkillable-scratch", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/sicp-info", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/prescient.el", UpstreamAlias: "origin"},
	{Folder: "~/.emacs.d/notElpa/pos-tip", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/powershell.el", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/highlight-indent-guides", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/icicles", UpstreamAlias: "mirror"},
	// {Folder: "~/.emacs.d/notElpa/hyperspec", UpstreamAlias: "nil"},
	{Folder: "~/.emacs.d/notElpa/FlamesOfFreedom", UpstreamAlias: "origin"},
	{Folder: "~/.emacs.d/notElpa/Indium", UpstreamAlias: "origin"},
	{Folder: "~/.emacs.d/notElpa/posframe", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/ivy-posframe", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/autothemer", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/company-mode", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/company-web", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/company-lsp", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/web-completion-data", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/fennel-mode", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/vertico", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/consult", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/elisp-bug-hunter", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/typescript.el", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/tide", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/compat.el", UpstreamAlias: "mirror"},
	{Folder: "~/.emacs.d/notElpa/slime-volleyball", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/macrostep", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/sx.el", UpstreamAlias: "upstream"},
	{Folder: "~/.emacs.d/notElpa/sunrise-commander", UpstreamAlias: "upstream"},
}

// get all the submodules
// git config --file .gitmodules --get-regexp path | awk '{ print $2 }'
// cmd := exec.Command("git", "config", "--file", ".gitmodules", "--get-regexp", "path", "|", "awk", "'{ print $2 }'")

func main() {
	fetchUpstreamRemotes()
}

func fetchUpstreamRemotes() {
	start := time.Now() // stop watch start

	reportFetched := make([]string, 0, len(DB)) // allocate for 100% fetch success
	reportFail := make([]string, 0, 4)          // allocate for low failure rate

	wg := sync.WaitGroup{}
	mut := sync.Mutex{}
	for i := 0; i < len(DB); i++ { // fetch upstream for each remote.
		wg.Add(1)
		go fetch(i, &reportFetched, &reportFail, &wg, &mut)
	}
	wg.Wait()

	duration := time.Since(start) // stop watch end
	fmt.Printf("\nFetched %d of %d remotes. time elapsed: %v\n",
		len(DB)-len(reportFail), len(DB), duration)

	// succes report. only includes repos that had new data to fetch.
	fmt.Printf("\nNEW repo data fetched: %d\n", len(reportFetched))
	for i := 0; i < len(reportFetched); i++ {
		fmt.Print(reportFetched[i])
	}
	// failure report
	fmt.Printf("\nFAILURES: %d\n", len(reportFail))
	for i := 0; i < len(reportFail); i++ {
		fmt.Print(reportFail[i])
	}
}

func fetch(i int, reportFetched *[]string, reportFail *[]string, wg *sync.WaitGroup, mut *sync.Mutex) {
	defer wg.Done()

	subMod := DB[i]

	// prepare fetch command
	cmd := exec.Command("git", "fetch", subMod.UpstreamAlias) // #nosec G204
	var err error
	cmd.Dir, err = expandPath(subMod.Folder)
	if err != nil { // issue with folder
		mut.Lock()
		*reportFail = append(*reportFail, fmt.Sprintf("%d: %s %v %s\n", i, subMod.Folder, cmd.Args, err.Error()))
		mut.Unlock()
		return
	}
	// Run git fetch! NOTE: cmd.Output() doesn't include the normal txt output when git fetch actually pulls new data.
	stdout, err := cmd.CombinedOutput() //cmd.Output()
	if err != nil {
		mut.Lock()
		*reportFail = append(*reportFail, fmt.Sprintf("%d: %s %v %s\n", i, subMod.Folder, cmd.Args, err.Error()))
		mut.Unlock()
		return
	}
	newDataFetched := len(stdout) > 0
	if newDataFetched {
		mut.Lock()
		*reportFetched = append(*reportFetched, fmt.Sprintf("%d: %s %v %s\n",
			i, subMod.Folder, cmd.Args, string(stdout)))
		mut.Unlock()
	}
}

func expandPath(path string) (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
		// log.Fatalf("failed to get user info for translating ~. error: %v", err.Error())
	}

	if path == "~" {
		// In case of "~", which won't be caught by the "else if"
		path = usr.HomeDir
	} else if strings.HasPrefix(path, "~/") {
		// Use strings.HasPrefix so we don't match paths like
		// "/something/~/something/"
		path = filepath.Join(usr.HomeDir, path[2:])
	}
	return path, nil
}
