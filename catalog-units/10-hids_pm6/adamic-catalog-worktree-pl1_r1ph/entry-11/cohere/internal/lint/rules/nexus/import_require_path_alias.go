package nexus

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/system-inc/cohere/internal/lint/ecmascript/imports"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// PathAlias maps a repository-relative directory to the alias that names it.
type PathAlias struct {
	Directory string `json:"directory"`
	Alias     string `json:"alias"`
}

// ImportRequirePathAliasOptions configures which directories map to which alias.
type ImportRequirePathAliasOptions struct {
	// Aliases are matched longest directory first, so a nested root is not shadowed by its parent.
	Aliases []PathAlias `json:"aliases"`

	// AliasesFromTsconfigPaths adds an alias for every directory the tsconfig's `paths` maps a
	// wildcard onto, beside any written in Aliases.
	//
	// A hand-written list is a second copy of what the tsconfig already says, and the copy is the
	// one that drifts: a project that renames a directory or adds an alias updates the tsconfig
	// because the compiler makes it, and nothing makes it update this list. It is also why the rule
	// stayed in ahra alone. Every Structure project spells the same three aliases in the same
	// tsconfig it extends, so a library can turn the rule on for all of them only by reading that.
	//
	// Off unless asked for, so a config that names no aliases means what it meant before, which is
	// nothing to suggest. An entry written in Aliases is kept beside the derived ones and wins a tie
	// on the same directory.
	AliasesFromTsconfigPaths bool `json:"aliasesFromTsconfigPaths"`

	// StrictRoots are directories where every relative import is reported, not only the deep ones.
	StrictRoots []string `json:"strictRoots"`

	// RepositoryRoot is the directory the aliases are relative to, absolute by the time the rule runs.
	//
	// The TypeScript original reads process.cwd() here. That makes the rule's verdict depend on
	// where the linter was invoked from, which is the same latent defect boundary-no-internal-import
	// carried: correct from the repository root, silently inert anywhere else. Passing the root
	// explicitly makes the dependency visible and testable.
	//
	// In a config it may be relative, resolving against the config file's directory, or absent,
	// defaulting to the project root cohere discovered. See decodeImportRequirePathAliasOptions.
	RepositoryRoot string `json:"repositoryRoot"`
}

// decodeImportRequirePathAliasOptions decodes the options and settles the root before any file runs.
//
// The root used to be taken as written, and a committed absolute path is right on one machine in one
// checkout. In every worktree and on every other machine no file sits under it, so the rule declined
// every file and reported nothing, which is indistinguishable from a tree with no deep imports. ahra
// pinned `/Users/kirkouimet/Projects/ahra` and the rule went dark in each of its worktrees.
//
// So the root is anchored the way a tsconfig anchors its paths: a relative one resolves against the
// directory of the config that wrote it, and an absent one is the project root. Whatever comes out
// must be a directory that exists. A root that names nothing, or that this decoder has no base to
// resolve, is refused by name, because refusing here is the only alternative to the silence.
func decodeImportRequirePathAliasOptions(raw []byte, base rule.OptionsBase) (any, error) {
	decoded, err := rule.DecodeOptionsInto[ImportRequirePathAliasOptions]()(raw)
	if err != nil {
		return nil, err
	}
	settings := decoded.(ImportRequirePathAliasOptions)

	root := settings.RepositoryRoot
	switch {
	case root == "":
		if base.ProjectRoot == "" {
			return nil, fmt.Errorf("repositoryRoot is absent and no project root is known to default it to, " +
				"so the rule could not tell which files its aliases describe. Write repositoryRoot, relative " +
				"to the config file, such as \".\"")
		}
		root = base.ProjectRoot

	case !filepath.IsAbs(root):
		if base.ConfigDirectory == "" {
			return nil, fmt.Errorf("repositoryRoot %q is relative and the config's directory is not known, "+
				"so there is nothing to resolve it against", root)
		}
		root = filepath.Join(base.ConfigDirectory, root)
	}
	root = filepath.Clean(root)

	information, statError := os.Stat(root)
	if statError != nil {
		return nil, fmt.Errorf("repositoryRoot %q resolves to %s, which cannot be read (%v), so no file "+
			"would be inside it and the rule would report nothing", settings.RepositoryRoot, root, statError)
	}
	if !information.IsDir() {
		return nil, fmt.Errorf("repositoryRoot %q resolves to %s, which is a file rather than a directory, "+
			"so no file would be inside it and the rule would report nothing", settings.RepositoryRoot, root)
	}

	// Existing is not enough, and the worktree is why. In an ahra worktree the pinned
	// `/Users/kirkouimet/Projects/ahra` exists, it is just another checkout, so not one file this run
	// checks is under it. A root is useful only if it holds the project or sits inside it; a root
	// beside the project makes the rule inert for every file it is handed.
	if base.ProjectRoot != "" && !pathsNest(root, base.ProjectRoot) {
		return nil, fmt.Errorf("repositoryRoot %q resolves to %s, which neither holds nor sits inside the "+
			"project at %s, so no file checked here would be inside it and the rule would report nothing. "+
			"A path pinned on one machine is another checkout in a worktree; write it relative to the "+
			"config file, such as \".\"", settings.RepositoryRoot, root, base.ProjectRoot)
	}

	settings.RepositoryRoot = root
	return settings, nil
}

// pathsNest reports whether one directory holds the other, either way round, or they are the same.
//
// Compared as written first and then with symlinks resolved, because one side can arrive through a
// link the other did not take (macOS spells its temporary directories both /var and /private/var),
// and refusing a root over a spelling would be a false alarm on a correct config.
func pathsNest(first string, second string) bool {
	nests := func(outer string, inner string) bool {
		relative, err := filepath.Rel(outer, inner)
		return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
	}
	if nests(first, second) || nests(second, first) {
		return true
	}
	resolvedFirst, firstError := filepath.EvalSymlinks(first)
	resolvedSecond, secondError := filepath.EvalSymlinks(second)
	if firstError != nil || secondError != nil {
		return false
	}
	return nests(resolvedFirst, resolvedSecond) || nests(resolvedSecond, resolvedFirst)
}

// importRequirePathAliasUseAliasText is the rule's `useAlias` message,
// whose wording lives in `policy/messages/import-require-path-alias.json`.
var importRequirePathAliasUseAliasText = policy.MessageOf("nexus/import-require-path-alias", "useAlias")

func messageUseAlias(importPath string, levels int, suggestion string) rule.Message {
	return rule.Message{
		Id: "useAlias",
		Description: importRequirePathAliasUseAliasText.Render(map[string]string{
			"importPath": importPath,
			"levels":     strconv.Itoa(levels),
			"suggestion": suggestion,
		}),
	}
}

// importRequirePathAliasUseAliasInStrictRootText is the rule's `useAliasInStrictRoot` message,
// whose wording lives in `policy/messages/import-require-path-alias.json`.
var importRequirePathAliasUseAliasInStrictRootText = policy.MessageOf("nexus/import-require-path-alias", "useAliasInStrictRoot")

func messageUseAliasInStrictRoot(importPath string, root string, suggestion string) rule.Message {
	return rule.Message{
		Id: "useAliasInStrictRoot",
		Description: importRequirePathAliasUseAliasInStrictRootText.Render(map[string]string{
			"importPath": importPath,
			"root":       root,
			"suggestion": suggestion,
		}),
	}
}

// ImportRequirePathAlias requires an alias for imports that leave their own neighbourhood.
//
//	valid:   import { Sibling } from './Sibling'
//	valid:   import { Parent } from '../Parent'
//	valid:   import { Thing } from '@base/source/foundation/Thing'
//	invalid: import { Thing } from '../../../foundation/Thing'
//
// A relative import that climbs two or more levels says how far up to walk rather than where it
// lands, so it stops carrying its own provenance. One level stays relative on purpose: `./Sibling`
// and `../Parent` are facts a reader already has from the file's own location, and spelling them as
// a full alias makes them longer without making them clearer.
//
// A strict root is a directory where even a sibling import must be aliased. A shared library read by
// several projects benefits from one spelling everywhere, so a path copied out of it lands correctly
// wherever it is pasted.
//
// A path holding an `internal` segment is exempt, and the reason is worth keeping: boundary-no-
// internal-import rejects an aliased path into an internal directory outright, because it wants a
// relative path to prove the importer is a neighbour. Reporting here would ask for the thing that
// rule rejects, and two rules that demand opposite edits are worse than either alone.
//
// The fix rewrites the specifier to the suggestion. An alias derived from the tsconfig's `paths`
// resolves by construction; one written in Aliases is exactly as right as the list it came from.
var ImportRequirePathAlias = rule.Rule{
	Name: "nexus/import-require-path-alias",
	// The tsconfig's `paths` and the directory they resolve against, read only when
	// aliasesFromTsconfigPaths asks for them.
	ProgramReads: rule.ReadsCompilerOptions,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		settings, hasSettings := rule.OptionsAs[ImportRequirePathAliasOptions](options)
		if !hasSettings || settings.RepositoryRoot == "" {
			// Without a root nothing can be made repository-relative. Declining is the honest
			// answer, and it keeps a misconfigured rule from guessing.
			return nil
		}
		repositoryRoot := strings.TrimSuffix(normalizedPathText(settings.RepositoryRoot), "/")

		// Written first, so a written alias wins a tie with a derived one on the same directory: the
		// sort below is stable.
		configured := append([]PathAlias{}, settings.Aliases...)
		if settings.AliasesFromTsconfigPaths && ctx.Program != nil {
			configured = append(configured,
				aliasesFromCompilerPaths(ctx.Program.Options(), ctx.Program.GetCurrentDirectory(), repositoryRoot)...)
		}
		if len(configured) == 0 {
			// Nothing to suggest, so nothing to report.
			return nil
		}

		// Longest directory first, so a nested root is not shadowed by its parent.
		aliases := make([]PathAlias, 0, len(configured))
		for _, alias := range configured {
			aliases = append(aliases, PathAlias{
				Directory: normalizeConfiguredDirectory(alias.Directory),
				Alias:     alias.Alias,
			})
		}
		sort.SliceStable(aliases, func(first int, second int) bool {
			return len(aliases[first].Directory) > len(aliases[second].Directory)
		})

		strictRoots := make([]string, 0, len(settings.StrictRoots))
		for _, root := range settings.StrictRoots {
			strictRoots = append(strictRoots, normalizeConfiguredDirectory(root))
		}

		importingFile := imports.NormalizedFileName(ctx.SourceFile)
		importingRelative, isInsideRepository := repositoryRelative(repositoryRoot, importingFile)
		if !isInsideRepository {
			return nil
		}

		check := func(node *ast.Node, specifierNode *ast.Node, importPath string) {
			if !strings.HasPrefix(importPath, ".") {
				return
			}

			resolved := tspath.NormalizePath(
				tspath.ResolvePath(tspath.GetDirectoryPath(importingFile), importPath))
			repositoryPath, isInside := repositoryRelative(repositoryRoot, resolved)
			// A path climbing out of the repository is not ours to name.
			if !isInside {
				return
			}
			// An internal import must stay relative; see the rule comment above.
			if imports.HasPathSegment(repositoryPath, "internal") {
				return
			}

			suggestion, hasAlias := aliasForPath(aliases, repositoryPath)
			if !hasAlias {
				return
			}

			specifierRange := rule.TokenRange(ctx.SourceFile, specifierNode)
			specifierText := ctx.SourceFile.Text()[specifierRange.Pos():specifierRange.End()]
			fix := ctx.ReplaceNode(specifierNode, quotedLike(specifierText, suggestion))

			for _, root := range strictRoots {
				if directoryContains(root, importingRelative) {
					ctx.ReportNodeWithFixes(specifierNode,
						messageUseAliasInStrictRoot(importPath, root, suggestion), fix)
					return
				}
			}

			levels := strings.Count(importPath, "../")
			if levels >= 2 {
				ctx.ReportNodeWithFixes(specifierNode, messageUseAlias(importPath, levels, suggestion), fix)
			}
		}

		return rule.Listeners{
			ast.KindImportDeclaration: func(node *ast.Node) {
				declaration := node.AsImportDeclaration()
				if declaration == nil || declaration.ModuleSpecifier == nil {
					return
				}
				if !ast.IsStringLiteralLike(declaration.ModuleSpecifier) {
					return
				}
				check(node, declaration.ModuleSpecifier, declaration.ModuleSpecifier.Text())
			},

			ast.KindCallExpression: func(node *ast.Node) {
				source, isImport := imports.CallExpressionSource(node)
				if !isImport {
					return
				}
				call := node.AsCallExpression()
				if call == nil || call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
					return
				}
				check(node, call.Arguments.Nodes[0], source)
			},
		}
	},
}

// aliasesFromCompilerPaths reads the aliases a tsconfig's `paths` declares, as repository-relative
// directories.
//
// Only the shape that names a directory becomes an alias: a wildcard key mapped to exactly one
// wildcard target, `"@structure/*": ["../../libraries/structure/*"]`. The rest are left out, each
// for a reason:
//
//   - An exact key such as `"@project/ProjectSettings": ["../../ProjectSettings.tsx"]` names one
//     file, so there is no directory for other imports to be suggested under.
//   - A key or target with its wildcard anywhere but the end, such as `"*-legacy": ["old/*"]`, does
//     not map a directory onto a prefix, so no suggestion could be built by appending the rest.
//   - More than one target is a fallback list: the compiler takes the first that has the file. A file
//     under the second target is reached through the alias only while the first lacks a file of the
//     same name, so suggesting the alias there would be right until someone adds that file, and then
//     silently resolve somewhere else.
//   - A target outside the repository is not one of this repository's directories.
//
// Targets resolve against the directory of the tsconfig that wrote `paths`, which is how the compiler
// reads them. Structure writes its aliases in a config two directories down that every project
// extends, so resolving against the project's own tsconfig instead would put every alias two levels
// above the repository.
func aliasesFromCompilerPaths(options *core.CompilerOptions, currentDirectory string, repositoryRoot string) []PathAlias {
	if options == nil || options.Paths.Size() == 0 {
		return nil
	}
	basePath := options.GetPathsBasePath(tspath.RootedDirectoryPath(currentDirectory))

	var aliases []PathAlias
	for pattern, targets := range options.Paths.Entries() {
		prefix, isWildcard := strings.CutSuffix(pattern, "/*")
		if !isWildcard || prefix == "" || strings.Contains(prefix, "*") || len(targets) != 1 {
			continue
		}
		targetDirectory, targetIsWildcard := strings.CutSuffix(targets[0], "/*")
		if !targetIsWildcard || strings.Contains(targetDirectory, "*") {
			continue
		}

		directory := tspath.NormalizePath(tspath.ResolvePath(normalizedPathText(basePath.AsString()), targetDirectory))
		relative, isInside := repositoryRelative(repositoryRoot, strings.TrimSuffix(directory, "/"))
		if !isInside {
			continue
		}
		if relative == "" {
			relative = "."
		}
		aliases = append(aliases, PathAlias{Directory: relative, Alias: prefix})
	}
	return aliases
}

// aliasForPath returns the aliased spelling of a repository-relative path.
func aliasForPath(aliases []PathAlias, repositoryPath string) (string, bool) {
	for _, alias := range aliases {
		if !directoryContains(alias.Directory, repositoryPath) {
			continue
		}
		rest := repositoryPath
		if alias.Directory != "" {
			rest = strings.TrimPrefix(strings.TrimPrefix(repositoryPath, alias.Directory), "/")
		}
		if rest == "" {
			return alias.Alias, true
		}
		return alias.Alias + "/" + rest, true
	}
	return "", false
}

// quotedLike wraps a replacement specifier in the quote character the original used.
//
// `ReplaceNode` takes the literal's whole token range, quotes included, so the replacement has to
// supply its own. Reading the quote off the source rather than hardcoding one keeps the fix from
// rewriting a file's quote style as a side effect: this codebase writes single quotes and Prettier
// would put them back either way, but a fixer that silently changed a byte nobody asked about is a
// fixer people stop trusting.
//
// Falls back to a single quote when the source text is too short to have quotes at all, which
// cannot happen for a parsed string literal and is handled rather than assumed away.
func quotedLike(originalText string, replacement string) string {
	quote := byte('\'')
	if len(originalText) > 0 && (originalText[0] == '"' || originalText[0] == '\'' || originalText[0] == '`') {
		quote = originalText[0]
	}
	return string(quote) + replacement + string(quote)
}

// normalizeConfiguredDirectory is a configured directory as the matcher wants it: no leading `./`,
// no trailing slash, and the repository root as the empty string.
//
// The root is the case this exists for, and it was a silent defect in both implementations. A
// tsconfig spells the root `"@project/*": ["../../*"]`, so a reader configuring this rule writes
// `{ directory: ".", alias: "@project" }` -- the only spelling that reads like what it means. The
// matcher compares against repository-relative paths like `app/Foo.tsx`, which equal neither `.` nor
// anything starting with `./`, so that entry matched nothing at all. Every deep relative import
// under the root went unreported and the tree read clean.
//
// Measured on Kirk's tree: 127 findings before, 365 after, and the 238 that appeared are exactly
// the `app/` and `modules/` imports a hand count had already found. Normalising here rather than at
// each comparison also means the alias matcher and the strict-root matcher cannot drift, and it
// accepts `./app` and `app/` for the same directory, which a reader has no reason to expect to
// differ.
func normalizeConfiguredDirectory(directory string) string {
	trimmed := directory
	for strings.HasPrefix(trimmed, "./") {
		trimmed = trimmed[len("./"):]
	}
	trimmed = strings.TrimRight(trimmed, "/")
	if trimmed == "." {
		return ""
	}
	return trimmed
}

// directoryContains reports whether a repository-relative path sits inside a normalised directory.
//
// The empty string is the repository root and contains everything, which is what makes a root alias
// work. Every other directory matches on a `/` boundary rather than as a bare prefix, so
// `libraries/structure` does not swallow `libraries/structure-tools`.
func directoryContains(directory string, repositoryPath string) bool {
	if directory == "" {
		return true
	}
	return repositoryPath == directory || strings.HasPrefix(repositoryPath, directory+"/")
}

// repositoryRelative expresses an absolute path relative to the repository root, reporting whether
// it is inside it at all.
func repositoryRelative(repositoryRoot string, path string) (string, bool) {
	if path == repositoryRoot {
		return "", true
	}
	if !strings.HasPrefix(path, repositoryRoot+"/") {
		return "", false
	}
	return strings.TrimPrefix(path, repositoryRoot+"/"), true
}

// normalizedPathText rewrites Windows separators in a plain string.
func normalizedPathText(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}
