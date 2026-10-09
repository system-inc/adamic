// Package corpus names the code outside this repository that some tests read, and is the only way a test
// finds it (#sycrdr6).
//
// cohere is public, and a test that reads a path on one developer's machine skips everywhere else, so its
// coverage existed only on that machine and nobody could tell. A test asks for a corpus by name instead.
// The corpus's variable says where it is. When the variable is unset the test skips with a line naming it,
// and `cohere-dev test` ends every run by naming the corpora it did not cover, so a green run says what it
// left out. No test names a path: internal/testpolicy fails one that names a home directory.
//
// `cohere-dev test` sets each unset variable from the user's corpora file, so a developer who has the
// checkouts writes the file once and every gate covers them, without exporting anything.
package corpus

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// Corpus is code outside this repository that tests read.
type Corpus struct {
	// Name is how the corpora file names it.
	Name string
	// Variable is the environment variable that holds its path.
	Variable string
	// What says what the path is, in the skip line.
	What string
}

var (
	// Ahra is a checkout of ahra, which is private: its CohereSettings.json and the chain it extends, its
	// design system's stylesheets, and its installed packages.
	Ahra = Corpus{Name: "ahra", Variable: "COHERE_CORPUS_AHRA", What: "a checkout of ahra (private)"}
	// Structure is a checkout of Structure, which is private, for the TypeScript under its source/.
	Structure = Corpus{Name: "structure", Variable: "COHERE_CORPUS_STRUCTURE", What: "a checkout of Structure (private)"}
	// Connected is a checkout of www-connected-app, which is private, for its design system.
	Connected = Corpus{Name: "connected", Variable: "COHERE_CORPUS_CONNECTED", What: "a checkout of www-connected-app (private)"}
	// PrettierFork is a checkout of the Prettier fork cohere's formatter is held to, with its node_modules
	// installed, for the oracles that run its parsers.
	PrettierFork = Corpus{Name: "prettier-fork", Variable: "COHERE_PRETTIER_FORK", What: "a checkout of the Prettier fork, with node_modules"}
)

// All is every corpus, in the order the gate names them.
var All = []Corpus{Ahra, Structure, Connected, PrettierFork}

// Root is the corpus's directory, or a skip naming the variable to set when it is unset. A variable set to
// something that is not a directory fails the test: the developer asked for coverage they are not getting.
func (corpus Corpus) Root(t testing.TB) string {
	t.Helper()
	root := os.Getenv(corpus.Variable)
	if root == "" {
		t.Skipf("%s: set %s to %s to cover this test", corpus.Name, corpus.Variable, corpus.What)
	}
	information, err := os.Stat(root)
	if err != nil || !information.IsDir() {
		t.Fatalf("%s is set to %s, which is not a directory, so this test cannot read %s", corpus.Variable, root, corpus.What)
	}
	return root
}

// Covered reports whether the corpus's variable is set, so a test whose cases skip one by one can count
// how many it should have run.
func (corpus Corpus) Covered() bool {
	return os.Getenv(corpus.Variable) != ""
}

// Path is a path inside the corpus, or the skip or failure Root gives.
func (corpus Corpus) Path(t testing.TB, elements ...string) string {
	t.Helper()
	return filepath.Join(append([]string{corpus.Root(t)}, elements...)...)
}

// A file that records a path inside a corpus, such as a fixture's entry point, spells it
// "<corpus>:<path in it>", as in "ahra:app/_theme/styles/theme.css", so the file reads the same on every
// machine. "ahra:" alone is the corpus's root. A corpus name is a word, never one letter, so a spelling
// cannot be read as a drive letter.
var spellingPattern = regexp.MustCompile(`^([a-z][a-z0-9-]+):(.*)$`)

// Spelled splits a "<corpus>:<path in it>" spelling into the corpus and the slash-separated path inside
// it. spelled is false for anything else, such as an absolute path. A spelling whose name is no corpus
// is an error, never a relative path, so a typo cannot read some other file.
func Spelled(spelling string) (corpus Corpus, inside string, spelled bool, err error) {
	match := spellingPattern.FindStringSubmatch(spelling)
	if match == nil {
		return Corpus{}, "", false, nil
	}
	for _, candidate := range All {
		if candidate.Name == match[1] {
			return candidate, match[2], true, nil
		}
	}
	return Corpus{}, "", true, fmt.Errorf("%s names the corpus %q, which is no corpus; the corpora are %s", spelling, match[1], corpusNames())
}

// Resolve is the path a "<corpus>:<path in it>" spelling names. When the corpus is unset it skips naming
// the variable, as Root does. It fails when the corpus is set and the path is not there, and when the
// spelling names no corpus or is not a spelling at all.
func Resolve(t testing.TB, spelling string) string {
	t.Helper()
	corpus, inside, spelled, err := Spelled(spelling)
	if err != nil {
		t.Fatal(err)
	}
	if !spelled {
		t.Fatalf("%q is not spelled <corpus>:<path in it>, so it names no corpus", spelling)
	}
	path := corpus.Path(t, filepath.FromSlash(inside))
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("%s is set, but %s is not in it: %v", corpus.Variable, spelling, err)
	}
	return path
}

// Locate is Resolve for a generator, which has no test to skip: the corpus's variable, or else the user's
// corpora file, gives the root, and an error says which to set when neither does. Anything that is not a
// spelling is a path, returned as it was given.
func Locate(spelling string) (string, error) {
	return locate(spelling, os.Getenv, ConfigPath)
}

func locate(spelling string, lookup func(string) string, configPath func() (string, error)) (string, error) {
	corpus, inside, spelled, err := Spelled(spelling)
	if err != nil || !spelled {
		return spelling, err
	}
	root := lookup(corpus.Variable)
	if root == "" {
		path, err := configPath()
		if err != nil {
			return "", err
		}
		config, err := ReadConfig(path)
		if err != nil {
			return "", err
		}
		root = config[corpus.Name]
	}
	if root == "" {
		return "", fmt.Errorf("%s needs %s: set %s, or name %q in the corpora file", spelling, corpus.What, corpus.Variable, corpus.Name)
	}
	return filepath.Join(root, filepath.FromSlash(inside)), nil
}

// Uncovered is every corpus whose variable lookup leaves unset, in All's order.
func Uncovered(lookup func(string) string) []Corpus {
	uncovered := []Corpus{}
	for _, corpus := range All {
		if lookup(corpus.Variable) == "" {
			uncovered = append(uncovered, corpus)
		}
	}
	return uncovered
}

// ConfigVariable names a corpora file to read in place of the default one.
const ConfigVariable = "COHERE_CORPORA_CONFIG"

// ConfigPath is the user's corpora file: ConfigVariable when it is set, and otherwise
// ~/.config/cohere/corpora.json.
func ConfigPath() (string, error) {
	if path := os.Getenv(ConfigVariable); path != "" {
		return path, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "cohere", "corpora.json"), nil
}

// ReadConfig reads a corpora file, a JSON object from corpus name to absolute path:
//
//	{ "ahra": "/path/to/ahra", "structure": "/path/to/ahra/libraries/structure" }
//
// A file that does not exist is no corpora. A name no corpus has, or a relative path, is refused, so a
// typo cannot quietly leave a corpus uncovered.
func ReadConfig(path string) (map[string]string, error) {
	contents, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return map[string]string{}, nil
	}
	if err != nil {
		return nil, err
	}
	paths := map[string]string{}
	if err := json.Unmarshal(contents, &paths); err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	known := map[string]bool{}
	for _, corpus := range All {
		known[corpus.Name] = true
	}
	names := make([]string, 0, len(paths))
	for name := range paths {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if !known[name] {
			return nil, fmt.Errorf("%s names %q, which is no corpus; the corpora are %s", path, name, corpusNames())
		}
		if !filepath.IsAbs(paths[name]) {
			return nil, fmt.Errorf("%s gives %q the relative path %s; give it absolutely", path, name, paths[name])
		}
	}
	return paths, nil
}

// Environment is the variables to add to a test run's environment: each corpus the config names whose
// variable lookup leaves unset. A variable already set wins, so one run can point a corpus elsewhere.
func Environment(config map[string]string, lookup func(string) string) []string {
	environment := []string{}
	for _, corpus := range All {
		if path, found := config[corpus.Name]; found && lookup(corpus.Variable) == "" {
			environment = append(environment, corpus.Variable+"="+path)
		}
	}
	return environment
}

func corpusNames() string {
	names := make([]string, 0, len(All))
	for _, corpus := range All {
		names = append(names, corpus.Name)
	}
	return strings.Join(names, ", ")
}
