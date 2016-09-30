package command

import (
	"fmt"
	"log"
	"os"

	"github.com/codegangsta/cli"
	_ "github.com/go-sql-driver/mysql" // mysql driver
	"github.com/kidonchu/gitcli/dbutil"
	"github.com/kidonchu/gitcli/gitutil"
	"github.com/libgit2/git2go"
)

// CmdDeleteStory deletes story
// First, it deletes local and remote branchs whose name matches with the `pattern`
// Then, it deletes the databases whose name matches with the `pattern`
func CmdDeleteStory(c *cli.Context) {

	pattern := c.String("pattern")

	// Get repo instance
	root, _ := os.Getwd()
	repo, err := gitutil.GetRepo(root)
	if err != nil {
		log.Fatal(err)
		return
	}

	branches, err := gitutil.FindBranches(repo, "^.*"+pattern+".*$", git.BranchLocal)
	if err != nil {
		log.Fatal(err)
		return
	}

	stashes := gitutil.FindStashes(repo, "^.*"+pattern+".*$")

	// Now handle databases
	var (
		host, _ = gitutil.ConfigString("story.hosteddb.host")
		port, _ = gitutil.ConfigInt32("story.hosteddb.port")
		user, _ = gitutil.ConfigString("story.hosteddb.user")
		pass, _ = gitutil.ConfigString("story.hosteddb.pass")
	)

	dbh, err := dbutil.Connect(host, port, user, pass)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer dbh.Close()

	// find list of dbs to delete
	dbs, err := dbutil.FindDbs(dbh, "^.*"+pattern+".*$")
	if err != nil {
		log.Fatal(err)
		return
	}

	if len(branches) < 1 && len(dbs) < 1 && len(stashes) < 1 {
		fmt.Println("Nothing to delete")
		return
	}

	// Show list of branches and databases to be deleted and get confirmation
	if len(branches) > 0 {
		fmt.Println("Following branches will be deleted")
		for _, b := range branches {
			name, _ := b.Name()
			fmt.Printf("* %s\n", name)
		}
		fmt.Println("")
	}
	if len(stashes) > 0 {
		fmt.Println("Following stashes will be deleted")
		for _, stash := range stashes {
			fmt.Printf("* %s\n", stash.Msg)
		}
		fmt.Println("")
	}
	if len(dbs) > 0 {
		fmt.Println("Following databases will be deleted")
		for _, database := range dbs {
			fmt.Printf("* %s\n", database)
		}
		fmt.Println("")
	}

	// If confirmed, delete branches and databases
	answer := GetUserInput("Continue? (nY): ")
	if answer == "Y" {

		remote, err := gitutil.GetRemote(repo, "origin")
		if err != nil {
			log.Fatal(err)
			return
		}

		err = gitutil.DeleteBranches(repo, remote, branches)
		if err != nil {
			log.Fatal(err)
			return
		}

		gitutil.DeleteStashes(repo, stashes)

		err = dbutil.Drop(dbh, dbs)
		if err != nil {
			log.Fatal(err)
			return
		}
	}

	fmt.Println("Done")
}
