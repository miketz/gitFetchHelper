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
	// return Remote{}, fmt.Errorf("no " + sym + " remote configured for " + r.Name + " in repos.jsonc")
	return Remote{}, fmt.Errorf("no %s remote configured for %s in repos.jsonc", sym, r.Name)
}

// get the hash of a branch in this GitRepo.
func (r *GitRepo) GetHash(branchName string) (string, error) {
	return GetHash(r.Folder, branchName)
}

// get the hash of a branch, tag, or "HEAD" in git repo folder.
func GetHash(repoFolder, branchTagOrHead string) (string, error) {
	cmd := exec.Command("git", "rev-parse", branchTagOrHead)
	cmd.Dir = expandPath(repoFolder)
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
	var err error
	DB, err = getRepoData()
	if err != nil {
		return err
	}

	homeDir, err = getHomeDir()
	if err != nil {
		return err
	}
	return nil
}

// get all the submodules
// git config --file .gitmodules --get-regexp path | awk '{ print $2 }'
// cmd := exec.Command("git", "config", "--file", ".gitmodules", "--get-regexp", "path", "|", "awk", "'{ print $2 }'")

func printCommands() {
	fmt.Printf(`Enter a command:
	fetchUpstream
	fetchDefault
	fetchMine
	mergeMine
	diffUpstream
	diffDefault
	diffMine
	init  (setUpstreamRemotesIfMissing)
	init2 (switchToBranches)
	init3 (cloneYoloRepos full-not-shallow)
	init3Shallow (cloneYoloRepos shallow)
	init4 (createLocalBranches)
`)
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
	case "fetchUpstream": // original
		fetchRemotes(RemoteUpstream)
	case "fetchDefault":
		fetchRemotes(RemoteDefault)
	case "fetchMine":
		fetchRemotes(RemoteMine)
	case "mergeMine":
		mergeMineRemotes()
	case "diffUpstream": // original diff
		listReposWithRemoteCodeToMerge(RemoteUpstream)
	case "diffDefault":
		listReposWithRemoteCodeToMerge(RemoteDefault)
	case "diffMine":
		listReposWithRemoteCodeToMerge(RemoteDefault)
	case "init":
		setUpstreamRemotesIfMissing()
	case "init2":
		switchToBranches()
	case "init3":
		cloneYoloRepos(false)
	case "init3Shallow":
		cloneYoloRepos(true)
	case "init4":
		createLocalBranches()
	default:
		printCommands()
	}
}

// read repos.jsonc into memory.
func getRepoData() ([]GitRepo, error) {
	jsonFile, err := os.Open("./repos.jsonc")
	if err != nil {
		fmt.Printf("opening json file: %v\n", err.Error())
		return nil, err
	}
	defer jsonFile.Close()

	repos := make([]GitRepo, 0, 256) // TODO: bump this if my repo count grows over 256
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

type RemoteType int

const (
	RemoteUpstream RemoteType = iota + 1
	RemoteMine
	RemoteDefault
)

// Fetch from remote for each repo, measure time, print reports. The main flow.
func fetchRemotes(remoteType RemoteType) { //nolint:dupl
	start := time.Now() // stop watch start

	reportFetched := make([]string, 0, len(DB)) // alloc 100%. no realloc on happy path.
	reportFail := make([]string, 0, 4)          // alloc for low failure rate

	wg := sync.WaitGroup{}
	mutFetched := sync.Mutex{}
	mutFail := sync.Mutex{}
	for i := 0; i < len(DB); i++ { // fetch upstream for each remote.
		wg.Add(1)
		go fetch(i, remoteType, &reportFetched, &reportFail, &wg, &mutFetched, &mutFail)
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

// Fetch remote for repo. Repo is identified by index i in DB.
func fetch(i int, remoteType RemoteType, reportFetched *[]string, reportFail *[]string,
	wg *sync.WaitGroup, mutFetched *sync.Mutex, mutFail *sync.Mutex,
) {
	defer wg.Done()

	repo := DB[i]

	// get remote info
	var remote Remote
	var err error
	if remoteType == RemoteUpstream {
		remote, err = repo.RemoteUpstream()
	} else if remoteType == RemoteDefault {
		remote, err = repo.RemoteDefault()
	} else if remoteType == RemoteMine {
		remote, err = repo.RemoteMine()
	} else {
		mutFail.Lock()
		*reportFail = append(*reportFail, fmt.Sprintf("%d: %s unknown repo type: %v\n", i, repo.Folder, remoteType))
		mutFail.Unlock()
		return
	}
	if err != nil {
		mutFail.Lock()
		*reportFail = append(*reportFail, fmt.Sprintf("%d: %s %s\n", i, repo.Folder, err.Error()))
		mutFail.Unlock()
		return
	}

	// prepare fetch command. example: git fetch upstream
	cmd := exec.Command("git", "fetch", remote.Alias) // #nosec G204
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

// merge in the code form "mine" remotes for BranchUse. the "mine" remotes are my forks
// or personal projects so it's OK for them to be merged without review.
func mergeMineRemotes() {
	start := time.Now() // stop watch start

	reportMerged := make([]string, 0, len(DB)) // alloc 100%. no realloc on happy path.
	reportFail := make([]string, 0, 4)         // alloc for low failure rate

	wg := sync.WaitGroup{}
	mutMerged := sync.Mutex{}
	mutFail := sync.Mutex{}
	for i := 0; i < len(DB); i++ { // fetch upstream for each remote.
		repo := DB[i]
		remoteMine, err := repo.RemoteMine()
		// this err just means no "mine" remote was configured in the
		// jsonc. so don't add to reportFail, just skip. TODO: make it return a bool, not err
		hasRemoteMine := err == nil
		if !hasRemoteMine {
			continue
		}
		wg.Add(1)
		go merge(i, &remoteMine, &reportMerged, &reportFail, &wg, &mutMerged, &mutFail)
	}
	wg.Wait()

	// summary report. print # of remotes merged, duration
	duration := time.Since(start) // stop watch end
	fmt.Printf("\nMerged %d of %d remotes. time elapsed: %v\n",
		len(reportMerged), len(DB), duration)

	// merge report. only includes repos that had new data to merge.
	fmt.Printf("\nRepos merged: %d\n", len(reportMerged))
	for i := 0; i < len(reportMerged); i++ {
		fmt.Print(reportMerged[i])
	}
	// failure report
	fmt.Printf("\nFAILURES: %d\n", len(reportFail))
	for i := 0; i < len(reportFail); i++ {
		fmt.Print(reportFail[i])
	}
}

func merge(i int, remoteMine *Remote, reportMerged *[]string, reportFail *[]string,
	wg *sync.WaitGroup, mutMerged *sync.Mutex, mutFail *sync.Mutex,
) {
	defer wg.Done()

	repo := DB[i]

	// in theory remote was already vetted to be a "mine" remote. but make sure
	if remoteMine.Sym != "mine" {
		return
	}
	currBranch, err := getCurrBranch(&repo)
	if err != nil {
		mutFail.Lock()
		*reportFail = append(*reportFail, fmt.Sprintf("%d: %s %s\n", i, repo.Folder, "problem getting current branch name: "+err.Error()))
		mutFail.Unlock()
		return
	}
	// verify BranchUse is checked out. don't switch to BranchUse as there may be
	// unstaged changes. just fail.
	if currBranch != repo.BranchUse {
		mutFail.Lock()
		*reportFail = append(*reportFail, fmt.Sprintf("%d: %s %s must be checked out before a merging from my remote.\n", i, repo.Folder, repo.BranchUse))
		mutFail.Unlock()
		return
	}

	// git merge origin/master
	cmd := exec.Command("git", "merge", remoteMine.Alias+"/"+repo.BranchUse) // #nosec G204
	cmd.Dir = expandPath(repo.Folder)
	// Run branch switch!
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		mutFail.Lock()
		*reportFail = append(*reportFail, fmt.Sprintf("%d: %s %v %s\n", i, repo.Folder, cmd.Args, err.Error()))
		mutFail.Unlock()
		return
	}
	// Merge and checking output is faster than checking hashes of master, origin/master in benchmarks.
	// At least with slow shelling out commands.
	output := string(stdout)
	if output == "Already up to date.\n" { // NOTE: this logic will break if msg changes in future
		return // nothing to merge, don't add to success report
	}
	lines := strings.Split(output, newLine)
	line2 := lines[1]
	// TODO: find a better way of detecting error or conflict. They could change the
	// message break this code.
	mergeFailed := strings.HasPrefix(line2, "error") || strings.HasPrefix(line2, "CONFLICT")
	if mergeFailed {
		mutFail.Lock()
		*reportFail = append(*reportFail, fmt.Sprintf("%d: %s %s\n", i, repo.Folder, output))
		mutFail.Unlock()
		return
	}
	// successful merge
	mutMerged.Lock()
	*reportMerged = append(*reportMerged, fmt.Sprintf("%d: %s %v %s\n",
		i, repo.Folder, cmd.Args, output))
	mutMerged.Unlock()
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

func listReposWithRemoteCodeToMerge(remoteType RemoteType) { //nolint:dupl
	start := time.Now() // stop watch start

	reportDiff := make([]string, 0, len(DB)) // alloc 100%. no realloc on happy path.
	reportFail := make([]string, 0, 4)       // alloc for low failure rate

	wg := sync.WaitGroup{}
	mutDiff := sync.Mutex{}
	mutFail := sync.Mutex{}
	for i := 0; i < len(DB); i++ { // check each repo for new upstream code
		wg.Add(1)
		go diff(i, remoteType, &reportDiff, &reportFail, &wg, &mutDiff, &mutFail)
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

func diff(i int, remoteType RemoteType, reportDiff *[]string, reportFail *[]string,
	wg *sync.WaitGroup, mutDiff *sync.Mutex, mutFail *sync.Mutex,
) {
	defer wg.Done()

	repo := DB[i]

	// get current checked out branch name.
	// It may be the configured repo.MainBranch, or custom "mine", or empty "" (detached head)
	// branchName, err := getCurrBranch(&repo)
	// if err != nil {
	// 	mutFail.Lock()
	// 	*reportFail = append(*reportFail, fmt.Sprintf("%d: %s %s\n", i, repo.Folder, "problem getting current branch name: "+err.Error()))
	// 	mutFail.Unlock()
	// 	return
	// }

	// get remote info
	var err error
	var remote Remote
	if remoteType == RemoteUpstream {
		remote, err = repo.RemoteUpstream()
	} else if remoteType == RemoteDefault {
		remote, err = repo.RemoteDefault()
	} else if remoteType == RemoteMine {
		remote, err = repo.RemoteMine()
	} else {
		mutFail.Lock()
		*reportFail = append(*reportFail, fmt.Sprintf("%d: %s unknown repo type: %v\n", i, repo.Folder, remoteType))
		mutFail.Unlock()
		return
	}
	if err != nil {
		mutFail.Lock()
		*reportFail = append(*reportFail, fmt.Sprintf("%d: %s %s\n", i, repo.Folder, err.Error()))
		mutFail.Unlock()
		return
	}

	// when comparing our current to upstream, we don't care about any custom changes in "mine" as those are expected difference from upstream.
	// instead compare repo.BranchMain if possible
	// or use HEAD if we are in a detached head state.
	// if we are comparieng against my remote fork or "default" remote then go ahead and use a non BranchMain in the
	// comparison
	var branchName string
	// if branchName == "" {
	// 	// detached head.
	// 	branchName = "HEAD"
	// }
	if remoteType == RemoteUpstream {
		// BranchUse is possibly a custom branch. Don't compare that when dealing with upstream as my custom branch won't exist there.
		branchName = repo.BranchMain
	} else if remoteType == RemoteMine {
		// BranchUse should always exist in my forked remote.
		// in this case we are interested in syncing up with the latest .emacs.d/ and that means BranchUse
		branchName = repo.BranchUse
	} else if remoteType == RemoteDefault {
		// in this case we are interested in syncing up with the latest .emacs.d/
		// but some of the remotes may use the upstream directly (no personal fork), for those continue to compare against the offical main branch
		isUpstreamRem := remote.Sym == "upstream"
		isMineRem := !isUpstreamRem && remote.Sym == "mine"
		if isUpstreamRem {
			branchName = repo.BranchMain
		} else if isMineRem {
			branchName = repo.BranchUse
		} else {
			// ad-hoc remote. not mine, not the official upstream.
			// maybe an old abondoned upstream remote configured in the json for informational purposes.
			// this case should rarely occur
			branchName = repo.BranchMain
		}
	}

	// prepare diff command. example: git diff master upstream/master
	// TODO: maybe compare git diff origin/master upstream/master
	//       to handle case where i'm on a "mine" branch and "master" only exists as a remote-tracking branch after a clone
	cmd := exec.Command("git", "diff",
		branchName,
		// remote.Alias+"/"+repo.BranchMain) // #nosec G204
		remote.Alias+"/"+branchName) // #nosec G204
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

// create local branches (ie featureX) for each remote tracking branch (ie origin/featureX).
// For all remote tracking of the default remote.
// this is needed for things like listReposWithUpstreamCodeToMerge() to work as it diffs
// the "local" branch (at least currently), and a differnet branch may be checked out (featureQ).
func createLocalBranches() {
	start := time.Now() // stop watch start

	reportBranch := make([]string, 0, len(DB)) // alloc 100%. no realloc on happy path.
	reportFail := make([]string, 0, 4)         // alloc for low failure rate

	wg := sync.WaitGroup{}
	mutBranch := sync.Mutex{}
	mutFail := sync.Mutex{}
	for i := 0; i < len(DB); i++ { // clone each "yolo" repo if missing
		wg.Add(1)
		go createLocalBranchesForRepo(i, &reportBranch, &reportFail, &wg, &mutBranch, &mutFail)
	}
	wg.Wait()

	// summary report. print # of branches checked out, duration
	duration := time.Since(start) // stop watch end
	fmt.Printf("\nChecked for existence of local branches in %d repos, create if not exist. time elapsed: %v\n",
		len(DB), duration)

	// clone report. only includes repos that needed to be cloned
	fmt.Printf("\nRepos with local branches created: %d\n", len(reportBranch))
	for i := 0; i < len(reportBranch); i++ {
		fmt.Print(reportBranch[i])
	}
	// failure report
	fmt.Printf("\nFAILURES: %d\n", len(reportFail))
	for i := 0; i < len(reportFail); i++ {
		fmt.Print(reportFail[i])
	}
}

// create "local" branches if they do not exist yet.
func createLocalBranchesForRepo(index int, reportBranch *[]string, reportFail *[]string,
	wg *sync.WaitGroup, mutBranch *sync.Mutex, mutFail *sync.Mutex,
) {
	defer wg.Done()

	repo := DB[index]

	// get current checked out branch name.
	// It may be the configured repo.MainBranch, repo.BranchUse (ie "mine"), or empty "" (detached head)
	// we will need to checkout this branch at the end as the act of creating branches will switch to them
	startingBranch, err := getCurrBranch(&repo)
	if err != nil {
		mutFail.Lock()
		*reportFail = append(*reportFail, fmt.Sprintf("%d: %s %s\n", index, repo.Folder, "problem getting current branch name: "+err.Error()))
		mutFail.Unlock()
		return
	}

	// default remote repo is using. usually my fork. sometimes direclty use the upstream.
	remoteDefault, err := repo.RemoteDefault()
	if err != nil {
		mutFail.Lock()
		*reportFail = append(*reportFail, fmt.Sprintf("%d: %s %s\n", index, repo.Folder, err.Error()))
		mutFail.Unlock()
		return
	}

	// // 1. get all remote branch names from the default remote
	// trackingBranches, err := TrackingBranches(repo.Folder, remoteDefault.Alias)
	// if err != nil {
	// 	mutFail.Lock()
	// 	*reportFail = append(*reportFail, fmt.Sprintf("%d: %s %s\n", i, repo.Folder, err.Error()))
	// 	mutFail.Unlock()
	// 	return
	// }
	// if len(trackingBranches) == 0 {
	// 	return // no branches to checkout!
	// }

	checkoutCnt := 0
	var collectOutput strings.Builder
	collectOutput.Grow(100 * 2) // allocate enough space up front

	// 2. for each remote tracking branch: create local branch if it does not exist
	// actually don't bother creating all remote tracking remoteBranches
	// only 2: repo.BranchMain, repo.BranchUse. if i need an odd branch I can manullay create as needed.
	remoteBranches := make([]string, 0, 2)
	remoteBranches = append(remoteBranches, remoteDefault.Alias+"/"+repo.BranchMain)
	// in theory BranchUse should always exist locally. And if BranchMain is the same branch then ditto.
	// but we will proceed with the checks and creation attempts anyway to fill in any gaps where
	// a branch doesn't exist for some reason.
	if repo.BranchMain != repo.BranchUse {
		remoteBranches = append(remoteBranches, remoteDefault.Alias+"/"+repo.BranchUse)
	}
	// check for, then create the branches
	for i := 0; i < len(remoteBranches); i++ {
		// remoteBranchName should be something like "origin/master"
		remoteBranchName := remoteBranches[i]
		// should be something like "master"
		branchName := removeRemoteFromBranchName(remoteBranchName)
		hasBranch, _ := hasLocalBranch(&repo, branchName)
		if hasBranch {
			continue // local branch already exists. no need to create it.
		}
		// create branch!
		// git checkout --track origin/featureX
		cmd := exec.Command("git", "checkout", "--track", remoteBranchName) // #nosec G204
		cmd.Dir = expandPath(repo.Folder)
		stdout, errOut := cmd.CombinedOutput()
		if errOut != nil {
			mutFail.Lock()
			*reportFail = append(*reportFail, fmt.Sprintf("%d: %s %v %s\n", i, repo.Folder, cmd.Args, errOut.Error()))
			mutFail.Unlock()
			return
		}
		collectOutput.WriteString(fmt.Sprintf("%d: %s %v %s\n",
			i, repo.Folder, cmd.Args, string(stdout)))
		checkoutCnt++
	}
	if checkoutCnt == 0 {
		return // don't write to the "success" report if we didn't do anything
	}
	// successfully checked out 1 or more branches
	mutBranch.Lock()
	*reportBranch = append(*reportBranch, collectOutput.String())
	mutBranch.Unlock()

	// 3. finally switch back to the starting branch. When creating "local" branches we
	// also checked them out!
	wasDetachedHead := startingBranch == ""
	if wasDetachedHead {
		// if we were in a detached head state, just stay where we are.
		// TODO: remember commit and switch back to commit of detatched head state
		return
	}
	// git checkout mine
	cmd := exec.Command("git", "checkout", startingBranch) // #nosec G204
	cmd.Dir = expandPath(repo.Folder)
	_, err = cmd.CombinedOutput()
	// possible for this function to be a success with local branch creation, but
	// fail when going back to starting branch
	if err != nil {
		mutFail.Lock()
		*reportFail = append(*reportFail, fmt.Sprintf("%d: %s %v %s\n", index, repo.Folder, cmd.Args, err.Error()))
		mutFail.Unlock()
		return
	}
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
// NOTE: git submodules dont' need to be cloned, they come with the .emacs.d/ repo.
func cloneYoloRepos(useShallowClone bool) {
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
		go cloneYolo(i, &reportClone, &reportFail, &wg, &mutClone, &mutFail, useShallowClone)
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
	wg *sync.WaitGroup, mutClone *sync.Mutex, mutFail *sync.Mutex, useShallowClone bool,
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

	var cmd *exec.Cmd
	if useShallowClone {
		// git clone --depth 1 --branch master --no-single-branch remoteUrl
		// using a shallow clone for performance. But still get the tip of each branch
		// with "--no-single-branch" to avoid a headache later when trying to switch to
		// other branches. git makes you go through convoluted steps if you don't get
		// the branches during the clone.
		// for full history manually run: git fetch --unshallow
		cmd = exec.Command("git", "clone", "--depth", "1", "--branch", repo.BranchUse, "--no-single-branch", remote.URL) // #nosec G204
	} else {
		// for now do not do shallow clone. although it's better for performance it messes up
		// subsequent merge/rebases (requireing fetch --unshallow).
		// The clone step in theory only executes 1 time ever on first setup of a new computer,
		// so it's OK if it's slower.
		cmd = exec.Command("git", "clone", "--branch", repo.BranchUse, remote.URL) // #nosec G204
	}

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

// get list of remote tracking branches for a remote.
func TrackingBranches(repoFolder, remoteAlias string) ([]string, error) {
	cmd := exec.Command("git", "branch", "-r") // #nosec G204
	cmd.Dir = expandPath(repoFolder)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	if len(output) == 0 { // no branches at all!
		return make([]string, 0), nil
	}
	// output might be something like:
	//     origin/master
	//     origin/mine
	//     upstream/master
	// split the raw shell output to a list of strings
	allTrackingBranches := strings.Split(string(output), newLine)
	// trim white space and * character from branch names
	for i, br := range allTrackingBranches {
		allTrackingBranches[i] = strings.Trim(br, "\n *")
	}
	// only include branches for THIS remote.
	remoteTrackingBranches := make([]string, 0, len(allTrackingBranches))
	remoteAliasSlash := remoteAlias + "/"
	remoteHEAD := remoteAliasSlash + "HEAD"
	for i := 0; i < len(allTrackingBranches); i++ {
		branchName := allTrackingBranches[i]
		isForThisRemote := strings.HasPrefix(branchName, remoteAliasSlash)
		if isForThisRemote {
			if strings.HasPrefix(branchName, remoteHEAD) {
				continue // not interested in the HEAD entry
			}
			remoteTrackingBranches = append(remoteTrackingBranches, branchName)
		}
	}
	return remoteTrackingBranches, nil
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

// returns true if file or directory exists.
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

// remove the "remote" prefix from a remote tracking branch name
// input:  "origin/km/reshelve-rewrite"
// output: "km/reshelve-rewrite"
func removeRemoteFromBranchName(remoteBranch string) string {
	i := strings.Index(remoteBranch, "/")
	return remoteBranch[i+1:]
}

// true if this program is running on MS Windows.
func isMsWindows() bool {
	return strings.HasPrefix(runtime.GOOS, "windows")
}

// Get user's home directory.
// If on MS Windows do a custom adjustment to my emacs config location.
func getHomeDir() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	myHomeDir := usr.HomeDir
	if isMsWindows() {
		// NOTE: this is a custom adjustment for my personal emacs config location
		// on MS Windows.
		myHomeDir += "/AppData/Local"
	}
	return myHomeDir, nil
}
