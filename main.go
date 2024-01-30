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

	"golang.org/x/exp/slices"
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
	{Folder: "~/.emacs.d/notElpa/paredit", UpstreamAlias: "upstream", UpstreamURL: "https://mumble.net/~campbell/git/paredit.git", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/combobulate", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/mickeynp/combobulate", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/emacs-buttercup", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/jorgenschaefer/emacs-buttercup", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/swiper", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/abo-abo/swiper", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/ivy-explorer", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/clemera/ivy-explorer", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/iedit", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/victorhge/iedit", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/lispy", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/abo-abo/lispy", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/evil", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/emacs-evil/evil", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/evil-leader", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/cofi/evil-leader", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/expand-region.el", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/magnars/expand-region.el", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/s.el", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/magnars/s.el", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/dash.el", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/magnars/dash.el", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/transient", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/magit/transient", MainBranch: "main"},
	{Folder: "~/.emacs.d/notElpa/with-editor", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/magit/with-editor", MainBranch: "main"},
	{Folder: "~/.emacs.d/notElpa/magit", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/magit/magit", MainBranch: "main"},
	{Folder: "~/.emacs.d/notElpa/libegit2", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/magit/libegit2", MainBranch: "main"},
	{Folder: "~/.emacs.d/notElpa/csharp-mode", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/josteink/csharp-mode", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/ctrlf", UpstreamAlias: "origin", UpstreamURL: "https://github.com/radian-software/ctrlf", MainBranch: "main"},
	{Folder: "~/.emacs.d/notElpa/spinner.el", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/Malabarba/spinner.el", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/ggtags", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/leoliu/ggtags", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/goto-chg", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/emacs-evil/goto-chg", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/mine/mor", UpstreamAlias: "origin", UpstreamURL: "https://github.com/miketz/mor", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/ido-grid.el", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/larkery/ido-grid.el", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/ov", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/emacsorphanage/ov", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/emacs-deferred", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/kiwanami/emacs-deferred", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/flx", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/lewang/flx", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/sallet", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/Fuco1/sallet", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/eros", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/xiongtx/eros", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/elisp-slime-nav", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/purcell/elisp-slime-nav", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/emacs-async", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/jwiegley/emacs-async", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/lua-mode", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/immerrr/lua-mode", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/slime", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/slime/slime", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/slime-company", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/anwyn/slime-company", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/sly", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/joaotavora/sly", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/eglot", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/joaotavora/eglot", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/lsp-mode", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/emacs-lsp/lsp-mode", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/f.el", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/rejeep/f.el", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/ht.el", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/Wilfred/ht.el", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/markdown-mode", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/jrblevin/markdown-mode", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/avy", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/abo-abo/avy", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/rust-mode", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/rust-lang/rust-mode", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/emacs-racer", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/racer-rust/emacs-racer", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/helm", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/emacs-helm/helm", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/rg.el", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/dajva/rg.el", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/rainbow-delimiters", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/Fanael/rainbow-delimiters", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/js2-mode", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/mooz/js2-mode", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/js2-highlight-vars.el", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/unhammer/js2-highlight-vars.el", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/json-mode", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/joshwnj/json-mode", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/json-snatcher", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/Sterlingg/json-snatcher", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/leerzeichen.el", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/fgeller/leerzeichen.el", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/citre", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/universal-ctags/citre", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/haskell-mode", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/haskell/haskell-mode", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/Emacs-wgrep", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/mhayashi1120/Emacs-wgrep", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/projectile", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/bbatsov/projectile", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/swift-mode", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/swift-emacs/swift-mode", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/dank-mode", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/john2x/dank-mode", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/darkroom", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/joaotavora/darkroom", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/smex", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/nonsequitur/smex", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/pkg-info", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/emacsorphanage/pkg-info", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/epl", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/cask/epl", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/erc-hl-nicks", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/leathekd/erc-hl-nicks", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/esxml", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/tali713/esxml", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/flycheck", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/flycheck/flycheck", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/smarttabs", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/jcsalomon/smarttabs", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/web-mode", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/fxbois/web-mode", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/puni", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/AmaiKinono/puni", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/ace-link", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/abo-abo/ace-link", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/ace-window", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/abo-abo/ace-window", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/adoc-mode", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/sensorflo/adoc-mode", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/markup-faces", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/sensorflo/markup-faces", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/hydra", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/abo-abo/hydra", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/nov.el", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/wasamasa/nov.el", MainBranch: "master"},
	// {Folder: "~/.emacs.d/notElpa/num3-mode", UpstreamAlias: "nil", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/nyan-mode", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/TeMPOraL/nyan-mode", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/php-mode", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/emacs-php/php-mode", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/emacs-reformatter", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/purcell/emacs-reformatter", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/zig-mode", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/ziglang/zig-mode", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/go-mode.el", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/dominikh/go-mode.el", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/zoutline", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/abo-abo/zoutline", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/yasnippet", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/joaotavora/yasnippet", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/yaml-mode", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/yoshiki/yaml-mode", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/vimrc-mode", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/mcandre/vimrc-mode", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/unkillable-scratch", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/EricCrosson/unkillable-scratch", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/sicp-info", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/webframp/sicp-info", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/prescient.el", UpstreamAlias: "origin", UpstreamURL: "https://github.com/radian-software/prescient.el", MainBranch: "main"},
	{Folder: "~/.emacs.d/notElpa/pos-tip", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/pitkali/pos-tip", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/powershell.el", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/jschaf/powershell.el", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/highlight-indent-guides", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/DarthFennec/highlight-indent-guides", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/icicles", UpstreamAlias: "mirror", UpstreamURL: "https://github.com/emacsmirror/icicles", MainBranch: "master"},
	// {Folder: "~/.emacs.d/notElpa/hyperspec", UpstreamAlias: "nil", UpstreamURL: "", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/FlamesOfFreedom", UpstreamAlias: "origin", UpstreamURL: "https://github.com/wiz21b/FlamesOfFreedom", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/Indium", UpstreamAlias: "origin", UpstreamURL: "https://github.com/NicolasPetton/Indium", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/posframe", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/tumashu/posframe", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/ivy-posframe", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/tumashu/ivy-posframe", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/autothemer", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/jasonm23/autothemer", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/company-mode", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/company-mode/company-mode", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/company-web", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/osv/company-web", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/company-lsp", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/tigersoldier/company-lsp", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/web-completion-data", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/osv/web-completion-data", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/fennel-mode", UpstreamAlias: "upstream", UpstreamURL: "https://git.sr.ht/~technomancy/fennel-mode", MainBranch: "main"},
	{Folder: "~/.emacs.d/notElpa/vertico", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/minad/vertico", MainBranch: "main"},
	{Folder: "~/.emacs.d/notElpa/consult", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/minad/consult", MainBranch: "main"},
	{Folder: "~/.emacs.d/notElpa/elisp-bug-hunter", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/Malabarba/elisp-bug-hunter", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/typescript.el", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/emacs-typescript/typescript.el", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/tide", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/ananthakumaran/tide", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/compat.el", UpstreamAlias: "mirror", UpstreamURL: "https://github.com/phikal/compat.el", MainBranch: "main"},
	{Folder: "~/.emacs.d/notElpa/slime-volleyball", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/fitzsim/slime-volleyball", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/macrostep", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/joddie/macrostep", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/sx.el", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/vermiculus/sx.el", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/sunrise-commander", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/sunrise-commander/sunrise-commander", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/visual-fill-column", UpstreamAlias: "upstream", UpstreamURL: "https://codeberg.org/joostkremers/visual-fill-column", MainBranch: "main"},
	{Folder: "~/.emacs.d/notElpa/Emacs-Klondike", UpstreamAlias: "upstream", UpstreamURL: "https://codeberg.org/WammKD/Emacs-Klondike", MainBranch: "primary"},
	{Folder: "~/.emacs.d/notElpa/stem-reading-mode.el", UpstreamAlias: "upstream", UpstreamURL: "https://gitlab.com/wavexx/stem-reading-mode.el.git", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/clojure-mode", UpstreamAlias: "upstream", UpstreamURL: "https://github.com/clojure-emacs/clojure-mode", MainBranch: "master"},
	{Folder: "~/.emacs.d/notElpa/mine/rapid-serial-visual-presentation", UpstreamAlias: "origin", UpstreamURL: "https://github.com/miketz/rapid-serial-visual-presentation", MainBranch: "master"},
}

var homeDir string
var isMsWindows = strings.HasPrefix(runtime.GOOS, "windows")

// initialize global variables. At the moment only homeDir.
func initGlobals() error {
	usr, err := user.Current()
	if err != nil {
		return err
	}
	homeDir = usr.HomeDir
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
	fmt.Printf("\nChecked for upstream remote on %d repos. time elapsed: %v\n",
		len(DB), duration)

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
	var aliases []string

	// prepare command to get remote aliases. example: git remote
	cmd := exec.Command("git", "remote") // #nosec G204
	cmd.Dir = expandPath(repo.Folder)
	remoteOutput, err := cmd.CombinedOutput()
	if err != nil {
		mutFail.Lock()
		*reportFail = append(*reportFail, fmt.Sprintf("%d: %s %v %s\n", i, repo.Folder, cmd.Args, err.Error()))
		mutFail.Unlock()
		return
	}
	hasRemotes := len(remoteOutput) > 0
	if !hasRemotes {
		goto CREATE_UPSTREAM
	}
	// stdout might be something like:
	//     origin
	//     upstream
	// split the raw shell output to a list of alias strings
	aliases = strings.Split(string(remoteOutput), "\n")
	if slices.Contains(aliases, repo.UpstreamAlias) {
		// check if url matches url in DB. git command: git remote get-url {upstream}
		cmd = exec.Command("git", "remote", "get-url", repo.UpstreamAlias) // #nosec G204
		cmd.Dir = expandPath(repo.Folder)
		urlOutput, err := cmd.CombinedOutput()
		if err != nil {
			mutFail.Lock()
			*reportFail = append(*reportFail, fmt.Sprintf("%d: %s %v %s\n", i, repo.Folder, cmd.Args, err.Error()))
			mutFail.Unlock()
			return
		}
		upstreamURL := string(urlOutput)
		newLine := "\n"
		if isMsWindows {
			newLine = "\r\n"
		}
		upstreamURL = strings.Trim(upstreamURL, newLine)
		mismatch := upstreamURL != repo.UpstreamURL
		if mismatch {
			mutFail.Lock()
			// note: in msg below config: and actual: are same len for visual alignment of url strings.
			*reportFail = append(*reportFail, fmt.Sprintf("%d: %s mismatched upstream URL.\nconfig: %s\nactual: %s\n\n",
				i, repo.Folder, repo.UpstreamURL, upstreamURL))
			mutFail.Unlock()
			return
		}
		return // no reporting needed for "normal" case when url matches.
	}
CREATE_UPSTREAM:
	// run git command: git remote add {alias} {url}
	cmd = exec.Command("git", "remote", "add", repo.UpstreamAlias, repo.UpstreamURL) // #nosec G204
	cmd.Dir = expandPath(repo.Folder)
	createOutput, err := cmd.CombinedOutput()
	if err != nil {
		mutFail.Lock()
		*reportFail = append(*reportFail, fmt.Sprintf("%d: %s %v %s\n", i, repo.Folder, cmd.Args, err.Error()))
		mutFail.Unlock()
		return
	}
	var createStr = string(createOutput)
	// TODO: find a better way of detecting error. They could change the error message to
	// not start with "error" and that would break this code.
	if strings.HasPrefix(createStr, "error") {
		mutFail.Lock()
		*reportFail = append(*reportFail, fmt.Sprintf("%d: %s %s\n", i, repo.Folder, createStr))
		mutFail.Unlock()
		return
	}
	// SUCCESS, remote created
	mutRemoteCreated.Lock()
	*reportRemoteCreated = append(*reportRemoteCreated, fmt.Sprintf("%d: %s %v\n",
		i, repo.Folder, cmd.Args))
	mutRemoteCreated.Unlock()
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
