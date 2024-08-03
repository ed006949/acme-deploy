package io_git

import (
	"errors"
	"os"
	"time"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"

	"acme-deploy/src/l"
)

func (r *GitDB) Load(path string, auth transport.AuthMethod, signKey *openpgp.Entity) (err error) {
	r.Path = path

	switch r.Repository, err = git.PlainOpen(path); {
	case err != nil:
		return
	}
	switch r.Worktree, err = r.Repository.Worktree(); {
	case err != nil:
		return
	}

	r.PullOptions = &git.PullOptions{
		Auth:     auth,
		Progress: os.Stderr, // FIXME why so quiet?
	}
	r.CommitOptions = &git.CommitOptions{
		All:               true,
		AllowEmptyCommits: false,
		Committer: func() (outbound *object.Signature) {
			switch {
			case signKey != nil:
				for _, b := range signKey.Identities { // use first available identity as a committer
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
	}
	r.PushOptions = &git.PushOptions{
		Auth:     auth,
		Progress: os.Stderr, // FIXME why so quiet?
		Atomic:   true,
	}
	return
}

func (r *GitDB) Commit(msg string) (err error) {
	var (
		gitStatus    git.Status
		plumbingHash plumbing.Hash
	)

	switch gitStatus, err = r.Worktree.Status(); {
	case err != nil:
		return
	case gitStatus.IsClean():
		return
	}

	switch err = r.Worktree.Pull(r.PullOptions); {
	case errors.Is(err, git.NoErrAlreadyUpToDate):
	case err != nil:
		return
	}

	switch gitStatus, err = r.Worktree.Status(); {
	case err != nil:
		return
	case gitStatus.IsClean():
		return
	}

	switch plumbingHash, err = r.Worktree.Add("."); {
	case err != nil:
		return
	}
	l.Z{"repo": r.Path, "plumbingHash": plumbingHash.String()}.Informational()

	switch {
	case r.CommitOptions != nil && r.CommitOptions.Committer != nil:
		r.CommitOptions.Committer.When = time.Now() // and again, wtf????
	}

	switch plumbingHash, err = r.Worktree.Commit(msg, r.CommitOptions); {
	case err != nil:
		return
	}
	l.Z{"repo": r.Path, "plumbingHash": plumbingHash.String()}.Informational()

	switch err = r.Repository.Push(r.PushOptions); {
	case errors.Is(err, git.NoErrAlreadyUpToDate):
	case err != nil:
		return
	}

	switch gitStatus, err = r.Worktree.Status(); {
	case err != nil:
		return
	case gitStatus.IsClean():
		return
	}

	return
}

func (r *AuthDB) WriteSSH(name string, user string, pemBytes []byte, password string) (err error) {
	switch _, ok := (*r)[name]; {
	case ok:
		return l.EDUPDATA
	}

	var (
		sshPublicKeys *ssh.PublicKeys
	)
	switch sshPublicKeys, err = ssh.NewPublicKeys(user, pemBytes, password); {
	case err != nil:
		return
	default:
		(*r)[name] = sshPublicKeys
		return
	}
}

func (r *AuthDB) WriteToken(name string, user string, tokenBytes []byte) (err error) {
	switch _, ok := (*r)[name]; {
	case ok:
		return l.EDUPDATA
	}

	(*r)[name] = &http.BasicAuth{
		Username: user,
		Password: string(tokenBytes),
	}
	return
}

func (r *AuthDB) ReadAuth(name string) (outbound transport.AuthMethod, err error) {
	switch value, ok := (*r)[name]; {
	case !ok:
		return nil, l.ENOTFOUND
	default:
		return value, nil
	}
}
