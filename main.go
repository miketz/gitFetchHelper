package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
	"sync"
	"time"
)

type GitRepo struct {
	Folder        string
	UpstreamAlias string
	MainBranch    string
}

// DUMMY repo project for testing fetch of new commits. https://github.com/miketz/dummyProj
// var DB = []GitRepo{
// 	{Folder: "~/proj/dummyProj2", UpstreamAlias: "origin"},
// }

var DB = []GitRepo{
	{Folder: "~/.emacs.d/notElpa/paredit", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/combobulate", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/emacs-buttercup", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/swiper", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/ivy-explorer", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/iedit", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/lispy", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/evil", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/evil-leader", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/expand-region.el", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/s.el", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/dash.el", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/transient", UpstreamAlias: "upstream", MainBranch: "main"},
	{Folder: "~/.emacs.d/notElpa/with-editor", UpstreamAlias: "upstream", MainBranch: "main"},
	{Folder: "~/.emacs.d/notElpa/magit", UpstreamAlias: "upstream", MainBranch: "main"},
	{Folder: "~/.emacs.d/notElpa/libegit2", UpstreamAlias: "upstream", MainBranch: "main"},
	{Folder: "~/.emacs.d/notElpa/csharp-mode", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/ctrlf", UpstreamAlias: "origin", MainBranch: "main"},
	{Folder: "~/.emacs.d/notElpa/spinner.el", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/ggtags", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/goto-chg", UpstreamAlias: "upstream", MainBranch: "master"},
	// {Folder: "~/.emacs.d/notElpa/mine/mor", UpstreamAlias: "nil", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/ido-grid.el", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/ov", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/emacs-deferred", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/flx", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/sallet", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/eros", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/elisp-slime-nav", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/emacs-async", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/lua-mode", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/slime", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/slime-company", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/sly", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/eglot", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/lsp-mode", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/f.el", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/ht.el", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/markdown-mode", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/avy", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/rust-mode", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/emacs-racer", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/helm", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/rg.el", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/rainbow-delimiters", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/js2-mode", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/js2-highlight-vars.el", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/json-mode", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/json-snatcher", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/leerzeichen.el", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/citre", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/haskell-mode", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/Emacs-wgrep", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/projectile", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/swift-mode", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/dank-mode", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/darkroom", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/smex", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/pkg-info", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/epl", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/erc-hl-nicks", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/esxml", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/flycheck", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/smarttabs", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/web-mode", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/puni", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/ace-link", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/ace-window", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/adoc-mode", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/markup-faces", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/hydra", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/nov.el", UpstreamAlias: "upstream", MainBranch: "master"},
	// {Folder: "~/.emacs.d/notElpa/num3-mode", UpstreamAlias: "nil", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/nyan-mode", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/php-mode", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/emacs-reformatter", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/zig-mode", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/go-mode.el", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/zoutline", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/yasnippet", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/yaml-mode", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/vimrc-mode", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/unkillable-scratch", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/sicp-info", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/prescient.el", UpstreamAlias: "origin", MainBranch: "main"},
	{Folder: "~/.emacs.d/notElpa/pos-tip", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/powershell.el", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/highlight-indent-guides", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/icicles", UpstreamAlias: "mirror", MainBranch: "master"},
	// {Folder: "~/.emacs.d/notElpa/hyperspec", UpstreamAlias: "nil", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/FlamesOfFreedom", UpstreamAlias: "origin", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/Indium", UpstreamAlias: "origin", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/posframe", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/ivy-posframe", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/autothemer", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/company-mode", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/company-web", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/company-lsp", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/web-completion-data", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/fennel-mode", UpstreamAlias: "upstream", MainBranch: "main"},
	{Folder: "~/.emacs.d/notElpa/vertico", UpstreamAlias: "upstream", MainBranch: "main"},
	{Folder: "~/.emacs.d/notElpa/consult", UpstreamAlias: "upstream", MainBranch: "main"},
	{Folder: "~/.emacs.d/notElpa/elisp-bug-hunter", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/typescript.el", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/tide", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/compat.el", UpstreamAlias: "mirror", MainBranch: "main"},
	{Folder: "~/.emacs.d/notElpa/slime-volleyball", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/macrostep", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/sx.el", UpstreamAlias: "upstream", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/sunrise-commander", UpstreamAlias: "upstream", MainBranch: "master"},
}

var homeDir string

// initialize global variables. At the moment only homeDir.
func initGlobals() error {
	usr, err := user.Current()
	if err != nil {
		return err
	}
	homeDir = usr.HomeDir
	var isMsWindows = strings.HasPrefix(runtime.GOOS, "windows")
	if isMsWindows {
		homeDir += "/AppData/Local"
	}
	return nil
}

// get all the submodules
// git config --file .gitmodules --get-regexp path | awk '{ print $2 }'
// cmd := exec.Command("git", "config", "--file", ".gitmodules", "--get-regexp", "path", "|", "awk", "'{ print $2 }'")

func printCommands() {
	fmt.Printf("Enter a command: [fetch, diff]\n")
}

func main() {
	err := initGlobals()
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		return
	}

	if len(os.Args) < 2 {
		printCommands()
		return
	}
	switch command := os.Args[1]; command {
	case "fetch":
		runAndReport(fetch,
			"\nFetched %d of %d remotes. time elapsed: %v\n",
			"\nNEW repo data fetched: %d\n",
			"\nFAILURES: %d\n")
	case "diff":
		runAndReport(diff,
			"\nDiffed %d of %d remotes. time elapsed: %v\n",
			"\nNEW upstream code: %d\n",
			"\nFAILURES: %d\n")
	default:
		printCommands()
	}
}

func runAndReport(
	fn func(i int, reportSuccess *[]string, reportFail *[]string, wg *sync.WaitGroup,
		mutDiff *sync.Mutex, mutFail *sync.Mutex),
	summaryHeader string,
	successHeader string,
	failHeader string,
) {
	start := time.Now() // stop watch start

	reportSuccess := make([]string, 0, len(DB)) // alloc 100%. no realloc on happy path.
	reportFail := make([]string, 0, 4)          // alloc for low failure rate

	wg := sync.WaitGroup{}
	mutDiff := sync.Mutex{}
	mutFail := sync.Mutex{}
	for i := 0; i < len(DB); i++ { // check each repo for new upstream code
		wg.Add(1)
		go fn(i, &reportSuccess, &reportFail, &wg, &mutDiff, &mutFail)
	}
	wg.Wait()

	// summary report. print # of remotes fetched, duration
	duration := time.Since(start) // stop watch end
	fmt.Printf(summaryHeader,
		len(DB)-len(reportFail), len(DB), duration)

	// diff report. only includes repos that have new data in upstream
	fmt.Printf(successHeader, len(reportSuccess))
	for i := 0; i < len(reportSuccess); i++ {
		fmt.Print(reportSuccess[i])
	}
	// failure report
	fmt.Printf(failHeader, len(reportFail))
	for i := 0; i < len(reportFail); i++ {
		fmt.Print(reportFail[i])
	}
}

// Fetch upstream remote for repo. Repo is identified by index i in DB.
func fetch(i int, reportFetched *[]string, reportFail *[]string,
	wg *sync.WaitGroup, mutFetched *sync.Mutex, mutFail *sync.Mutex) {
	defer wg.Done()

	repo := DB[i]

	// prepare fetch command. example: git fetch upstream
	cmd := exec.Command("git", "fetch", repo.UpstreamAlias) // #nosec G204
	cmd.Dir = expandPath(repo.Folder)
	// Run git fetch! NOTE: cmd.Output() doesn't include the output when git fetch pulls new data.
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		mutFail.Lock()
		*reportFail = append(*reportFail, fmt.Sprintf("%d: %s %v %s\n", i, repo.Folder, cmd.Args, err.Error()))
		mutFail.Unlock()
		return
	}
	newDataFetched := len(stdout) > 0
	if !newDataFetched {
		return
	}
	mutFetched.Lock()
	*reportFetched = append(*reportFetched, fmt.Sprintf("%d: %s %v %s\n",
		i, repo.Folder, cmd.Args, string(stdout)))
	mutFetched.Unlock()
}

func diff(i int, reportDiff *[]string, reportFail *[]string,
	wg *sync.WaitGroup, mutDiff *sync.Mutex, mutFail *sync.Mutex,
) {
	defer wg.Done()

	repo := DB[i]

	// prepare diff command. example: git diff master upstream/master
	cmd := exec.Command("git", "diff", repo.MainBranch, repo.UpstreamAlias+"/"+repo.MainBranch) // #nosec G204
	cmd.Dir = expandPath(repo.Folder)
	// Run git diff!
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		mutFail.Lock()
		*reportFail = append(*reportFail, fmt.Sprintf("%d: %s %v %s\n", i, repo.Folder, cmd.Args, err.Error()))
		mutFail.Unlock()
		return
	}
	hasDifference := len(stdout) > 0
	if !hasDifference {
		return
	}
	mutDiff.Lock()
	// don't incldue the diff output in stdout as it's too verbose to display
	*reportDiff = append(*reportDiff, fmt.Sprintf("%d: %s %v\n",
		i, repo.Folder, cmd.Args))
	mutDiff.Unlock()
}

// expand "~" in path to user's home dir.
func expandPath(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}
	// replace 1st instance of ~ only.
	path = strings.Replace(path, "~", homeDir, 1)
	return path
}
