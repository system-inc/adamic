package main

import (
	"path/filepath"
	"strings"
	"testing"
)

/*
 * `nexus/import-require-path-alias` must not go dark because of where the config was written.
 *
 * The rule took its repositoryRoot as written. ahra committed `/Users/kirkouimet/Projects/ahra`, and
 * in every worktree and on every other machine no checked file was under it, so the rule declined
 * every file and the run printed the same thing a tree with no deep imports prints. A relative root
 * was worse: it never matched an absolute file name at all.
 *
 * These run the real binary from a directory below the project, so a root resolved against the
 * process's working directory would land in the wrong place and the finding would vanish.
 */

const pathAliasImporter = "import { thing } from '../../foundation/Thing';\n\nexport const used = thing;\n"

func pathAliasProject(t *testing.T, ruleOptions string) string {
	t.Helper()
	root := t.TempDir()
	writeTree(t, root, map[string]string{
		"tsconfig.json":               fixScopeTsconfig,
		"CohereSettings.json":         `{"rules":{"nexus/import-require-path-alias":["error",` + ruleOptions + `]}}`,
		"foundation/Thing.ts":         "export const thing = 1;\n",
		"source/features/Importer.ts": pathAliasImporter,
	})
	return root
}

func TestPathAliasRootIsAnchoredToTheConfig(t *testing.T) {
	t.Parallel()
	binary := buildCohere(t)
	const aliases = `"aliases":[{"directory":".","alias":"@project"}]`

	for name, options := range map[string]string{
		"a relative root, run from below the project": `{` + aliases + `,"repositoryRoot":"."}`,
		"an absent root defaults to the project":      `{` + aliases + `}`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			root := pathAliasProject(t, options)
			output, _ := runCohere(t, binary, filepath.Join(root, "source", "features"), "--no-fix", "--lint")
			if !strings.Contains(output, "import-require-path-alias") || !strings.Contains(output, "@project/foundation/Thing") {
				t.Errorf("the rule reported nothing on an import that climbs two levels, so it ran inert:\n%s", output)
			}
		})
	}

	// Never silently inert: a root that cannot hold any checked file stops the run and names itself.
	t.Run("an impossible root fails loudly", func(t *testing.T) {
		t.Parallel()
		root := pathAliasProject(t, `{`+aliases+`,"repositoryRoot":"./does-not-exist"}`)
		output, code := runCohere(t, binary, root, "--no-fix", "--lint")
		if code == 0 {
			t.Fatalf("a run with an impossible repositoryRoot exited 0:\n%s", output)
		}
		if !strings.Contains(output, "repositoryRoot") || !strings.Contains(output, "does-not-exist") {
			t.Errorf("the failure does not name the option and its value:\n%s", output)
		}
	})
}
