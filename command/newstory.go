package command

import "github.com/codegangsta/cli"

func CmdNew(c *cli.Context) {
	// config := getGitconfig()

	// repoRoots := []string{
	// 	"acroots.ember",
	// 	"acroots.hosted",
	// }

	// var repos []*git.Repository
	// for _, repoRoot := range repoRoots {
	// 	root, err := config.LookupString(repoRoot)
	// 	if err != nil {
	// 		fmt.Printf("Unable to find root for %s", repoRoot)
	// 		return
	// 	}
	// 	repo := getRepo(root)
	// 	if repo == nil {
	// 		return
	// 	}

	// 	fetchOptions := &git.FetchOptions{
	// 		RemoteCallbacks: git.RemoteCallbacks{
	// 			CredentialsCallback: func(url string, username string, allowedTypes git.CredType) (git.ErrorCode, *git.Cred) {
	// 				ret, cred := git.NewCredSshKey(
	// 					"git", "/Users/kchu/.ssh/id_rsa_github.pub",
	// 					"/Users/kchu/.ssh/id_rsa_github", "")
	// 				return git.ErrorCode(ret), &cred
	// 			},
	// 			CertificateCheckCallback: func(cert *git.Certificate, valid bool, hostname string) git.ErrorCode {
	// 				return 0
	// 			},
	// 		},
	// 	}

	// 	remotes, err := repo.Remotes.List()
	// 	for _, remote := range remotes {
	// 		r, err := repo.Remotes.Lookup(remote)
	// 		if err != nil {
	// 			log.Fatal(err)
	// 		}

	// 		err = r.Fetch([]string{}, fetchOptions, "")
	// 		if err != nil {
	// 			log.Fatal(err)
	// 		}
	// 	}

	// 	repos = append(repos, repo)
	// }

	// branch := c.String("branch")
	// if branch == "" {
	// 	fmt.Println("Branch to create is not specified")
	// 	return
	// }

	// from := c.String("from")
	// source := lookupBranchSource(config, from)
	// if source == "" {
	// 	return
	// }

	// fmt.Printf("* Create new branch %s from %s in ember-app\n", branch, source)
	// fmt.Printf("* Create new branch %s from %s in Hosted\n", branch, source)

	// answer := getUserInput("Proceed with above items? (nY): ")
	// if answer != "Y" {
	// 	fmt.Println("What? Are you kidding? You got me working so hard and now you are quitting. FCUK")
	// 	return
	// }

	// for _, repo := range repos {
	// 	stashChanges(repo)
	// 	createBranch(repo, branch, source)
	// 	push(repo, branch)
	// }

}

// func lookupBranchSource(config *git.Config, from string) string {

// 	// default source: contact-deal
// 	if from == "" {
// 		from = "contact"
// 	}

// 	source, err := config.LookupString("storysource." + from)
// 	if err != nil {
// 		fmt.Printf("Unable to find source for %s", from)
// 		return ""
// 	}

// 	return source
// }

// func createBranch(repo *git.Repository, branch string, source string) *git.Branch {

// 	newBranch, err := repo.LookupBranch(branch, git.BranchLocal)
// 	if err != nil {
// 		// find source branch to create new branch from
// 		sourceBranch, err := repo.LookupBranch(source, git.BranchRemote)
// 		if err != nil {
// 			log.Fatal(err)
// 		}

// 		sourceCommit, err := repo.LookupCommit(sourceBranch.Target())
// 		if err != nil {
// 			log.Fatal(err)
// 		}

// 		newBranch, err = repo.CreateBranch(branch, sourceCommit, false)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 	}

// 	repo.References.Lookup(branch)
// 	_, err = repo.References.CreateSymbolic("HEAD", "refs/heads/"+branch, true, "headOne")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	opts := &git.CheckoutOpts{
// 		Strategy: git.CheckoutSafe | git.CheckoutRecreateMissing,
// 	}
// 	if err := repo.CheckoutHead(opts); err != nil {
// 		log.Fatal(err)
// 	}

// 	return newBranch
// }
