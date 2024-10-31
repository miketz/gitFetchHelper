package main

import (
	"bufio"
	"path/filepath"
	// "encoding/json".
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/komkom/jsonc/jsonc"
	"golang.org/x/exp/slices"
)

// Info about the server side remote.
type Remote struct {
	// A special tag to identify the meaning of the Remote.
	// Alias is not enough to convey meaning as it's often "origin" by default after a git clone.
	// "upstream" represents the original or canonical repo of the project.
	// "mine" is my fork.
	Sym string `json:"sym"`
	// Git remote URL
	URL string `json:"url"`
	// The alias used by git to reference the remote. May match the Sym value
	// but not always. For example my fork will usually have an alias of "origin" with a
	// Sym of "mine"
	Alias string `json:"alias"`
}

// GitRepo holds info about a git repo. In this case my .emacs.d/notElpa submodules.
type GitRepo struct {
	// Simple short name of the project. In the case of Emacs packages make this
	// the feature symbol used by (require 'feature).
	Name string `json:"name"`
	// Top level root folder of the project.
	Folder string `json:"folder"`
	// List of remotes. Usually will be 2 remotes. It's expected that most repos will have
	// a remote of Sym "mine" and "upstream", however there can be unlimited remotes. The
	// Sym field is used to identify the special remotes in the slice.
	Remotes []Remote `json:"remotes"`
	// The remote we are tracking against. In my case this is usually my fork specified via sym "mine".
	RemoteDefaultSym string `json:"remoteDefault"`
	// The branch we are interested in following for this Emacs package.
	// It may be a "develop" branch if we are interested in the bleeding edge.
	BranchMain string `json:"branchMain"`
	// The branch we will use. Usually the same as BranchMain. But sometimes I
	// will use a custom branch derived from BranchMain for small modifications,
	// even if it's a minor change like adding to .gitignore.
	BranchUse string `json:"branchUse"`
	// not a git submodule
	IsYolo bool `json:"isYolo"`
}

// get the "upstream" remote for the git repo.
func (r *GitRepo) RemoteUpstream() (Remote, error) {
	return r.GetRemoteBySym("upstream")
}

// get the "mine" remote for the git repo. This is usually my fork or my own project.
func (r *GitRepo) RemoteMine() (Remote, error) {
	return r.GetRemoteBySym("mine")
}

// get the "default" remote specified by "RemoteDefaultSym" for the git repo.
// Sometimes this may be the upstream, but usually my fork or my own project.
func (r *GitRepo) RemoteDefault() (Remote, error) {
	return r.GetRemoteBySym(r.RemoteDefaultSym)
}

// get the remote based on symbol "sym".
// sym is a semantic meaning for the remote separate from it's alias name.
func (r *GitRepo) GetRemoteBySym(sym string) (Remote, error) {
	for _, rem := range r.Remotes {
		if rem.Sym == sym {
			return rem, nil
		}
	}
	return Remote{}, fmt.Errorf("no " + sym + " remote configured for " + r.Name + " in repos.jsonc")
}

// get the hash of a branch in this GitRepo.
func (r *GitRepo) GetHash(branchName string) (string, error) {
	cmd := exec.Command("git", "rev-parse", branchName)
	cmd.Dir = expandPath(r.Folder)
	hash, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	hashStr := strings.Trim(string(hash), newLine)
	return hashStr, nil
}

// DB is a database (as a slice) of relevant GitRepos. In this case my .emacs.d/ submodules.
var DB = []GitRepo{}

// my home directory. where .emacs.d/ is stored.
var homeDir string

// new line character.
var newLine = "\n"

// initialize global variables.
func initGlobals() error {
	isMsWindows := strings.HasPrefix(runtime.GOOS, "windows")

	var err error
	DB, err = getRepoData()
	if err != nil {
		return err
	}

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
	fmt.Printf("Enter a command: [fetch, diff, init, init2, init3]\n")
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
	case "init2":
		switchToBranches()
	case "init3":
		cloneYoloRepos()
	default:
		printCommands()
	}
}

// read repos.jsonc into memory
func getRepoData() ([]GitRepo, error) {
	jsonFile, err := os.Open("./repos.jsonc")
	if err != nil {
		fmt.Printf("opening json file: %v\n", err.Error())
		return nil, err
	}
	defer jsonFile.Close()

	repos := make([]GitRepo, 0, 256)
	// jsonParser := json.NewDecoder(jsonFile)
	reader := bufio.NewReader(jsonFile)
	jsonParser, err := jsonc.NewDecoder(reader)
	if err != nil {
		fmt.Printf("failed to create jsonc decoder: %v\n", err.Error())
		return nil, err
	}
	err = jsonParser.Decode(&repos)
	if err != nil {
		fmt.Printf("parsing config file: %v\n", err.Error())
		return nil, err
	}
	return repos, nil
}

// Fetch upstream repos, measure time, print reports. The main flow.
func fetchUpstreamRemotes() { //nolint:dupl
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

	// get upstream remote info
	upstream, err := repo.RemoteUpstream()
	if err != nil {
		mutFail.Lock()
		*reportFail = append(*reportFail, fmt.Sprintf("%d: %s %s\n", i, repo.Folder, err.Error()))
		mutFail.Unlock()
		return
	}
	// prepare fetch command. example: git fetch upstream
	cmd := exec.Command("git", "fetch", upstream.Alias) // #nosec G204
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
func setUpstreamRemotesIfMissing() { //nolint:dupl
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

	// get configured upstream remote info
	upstream, err := repo.RemoteUpstream()
	if err != nil {
		mutFail.Lock()
		*reportFail = append(*reportFail, fmt.Sprintf("%d: %s %s\n", i, repo.Folder, err.Error()))
		mutFail.Unlock()
		return
	}

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
	// remoteOutput might be something like:
	//     origin
	//     upstream
	// split the raw shell output to a list of alias strings
	aliases = strings.Split(string(remoteOutput), newLine)
	if slices.Contains(aliases, upstream.Alias) {
		// check if URL matches URL in DB. git command: git remote get-url {upstream}
		cmd = exec.Command("git", "remote", "get-url", upstream.Alias) // #nosec G204
		cmd.Dir = expandPath(repo.Folder)
		urlOutput, err := cmd.CombinedOutput() //nolint:govet
		if err != nil {
			mutFail.Lock()
			*reportFail = append(*reportFail, fmt.Sprintf("%d: %s %v %s\n", i, repo.Folder, cmd.Args, err.Error()))
			mutFail.Unlock()
			return
		}
		upstreamURL := strings.Trim(string(urlOutput), newLine)
		mismatch := upstreamURL != upstream.URL
		if mismatch {
			mutFail.Lock()
			// note: in msg below config: and actual: are same len for visual alignment of url strings.
			*reportFail = append(*reportFail, fmt.Sprintf("%d: %s mismatched upstream URL.\nconfig: %s\nactual: %s\n\n",
				i, repo.Folder, upstream.URL, upstreamURL))
			mutFail.Unlock()
			return
		}
		return // no reporting needed for "normal" case when url matches.
	}
CREATE_UPSTREAM:
	// run git command: git remote add {alias} {url}
	cmd = exec.Command("git", "remote", "add", upstream.Alias, upstream.URL) //nolint:gosec // #nosec G204
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

func listReposWithUpstreamCodeToMerge() { //nolint:dupl
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

	// get current checked out branch name.
	// It may be the configured repo.MainBranch, or custom "mine", or empty "" (detached head)
	branchName, err := getCurrBranch(&repo)
	if err != nil {
		mutFail.Lock()
		*reportFail = append(*reportFail, fmt.Sprintf("%d: %s %s\n", i, repo.Folder, "problem getting current branch name: "+err.Error()))
		mutFail.Unlock()
		return
	}

	// when comparing our current to upstream, we don't care about any custom changes in "mine" as those are expected difference from upstream.
	// instead compare repo.BranchMain if possible
	// or use HEAD if we are in a detached head state.
	if branchName == "" {
		// detached head.
		branchName = "HEAD"
	} else if branchName != repo.BranchMain {
		// on my custom branch. Don't' compare that.
		branchName = repo.BranchMain
	}

	// get configured upstream remote info
	upstream, err := repo.RemoteUpstream()
	if err != nil {
		mutFail.Lock()
		*reportFail = append(*reportFail, fmt.Sprintf("%d: %s %s\n", i, repo.Folder, err.Error()))
		mutFail.Unlock()
		return
	}

	// prepare diff command. example: git diff master upstream/master
	cmd := exec.Command("git", "diff",
		branchName,
		upstream.Alias+"/"+repo.BranchMain) // #nosec G204
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
	// don't include the diff output in stdout as it's too verbose to display
	*reportDiff = append(*reportDiff, fmt.Sprintf("%d: %s %v\n",
		i, repo.Folder, cmd.Args))
	mutDiff.Unlock()
}

// Checkout the "UseBranch" for each git submodule.
// Useful after a fresh emacs config clone to a new computer to avoid detached head state.
func switchToBranches() { //nolint:dupl
	start := time.Now() // stop watch start

	reportBranchChange := make([]string, 0, len(DB)) // alloc 100%. no realloc on happy path.
	reportFail := make([]string, 0, 4)               // alloc for low failure rate

	wg := sync.WaitGroup{}
	mutBranchChange := sync.Mutex{}
	mutFail := sync.Mutex{}
	for i := 0; i < len(DB); i++ { // check each repo for upstream remote, create if missing
		wg.Add(1)
		go switchToBranch(i, &reportBranchChange, &reportFail, &wg, &mutBranchChange, &mutFail)
	}
	wg.Wait()

	// summary report. print # of branches checked out, duration
	duration := time.Since(start) // stop watch end
	fmt.Printf("\nChecked for UseBranch on %d repos. time elapsed: %v\n",
		len(DB), duration)

	// branch change report. only includes repos that needed a switch to UseBranch.
	fmt.Printf("\nBranch change actions: %d\n", len(reportBranchChange))
	for i := 0; i < len(reportBranchChange); i++ {
		fmt.Print(reportBranchChange[i])
	}
	// failure report
	fmt.Printf("\nFAILURES: %d\n", len(reportFail))
	for i := 0; i < len(reportFail); i++ {
		fmt.Print(reportFail[i])
	}
}

// Checkout the "UseBranch" for a git repo. Git repo identified by index i from DB.
func switchToBranch(i int, reportBranchChange *[]string, reportFail *[]string,
	wg *sync.WaitGroup, mutBranchChange *sync.Mutex, mutFail *sync.Mutex,
) {
	defer wg.Done()

	repo := DB[i]

	// get current checked out branch name.
	// It may be the configured repo.MainBranch, or custom "mine", or empty "" (detached head)
	branchName, err := getCurrBranch(&repo)
	if err != nil {
		mutFail.Lock()
		*reportFail = append(*reportFail, fmt.Sprintf("%d: %s %s\n", i, repo.Folder, "problem getting current branch name: "+err.Error()))
		mutFail.Unlock()
		return
	}

	remoteDefault, err := repo.RemoteDefault()
	if err != nil {
		mutFail.Lock()
		*reportFail = append(*reportFail, fmt.Sprintf("%d: %s %s\n", i, repo.Folder, err.Error()))
		mutFail.Unlock()
		return
	}
	// switch to branch if not already on it.
	if branchName != repo.BranchUse {
		hasLocalBranch, err2 := hasLocalBranch(&repo, repo.BranchUse)
		if err2 != nil {
			mutFail.Lock()
			*reportFail = append(*reportFail, fmt.Sprintf("%d: %s %s\n", i, repo.Folder, "problem checking for local branch existence: "+err2.Error()))
			mutFail.Unlock()
			return
		}
		// Action #1
		// prepare branch switch command. example: git checkout --track origin/master
		var cmd *exec.Cmd
		if hasLocalBranch {
			cmd = exec.Command("git", "checkout", repo.BranchUse) // #nosec G204
		} else {
			cmd = exec.Command("git", "checkout", "--track", remoteDefault.Alias+"/"+repo.BranchUse) // #nosec G204
		}
		cmd.Dir = expandPath(repo.Folder)
		// Run branch switch!
		_, err2 = cmd.CombinedOutput()
		if err2 != nil {
			mutFail.Lock()
			*reportFail = append(*reportFail, fmt.Sprintf("%d: %s %v %s\n", i, repo.Folder, cmd.Args, err2.Error()))
			mutFail.Unlock()
			return
		}

		// track the fact we just switched branches
		mutBranchChange.Lock()
		*reportBranchChange = append(*reportBranchChange, fmt.Sprintf("%d: %s %v\n",
			i, repo.Folder, cmd.Args))
		mutBranchChange.Unlock()
	}

	// make sure branch is up to date with origin
	hashLocalUseBranch, err := repo.GetHash(repo.BranchUse)
	if err != nil {
		mutFail.Lock()
		*reportFail = append(*reportFail, fmt.Sprintf("%d: %s %s\n", i, repo.Folder, err.Error()))
		mutFail.Unlock()
		return
	}
	hashRemoteUseBranch, err := repo.GetHash(remoteDefault.Alias + "/" + repo.BranchUse)
	if err != nil {
		mutFail.Lock()
		*reportFail = append(*reportFail, fmt.Sprintf("%d: %s %s\n", i, repo.Folder, err.Error()))
		mutFail.Unlock()
		return
	}

	if hashLocalUseBranch != hashRemoteUseBranch {
		// Action #2.
		// force reset to remote version of branch
		cmd := exec.Command("git", "reset", "--hard", remoteDefault.Alias+"/"+repo.BranchUse) // #nosec G204
		cmd.Dir = expandPath(repo.Folder)
		// Run branch switch!
		_, err = cmd.CombinedOutput()
		if err != nil {
			mutFail.Lock()
			*reportFail = append(*reportFail, fmt.Sprintf("%d: %s %v %s\n", i, repo.Folder, cmd.Args, err.Error()))
			mutFail.Unlock()
			return
		}
		// track the fact we just reset the branch to match origin
		mutBranchChange.Lock()
		*reportBranchChange = append(*reportBranchChange, fmt.Sprintf("%d: %s %v\n",
			i, repo.Folder, cmd.Args))
		mutBranchChange.Unlock()
	}
}

// for each "yolo" repo, clone it if it does not yet exist
func cloneYoloRepos() {
	start := time.Now() // stop watch start

	reportClone := make([]string, 0, len(DB)) // alloc 100%. no realloc on happy path.
	reportFail := make([]string, 0, 4)        // alloc for low failure rate

	yoloFolder := expandPath("~/.emacs.d/notElpaYolo")
	yoloFolderExists, _ := exists(yoloFolder)
	if !yoloFolderExists {
		if err := os.Mkdir(yoloFolder, os.ModePerm); err != nil {
			fmt.Printf("Failed to create folder %s, err: %v\n", yoloFolder, err)
			return
		}
	}

	wg := sync.WaitGroup{}
	mutClone := sync.Mutex{}
	mutFail := sync.Mutex{}
	yoloCnt := 0
	for i := 0; i < len(DB); i++ { // clone each "yolo" repo if missing
		if !DB[i].IsYolo {
			continue
		}
		yoloCnt++
		wg.Add(1)
		go cloneYolo(i, &reportClone, &reportFail, &wg, &mutClone, &mutFail)
	}
	wg.Wait()

	// summary report. print # of branches checked out, duration
	duration := time.Since(start) // stop watch end
	fmt.Printf("\nChecked for existence of %d yolo repos, clone if not exist. time elapsed: %v\n",
		yoloCnt, duration)

	// clone report. only includes repos that needed to be cloned
	fmt.Printf("\nClones performed: %d\n", len(reportClone))
	for i := 0; i < len(reportClone); i++ {
		fmt.Print(reportClone[i])
	}
	// failure report
	fmt.Printf("\nFAILURES: %d\n", len(reportFail))
	for i := 0; i < len(reportFail); i++ {
		fmt.Print(reportFail[i])
	}
}

// clone the "yolo" repo if it does not exist in target location.
func cloneYolo(i int, reportClone *[]string, reportFail *[]string,
	wg *sync.WaitGroup, mutClone *sync.Mutex, mutFail *sync.Mutex,
) {
	defer wg.Done()

	repo := DB[i]
	if !repo.IsYolo { // GUARD: for "yolo" repos only, not submodules
		return
	}

	folder := expandPath(repo.Folder)
	folderExists, _ := exists(folder)
	if folderExists {
		// assume folder is the cloned repo. don't bother verifying git repo status, etc. maybe later?
		// return early early, nothing to clone
		return
	}

	// get default remote
	remote, err := repo.RemoteDefault()
	if err != nil {
		mutFail.Lock()
		*reportFail = append(*reportFail, fmt.Sprintf("%d: %s %s\n", i, repo.Folder, err.Error()))
		mutFail.Unlock()
		return
	}

	// git clone --depth 1 --branch master --no-single-branch remoteUrl
	// using a shallow clone for performance. But still get the tip of each branch
	// with "--no-single-branch" to avoid a headache later when trying to switch to
	// other branches. git makes you go through convoluted steps if you don't get
	// the branches during the clone.
	// for full history manually run: git fetch --unshallow
	cmd := exec.Command("git", "clone", "--depth", "1", "--branch", repo.BranchUse, "--no-single-branch", remote.URL) // #nosec G204
	// go to parent folder 1 level up to execute the clone command.
	// because the target folder does not exist until after clone
	cmd.Dir = parentDir(folder)
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		mutFail.Lock()
		*reportFail = append(*reportFail, fmt.Sprintf("%d: %s %v %s\n", i, repo.Folder, cmd.Args, err.Error()))
		mutFail.Unlock()
		return
	}
	// TODO: make sure there's nothing else i need to check for clone success/fail
	mutClone.Lock()
	*reportClone = append(*reportClone, fmt.Sprintf("%d: %s %v %s\n",
		i, repo.Folder, cmd.Args, string(stdout)))
	mutClone.Unlock()
}

// get current checked out branch name for a GitRepo.
func getCurrBranch(repo *GitRepo) (string, error) {
	// get current checked out branch name.
	// It may be the configured repo.MainBranch, or custom "mine", or empty "" (detached head)
	cmdBranch := exec.Command("git", "branch", "--show-current")
	cmdBranch.Dir = expandPath(repo.Folder)
	branchOut, err := cmdBranch.CombinedOutput()
	if err != nil {
		return "", err
	}
	branchName := strings.Trim(string(branchOut), newLine)
	return branchName, nil
}

// True if the repo has a local version of the branch. (ignore remote tracking branches).
func hasLocalBranch(repo *GitRepo, branchName string) (bool, error) {
	cmd := exec.Command("git", "branch") // #nosec G204
	cmd.Dir = expandPath(repo.Folder)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, err
	}
	if len(output) == 0 { // no branches at all!
		return false, nil
	}
	// output might be something like:
	//     master
	//     mine
	// split the raw shell output to a list of strings
	branches := strings.Split(string(output), newLine)
	// trim white space and * character from branch names
	for i, br := range branches {
		branches[i] = strings.Trim(br, "\n *")
	}
	hasBranch := slices.Contains(branches, branchName)
	return hasBranch, nil
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

// Returns true if folder path is inside a git repo.
func isInGitRepo(path string) bool {
	// git rev-parse --is-inside-work-tree
	// "true\n"
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree") // #nosec G204
	cmd.Dir = expandPath(path)
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		// git rev-parse throws a fatal err if not in a git repo.
		// So just interpret err as not in a repo. (don't propagate the err up the chain)
		return false
	}
	return string(stdout) == "true\n"
}

// Returns true if folder path is inside a git submodule.
func isInGitSubmodule(path string) bool {
	// git rev-parse --show-superproject-working-tree
	// len(output) > 0
	cmd := exec.Command("git", "rev-parse", "--show-superproject-working-tree") // #nosec G204
	cmd.Dir = expandPath(path)
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		// git rev-parse throws a fatal err if not in a git repo.
		// So just interpret err as not in a git submodule. (don't propagate the err up the chain)
		return false
	}
	return len(stdout) > 0
}

// returns true if file or directory exists
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// get a path string to the parent dir of path.
// string manipulation, directories do not need to exist on disk.
func parentDir(path string) string {
	path = expandPath(path)
	parentDir := filepath.Join(path, "../")
	return parentDir
}
