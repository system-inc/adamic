package main

import (
	"github.com/system-inc/cohere/internal/gitignore"
)

// gitignoreScope is internal/gitignore's matcher for one directory (#ndtgy1w), as discovery asks it.
type gitignoreScope struct {
	matcher *gitignore.Matcher
}

func newGitignoreScope(repositoryRoot string) (ignoreScope, error) {
	matcher, err := gitignore.New(repositoryRoot)
	if err != nil {
		return nil, err
	}
	return gitignoreScope{matcher: matcher}, nil
}

func (scope gitignoreScope) Enter(relativeDirectory string) (ignoreScope, error) {
	entered, err := scope.matcher.Enter(relativeDirectory)
	if err != nil {
		return nil, err
	}
	return gitignoreScope{matcher: entered}, nil
}

func (scope gitignoreScope) Ignored(relativePath string, isDirectory bool) bool {
	ignored, _ := scope.matcher.Ignored(relativePath, isDirectory)
	return ignored
}
