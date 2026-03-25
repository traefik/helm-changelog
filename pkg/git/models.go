// Package git provides utilities for interacting with git repositories.
package git

import "time"

// gitformat defines a yaml formatting string for git.
var gitformat = `--pretty=format:
---
commit: "%H"
parent: "%P"
refs: "%D"
subject: |-
  %s

author: { "name": "%aN", "email": "%aE", "date": "%ad" }
commiter: { "name": "%cN", "email": "%cE", "date": "%cd" }
`

// Person represents a git actor (author or committer).
type Person struct {
	Name  string     `yaml:"name"`
	Email string     `yaml:"email"`
	Date  *time.Time `yaml:"date"`
}

// Commit represents a single git commit.
type Commit struct {
	Hash     string `yaml:"commit"`
	Parent   string `yaml:"parent"`
	Refs     string `yaml:"refs"`
	Subject  string `yaml:"subject"`
	Author   Person `yaml:"author"`
	Commiter Person `yaml:"commiter"`
}
