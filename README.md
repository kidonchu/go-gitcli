A git wrapper to provide additional features around working with stories.

## Overview

gitcli is an extended git CLI tool written in Go. It is written to stop fiddling around creating
new stories, managing remote branches, switching between working branches while reserving the
context, and cleaning up stories, including remote branches, stashes, and databases.

gitcli is designed to work well with any git repository.

## Installation

Currently only supporting MacOS 64bit binary. More binaries for different OS will be supported later.

### Installing Binary

Download the released binary from [releases](https://github.com/kidonchu/gitcli/releases) and place it
in your PATH to run globally.

### Building Binary from Source (Advanced)

To be added.

## Usages

### Creating new story

Add *source* to git config

	$> git config story.source.SOURCE_IDENTIFIER SOURCE_BRANCH_NAME
    
Add *remote target* to git config

	$> git config story.remote.target TARGET_REMOTE_NAME

Then run the gitcli command to create new branch.

	$> gitcli story new --source SOURCE_IDENTIFIER --branch NEW_BRANCH_NAME

Following operations will be executed in the order.

* Stash any changes on current branch
* Fetch most recent changes of `SOURCE_BRANCH_NAME` from its remote
* Create new branch `NEW_BRANCH_NAME` off of `SOURCE_BRANCH_NAME` by referencing it via `SOURCE_IDENTIFIER`
* Push new local branch to `TARGET_REMOTE_NAME`

For example, in the following case,

* I forked my repo from FooBar/master into origin/master
* I need to push local branch to origin/feature-branch-1
* I need to create a pull request against FooBar/master

I run these commands to create my new local branch.

	$> git config story.source.master FooBar/master
    $> git config story.remote.target origin
    $> gitcli story new --source master --branch feature-branch-1

### Switching to story

Switch to an existing local branch.

	$> gitcli story switch -p PATTERN
    
This command displays an enumerated list of local branches. If a branch is selected, the following operations will be executed.

* Stash any changes on current branch
* Checkout the chosen branch and make it HEAD
* Pop the last stash this checked-out branch has, if any

If `PATTERN` is not specified, all local branches will be presented.

If `PATTERN` is specified, only local branches whose names regex-match with the PATTERN will be presented.

### Deleting story

Add database access info to git config

	$> git config story.hosteddb.host 127.0.0.1
    $> git config story.hosteddb.port 3306
    $> git config story.hosteddb.user foo
    $> git config story.hosteddb.pass bar

Then run the gitcli command to delete/drop branches, stashes, and dbs.

	$> gitcli story delete -p PATTERN
    
This command displays a bullet list of items that will be deleted/dropped. If confirmed by typing "Y", following operations will be executed.

* Deletes local branches
* Deletes remote branches
* Drops stashes
* Drops databases

If `PATTERN` is not specified, all local branches, remote branches, stashes, and databases will be deleted.

If `PATTERN` is specified, only iterms that regex-match with the PATTERN will be deleted.

### Pulling recent changes

Add *source* to git config

	$> git config story.source.SOURCE_IDENTIFIER SOURCE_BRANCH_NAME
    
Then run the gitcli command to pull most rencent changes into current branch

	$> gitcli story pull -s SOURCE_IDENTIFIER
    
Then following operations will be executed in the order.

* Fetch most recent changes of `SOURCE_BRANCH_NAME` from its remote
* Attempt to merge `SOURCE_BRANCH_NAME` into current local branch

For example, in the following case,

* I forked my repo from FooBar/master into origin/master
* I need to create a pull request against FooBar/master
* I need to pull most recent changes of FooBar/master to make sure I don't have any conflict in my PR

I run these commands to pull.

	$> git config story.source.master FooBar/master
    $> gitcli story pull --source master
    
If there are any conflicts, you will get a message saying you have conflicts. These conflicts will need to be handled manually and comitted.

## Bonus

I have my `git` command setup in the following way.

```
git () {
        customCmds=("s" "story")
        if [[ ${customCmds[(i)$1]} -le ${#customCmds} ]]
        then
                gitcli "$@"
        else
                /usr/bin/git "$@"
        fi
}
```

This lets me use `git story new/switch/delete/pull` as if I am using the native git client.

## Author

[kidonchu](https://github.com/kidonchu)
