package io_git

import (
	"errors"
	"os"
	"time"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"

	"acme-deploy/src/l"
)

func (receiver *GitDB) MustLoad(path string, auth transport.AuthMethod, signKey *openpgp.Entity) {
	var (
		r = mustPlainOpen(path)
	)
	*receiver = GitDB{
		Path:       path,
		Repository: r,
		Worktree:   mustWorktree(r),
		PullOptions: &git.PullOptions{
			Auth:     auth,
			Progress: os.Stderr, // FIXME why so quiet?
		},
		CommitOptions: &git.CommitOptions{
			All:               true,
			AllowEmptyCommits: false,
			Committer: func() (outbound *object.Signature) {
				switch {
				case signKey != nil:
					// use first available identity as a committer
					for _, b := range signKey.Identities {
						outbound = &object.Signature{
							Name:  b.UserId.Name,
							Email: b.UserId.Email,
							When:  time.Now(), // wtf????
						}
						break
					}
				}
				return
			}(),
			SignKey: signKey,
			Signer:  nil,
			Amend:   false,
		},
		PushOptions: &git.PushOptions{
			Auth:     auth,
			Progress: os.Stderr, // FIXME why so quiet?
			Atomic:   true,
		},
	}
}

func (receiver *GitDB) MustCommit(msg string) {

	l.Informational.L(l.F{"repo": receiver.Path, "action": "pull&commit"})

	switch {
	case mustIsClean(receiver.Worktree):
		return
	}

	l.Informational.L(l.F{"repo": receiver.Path, "status": mustStatus(receiver.Worktree).String()})

	l.Informational.L(l.F{"repo": receiver.Path, "action": "pull"})
	mustPull(receiver.Worktree, receiver.PullOptions)

	switch {
	case mustIsClean(receiver.Worktree):
		return
	}

	l.Informational.L(l.F{"repo": receiver.Path, "status": mustStatus(receiver.Worktree).String()})

	l.Informational.L(l.F{"repo": receiver.Path, "action": "add"})
	mustAdd(receiver.Worktree, ".")
	l.Informational.L(l.F{"repo": receiver.Path, "action": "commit"})
	mustCommit(receiver.Worktree, msg, receiver.CommitOptions)
	l.Informational.L(l.F{"repo": receiver.Path, "action": "push"})
	mustPush(receiver.Repository, receiver.PushOptions)

	l.Informational.L(l.F{"repo": receiver.Path, "status": mustStatus(receiver.Worktree).String()})
}

func (receiver *AuthDB) WriteSSH(name string, user string, pemBytes []byte, password string) error {
	switch _, ok := (*receiver)[name]; {
	case ok:
		return l.EDUPDATA
	}

	switch outbound, err := ssh.NewPublicKeys(user, pemBytes, password); {
	case err != nil:
		return err
	default:
		(*receiver)[name] = outbound
		return nil
	}
}
func (receiver *AuthDB) MustWriteSSH(name string, user string, pemBytes []byte, password string) {
	switch err := receiver.WriteSSH(name, user, pemBytes, password); {
	case err != nil && errors.Is(err, l.EDUPDATA):
		l.Warning.E(err, l.F{"ssh key": name})
	case err != nil:
		l.Critical.E(err, l.F{"ssh key": name})
	}
}

func (receiver *AuthDB) WriteToken(name string, user string, tokenBytes []byte) error {
	switch _, ok := (*receiver)[name]; {
	case ok:
		return l.EDUPDATA
	}

	(*receiver)[name] = &http.BasicAuth{
		Username: user,
		Password: string(tokenBytes),
	}
	return nil
}
func (receiver *AuthDB) MustWriteToken(name string, user string, tokenBytes []byte) {
	switch err := receiver.WriteToken(name, user, tokenBytes); {
	case err != nil && errors.Is(err, l.EDUPDATA):
		l.Warning.E(err, l.F{"token": name})
	case err != nil:
		l.Critical.E(err, l.F{"token": name})
	}
}

func (receiver *AuthDB) ReadAuth(name string) (transport.AuthMethod, error) {
	switch value, ok := (*receiver)[name]; {
	case !ok:
		return nil, l.ENOTFOUND
	default:
		return value, nil
	}
}
func (receiver *AuthDB) MustReadAuth(name string) transport.AuthMethod {
	switch value, err := receiver.ReadAuth(name); {
	case err != nil:
		l.Critical.E(err, l.F{"auth": name})
		return nil
	default:
		return value
	}
}
