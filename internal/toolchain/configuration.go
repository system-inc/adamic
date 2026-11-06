package toolchain

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// configurationLine begins each line clang's --version writes for a config file it read.
const configurationLine = "Configuration file: "

// ConfigurationsByContent is a clang report with every config file it names keyed by what the config holds, never by
// where it sits (Node's rule on #t3rw158): each `Configuration file: <path>` line becomes the lines ConfigurationFile
// gives for that path. An arm64 clang carries its sysroot and lld in its own target config, and keyed by its path one
// clang kept at two places keyed every product apart; keyed by its bytes, a box that keeps the same clang and configs
// somewhere else moves no key, and an edit of any config, or of a file one names, moves every key it reaches. A config
// that can't be read is named by its error, so the report moves rather than reading as another's.
func ConfigurationsByContent(report string) string {
	var lines []string
	for _, line := range strings.Split(report, "\n") {
		path, found := strings.CutPrefix(strings.TrimSpace(line), configurationLine)
		if !found {
			lines = append(lines, line)
			continue
		}
		configuration, err := ConfigurationFile(path)
		if err != nil {
			lines = append(lines, configurationLine+"unreadable ("+err.Error()+")")
			continue
		}
		lines = append(lines, configuration...)
	}
	return strings.Join(lines, "\n")
}

// ConfigurationFile is a clang config file at path as a key holds it, with no path in it: `Configuration file: sha256
// <digest>` for its bytes, then a line for each file it names, in its order. A config it includes (@file, from the
// including config's directory) is its own `Configuration file:` line and its names, after the including one's; any
// other file it names by an absolute path, after clang's <CFGDIR> is spelled out (lld's binary, by --ld-path= or
// -fuse-ld=), is `Configuration names: sha256 <digest>`. A directory it names (the sysroot, -B's) is keyed where the
// recipe keys it, by its archive's digest, not here, and a name that's no file (-fuse-ld=lld) is found by clang where
// the key already reads it.
func ConfigurationFile(path string) ([]string, error) {
	return configurationFile(path, map[string]bool{})
}

// configurationFile is ConfigurationFile, with seen the configs already read, so an include cycle ends.
func configurationFile(path string, seen map[string]bool) ([]string, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	seen[path] = true
	lines := []string{fmt.Sprintf("%ssha256 %x", configurationLine, sha256.Sum256(contents))}
	directory := filepath.Dir(path)
	for _, word := range configurationWords(string(contents)) {
		word = strings.ReplaceAll(word, "<CFGDIR>", directory)
		if included, found := strings.CutPrefix(word, "@"); found {
			if !filepath.IsAbs(included) {
				included = filepath.Join(directory, included)
			}
			if seen[included] {
				continue
			}
			nested, err := configurationFile(included, seen)
			if err != nil {
				return nil, err
			}
			lines = append(lines, nested...)
			continue
		}
		candidates := []string{word}
		if _, value, found := strings.Cut(word, "="); found {
			candidates = append(candidates, value)
		}
		for _, candidate := range candidates {
			if !filepath.IsAbs(candidate) {
				continue
			}
			if info, err := os.Stat(candidate); err != nil || !info.Mode().IsRegular() {
				continue
			}
			named, err := os.ReadFile(candidate)
			if err != nil {
				return nil, err
			}
			lines = append(lines, fmt.Sprintf("Configuration names: sha256 %x", sha256.Sum256(named)))
		}
	}
	return lines, nil
}

// configurationWords are a config's words as clang reads them: a line beginning with # is a comment, a backslash at a
// line's end joins it to the next, and words are split at white space, a quoted one kept whole without its quotes.
func configurationWords(contents string) []string {
	var kept []string
	for _, line := range strings.Split(strings.ReplaceAll(contents, "\\\n", ""), "\n") {
		if !strings.HasPrefix(strings.TrimSpace(line), "#") {
			kept = append(kept, line)
		}
	}
	var words []string
	var word strings.Builder
	quote := rune(0)
	inWord := false
	for _, character := range strings.Join(kept, "\n") {
		switch {
		case quote != 0 && character == quote:
			quote = 0
		case quote != 0:
			word.WriteRune(character)
		case character == '"' || character == '\'':
			quote, inWord = character, true
		case character == ' ' || character == '\t' || character == '\n' || character == '\r':
			if inWord {
				words = append(words, word.String())
				word.Reset()
				inWord = false
			}
		default:
			word.WriteRune(character)
			inWord = true
		}
	}
	if inWord {
		words = append(words, word.String())
	}
	return words
}
