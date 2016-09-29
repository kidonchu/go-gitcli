package command

import (
	"fmt"
	"log"
	"os"

	"github.com/codegangsta/cli"
	"github.com/kidonchu/gitcli/gitutil"
)

// CmdNewStory creates new branchName for story
func CmdNewStory(c *cli.Context) {

	branchName := c.String("branch")
	if branchName == "" {
		log.Fatal("Branch to create is not specified")
	}

	from := c.String("source")
	source, err := gitutil.LookupBranchSource(from)
	if err != nil {
		log.Fatal(err)
	}

	// Get repo instance
	root, _ := os.Getwd()
	repo, err := gitutil.GetRepo(root)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("* %s => %s\n", source, branchName)

	answer := GetUserInput("Proceed with above items? (nY): ")
	if answer != "Y" {
	}

	// Stash all changes for current branch, if any
	fmt.Println("Stashing changes on current branch")
	err = gitutil.Stash(repo)
	if err != nil {
		log.Fatal(err)
	}

	// Fetch from main repo before creating new branch
	fmt.Println("Fetching most recent branches")
	if err = gitutil.Fetch(repo, "ActiveCampaign"); err != nil {
		// do not fail entire app even if fetch fails
		log.Fatal(err)
	}

	fmt.Println("Creating new branch")
	newBranch, err := gitutil.CreateBranch(repo, branchName, source)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Finding remote `origin`")
	remote, err := gitutil.GetRemote(repo, "origin")
	if err != nil {
		log.Fatalf("Unable to find remote: %+v\n", err)
	}

	fmt.Println("Pushing to remote")
	ref := "refs/heads/" + branchName
	err = gitutil.Push(repo, remote, ref)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Setting upstream to remote branch")
	err = gitutil.SetUpstream(newBranch, "origin")
	if err != nil {
		log.Fatal(err)
	}
}
