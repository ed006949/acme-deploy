package io_git

import (
	"errors"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"

	"acme-deploy/src/l"
)

func mustPlainOpen(path string) *git.Repository {
	switch outbound, err := git.PlainOpen(path); {
	case err != nil:
		l.Critical.E(err, l.F{"name": path})
		return nil
	default:
		return outbound
	}
}
func mustWorktree(r *git.Repository) *git.Worktree {
	switch outbound, err := r.Worktree(); {
	case err != nil:
		l.Critical.E(err, nil)
		return nil
	default:
		return outbound
	}
}
func mustGetConfig(r *git.Repository) *config.Config {
	switch outbound, err := r.Config(); {
	case err != nil:
		l.Critical.E(err, nil)
		return nil
	default:
		return outbound
	}
}
func mustSetConfig(r *git.Repository, c *config.Config) {
	switch err := r.SetConfig(c); {
	case err != nil:
		l.Critical.E(err, nil)
	}
}

func mustPull(w *git.Worktree, o *git.PullOptions) {
	switch err := w.Pull(o); {
	case errors.Is(err, git.NoErrAlreadyUpToDate):
	case err != nil:
		l.Critical.E(err, nil)
	}
}

func mustAdd(w *git.Worktree, path string) {
	switch _, err := w.Add(path); {
	case err != nil:
		l.Critical.E(err, l.F{"name": path})
	}
}
func mustCommit(w *git.Worktree, msg string, opts *git.CommitOptions) {
	switch {
	case opts != nil && opts.Committer != nil:
		opts.Committer.When = time.Now() // and again, wtf????
	}

	switch _, err := w.Commit(msg, opts); {
	case err != nil:
		l.Critical.E(err, l.F{"commit message": msg})
	}
}
func mustPush(r *git.Repository, o *git.PushOptions) {
	switch err := r.Push(o); {
	case errors.Is(err, git.NoErrAlreadyUpToDate):
	case err != nil:
		l.Critical.E(err, nil)
	}
}
func mustStatus(w *git.Worktree) git.Status {
	switch value, err := w.Status(); {
	case err != nil:
		l.Critical.E(err, nil)
		return nil
	default:
		return value
	}
}
func mustIsClean(w *git.Worktree) bool {
	return mustStatus(w).IsClean()
}
