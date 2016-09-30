package gitutil

import (
	"fmt"
	"log"
	"regexp"
	"sort"
	"time"

	git "github.com/libgit2/git2go"
)

// acting as remote cache
var (
	branches map[string]*git.Branch
	remotes  map[string]*git.Remote
)

// CredentialsCallback creates ssh key for github creds
func CredentialsCallback(url string, username string, allowedTypes git.CredType) (git.ErrorCode, *git.Cred) {
	ret, cred := git.NewCredSshKey(
		"git", "/Users/kchu/.ssh/id_rsa_github.pub",
		"/Users/kchu/.ssh/id_rsa_github", "")
	return git.ErrorCode(ret), &cred
}

// CertificateCheckCallback for placeholder
func CertificateCheckCallback(cert *git.Certificate, valid bool, hostname string) git.ErrorCode {
	return 0
}

// DeleteBranch deletes branch
func DeleteBranch(repo *git.Repository, remote *git.Remote, branch *git.Branch) error {

	name, _ := branch.Name()
	fmt.Printf("Deleting branch: `%s`...\n", name)

	err := branch.Delete()
	if err != nil {
		return fmt.Errorf("\tDeleteBranch: Unable to delete branch: `%s`\n%+v", name, err)
	}

	// if remote branch, need to push to update remote repo
	_, err = repo.LookupBranch(fmt.Sprintf("%s/%s", remote.Name(), name), git.BranchRemote)
	if err == nil {
		ref := fmt.Sprintf(":refs/heads/%s", name)
		fmt.Printf("Deleting remote branch: `%s`...\n", name)
		err = Push(repo, remote, ref)
		if err != nil {
			return fmt.Errorf("Unable to push refspec: `%s`\n%+v", ref, err)
		}
	}

	return nil
}

// DeleteBranches deletes branches passed in
func DeleteBranches(repo *git.Repository, remote *git.Remote, branches []*git.Branch) error {

	for _, branch := range branches {
		err := DeleteBranch(repo, remote, branch)
		if err != nil { // do not stop even if delete for one branch fails
			log.Println(err)
		}
	}

	return nil
}

// Fetch fetches all delta from remote repo
func Fetch(repo *git.Repository, remoteName string) error {

	fetchOptions := &git.FetchOptions{
		RemoteCallbacks: git.RemoteCallbacks{
			CredentialsCallback:      CredentialsCallback,
			CertificateCheckCallback: CertificateCheckCallback,
		},
	}

	remote, err := repo.Remotes.Lookup(remoteName)
	if err != nil {
		return fmt.Errorf("Unable to lookup remote: `%s`\n%+v\n", remoteName, err)
	}

	err = remote.Fetch([]string{}, fetchOptions, "")
	if err != nil {
		return fmt.Errorf("Unable to fetch for remote: `%s`\n%+v\n", remoteName, err)
	}

	return nil
}

// FindBranches returns a list of branches that matches pattern
func FindBranches(repo *git.Repository, pattern string, branchType git.BranchType) ([]*git.Branch, error) {

	var branches []*git.Branch

	it, err := repo.NewBranchIterator(branchType)
	if err != nil {
		return nil, fmt.Errorf("Unable to iterate through branches\n%+v\n", err)
	}

	branchRegex, _ := regexp.Compile(pattern)
	it.ForEach(func(b *git.Branch, t git.BranchType) error {
		name, _ := b.Name()
		matched := branchRegex.FindString(name)
		if matched != "" {
			branches = append(branches, b)
		}
		return nil
	})

	return branches, nil
}

// FindBranch looks up the repo and fetch branch with given name
func FindBranch(repo *git.Repository, name string, bt git.BranchType) (*git.Branch, error) {

	key := fmt.Sprintf("%s-%d", name, bt)

	// First, look at cached instances
	if branch := branches[key]; branch != nil {
		return branch, nil
	}

	// Otherwise, lookup the repo
	branch, err := repo.LookupBranch(name, bt)
	if err != nil {
		return nil, fmt.Errorf("Unable to find branch `%s`\n%+v\n", name, err)
	}

	if branches == nil {
		branches = make(map[string]*git.Branch)
	}

	// Store result into cache
	branches[key] = branch

	return branch, nil
}

// GetRemote searches the repo for remote with given name and returns remote instance
func GetRemote(repo *git.Repository, name string) (*git.Remote, error) {

	// First, look at cached instances
	if remote := remotes[name]; remote != nil {
		return remote, nil
	}

	// Otherwise, lookup the repo
	remote, err := repo.Remotes.Lookup(name)
	if err != nil {
		return nil, fmt.Errorf("Unable to find remote `%s`\n%+v\n", name, err)
	}

	if remotes == nil {
		remotes = make(map[string]*git.Remote)
	}

	// Store result into cache
	remotes[name] = remote

	return remote, nil
}

// GetRepo creates a Repository instance
func GetRepo(repoName string) (*git.Repository, error) {
	repo, err := git.OpenRepository(repoName)
	if err != nil {
		return nil, fmt.Errorf("Unable to open repository: `%s`\n%+v\n", repoName, err)
	}
	return repo, nil
}

// Push pushes given ref to remote repo
func Push(repo *git.Repository, remote *git.Remote, ref string) error {

	// execute push
	err := remote.Push([]string{ref}, &git.PushOptions{
		RemoteCallbacks: git.RemoteCallbacks{
			CredentialsCallback:      CredentialsCallback,
			CertificateCheckCallback: CertificateCheckCallback,
		},
	})
	if err != nil {
		return fmt.Errorf("Unable to push `%s` to remote `%s`\n%+v\n", ref, remote.Name(), err)
	}

	return nil
}

// Stash stashes changes
func Stash(repo *git.Repository) error {

	// check if there are any changes to be stashed
	opts := &git.StatusOptions{}
	opts.Flags = git.StatusOptIncludeUntracked
	statusList, err := repo.StatusList(opts)
	if err != nil {
		return fmt.Errorf("Could not obtain status list: %+v", err)
	}
	entryCount, err := statusList.EntryCount()
	if err != nil {
		return err
	}

	// if there isn't any changed files, no need for stashing
	if entryCount <= 0 {
		fmt.Println("\tStash: No changes to stash")
		return nil
	}

	fmt.Printf("\tStash: There are %d updated files. Start stashing...\n", entryCount)

	var name, email string
	name, email, err = gitUser()
	if err != nil {
		log.Fatal(err)
		// Use default signature if not found
		name = "Kidon Chu"
		email = "kidonchu@gmail.com"
	}

	fmt.Printf("\tStash: Generating signature for stashing with %s and %s\n", name, email)
	sig := &git.Signature{
		Name:  name,
		Email: email,
		When:  time.Now(),
	}

	// add every changes and new files into index
	index, err := repo.Index()
	if err != nil {
		return err
	}

	fmt.Println("\tStash: Adding all changes to index")
	err = index.AddAll([]string{}, git.IndexAddDefault, nil)
	if err != nil {
		return err
	}
	fmt.Println("\tStash: Writing tree")
	_, err = index.WriteTree()
	if err != nil {
		return err
	}
	fmt.Println("\tStash: Writing index")
	err = index.Write()
	if err != nil {
		return err
	}

	branchName, err := currentBranchName(repo)
	if err != nil {
		return err
	}

	fmt.Printf("\tStash: Creating stash commit for %s\n", branchName)
	oid, err := repo.Stashes.Save(
		sig,
		fmt.Sprintf("WIP on %s", branchName),
		git.StashDefault,
	)
	if err != nil {
		return err
	}

	// store last stashed commit to config
	stashConfigPath := fmt.Sprintf("branch.%s.laststash", branchName)
	fmt.Printf("\tStash: Storing last stash commit to '%s'\n", stashConfigPath)
	err = SetConfigString(stashConfigPath, oid.String())
	if err != nil {
		return err
	}

	return nil
}

// PopLastStash pops stash for current branch
func PopLastStash(repo *git.Repository) error {

	branchName, err := currentBranchName(repo)
	if err != nil {
		return err
	}

	stashConfigPath := fmt.Sprintf("branch.%s.laststash", branchName)

	stashCommit, err := ConfigString(stashConfigPath)
	if err != nil {
		// if no stash is found, nothing to pop
		fmt.Println("\tPop: Nothing to pop")
		return nil
	}

	// find last stash's stash index
	stashIndex := -1
	repo.Stashes.Foreach(func(index int, msg string, id *git.Oid) error {
		if id.String() == stashCommit {
			stashIndex = index
		}
		return nil
	})

	// if stash is found, pop it
	if stashIndex > -1 {
		fmt.Printf("\tPop: Last stash found with index: %d, Oid: %s. Popping...\n", stashIndex, stashCommit)
		opts, _ := git.DefaultStashApplyOptions()
		err = repo.Stashes.Pop(stashIndex, opts)
		if err != nil {
			return err
		}
	}

	// clear out last stash info
	if stashCommit != "" {
		SetConfigString(stashConfigPath, "")
	}

	return nil
}

func gitUser() (string, string, error) {
	name, err := ConfigString("user.name")
	if err != nil {
		return "", "", err
	}

	email, err := ConfigString("user.email")
	if err != nil {
		return "", "", err
	}

	return name, email, nil
}

func currentBranchName(repo *git.Repository) (string, error) {

	head, err := repo.Head()
	if err != nil {
		return "", err
	}

	branchName, err := head.Branch().Name()
	if err != nil {
		return "", err
	}

	return branchName, nil
}

// LookupBranchSource looks up branch source
func LookupBranchSource(from string) (string, error) {

	// default source: contact-deal
	if from == "" {
		from = "contact"
	}

	source, err := ConfigString("story.source." + from)
	if err != nil {
		return "", fmt.Errorf("Unable to find source for %s", from)
	}

	return source, nil
}

// CreateBranch creates new branch
func CreateBranch(repo *git.Repository, branchName string, source string) (*git.Branch, error) {

	var newBranch *git.Branch
	var err error

	newBranch, err = repo.LookupBranch(branchName, git.BranchLocal)
	if err != nil {
		// find source branch to create new branch from
		sourceBranch, err := repo.LookupBranch(source, git.BranchRemote)
		if err != nil {
			return nil, err
		}

		sourceCommit, err := repo.LookupCommit(sourceBranch.Target())
		if err != nil {
			return nil, err
		}

		newBranch, err = repo.CreateBranch(branchName, sourceCommit, false)
		if err != nil {
			return nil, err
		}
	}

	// Checkout new branch as HEAD
	err = Checkout(repo, branchName)
	if err != nil {
		log.Printf("Unable to checkout new branch as HEAD: %+v", err)
	}

	return newBranch, nil
}

// Checkout checks out given branch
func Checkout(repo *git.Repository, branchName string) error {

	// see if we have a branch named with given branchName
	_, err := repo.References.Lookup("refs/heads/" + branchName)
	if err != nil {
		return err
	}

	// mark HEAD as given branch
	fmt.Printf("\tCheckout: Creating symbolic reference to HEAD for %s\n", branchName)
	_, err = repo.References.CreateSymbolic("HEAD", "refs/heads/"+branchName, true, "")
	if err != nil {
		return err
	}
	opts := &git.CheckoutOpts{
		Strategy: git.CheckoutForce,
	}
	if err := repo.CheckoutHead(opts); err != nil {
		return err
	}

	return nil
}

// SetUpstream sets upstream for local branch
func SetUpstream(branch *git.Branch, remoteName string) error {

	branchName, _ := branch.Name()
	upstream := fmt.Sprintf("%s/%s", remoteName, branchName)
	err := branch.SetUpstream(upstream)
	if err != nil {
		return fmt.Errorf("\tSetUpstream: Unable to set upstream to `%s`: %+v", upstream, err)
	}

	return nil
}

// DeleteStashes deletes given stashes in repo
func DeleteStashes(repo *git.Repository, stashes map[int]*StashInfo) {

	// before deleting, sort by index descending order
	// since dropping stash will reduce index by 1 each time
	keys := make([]int, len(stashes))
	i := 0
	for k := range stashes {
		keys[i] = k
		i++
	}
	sort.Sort(sort.Reverse(sort.IntSlice(keys)))

	for _, k := range keys {
		stashInfo := stashes[k]
		err := repo.Stashes.Drop(stashInfo.Index)
		if err != nil {
			fmt.Printf("%+v", err)
		}
	}
}

// FindStashes fines stashes in repo that matches given branch
func FindStashes(repo *git.Repository, pattern string) map[int]*StashInfo {

	regex, _ := regexp.Compile(pattern)

	// delete stashes that match the pattern
	stashes := make(map[int]*StashInfo)
	repo.Stashes.Foreach(func(index int, msg string, id *git.Oid) error {
		matched := regex.FindString(msg)
		if matched != "" {
			stashInfo := &StashInfo{Index: index, ID: id, Msg: msg}
			stashes[index] = stashInfo
		}
		return nil
	})

	return stashes
}

// StashInfo stores information about stash
type StashInfo struct {
	Index int
	ID    *git.Oid
	Msg   string
}
