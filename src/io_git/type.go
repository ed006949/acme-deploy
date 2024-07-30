package io_git

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

type AuthDB map[string]transport.AuthMethod

type GitDB struct {
	Path          string
	Repository    *git.Repository
	Worktree      *git.Worktree
	PullOptions   *git.PullOptions
	CommitOptions *git.CommitOptions
	PushOptions   *git.PushOptions
}
