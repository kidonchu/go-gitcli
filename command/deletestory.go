package command

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

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

	// find branches to delete
	branches, err := gitutil.FindBranches(repo, "^.*"+pattern+".*$", git.BranchLocal)
	if err != nil {
		log.Fatal(err)
		return
	}

	// find stashes to delete
	stashes := gitutil.FindStashes(repo, "^.*"+pattern+".*$")

	// Now handle databases
	dbh, err := getDbConnection()
	if err == nil {
		defer dbh.Close()
	}

	var dbs []string
	if dbh != nil {
		// find dbs to delete
		dbs, _ = dbutil.FindDbs(dbh, "^.*"+pattern+".*$")
	}

	if len(branches) < 1 && len(stashes) < 1 && len(dbs) < 1 {
		fmt.Println("Nothing to delete")
		return
	}

	branchesToDelete, stashesToDelete, dbsToDelete, err := getItemsToDelete(branches, stashes, dbs)
	if err != nil {
		log.Fatal(err)
	}

	if len(branchesToDelete) > 0 {
		remote, err := gitutil.GetRemote(repo, "origin")
		if err != nil {
			fmt.Printf("%+v", err)
		}

		err = gitutil.DeleteBranches(repo, remote, branchesToDelete)
		if err != nil {
			fmt.Printf("%+v", err)
		}
	}

	if len(stashesToDelete) > 0 {
		gitutil.DeleteStashes(repo, stashesToDelete)
	}

	if len(dbsToDelete) > 0 {
		err = dbutil.Drop(dbh, dbsToDelete)
		if err != nil {
			fmt.Printf("%+v", err)
		}
	}
}

func getDbConnection() (*sql.DB, error) {
	var (
		host, _ = gitutil.ConfigString("story.hosteddb.host")
		port, _ = gitutil.ConfigInt32("story.hosteddb.port")
		user, _ = gitutil.ConfigString("story.hosteddb.user")
		pass, _ = gitutil.ConfigString("story.hosteddb.pass")
	)

	dbh, err := dbutil.Connect(host, port, user, pass)
	if err != nil {
		return nil, err
	}
	return dbh, nil
}

func getItemsToDelete(
	branches gitutil.Branches,
	stashes map[int]*gitutil.StashInfo,
	dbs []string,
) (
	[]*git.Branch,
	map[int]*gitutil.StashInfo,
	[]string,
	error,
) {

	// prepare delete options
	options := make(map[int]string)
	optionIndex := 0

	if len(branches) > 0 {
		sort.Sort(branches)
		fmt.Println("Branches:")
		for i, b := range branches {
			name, _ := b.Name()
			options[optionIndex] = "branch-" + strconv.Itoa(i)
			optionIndex++
			fmt.Printf("%d. %s\n", optionIndex, name)
		}
		fmt.Println("")
	}
	if len(stashes) > 0 {
		fmt.Println("Stashes:")
		for i, stash := range stashes {
			options[optionIndex] = "stash-" + strconv.Itoa(i)
			optionIndex++
			fmt.Printf("%d. %s\n", optionIndex, stash.Msg)
		}
		fmt.Println("")
	}
	if len(dbs) > 0 {
		sort.Strings(dbs)
		fmt.Println("Databases:")
		for i, database := range dbs {
			options[optionIndex] = "database-" + strconv.Itoa(i)
			optionIndex++
			fmt.Printf("%d. %s\n", optionIndex, database)
		}
		fmt.Println("")
	}

	answer := GetUserInput("Choose options to delete (separted by spaces): ")
	choices := strings.Split(answer, " ")

	var branchesToDelete []*git.Branch
	stashesToDelete := make(map[int]*gitutil.StashInfo)
	var dbsToDelete []string

	for _, i := range choices {
		optionIndex, _ := strconv.Atoi(i)
		optionIndex--
		option := strings.Split(options[optionIndex], "-")
		index, _ := strconv.Atoi(option[1])

		switch option[0] {
		case "branch":
			branchesToDelete = append(branchesToDelete, branches[index])
		case "stash":
			stashesToDelete[index] = stashes[index]
		case "database":
			dbsToDelete = append(dbsToDelete, dbs[index])
		}
	}

	return branchesToDelete, stashesToDelete, dbsToDelete, nil
}
