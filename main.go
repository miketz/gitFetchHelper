package main

import (
	"encoding/json"
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

// Info about the server side remote
type Remote struct {
	Sym   string `json:"sym"`
	Url   string `json:"url"`
	Alias string `json:"alias"`
}

// GitRepo holds info about a git repo. In this case my .emacs.d/notElpa submodules.
type GitRepo struct {
	Name          string   `json:"name"`
	Folder        string   `json:"folder"`
	Remotes       []Remote `json:"remotes"`
	RemoteDefault string   `json:"remoteDefault"`
	BranchMain    string   `json:"branchMain"`
	BranchUse     string   `json:"branchUse"`
}

// get the "upstream" remote for the git repo
func (r *GitRepo) RemoteUpstream() Remote {
	for _, rem := range r.Remotes {
		if rem.Sym == "upstream" {
			return rem
		}
	}
	return nil // no upstream remote configured.
}

// get the "mine" remote for the git repo. This is usually my fork or my own project.
func (r *GitRepo) RemoteMine() Remote {
	for _, rem := range r.Remotes {
		if rem.Sym == "mine" {
			return rem
		}
	}
	return nil // no mine remote configured.
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
	fmt.Printf("Enter a command: [fetch, diff, init, init2]\n")
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
	default:
		printCommands()
	}
}

func getRepoData() ([]GitRepo, error) {
	jsonFile, err := os.Open("./repos.json")
	if err != nil {
		fmt.Printf("opening json file: %v\n", err.Error())
		return nil, err
	}
	defer jsonFile.Close()

	repos := make([]GitRepo, 0, 256)
	jsonParser := json.NewDecoder(jsonFile)
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
	if slices.Contains(aliases, repo.UpstreamAlias) {
		// check if url matches url in DB. git command: git remote get-url {upstream}
		cmd = exec.Command("git", "remote", "get-url", repo.UpstreamAlias) // #nosec G204
		cmd.Dir = expandPath(repo.Folder)
		urlOutput, err := cmd.CombinedOutput() //nolint:govet
		if err != nil {
			mutFail.Lock()
			*reportFail = append(*reportFail, fmt.Sprintf("%d: %s %v %s\n", i, repo.Folder, cmd.Args, err.Error()))
			mutFail.Unlock()
			return
		}
		upstreamURL := strings.Trim(string(urlOutput), newLine)
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
	// It may be the configured repo.MainBranch, or custom "mine", or empty "" (detatched head)
	branchName, err := getCurrBranch(&repo)
	if err != nil {
		mutFail.Lock()
		*reportFail = append(*reportFail, fmt.Sprintf("%d: %s %s\n", i, repo.Folder, "problem getting current brnach name: "+err.Error()))
		mutFail.Unlock()
		return
	}

	// when comparing our current to upstream, we don't care about any custom changes in "mine" as those are expected difference from upstream.
	// instead compare repo.Mainbranch if possible
	// or use HEAD if we are in a detached head state.
	if branchName == "" {
		// detached head.
		branchName = "HEAD"
	} else if branchName != repo.MainBranch {
		// on my custom branch. Dont' compare that.
		branchName = repo.MainBranch
	}

	// prepare diff command. example: git diff master upstream/master
	cmd := exec.Command("git", "diff",
		branchName,
		repo.UpstreamAlias+"/"+repo.MainBranch) // #nosec G204
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
	fmt.Printf("\nCHECKED OUT UseBranches: %d\n", len(reportBranchChange))
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
	// It may be the configured repo.MainBranch, or custom "mine", or empty "" (detatched head)
	branchName, err := getCurrBranch(&repo)
	if err != nil {
		mutFail.Lock()
		*reportFail = append(*reportFail, fmt.Sprintf("%d: %s %s\n", i, repo.Folder, "problem getting current brnach name: "+err.Error()))
		mutFail.Unlock()
		return
	}
	if branchName == repo.UseBranch {
		return // already using the desired branch. return early.
	}

	// prepare branch switch command. example: git checkout master
	cmd := exec.Command("git", "checkout", repo.UseBranch) // #nosec G204
	cmd.Dir = expandPath(repo.Folder)
	// Run branch switch!
	_, err = cmd.CombinedOutput()
	if err != nil {
		mutFail.Lock()
		*reportFail = append(*reportFail, fmt.Sprintf("%d: %s %v %s\n", i, repo.Folder, cmd.Args, err.Error()))
		mutFail.Unlock()
		return
	}

	mutBranchChange.Lock()
	*reportBranchChange = append(*reportBranchChange, fmt.Sprintf("%d: %s %v\n",
		i, repo.Folder, cmd.Args))
	mutBranchChange.Unlock()
}

// get current checked out branch name for a GitRepo.
func getCurrBranch(repo *GitRepo) (string, error) {
	// get current checked out branch name.
	// It may be the configured repo.MainBranch, or custom "mine", or empty "" (detatched head)
	cmdBranch := exec.Command("git", "branch", "--show-current")
	cmdBranch.Dir = expandPath(repo.Folder)
	branchOut, err := cmdBranch.CombinedOutput()
	if err != nil {
		return "", err
	}
	branchName := strings.Trim(string(branchOut), newLine)
	return branchName, nil
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
