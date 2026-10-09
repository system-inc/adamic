package main

import (
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/system-inc/cohere/internal/corpus"
)

// Corpora: some tests read code outside this repository, by name (internal/corpus). Every run sets each
// corpus's variable the user's corpora file names and the environment leaves unset, so a developer who has
// the checkouts writes the file once and every gate covers them. Every run ends by naming the corpora it did
// not cover, whether its tests ran or answered from Go's cache, since a skip in a cached package prints
// nothing (#sycrdr6).

// testEnvironment is the environment a go test run gets, base with the corpora file's variables added, and
// the corpora that stay uncovered.
func testEnvironment(base []string) ([]string, []corpus.Corpus, error) {
	environment := slices.Clone(base)
	path, err := corpus.ConfigPath()
	if err != nil {
		return nil, nil, err
	}
	config, err := corpus.ReadConfig(path)
	if err != nil {
		return nil, nil, err
	}
	added := corpus.Environment(config, os.Getenv)
	environment = append(environment, added...)
	set := map[string]string{}
	for _, variable := range environment {
		if name, value, found := strings.Cut(variable, "="); found {
			set[name] = value
		}
	}
	return environment, corpus.Uncovered(func(name string) string { return set[name] }), nil
}

// printUncovered names the corpora a run did not cover, and how to cover them, or nothing when it covered
// every one.
func printUncovered(writer io.Writer, uncovered []corpus.Corpus) {
	if len(uncovered) == 0 {
		return
	}
	names := make([]string, 0, len(uncovered))
	for _, uncoveredCorpus := range uncovered {
		names = append(names, fmt.Sprintf("%s (%s)", uncoveredCorpus.Name, uncoveredCorpus.Variable))
	}
	path, _ := corpus.ConfigPath()
	fmt.Fprintf(writer, "cohere-dev: not covered, so their tests skipped: %s. Set the variables, or name the "+
		"checkouts in %s.\n", strings.Join(names, ", "), path)
}
