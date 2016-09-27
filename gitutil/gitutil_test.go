package gitutil

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"runtime"
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

func TestStash(t *testing.T) {
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	prepareStashRepo(t, repo)

	sig := &git.Signature{
		Name:  "Rand Om Hacker",
		Email: "random@hacker.com",
		When:  time.Now(),
	}

	stash1, err := repo.Stashes.Save(sig, "First stash", git.StashDefault)
	testutil.CheckFatal(t, err)

	_, err = repo.LookupCommit(stash1)
	testutil.CheckFatal(t, err)

	b, err := ioutil.ReadFile(pathInRepo(repo, "README"))
	testutil.CheckFatal(t, err)
	if string(b) == "Update README goes to stash\n" {
		t.Errorf("README still contains the uncommitted changes")
	}

	if !fileExistsInRepo(repo, "untracked.txt") {
		t.Errorf("untracked.txt doesn't exist in the repo; should be untracked")
	}

	// Apply: default

	opts, err := git.DefaultStashApplyOptions()
	testutil.CheckFatal(t, err)

	err = repo.Stashes.Apply(0, opts)
	testutil.CheckFatal(t, err)

	b, err = ioutil.ReadFile(pathInRepo(repo, "README"))
	testutil.CheckFatal(t, err)
	if string(b) != "Update README goes to stash\n" {
		t.Errorf("README changes aren't here")
	}

	// Apply: no stash for the given index

	err = repo.Stashes.Apply(1, opts)
	if !git.IsErrorCode(err, git.ErrNotFound) {
		t.Errorf("expecting GIT_ENOTFOUND error code %d, got %v", git.ErrNotFound, err)
	}

	// Apply: callback stopped

	opts.ProgressCallback = func(progress git.StashApplyProgress) error {
		if progress == git.StashApplyProgressCheckoutModified {
			return fmt.Errorf("Stop")
		}
		return nil
	}

	err = repo.Stashes.Apply(0, opts)
	if err.Error() != "Stop" {
		t.Errorf("expecting error 'Stop', got %v", err)
	}

	// Create second stash with ignored files

	os.MkdirAll(pathInRepo(repo, "tmp"), os.ModeDir|os.ModePerm)
	err = ioutil.WriteFile(pathInRepo(repo, "tmp/ignored.txt"), []byte("Ignore me\n"), 0644)
	testutil.CheckFatal(t, err)

	stash2, err := repo.Stashes.Save(sig, "Second stash", git.StashIncludeIgnored)
	testutil.CheckFatal(t, err)

	if fileExistsInRepo(repo, "tmp/ignored.txt") {
		t.Errorf("tmp/ignored.txt should not exist anymore in the work dir")
	}

	// Stash foreach

	expected := []stash{
		{0, "On master: Second stash", stash2.String()},
		{1, "On master: First stash", stash1.String()},
	}
	checkStashes(t, repo, expected)

	// Stash pop

	opts, _ = git.DefaultStashApplyOptions()
	err = repo.Stashes.Pop(1, opts)
	testutil.CheckFatal(t, err)

	b, err = ioutil.ReadFile(pathInRepo(repo, "README"))
	testutil.CheckFatal(t, err)
	if string(b) != "Update README goes to stash\n" {
		t.Errorf("README changes aren't here")
	}

	expected = []stash{
		{0, "On master: Second stash", stash2.String()},
	}
	checkStashes(t, repo, expected)

	// Stash drop

	err = repo.Stashes.Drop(0)
	testutil.CheckFatal(t, err)

	expected = []stash{}
	checkStashes(t, repo, expected)
}

type stash struct {
	index int
	msg   string
	id    string
}

func checkStashes(t *testing.T, repo *git.Repository, expected []stash) {
	var actual []stash

	repo.Stashes.Foreach(func(index int, msg string, id *git.Oid) error {
		stash := stash{index, msg, id.String()}
		if len(expected) > len(actual) {
			if s := expected[len(actual)]; s.id == "" {
				stash.id = "" //  don't check id
			}
		}
		actual = append(actual, stash)
		return nil
	})

	if len(expected) > 0 && !reflect.DeepEqual(expected, actual) {
		// The failure happens at wherever we were called, not here
		_, file, line, ok := runtime.Caller(1)
		if !ok {
			t.Fatalf("Unable to get caller")
		}
		t.Errorf("%v:%v: expecting %#v\ngot %#v", path.Base(file), line, expected, actual)
	}
}

func prepareStashRepo(t *testing.T, repo *git.Repository) {
	seedTestRepo(t, repo)

	// Write .gitignore file
	err := ioutil.WriteFile(pathInRepo(repo, ".gitignore"), []byte("tmp\n"), 0644)
	testutil.CheckFatal(t, err)

	sig := &git.Signature{
		Name:  "Rand Om Hacker",
		Email: "random@hacker.com",
		When:  time.Now(),
	}

	idx, err := repo.Index()
	testutil.CheckFatal(t, err)
	err = idx.AddByPath(".gitignore")
	testutil.CheckFatal(t, err)
	treeID, err := idx.WriteTree()
	testutil.CheckFatal(t, err)
	err = idx.Write()
	testutil.CheckFatal(t, err)

	currentBranch, err := repo.Head()
	testutil.CheckFatal(t, err)
	currentTip, err := repo.LookupCommit(currentBranch.Target())
	testutil.CheckFatal(t, err)

	message := "Add .gitignore\n"
	tree, err := repo.LookupTree(treeID)
	testutil.CheckFatal(t, err)
	_, err = repo.CreateCommit("HEAD", sig, sig, message, tree, currentTip)
	testutil.CheckFatal(t, err)

	err = ioutil.WriteFile(pathInRepo(repo, "README"), []byte("Update README goes to stash\n"), 0644)
	testutil.CheckFatal(t, err)

	err = ioutil.WriteFile(pathInRepo(repo, "untracked.txt"), []byte("Hello, World\n"), 0644)
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
	treeID, err := idx.WriteTree()
	testutil.CheckFatal(t, err)

	message := "This is a commit\n"
	tree, err := repo.LookupTree(treeID)
	testutil.CheckFatal(t, err)
	commitID, err := repo.CreateCommit("HEAD", sig, sig, message, tree)
	testutil.CheckFatal(t, err)

	return commitID, treeID
}

func printBranches(t *testing.T, repo *git.Repository) {
	iter, _ := repo.NewBranchIterator(git.BranchAll)
	iter.ForEach(func(b *git.Branch, bt git.BranchType) error {
		name, _ := b.Name()
		t.Log(name)
		return nil
	})
}

func fileExistsInRepo(repo *git.Repository, name string) bool {
	if _, err := os.Stat(pathInRepo(repo, name)); err != nil {
		return false
	}
	return true
}

func pathInRepo(repo *git.Repository, name string) string {
	return path.Join(path.Dir(path.Dir(repo.Path())), name)
}
