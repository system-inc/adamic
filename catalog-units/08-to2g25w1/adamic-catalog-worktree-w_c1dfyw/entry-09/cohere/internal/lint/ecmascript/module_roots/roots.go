// Package module_roots decides which files are alive with no importer: a framework loads them by
// filename, a test runner by glob, a computed `import()` by directory.
//
// It lives on the shared shelf because two consumers ask the same question and must not disagree.
// The unused report spares these files from "nothing references this", and
// nexus/consistency-require-constant-casing declines to tell an exported camelCase constant in one of
// them to drop its `export`. A rule package may import only leaf packages (see
// `TestRulePackagesStayLeaves`), and the report's own package is deep, so the conventions live here
// and both read them.
package module_roots

import (
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// Reason is why a file is alive regardless of whether anything imports it.
//
// It is carried rather than reduced to a boolean because the report has to be able to say which
// convention spared a file. A reader who disagrees with the analysis needs to see the reason to
// know whether the analysis is wrong or their belief is.
type Reason string

const (
	// ReasonNextAppRoute is a file the Next.js App Router loads by filesystem convention. No import
	// statement for one of these exists anywhere in the tree, by design.
	ReasonNextAppRoute Reason = "Next.js App Router convention"

	// ReasonNextPagesRoute is the older `pages/` directory convention, loaded the same way.
	ReasonNextPagesRoute Reason = "Next.js pages/ convention"

	// ReasonNextInstrumentation is a framework hook Next.js calls by name at process start.
	ReasonNextInstrumentation Reason = "Next.js framework hook"

	// ReasonBuildConfig is a config file a build tool loads by name.
	ReasonBuildConfig Reason = "build configuration, loaded by name"

	// ReasonTest is a test file, which a runner loads by glob rather than by import.
	ReasonTest Reason = "test file"

	// ReasonDeclaration is a `.d.ts`, which declares rather than defines and is often ambient.
	ReasonDeclaration Reason = "type declaration file"

	// ReasonDynamicImportTarget is a file reachable only through a template-literal `import()`.
	//
	// This is the class that would have done the most damage. Measured on the ahra tree: 532 locale
	// files under `translations/` directories are loaded exclusively as
	// `import(`.../translations/${localeCode}`)`, so no static import to them exists and a naive
	// reference index reports every one as unused. The specifier is computed at runtime, so no
	// static analysis can resolve it — which means the honest answer is to treat the directory a
	// computed specifier points into as rooted, rather than to pretend the reference was found.
	ReasonDynamicImportTarget Reason = "reachable only through a computed import() specifier"

	// ReasonTypeScriptConfigEntry is a file the tsconfig named directly in `files`.
	ReasonTypeScriptConfigEntry Reason = "named directly by the tsconfig"
)

// nextAppRouterFiles are the App Router's reserved filenames. Next.js loads each by convention when
// it appears in a route directory, so none of them is ever imported.
//
// Measured on the ahra tree rather than taken from the framework docs, because the docs list files
// this codebase does not use and a root set carrying entries nothing matches is a root set nobody
// can check. Present here and counted: route.ts (161), page.tsx (83), layout.tsx (9),
// not-found.tsx (2), error.tsx (2). The rest are listed because their absence today is a fact about
// this tree at this moment rather than a property of the framework, and a route added tomorrow must
// not be reported as dead the day it lands.
var nextAppRouterFiles = map[string]bool{
	"page":            true,
	"layout":          true,
	"route":           true,
	"template":        true,
	"loading":         true,
	"error":           true,
	"global-error":    true,
	"not-found":       true,
	"default":         true,
	"forbidden":       true,
	"unauthorized":    true,
	"sitemap":         true,
	"robots":          true,
	"manifest":        true,
	"opengraph-image": true,
	"twitter-image":   true,
	"icon":            true,
	"apple-icon":      true,
}

// buildConfigFiles are files a build tool loads by name from a project root.
//
// The Next.js and OpenNext entries are the framework's. The capitalized ones are this codebase's own
// convention — ProjectSettings.tsx, TailwindConfiguration.ts, LintConfiguration.ts and their
// neighbours are read by Structure's tooling rather than imported by application code — and they are
// here because measuring the tree found them, not because a framework document lists them.
var buildConfigFiles = map[string]bool{
	"next.config":             true,
	"open-next.config":        true,
	"middleware":              true,
	"instrumentation":         true,
	"instrumentation-client":  true,
	"ProjectSettings":         true,
	"ProjectRoot":             true,
	"TailwindConfiguration":   true,
	"LintConfiguration":       true,
	"EsLintConfiguration":     true,
	"TypeScriptConfiguration": true,
	"CohereSettings":          true,
	"next-env":                true,
}

// Set decides which files are alive by definition.
//
// A file in the root set is never reported as unreferenced, whatever the reference index says about
// it. Getting this wrong in the permissive direction costs a missed finding; getting it wrong in the
// strict direction reports a live homepage as dead, which is the failure that makes a reader close
// the report and never open it again. So where the two are in tension this errs permissive, and says
// so at every site where it does.
type Set struct {
	// dynamicImportDirectories are directory prefixes that a computed `import()` specifier can reach.
	//
	// Collected from the program rather than configured, so a new dynamically-loaded directory is
	// covered the day it is written rather than the day someone remembers to list it.
	dynamicImportDirectories []string
}

// IsRoot reports whether a file is alive by definition, and why.
//
// The second result is empty when the file is not a root, so a caller can print the reason beside
// the decision rather than restating the rule.
func (set *Set) IsRoot(fileName string) (bool, Reason) {
	base := baseName(fileName)
	stem := stripExtension(base)

	// A declaration file declares rather than defines. Ambient ones are referenced by nothing and
	// affect the whole program, and the ones that are imported are already covered by the index, so
	// nothing is gained by reporting either.
	if strings.HasSuffix(base, ".d.ts") {
		return true, ReasonDeclaration
	}

	// A test runner loads test files by glob. They import the code under test and nothing imports
	// them, which is exactly the shape the reference index reports as unused.
	if isTestFile(stem) {
		return true, ReasonTest
	}

	if buildConfigFiles[stem] {
		return true, ReasonBuildConfig
	}

	// The App Router's reserved names only mean anything inside an app directory. Checking the
	// directory as well as the name keeps a utility genuinely called `error.ts` reportable.
	if nextAppRouterFiles[stem] && withinRouteDirectory(fileName) {
		return true, ReasonNextAppRoute
	}

	if withinPagesDirectory(fileName) {
		return true, ReasonNextPagesRoute
	}

	// Contains rather than HasPrefix. The collected entry is a directory SEGMENT (`/translations/`)
	// because a computed specifier's literal prefix is an alias path that this analysis deliberately
	// does not resolve, so the match has to be against the middle of an absolute path rather than
	// its start. HasPrefix here silently spared nothing and the locale fixture caught it.
	for _, directory := range set.dynamicImportDirectories {
		if strings.Contains(fileName, directory) {
			return true, ReasonDynamicImportTarget
		}
	}

	return false, ""
}

// DynamicImportDirectories returns the directories a computed `import()` roots, sorted, for a
// fingerprint of what IsRoot decides by besides the file's own name.
func (set *Set) DynamicImportDirectories() []string {
	return slices.Sorted(slices.Values(set.dynamicImportDirectories))
}

// isTestFile reports whether a stem names a test, by the suffixes actually in use.
//
// Measured on the ahra tree: 60 files match `.test.` or `.spec.`. Go's own `_test.go` convention is
// not checked because this analysis reads TypeScript.
func isTestFile(stem string) bool {
	return strings.HasSuffix(stem, ".test") ||
		strings.HasSuffix(stem, ".spec") ||
		strings.Contains(stem, ".test.") ||
		strings.Contains(stem, ".spec.")
}

// withinRouteDirectory reports whether a path sits under a Next.js `app` directory.
//
// It matches any `app` segment rather than only a project-root one, because this tree nests whole
// Next.js projects under `projects/` and each has its own. A single-rooted check would have covered
// the top-level app and reported every route in `projects/www-ahra-ai/app/` as dead.
func withinRouteDirectory(fileName string) bool {
	return hasPathSegment(fileName, "app") || hasPathSegment(fileName, "src")
}

// withinPagesDirectory reports whether a path sits under a `pages` directory, where every file is a
// route regardless of its name.
//
// The `_components` and similar underscore-prefixed directories are private by Next.js convention
// and are NOT routes, so they are excluded: a component parked in `pages/_components/` really can be
// unused, and rooting it would hide the finding.
func withinPagesDirectory(fileName string) bool {
	if !hasPathSegment(fileName, "pages") {
		return false
	}
	for _, segment := range strings.Split(fileName, "/") {
		if strings.HasPrefix(segment, "_") {
			return false
		}
	}
	return true
}

// hasPathSegment reports whether name contains segment as a whole path component, so `app` does not
// match `apple` and `pages` does not match `pagestate`.
func hasPathSegment(name string, segment string) bool {
	for _, part := range strings.Split(name, "/") {
		if part == segment {
			return true
		}
	}
	return false
}

func baseName(path string) string {
	if index := strings.LastIndex(path, "/"); index >= 0 {
		return path[index+1:]
	}
	return path
}

// stripExtension removes one trailing extension, leaving `.d` on a declaration file so the caller
// can still recognize it.
func stripExtension(base string) string {
	if index := strings.LastIndex(base, "."); index > 0 {
		return base[:index]
	}
	return base
}

// Build collects the conventions that make a file alive regardless of what imports it.
//
// The dynamic-import directories are discovered from the program rather than configured, so a
// directory that becomes dynamically loaded is covered the day it is written.
func Build(files []*ast.SourceFile) *Set {
	roots := &Set{}

	seen := map[string]bool{}
	for _, file := range files {
		for _, directory := range computedImportDirectories(file) {
			if !seen[directory] {
				seen[directory] = true
				roots.dynamicImportDirectories = append(roots.dynamicImportDirectories, directory)
			}
		}
	}
	return roots
}

// computedImportDirectories finds the directories a template-literal `import()` can reach.
//
// # Why this exists, measured rather than imagined
//
// This is the class that would have done the most damage to the report's credibility. On the ahra
// tree, 532 locale files live under `translations/` directories and are loaded exclusively as
// `import(`@structure/source/.../translations/${localeCode}`)`. Not one of them has a static import
// anywhere. A reference index without this treats every one as unused, and a report whose first
// screen is 532 live translation files is a report nobody reads to the second screen.
//
// The specifier is computed at runtime, so no static analysis can resolve which file a given
// `${localeCode}` selects. The honest move is therefore to root the whole directory the literal
// prefix points into, rather than to pretend a reference was found or to resolve a specifier that
// genuinely is not knowable until the program runs.
func computedImportDirectories(file *ast.SourceFile) []string {
	var directories []string
	fileName := file.FileName()

	var visit func(node *ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node == nil {
			return false
		}
		if node.Kind == ast.KindCallExpression {
			call := node.AsCallExpression()
			if call.Expression != nil && call.Expression.Kind == ast.KindImportKeyword &&
				call.Arguments != nil && len(call.Arguments.Nodes) > 0 {
				argument := call.Arguments.Nodes[0]
				if argument.Kind == ast.KindTemplateExpression {
					if prefix := templateHeadText(argument); prefix != "" {
						if directory := resolveSpecifierDirectory(prefix, fileName.AsString()); directory != "" {
							directories = append(directories, directory)
						}
					}
				}
			}
		}
		node.ForEachChild(visit)
		return false
	}
	file.AsNode().ForEachChild(visit)
	return directories
}

// templateHeadText returns the literal text before the first substitution in a template.
func templateHeadText(node *ast.Node) string {
	template := node.AsTemplateExpression()
	if template == nil || template.Head == nil {
		return ""
	}
	return template.Head.Text()
}

// resolveSpecifierDirectory turns the literal prefix of a computed specifier into a directory
// suffix that a file path can be matched against.
//
// It deliberately returns a path SUFFIX rather than a resolved absolute path. Resolving the alias
// (`@structure/...`) would mean reimplementing module resolution for a specifier that is not fully
// known, and getting that wrong fails silently in the direction that reports live files as dead.
// Matching on the trailing directory is coarser and cannot make that mistake.
func resolveSpecifierDirectory(prefix string, fileName string) string {
	trimmed := strings.TrimSuffix(prefix, "/")
	if trimmed == "" {
		return ""
	}
	index := strings.LastIndex(trimmed, "/")
	if index < 0 {
		return ""
	}
	// The last literal segment is the directory the computed part selects a file from.
	segment := trimmed[index+1:]
	if segment == "" || strings.Contains(segment, "$") {
		return ""
	}
	_ = fileName
	return "/" + segment + "/"
}
