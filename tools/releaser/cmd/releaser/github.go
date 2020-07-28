package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/go-git/go-git/v5/plumbing/transport"
	gogithttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	gogitssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

func (r *releaser) readGithubToken() error {
	if r.opts.githubToken == "" {
		return nil
	}
	f, err := os.Open(r.opts.githubToken)
	if err != nil {
		return errors.Wrap(err, "failed opening github token file")
	}
	token, err := ioutil.ReadAll(f)
	if err != nil {
		return errors.Wrap(err, "failed reading github token file")
	}
	r.githubToken = string(token)
	return nil
}

type gitConfig struct {
	auth transport.AuthMethod
	url  string
}

func (r *releaser) composeGitSSHAuth() (transport.AuthMethod, error) {
	auth, err := gogitssh.NewPublicKeysFromFile("git", r.opts.githubSSHKey, "")
	if err != nil {
		return nil, errors.Wrap(err, "failed composing new ssh key for git client")
	}
	auth.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	return auth, nil
}

func (r *releaser) composeGitHTTPTokenAuth() (transport.AuthMethod, error) {
	auth := gogithttp.TokenAuth{Token: r.githubToken}
	return &auth, nil
}

func (r *releaser) composeGitHTTPBasicAuth() (transport.AuthMethod, error) {
	auth := gogithttp.BasicAuth{Username: r.opts.githubUser, Password: r.githubToken}
	return &auth, nil
}

func (r *releaser) composeGitSSHRepoURL() string {
	return fmt.Sprintf("ssh://git@%s/%s/%s.git", r.opts.githubServer, r.opts.githubOrg, r.opts.githubRepo)
}

func (r *releaser) composeGitHTTPRepoURL() string {
	return fmt.Sprintf("%s://%s/%s/%s.git", r.opts.githubHTTPSchema, r.opts.githubServer, r.opts.githubOrg, r.opts.githubRepo)
}

func (r *releaser) composeGitConfig() (gitConfig, error) {
	var composeGitAuth func(*releaser) (transport.AuthMethod, error)
	var composeGitRepoURL func(*releaser) string
	if r.opts.githubToken != "" {
		logrus.Info("Selecting HTTP Basic Auth")
		composeGitAuth = (*releaser).composeGitHTTPBasicAuth
		composeGitRepoURL = (*releaser).composeGitHTTPRepoURL
	} else {
		logrus.Info("Selecting SSH Auth")
		composeGitAuth = (*releaser).composeGitSSHAuth
		composeGitRepoURL = (*releaser).composeGitSSHRepoURL
	}
	auth, err := composeGitAuth(r)
	if err != nil {
		return gitConfig{}, errors.Wrap(err, "failed composing git auth")
	}
	return gitConfig{
		auth: auth,
		url:  composeGitRepoURL(r),
	}, nil
}
