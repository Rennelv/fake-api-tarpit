package generator

import (
	"fmt"
)

// GitGenerator produces .git/config and .git-credentials contents.
type GitGenerator struct {
	lineCount int
	isCredentials bool
}

// NewGitGenerator creates a generator for Git repository probes.
func NewGitGenerator(isCredentials bool) *GitGenerator {
	return &GitGenerator{isCredentials: isCredentials}
}

// NextChunk returns the next line of git config or credentials.
func (g *GitGenerator) NextChunk() string {
	idx := g.lineCount
	g.lineCount++

	if g.isCredentials {
		// Output git-credentials lines with realistic tokens
		token := "ghp_" + RandomAlphanumeric(36)
		user := RandomChoice([]string{"deploy-bot", "ci-builder", "git-sync", "admin"})
		return fmt.Sprintf("https://%s:%s@github.com\n", user, token)
	}

	// Authentic .git/config format
	gitConfigLines := []string{
		"[core]",
		"\trepositoryformatversion = 0",
		"\tfilemode = true",
		"\tbare = false",
		"\tlogallrefupdates = true",
		"[remote \"origin\"]",
		fmt.Sprintf("\turl = https://ghp_%s@github.com/organization/production-core.git", RandomAlphanumeric(36)),
		"\tfetch = +refs/heads/*:refs/remotes/origin/*",
		"[branch \"main\"]",
		"\tremote = origin",
		"\tmerge = refs/heads/main",
		"[user]",
		"\tname = Production Deployment Agent",
		"\temail = devops@internal-infra.net",
		"[credential]",
		"\thelper = store --file /root/.git-credentials",
	}

	if idx < len(gitConfigLines) {
		return gitConfigLines[idx] + "\n"
	}

	// Continue streaming realistic submodules and extra remotes
	subIndex := idx - len(gitConfigLines)
	if subIndex%5 == 0 {
		return fmt.Sprintf("[submodule \"infra-shared-%d\"]\n\tpath = vendor/infra-%d\n\turl = https://ghp_%s@github.com/organization/infra-%d.git\n",
			subIndex, subIndex, RandomAlphanumeric(36), subIndex)
	}

	return fmt.Sprintf("\ttoken = ghp_%s\n", RandomAlphanumeric(36))
}
