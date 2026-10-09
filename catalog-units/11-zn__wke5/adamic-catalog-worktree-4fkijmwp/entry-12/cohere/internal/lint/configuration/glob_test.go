package configuration

import (
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// referenceMatch is Match as it was before patterns were compiled (#hsd2dfb): braces expanded and both
// sides split on every call. The compiled path must answer exactly as it does.
func referenceMatch(pattern string, path string) bool {
	for _, expanded := range expandBraces(pattern) {
		if matchSegments(splitPath(expanded), splitPath(path)) {
			return true
		}
	}
	return false
}

// globProbePaths are paths each glob form is asked about: a path built from the pattern itself, so a
// pattern meets at least one path it should match, and a fixed set of shapes from the trees cohere checks.
func globProbePaths(patterns []string) []string {
	paths := []string{
		"index.ts", "source/Thing.ts", "source/Thing.tsx", "source/Thing.test.ts", "source/deep/nested/Thing.ts",
		"app/page.tsx", "app/(group)/route/page.tsx", "modules/os/Os.ts", "libraries/structure/source/api/Fetch.ts",
		"source/api/graphql/generated/Types.generated.ts", "workers/email/Template.tsx", "types.d.ts",
		"scripts/build.cjs", "next.config.mjs", "node_modules/package/index.js", ".cache/x.json",
	}
	for _, pattern := range patterns {
		for _, alternative := range expandBraces(pattern) {
			concrete := strings.NewReplacer("**", "a/b", "*", "x", "?", "q").Replace(alternative)
			paths = append(paths, strings.Trim(concrete, "/"))
		}
	}
	return paths
}

/*
 * Every glob form the sets cohere carries use answers as the uncompiled Match did, through both readers.
 *
 * Resolve and ValidateSelectors now read compiled patterns and paths split once. The matching underneath is
 * the same matchSegments, so the risk is in the wiring: an index that names the wrong override, an ignore
 * pattern reported by the wrong name, a pattern compiled from the wrong list. Each set is loaded as a
 * project would extend it, and each path's ignore verdict and override match is compared with what the
 * reference gives pattern by pattern.
 */
func TestCompiledGlobsAnswerAsMatchDid(t *testing.T) {
	t.Parallel()
	forms := 0
	for _, name := range SetNames() {
		directory := writeConfigs(t, map[string]string{"CohereSettings.json": `{"extends": "` + name + `"}`})
		loaded := loadOrFail(t, filepath.Join(directory, "CohereSettings.json"))

		patterns := append([]string{}, loaded.IgnorePatterns...)
		for _, override := range loaded.Overrides {
			patterns = append(patterns, override.Files...)
		}
		forms += len(patterns)

		for _, path := range globProbePaths(patterns) {
			wantIgnoredBy := ""
			for _, pattern := range loaded.IgnorePatterns {
				if referenceMatch(pattern, path) {
					wantIgnoredBy = pattern
					break
				}
			}
			resolved := loaded.Resolve(filepath.Join(directory, path))
			if resolved.Ignored != (wantIgnoredBy != "") || resolved.IgnoredBy != wantIgnoredBy {
				t.Errorf("%s: %q resolved ignored=%v by %q, want by %q", name, path, resolved.Ignored, resolved.IgnoredBy, wantIgnoredBy)
			}
			if resolved.Ignored {
				continue
			}

			var want []string
			for index, override := range loaded.Overrides {
				for _, pattern := range override.Files {
					if referenceMatch(pattern, path) {
						want = append(want, strconv.Itoa(index))
						break
					}
				}
			}
			var got []string
			globs := loaded.globs()
			segments := splitPath(path)
			for index, files := range globs.overrides {
				if matchesAny(files, segments) {
					got = append(got, strconv.Itoa(index))
				}
			}
			if strings.Join(got, ",") != strings.Join(want, ",") {
				t.Errorf("%s: %q matched overrides %v, want %v", name, path, got, want)
			}
			for _, pattern := range patterns {
				if Match(pattern, path) != referenceMatch(pattern, path) {
					t.Errorf("Match(%q, %q) disagrees with the reference", pattern, path)
				}
			}
		}
	}
	// The control: the sets carry glob forms at all, so the loop above compared something.
	if forms == 0 {
		t.Fatal("no set carries an ignore pattern or an override, so nothing was compared")
	}
}

// A brace form and a doubled star are the two shapes most easily compiled wrong, so they are pinned on
// their own beside the sets' forms, through MatchAny's single split.
func TestCompiledGlobFormsMatchAsBefore(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		pattern string
		paths   []string
	}{
		{"**/generated/**/*.{ts,tsx}", []string{"a/generated/b/c.ts", "generated/c.tsx", "a/generated/c.js", "generated.ts"}},
		{"modules/*", []string{"modules/os", "modules/os/Os.ts", "modules"}},
		{"node_modules/**", []string{"node_modules", "node_modules/a/b.js", "x/node_modules"}},
		{"{a,{b,c}}/x.ts", []string{"a/x.ts", "b/x.ts", "c/x.ts", "d/x.ts"}},
		{"*.d.ts", []string{"types.d.ts", "a/types.d.ts", "d.ts"}},
		{"?.ts", []string{"a.ts", "ab.ts"}},
	} {
		for _, path := range testCase.paths {
			want := referenceMatch(testCase.pattern, path)
			if got := MatchAny([]string{"never/**", testCase.pattern}, path); got != want {
				t.Errorf("MatchAny(%q, %q) = %v, want %v", testCase.pattern, path, got, want)
			}
		}
	}
}
