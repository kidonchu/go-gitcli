package command

import (
	"fmt"
	"log"

	"github.com/codegangsta/cli"
	_ "github.com/go-sql-driver/mysql" // mysql driver
	"github.com/kidonchu/gitcli/dbutil"
	"github.com/kidonchu/gitcli/gitutil"
	"github.com/kidonchu/gitcli/util"
	"github.com/libgit2/git2go"
)

var (
	branches []*git.Branch
	remotes  []*git.Remote
)

// CmdDeleteStory deletes story
// First, it deletes local and remote branchs whose name matches with the ticket number
// Then, it deletes the databases whose name matches with the ticket number
func CmdDeleteStory(c *cli.Context) {
	// Get ticket number to delete
	ticket := c.String("ticket")

	// Prepare gitconfig
	config, err := gitutil.GetGitconfig()
	if err != nil {
		log.Fatal(err)
		return
	}

	// Get repo instance
	root, err := config.LookupString("story.acroot.ember")
	if err != nil {
		log.Fatal(err)
		return
	}
	repo, err := gitutil.GetRepo(root)
	if err != nil {
		log.Fatal(err)
		return
	}

	branches, err := gitutil.FindBranches(repo, "^.*"+ticket+".*$", git.BranchLocal)
	if err != nil {
		log.Fatal(err)
		return
	}
	if len(branches) < 1 {
		fmt.Println("No branch to delete")
		// we don't want to return here because there can be databases that don't have branches for it
	} else {
		// Show list of branches to be deleted and get confirmation
		fmt.Println("Following branches will be deleted")
		for _, b := range branches {
			name, _ := b.Name()
			fmt.Printf("* %s\n", name)
		}
		answer := util.GetUserInput("Continue? (nY): ")
		if answer == "Y" {
			remote, err := gitutil.FindRemote(repo, "origin")
			if err != nil {
				log.Fatal(err)
				return
			}

			err = gitutil.DeleteBranches(repo, remote, branches)
			if err != nil {
				log.Fatal(err)
				return
			}
			fmt.Println("Successfully deleted branch(es)")
		}
	}

	// Now handle databases
	var (
		host, _ = config.LookupString("story.hosteddb.host")
		port, _ = config.LookupInt32("story.hosteddb.port")
		user, _ = config.LookupString("story.hosteddb.user")
		pass, _ = config.LookupString("story.hosteddb.pass")
	)

	dbh, err := dbutil.Connect(host, port, user, pass)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer dbh.Close()

	// find list of dbs to delete
	dbs, err := dbutil.FindDbs(dbh, "^.*"+ticket+".*$")
	if err != nil {
		log.Fatal(err)
		return
	}
	if len(dbs) < 1 {
		fmt.Println("No database to delete")
		return
	}

	// Show list of dbs to be deleted and get confirmation
	fmt.Println("Following databases will be deleted")
	for _, database := range dbs {
		fmt.Printf("* %s\n", database)
	}
	answer := util.GetUserInput("Continue? (nY): ")
	if answer == "Y" {
		err = dbutil.Drop(dbh, dbs)
		if err != nil {
			log.Fatal(err)
			return
		}
	}
	fmt.Println("All Done")
}
