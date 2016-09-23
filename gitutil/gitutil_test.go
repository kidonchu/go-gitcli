package gitutil

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/kidonchu/gitcli/testutil"
	git "github.com/libgit2/git2go"
)

func TestFindBranches(t *testing.T) {

	localRepo := createTestRepo(t)
	defer cleanupTestRepo(t, localRepo)

	head, _ := seedTestRepo(t, localRepo)
	commit, _ := localRepo.LookupCommit(head)

	// Create few more branches
	newBranches := []string{"12345", "45678", "12389", "14532", "11111"}
	for _, val := range newBranches {
		localRepo.CreateBranch("test-"+val, commit, true)
	}

	branches, err := FindBranches(localRepo, "^.*45.*$", git.BranchLocal)
	testutil.CheckFatal(t, err)

	if len(branches) != 3 {
		testutil.CheckFatal(t, fmt.Errorf("Expected 3 branches, but got %d branches", len(branches)))
	}

	iter, _ := localRepo.NewBranchIterator(git.BranchAll)
	iter.ForEach(func(b *git.Branch, bt git.BranchType) error {
		b.Delete()
		return nil
	})
}

func TestPush(t *testing.T) {

	repo := createBareTestRepo(t)
	defer cleanupTestRepo(t, repo)

	localRepo := createTestRepo(t)
	defer cleanupTestRepo(t, localRepo)

	remote, err := localRepo.Remotes.Create("test_push", repo.Path())
	testutil.CheckFatal(t, err)

	seedTestRepo(t, localRepo)

	err = Push(localRepo, remote, "refs/heads/master")
	testutil.CheckFatal(t, err)

	_, err = localRepo.References.Lookup("refs/remotes/test_push/master")
	testutil.CheckFatal(t, err)

	_, err = repo.References.Lookup("refs/heads/master")
	testutil.CheckFatal(t, err)
}

func TestDeleteBranch(t *testing.T) {

	// Prepare repos for testing
	repo := createBareTestRepo(t)
	defer cleanupTestRepo(t, repo)
	localRepo := createTestRepo(t)
	defer cleanupTestRepo(t, localRepo)
	remote, _ := localRepo.Remotes.Create("test_push", repo.Path())
	head, _ := seedTestRepo(t, localRepo)
	commit, _ := localRepo.LookupCommit(head)
	Push(localRepo, remote, "refs/heads/master")

	// Switch HEAD to different branch so that we can delete master branch
	_, err := localRepo.CreateBranch("develop", commit, true)
	testutil.CheckFatal(t, err)
	err = localRepo.SetHead("refs/heads/develop")
	testutil.CheckFatal(t, err)

	localBranch, err := localRepo.LookupBranch("master", git.BranchLocal)
	testutil.CheckFatal(t, err)

	_, err = localRepo.LookupBranch("test_push/master", git.BranchRemote)
	testutil.CheckFatal(t, err)

	err = DeleteBranch(localRepo, remote, localBranch)
	testutil.CheckFatal(t, err)

	_, err = localRepo.References.Lookup("refs/heads/master")
	if err == nil {
		err = errors.New("Lookup should have thrown the error since local branch is deleted")
		testutil.CheckFatal(t, err)
	}

	_, err = repo.References.Lookup("refs/heads/master")
	if err == nil {
		err = errors.New("Lookup should have thrown the error since remote branch is deleted")
		testutil.CheckFatal(t, err)
	}
}

func cleanupTestRepo(t *testing.T, r *git.Repository) {
	var err error
	if r.IsBare() {
		err = os.RemoveAll(r.Path())
	} else {
		err = os.RemoveAll(r.Workdir())
	}
	testutil.CheckFatal(t, err)

	r.Free()
}

func createTestRepo(t *testing.T) *git.Repository {
	// figure out where we can create the test repo
	path, err := ioutil.TempDir("", "gitcli")
	testutil.CheckFatal(t, err)
	repo, err := git.InitRepository(path, false)
	testutil.CheckFatal(t, err)

	tmpfile := "README"
	err = ioutil.WriteFile(path+"/"+tmpfile, []byte("foo\n"), 0644)
	testutil.CheckFatal(t, err)

	return repo
}

func createBareTestRepo(t *testing.T) *git.Repository {
	// figure out where we can create the test repo
	path, err := ioutil.TempDir("", "gitcli")
	testutil.CheckFatal(t, err)
	repo, err := git.InitRepository(path, true)
	testutil.CheckFatal(t, err)

	return repo
}

func seedTestRepo(t *testing.T, repo *git.Repository) (*git.Oid, *git.Oid) {
	loc, err := time.LoadLocation("America/Chicago")
	testutil.CheckFatal(t, err)
	sig := &git.Signature{
		Name:  "Kidon Chu",
		Email: "kidonchu@gmail.com",
		When:  time.Date(2016, 9, 20, 14, 30, 0, 0, loc),
	}

	idx, err := repo.Index()
	testutil.CheckFatal(t, err)
	err = idx.AddByPath("README")
	testutil.CheckFatal(t, err)
	treeId, err := idx.WriteTree()
	testutil.CheckFatal(t, err)

	message := "This is a commit\n"
	tree, err := repo.LookupTree(treeId)
	testutil.CheckFatal(t, err)
	commitId, err := repo.CreateCommit("HEAD", sig, sig, message, tree)
	testutil.CheckFatal(t, err)

	return commitId, treeId
}

func printBranches(t *testing.T, repo *git.Repository) {
	iter, _ := repo.NewBranchIterator(git.BranchAll)
	iter.ForEach(func(b *git.Branch, bt git.BranchType) error {
		name, _ := b.Name()
		t.Log(name)
		return nil
	})
}
