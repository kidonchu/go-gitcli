package gitutil

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"

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
		return fmt.Errorf("Unable to delete branch: `%s`\n%+v\n", name, err)
	}

	// if remote branch, need to push to update remote repo
	_, err = repo.LookupBranch(fmt.Sprintf("%s/%s", remote.Name(), name), git.BranchRemote)
	if err == nil {
		ref := fmt.Sprintf(":refs/heads/%s", name)
		err = Push(repo, remote, ref)
		if err != nil {
			return fmt.Errorf("Unable to push refspec: `%s`\n%+v\n", ref, err)
		}
	}

	return nil
}

// DeleteBranches deletes branches passed in
func DeleteBranches(repo *git.Repository, remote *git.Remote, branches []*git.Branch) error {

	for _, branch := range branches {
		err := DeleteBranch(repo, remote, branch)
		if err != nil {
			return err
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

// FindRemote searches the repo for remote with given name and returns remote instance
func FindRemote(repo *git.Repository, name string) (*git.Remote, error) {

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

// GetGitconfig returns git config object
func GetGitconfig() (*git.Config, error) {
	configPath, err := git.ConfigFindGlobal()
	if err != nil {
		return nil, fmt.Errorf("Global .gitconfig could not be found\n%+v\n", err)
	}

	config, err := git.OpenOndisk(nil, configPath)
	if err != nil {
		return nil, fmt.Errorf("Unable to open `%s`\n%+v\n", configPath, err)
	}

	return config, nil
}

// func getRemote(repo *git.Repository, remoteName string) (*git.Remote, error) {

// 	// First, look at cached branch instances
// 	if remotes[remoteName] != nil {
// 		return remotes[remoteName], nil
// 	}

// 	remote, err := repo.Remotes.Lookup(remoteName)
// 	if err != nil {
// 		return nil, fmt.Errorf("Unable to find remote: %s\n%+v\n", remoteName, err)
// 	}

// 	return remote, nil
// }

// GetRepo creates a Repository instance
func GetRepo(repoName string) (*git.Repository, error) {
	repo, err := git.OpenRepository(repoName)
	if err != nil {
		return nil, fmt.Errorf("Unable to open repository: `%s`\n%+v\n", repoName, err)
	}
	return repo, nil
}

// SetUpstream sets upstream for local branch
func SetUpstream(repo *git.Repository, remoteName string, branchName string) error {

	branch, err := repo.LookupBranch(branchName, git.BranchLocal)
	if err != nil {
		return fmt.Errorf("Unable to lookup breanch `%s`\n%+v\n", branchName, err)
	}
	ref := fmt.Sprintf("%s/%s", remoteName, branchName)
	err = branch.SetUpstream(ref)
	if err != nil {
		return fmt.Errorf("Unable to set upstream for `%s`\n%+v\n", ref, err)
	}

	return nil
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

// StashChanges stashes changes
func StashChanges(repo *git.Repository) {

	// check if there are any changes to be stashed
	opts := &git.StatusOptions{}
	opts.Flags = git.StatusOptIncludeUntracked
	statusList, err := repo.StatusList(opts)
	if err != nil {
		log.Fatal(err)
	}
	entryCount, err := statusList.EntryCount()
	if err != nil {
		log.Fatal(err)
	}

	// if there is any change to the current branch, stash theme
	if entryCount > 0 {

		// store current working directory to come back later
		origWd, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}

		// go to repo dir
		repoPath := path.Dir(path.Dir(repo.Path()))
		err = os.Chdir(repoPath)
		if err != nil {
			log.Fatal(err)
		}

		// add every changes and new files into index
		index, err := repo.Index()
		if err != nil {
			log.Fatal(err)
		}
		err = index.AddAll([]string{}, git.IndexAddDefault, nil)
		if err != nil {
			log.Fatal(err)
		}
		_, err = index.WriteTree()
		err = index.Write()
		if err != nil {
			log.Fatal(err)
		}

		// run actual stash command
		cmd := exec.Command("git", "stash")
		out, err := cmd.Output()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("%s\n", out)

		// return to original working directory
		err = os.Chdir(origWd)
		if err != nil {
			log.Fatal(err)
		}
	}
}
