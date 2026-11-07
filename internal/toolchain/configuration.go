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
// that can't be read is refused, as the unit key refuses it, so no key reads as one whose configs it didn't read.
func ConfigurationsByContent(report string) (string, error) {
	var lines []string
	for _, line := range strings.Split(report, "\n") {
		path, found := strings.CutPrefix(strings.TrimSpace(line), configurationLine)
		if !found {
			lines = append(lines, line)
			continue
		}
		configuration, err := ConfigurationFile(path)
		if err != nil {
			return "", fmt.Errorf("reading clang's config %s: %w", path, err)
		}
		lines = append(lines, configuration...)
	}
	return strings.Join(lines, "\n"), nil
}

// ConfigurationFile is a clang config file at path as a key holds it, with no path in it: `Configuration file: sha256
// <digest>` for its bytes, then a line for each file it names, in its order. A config it includes (@file, from the
// including config's directory) is its own `Configuration file:` line and its names, after the including one's; any
// other file it names by an absolute path, after clang's <CFGDIR> is spelled out (lld's binary, by --ld-path= or
// -fuse-ld=), is `Configuration names: sha256 <digest>`. A directory a config names (directoryFlag) is taken only
// inside the compiler's own installation (insideInstallation), and refused by name anywhere else until the directory is
// keyed by its archive's digest (#5xsd2ha's wave 2b). A name that's no file (-fuse-ld=lld) is found by clang where the
// key already reads it.
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
	words := configurationWords(string(contents))
	for index := 0; index < len(words); index++ {
		word := strings.ReplaceAll(words[index], "<CFGDIR>", directory)
		if flag, named, found := directoryFlag(word); found {
			// The directory is joined to its flag or the next word.
			if named == "" && index+1 < len(words) {
				index++
				named = strings.ReplaceAll(words[index], "<CFGDIR>", directory)
			}
			if !insideInstallation(path, named) {
				return nil, fmt.Errorf("a clang config naming a directory, %s, isn't keyed until its archive digest is pinned (#5xsd2ha)", flag)
			}
			continue
		}
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

// directoryFlag is the flag a config's word names a directory by, joined to it or with the directory the next word: the
// sysroot (--sysroot=, -isysroot), a prefix clang finds its tools and libraries under (-B, and its alias --prefix), or a
// GCC installation (--gcc-toolchain=). Nothing keys what such a directory holds yet: a sysroot is keyed by its
// archive's digest once wave 2b pins it, so until then a config naming one outside the compiler's own installation is
// refused rather than keyed by its path.
//
// It gives the flag as a refusal names it and the directory joined to it, or "" where the directory is the next word.
func directoryFlag(word string) (flag string, named string, found bool) {
	for _, flag := range []string{"--sysroot", "--gcc-toolchain", "--prefix"} {
		if word == flag {
			return flag + "=", "", true
		}
		if named, joined := strings.CutPrefix(word, flag+"="); joined {
			return flag + "=", named, true
		}
	}
	for _, flag := range []string{"-isysroot", "-B"} {
		if named, joined := strings.CutPrefix(word, flag); joined {
			return flag, named, true
		}
	}
	return "", "", false
}

// insideInstallation says whether directory, named by the config at path, lies inside the compiler's own installation:
// the directory above the bin/ the config sits in, with <CFGDIR> spelled out and `..` resolved. wasi-sdk 27 ships
// bin/clang.cfg and bin/clang++.cfg naming `--sysroot=<CFGDIR>/../share/wasi-sysroot`, its own sysroot. Such a
// directory's bytes enter no key directly: the installation is keyed by its release, its clang's version report, which
// names the release (wasi-sdk's pinned by WASIClang), so the sysroot it ships moves with it. A directory outside, or a
// config that sits in no bin/, is no part of a release a key names.
func insideInstallation(path, directory string) bool {
	bin := filepath.Dir(path)
	if filepath.Base(bin) != "bin" || !filepath.IsAbs(directory) {
		return false
	}
	relative, err := filepath.Rel(filepath.Dir(bin), filepath.Clean(directory))
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
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
