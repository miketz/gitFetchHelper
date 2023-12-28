package main

import (
	"fmt"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
	"sync"
	"time"
)

type RemoteSym int

const (
	mine RemoteSym = iota
	upstream
	mirrorUpstream
)

type Remote struct {
	Sym   RemoteSym
	Alias string
	URL   string
}

type Dependency struct {
	Name    string
	Version string
}

type RepoTable struct {
	Name          []string
	Comment       []string
	Folder        []string
	Remotes       [][]Remote
	RemoteDefault []RemoteSym // default remote to pull/push
	// The canonical branch used by the upstream. Usually "master" or "main".
	MainBranch []string
	// Usually the same as `main-branch' but sometimes a private "mine" branch
	// with a few odd tweaks. This is the branch I use locally on my side.
	UseBranch []string
	/* git SHA (or other vc equiv) to use. Alternative to branch as branch means
	   you are following latest tip of that branch. A specific commit is more
	   exact. However this will be rarely used as the moment I start developing
	   my own code I'll need a branch to avoid a detached HEAD state. This is
	   more for if I want *this* config to control the state of the package
	   rather than the git branch itself.*/
	UseCommit  []string
	DependHard [][]Dependency // required or important dependencies.
	DependSoft [][]Dependency // optional dependencies. Or only needed for the tests.
	/* For when packages bundle dependencies. For informational purposes so I
	   don't try to install something when I don't need to. */
	DependBundled [][]Dependency
}

// From the configured remotes, find the "upstream", return it's alias.
func (t *RepoTable) UpstreamAlias(i int) (string, error) {
	for _, rem := range t.Remotes[i] {
		if rem.Sym == upstream {
			alias := strings.TrimSpace(rem.Alias)
			if len(alias) == 0 {
				return "", fmt.Errorf("Upstream alias not configured for %s", t.Name[i])
			}
			return alias, nil
		}
	}
	return "", fmt.Errorf("Upstream not configured for %s", t.Name[i])
}

// Number of repos configured.
func (t *RepoTable) Count() int {
	return len(t.Name)
}

// DUMMY repo project for testing fetch of new commits. https://github.com/miketz/dummyProj
// var DB = []GitRepo{
// 	{Folder: "~/proj/dummyProj2", UpstreamAlias: "origin"},
// }

var numRepos int = 1
var DB = RepoTable{
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

// {Folder: "~/.emacs.d/notElpa/paredit", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/combobulate", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/emacs-buttercup", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/swiper", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/ivy-explorer", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/iedit", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/lispy", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/evil", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/evil-leader", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/expand-region.el", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/s.el", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/dash.el", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/transient", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/with-editor", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/magit", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/libegit2", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/csharp-mode", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/ctrlf", UpstreamAlias: "origin"},
// {Folder: "~/.emacs.d/notElpa/spinner.el", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/ggtags", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/goto-chg", UpstreamAlias: "upstream"},
// // {Folder: "~/.emacs.d/notElpa/mine/mor", UpstreamAlias: "nil"},
// {Folder: "~/.emacs.d/notElpa/ido-grid.el", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/ov", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/emacs-deferred", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/flx", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/sallet", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/eros", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/elisp-slime-nav", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/emacs-async", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/lua-mode", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/slime", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/slime-company", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/sly", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/eglot", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/lsp-mode", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/f.el", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/ht.el", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/markdown-mode", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/avy", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/rust-mode", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/emacs-racer", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/helm", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/rg.el", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/rainbow-delimiters", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/js2-mode", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/js2-highlight-vars.el", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/json-mode", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/json-snatcher", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/leerzeichen.el", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/citre", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/haskell-mode", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/Emacs-wgrep", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/projectile", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/swift-mode", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/dank-mode", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/darkroom", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/smex", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/pkg-info", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/epl", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/erc-hl-nicks", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/esxml", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/flycheck", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/smarttabs", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/web-mode", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/puni", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/ace-link", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/ace-window", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/adoc-mode", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/markup-faces", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/hydra", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/nov.el", UpstreamAlias: "upstream"},
// // {Folder: "~/.emacs.d/notElpa/num3-mode", UpstreamAlias: "nil"},
// {Folder: "~/.emacs.d/notElpa/nyan-mode", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/php-mode", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/emacs-reformatter", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/zig-mode", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/go-mode.el", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/zoutline", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/yasnippet", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/yaml-mode", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/vimrc-mode", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/unkillable-scratch", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/sicp-info", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/prescient.el", UpstreamAlias: "origin"},
// {Folder: "~/.emacs.d/notElpa/pos-tip", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/powershell.el", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/highlight-indent-guides", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/icicles", UpstreamAlias: "mirror"},
// // {Folder: "~/.emacs.d/notElpa/hyperspec", UpstreamAlias: "nil"},
// {Folder: "~/.emacs.d/notElpa/FlamesOfFreedom", UpstreamAlias: "origin"},
// {Folder: "~/.emacs.d/notElpa/Indium", UpstreamAlias: "origin"},
// {Folder: "~/.emacs.d/notElpa/posframe", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/ivy-posframe", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/autothemer", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/company-mode", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/company-web", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/company-lsp", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/web-completion-data", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/fennel-mode", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/vertico", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/consult", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/elisp-bug-hunter", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/typescript.el", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/tide", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/compat.el", UpstreamAlias: "mirror"},
// {Folder: "~/.emacs.d/notElpa/slime-volleyball", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/macrostep", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/sx.el", UpstreamAlias: "upstream"},
// {Folder: "~/.emacs.d/notElpa/sunrise-commander", UpstreamAlias: "upstream"},

var homeDir string

// initialize global variables.
func initGlobals() error {
	// homeDir
	usr, err := user.Current()
	if err != nil {
		return err
	}
	homeDir = usr.HomeDir
	var isMsWindows = strings.HasPrefix(runtime.GOOS, "windows")
	if isMsWindows {
		homeDir += "/AppData/Local"
	}

	// DB
	DB.Name[0] = "paredit"
	DB.Comment[0] = ""
	DB.Folder[0] = "~/.emacs.d/notElpa/paredit"
	DB.Remotes[0] = append(DB.Remotes[0], Remote{Sym: mine, Alias: "origin", URL: "https://github.com/miketz/paredit"})
	DB.Remotes[0] = append(DB.Remotes[0], Remote{Sym: upstream, Alias: "upstream", URL: "https://mumble.net/~campbell/git/paredit.git"})
	DB.RemoteDefault[0] = mine
	DB.MainBranch[0] = "master"
	DB.UseBranch[0] = "master"
	DB.DependHard[0] = nil
	DB.DependSoft[0] = nil
	DB.DependBundled[0] = nil
	return nil
}

// get all the submodules
// git config --file .gitmodules --get-regexp path | awk '{ print $2 }'
// cmd := exec.Command("git", "config", "--file", ".gitmodules", "--get-regexp", "path", "|", "awk", "'{ print $2 }'")

func main() {
	err := initGlobals()
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		return
	}
	fetchUpstreamRemotes()
}

// Fetch upstream repos, measure time, print reports. The main flow.
func fetchUpstreamRemotes() {
	start := time.Now() // stop watch start

	reportFetched := make([]string, 0, DB.Count()) // alloc 100%. no realloc on happy path.
	reportFail := make([]string, 0, 4)             // alloc for low failure rate

	wg := sync.WaitGroup{}
	mutFetched := sync.Mutex{}
	mutFail := sync.Mutex{}
	for i := 0; i < DB.Count(); i++ { // fetch upstream for each remote.
		wg.Add(1)
		go fetch(i, &reportFetched, &reportFail, &wg, &mutFetched, &mutFail)
	}
	wg.Wait()

	// summary report. print # of remotes fetched, duration
	duration := time.Since(start) // stop watch end
	fmt.Printf("\nFetched %d of %d remotes. time elapsed: %v\n",
		DB.Count()-len(reportFail), DB.Count(), duration)

	// fetch report. only includes repos that had new data to fetch.
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

// Fetch upstream remote for repo. Repo is identified by index i in DB.
func fetch(i int, reportFetched *[]string, reportFail *[]string,
	wg *sync.WaitGroup, mutFetched *sync.Mutex, mutFail *sync.Mutex,
) {
	defer wg.Done()

	// prepare fetch command. example: git fetch upstream
	upstreamAlias, err := DB.UpstreamAlias(i)
	if err != nil {
		mutFail.Lock()
		*reportFail = append(*reportFail, fmt.Sprintf("%d: %s %s\n", i, DB.Folder[i], err.Error()))
		mutFail.Unlock()
		return
	}
	cmd := exec.Command("git", "fetch", upstreamAlias) // #nosec G204
	cmd.Dir = expandPath(DB.Folder[i])
	// Run git fetch! NOTE: cmd.Output() doesn't include the output when git fetch pulls new data.
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		mutFail.Lock()
		*reportFail = append(*reportFail, fmt.Sprintf("%d: %s %v %s\n", i, DB.Folder[i], cmd.Args, err.Error()))
		mutFail.Unlock()
		return
	}
	newDataFetched := len(stdout) > 0
	if !newDataFetched {
		return
	}
	mutFetched.Lock()
	*reportFetched = append(*reportFetched, fmt.Sprintf("%d: %s %v %s\n",
		i, DB.Folder[i], cmd.Args, string(stdout)))
	mutFetched.Unlock()
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
