package gitutil

import (
	"regexp"
	"strings"

	"github.com/containerd/containerd/errdefs"
)

type GitRef struct {
	Repo   string
	Commit string // Optional
	Dir    string
}

var httpPrefix = regexp.MustCompile(`^https?://`)
var gitURLPathWithFragmentSuffix = regexp.MustCompile(`\.git(?:#.+)?$`)

func IsGitRef(ref string) bool {
	if httpPrefix.MatchString(ref) && gitURLPathWithFragmentSuffix.MatchString(ref) {
		return true
	}
	for _, prefix := range []string{"git://", "github.com/", "git@"} {
		if strings.HasPrefix(ref, prefix) {
			return true
		}
	}
	return false
}

func ParseGitRef(ref string) (*GitRef, error) {
	if !IsGitRef(ref) {
		return nil, errdefs.ErrInvalidArgument
	}
	refSplitBySharp := strings.SplitN(ref, "#", 2)
	repo := refSplitBySharp[0]
	commit := ""
	if len(refSplitBySharp) > 1 {
		commit = refSplitBySharp[1]
	}
	repoSplitBySlash := strings.Split(ref, "/")
	dir := strings.TrimSuffix(repoSplitBySlash[len(repoSplitBySlash)-1], ".git")
	return &GitRef{
		Repo:   repo,
		Commit: commit,
		Dir:    dir,
	}, nil
}
