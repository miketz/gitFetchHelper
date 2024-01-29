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

// GitRepo holds info about a git repo.
type GitRepo struct {
	Folder        string
	UpstreamAlias string
	UpstreamURL   string
	MainBranch    string
}

// DUMMY repo project for testing fetch of new commits. https://github.com/miketz/dummyProj
// var DB = []GitRepo{
// 	{Folder: "~/proj/dummyProj2", UpstreamAlias: "origin"},
// }

// DB is a database (as a slice) of relevant GitRepos. In this case my .emacs.d/ submodules.
var DB = []GitRepo{
	{Folder: "~/.emacs.d/notElpa/paredit", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/combobulate", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/emacs-buttercup", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/swiper", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/ivy-explorer", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/iedit", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/lispy", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/evil", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/evil-leader", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/expand-region.el", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/s.el", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/dash.el", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/transient", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "main"},
	{Folder: "~/.emacs.d/notElpa/with-editor", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "main"},
	{Folder: "~/.emacs.d/notElpa/magit", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "main"},
	{Folder: "~/.emacs.d/notElpa/libegit2", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "main"},
	{Folder: "~/.emacs.d/notElpa/csharp-mode", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/ctrlf", UpstreamAlias: "origin", UpstreamURL: "", MainBranch: "main"},
	{Folder: "~/.emacs.d/notElpa/spinner.el", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/ggtags", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/goto-chg", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/mine/mor", UpstreamAlias: "origin", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/ido-grid.el", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/ov", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/emacs-deferred", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/flx", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/sallet", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/eros", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/elisp-slime-nav", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/emacs-async", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/lua-mode", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/slime", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/slime-company", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/sly", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/eglot", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/lsp-mode", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/f.el", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/ht.el", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/markdown-mode", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/avy", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/rust-mode", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/emacs-racer", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/helm", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/rg.el", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/rainbow-delimiters", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/js2-mode", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/js2-highlight-vars.el", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/json-mode", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/json-snatcher", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/leerzeichen.el", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/citre", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/haskell-mode", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/Emacs-wgrep", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/projectile", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/swift-mode", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/dank-mode", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/darkroom", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/smex", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/pkg-info", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/epl", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/erc-hl-nicks", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/esxml", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/flycheck", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/smarttabs", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/web-mode", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/puni", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/ace-link", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/ace-window", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/adoc-mode", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/markup-faces", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/hydra", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/nov.el", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	// {Folder: "~/.emacs.d/notElpa/num3-mode", UpstreamAlias: "nil", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/nyan-mode", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/php-mode", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/emacs-reformatter", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/zig-mode", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/go-mode.el", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/zoutline", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/yasnippet", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/yaml-mode", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/vimrc-mode", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/unkillable-scratch", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/sicp-info", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/prescient.el", UpstreamAlias: "origin", UpstreamURL: "", MainBranch: "main"},
	{Folder: "~/.emacs.d/notElpa/pos-tip", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/powershell.el", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/highlight-indent-guides", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/icicles", UpstreamAlias: "mirror", UpstreamURL: "", MainBranch: "master"},
	// {Folder: "~/.emacs.d/notElpa/hyperspec", UpstreamAlias: "nil", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/FlamesOfFreedom", UpstreamAlias: "origin", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/Indium", UpstreamAlias: "origin", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/posframe", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/ivy-posframe", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/autothemer", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/company-mode", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/company-web", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/company-lsp", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/web-completion-data", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/fennel-mode", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "main"},
	{Folder: "~/.emacs.d/notElpa/vertico", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "main"},
	{Folder: "~/.emacs.d/notElpa/consult", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "main"},
	{Folder: "~/.emacs.d/notElpa/elisp-bug-hunter", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/typescript.el", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/tide", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/compat.el", UpstreamAlias: "mirror", UpstreamURL: "", MainBranch: "main"},
	{Folder: "~/.emacs.d/notElpa/slime-volleyball", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/macrostep", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/sx.el", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/sunrise-commander", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/visual-fill-column", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "main"},
	{Folder: "~/.emacs.d/notElpa/Emacs-Klondike", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "primary"},
	{Folder: "~/.emacs.d/notElpa/stem-reading-mode.el", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/clojure-mode", UpstreamAlias: "upstream", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/mine/rapid-serial-visual-presentation", UpstreamAlias: "origin", UpstreamURL: "", MainBranch: "master"},
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
	fmt.Printf("Enter a command: [fetch, diff, init]\n")
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
		fetchUpstreamRemotes()
	case "diff":
		listReposWithUpstreamCodeToMerge()
	case "init":
		setUpstreamRemotesIfMissing()
	default:
		printCommands()
	}
}

// Fetch upstream repos, measure time, print reports. The main flow.
func fetchUpstreamRemotes() {
	start := time.Now() // stop watch start

	reportFetched := make([]string, 0, len(DB)) // alloc 100%. no realloc on happy path.
	reportFail := make([]string, 0, 4)          // alloc for low failure rate

	wg := sync.WaitGroup{}
	mutFetched := sync.Mutex{}
	mutFail := sync.Mutex{}
	for i := 0; i < len(DB); i++ { // fetch upstream for each remote.
		wg.Add(1)
		go fetch(i, &reportFetched, &reportFail, &wg, &mutFetched, &mutFail)
	}
	wg.Wait()

	// summary report. print # of remotes fetched, duration
	duration := time.Since(start) // stop watch end
	fmt.Printf("\nFetched %d of %d remotes. time elapsed: %v\n",
		len(DB)-len(reportFail), len(DB), duration)

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

// Set up upstream remotes.
// Useful after a fresh emacs config clone to a new computer. Or after getting latest
// when a new package has been added.
func setUpstreamRemotesIfMissing() {
	start := time.Now() // stop watch start

	reportRemoteCreated := make([]string, 0, len(DB)) // alloc 100%. no realloc on happy path.
	reportFail := make([]string, 0, 4)                // alloc for low failure rate

	wg := sync.WaitGroup{}
	mutRemoteCreated := sync.Mutex{}
	mutFail := sync.Mutex{}
	for i := 0; i < len(DB); i++ { // check each repo for upstream remote, create if missing
		wg.Add(1)
		go setUpstreamRemote(i, &reportRemoteCreated, &reportFail, &wg, &mutRemoteCreated, &mutFail)
	}
	wg.Wait()

	// summary report. print # of remotes checked, duration
	duration := time.Since(start) // stop watch end
	fmt.Printf("\nChecked for upstream remote on %d of %d repos. time elapsed: %v\n",
		len(DB)-len(reportFail), len(DB), duration)

	// remote created report. only includes repos that had a missing upstream remote set.
	fmt.Printf("\nNEW upstream remote set: %d\n", len(reportRemoteCreated))
	for i := 0; i < len(reportRemoteCreated); i++ {
		fmt.Print(reportRemoteCreated[i])
	}
	// failure report
	fmt.Printf("\nFAILURES: %d\n", len(reportFail))
	for i := 0; i < len(reportFail); i++ {
		fmt.Print(reportFail[i])
	}
}

func setUpstreamRemote(i int, reportRemoteCreated *[]string, reportFail *[]string,
	wg *sync.WaitGroup, mutRemoteCreated *sync.Mutex, mutFail *sync.Mutex,
) {
	defer wg.Done()

	repo := DB[i]
	fmt.Printf("repo: %v\n", repo.Folder)
	/*
		"For MOD, create the configured REMOTE-SYM as a remote on the git side."
		  ;; GUARD: mod must be provided
		  (when (null mod)
		    (cl-return-from my-git-remote-create nil))

		  (let* ((remote (my-get-remote mod remote-sym)))
		    ;; GUARD: remote-sym must be configured in `my-modules'
		    (when (null remote)
		      (cl-return-from my-git-remote-create 'remote-not-configured-in-my-modules))

		    ;; GUARD: don't create the remote if it's already setup
		    (when (my-git-remote-setup-p mod remote-sym)
		      (cl-return-from my-git-remote-create 'already-created))

		    ;; OK, now it's safe to create the remote.
		    (let* ((default-directory (module-folder mod))
		           (remote (my-get-remote mod remote-sym))
		           ;; creating the remote here
		           (shell-output (shell-command-to-string (concat "git remote add "
		                                                          (cl-getf remote :alias) " "
		                                                          (cl-getf remote :url)))))
		      ;; TODO: find a better way of detecting error. They could change the error message to
		      ;; not start with "error" and that would break this code.
		      (if (s-starts-with-p "error" shell-output)
		          ;; just return the error msg itself. This string is inconsistent with
		          ;; the symbol return types, but it should be OK as it's just a report
		          ;; of what happened. No real processing on it.
		          (s-trim shell-output)
		          ;; else SUCCESS
		          'remote-created)))
	*/
}

func listReposWithUpstreamCodeToMerge() {
	start := time.Now() // stop watch start

	reportDiff := make([]string, 0, len(DB)) // alloc 100%. no realloc on happy path.
	reportFail := make([]string, 0, 4)       // alloc for low failure rate

	wg := sync.WaitGroup{}
	mutDiff := sync.Mutex{}
	mutFail := sync.Mutex{}
	for i := 0; i < len(DB); i++ { // check each repo for new upstream code
		wg.Add(1)
		go diff(i, &reportDiff, &reportFail, &wg, &mutDiff, &mutFail)
	}
	wg.Wait()

	// summary report. print # of remotes fetched, duration
	duration := time.Since(start) // stop watch end
	fmt.Printf("\nDiffed %d of %d remotes. time elapsed: %v\n",
		len(DB)-len(reportFail), len(DB), duration)

	// diff report. only includes repos that have new data in upstream
	fmt.Printf("\nNEW upstream code: %d\n", len(reportDiff))
	for i := 0; i < len(reportDiff); i++ {
		fmt.Print(reportDiff[i])
	}
	// failure report
	fmt.Printf("\nFAILURES: %d\n", len(reportFail))
	for i := 0; i < len(reportFail); i++ {
		fmt.Print(reportFail[i])
	}
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
