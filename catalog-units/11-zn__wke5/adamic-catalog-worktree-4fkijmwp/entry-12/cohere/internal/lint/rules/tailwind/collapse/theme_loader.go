// Building a Theme from a repository's stylesheets: `@import` resolution and `@theme` ingestion.
//
// Ported from the `@theme` and `@import` handling in `src/index.ts` and `src/at-import.ts` at
// Tailwind 4.3.3, read at the pinned tag rather than from the minified bundle.
//
// # Where the boundary of this file is, and why it is here rather than further out
//
// `theme.go` is the store. It answers lookups and knows nothing about CSS. This file is the only
// thing that fills it, and it exists because a Theme that cannot be built from a repository proves
// nothing: the exit criterion for this component is a resolved namespace map diffed against the
// engine's own, on two repositories, and there is no way to reach that without walking the real
// `@import` graph.
//
// It is deliberately narrower than upstream's `index.ts`, which is 867 lines because it also builds
// utilities, variants, and the compiled output. What is here is the path from a stylesheet on disk
// to a filled Theme, and nothing else. The pieces that belong to sibling tasks are named in
// "Not handled" below rather than stubbed, so a reader can see the edge of this component instead
// of discovering it.
//
// # Not handled, and what each would take
//
//   - `@import` modifiers other than a bare specifier: `theme(...)`, `prefix(...)`, `reference`,
//     `layer(...)`, `supports(...)`, and media-query conditions. Upstream turns each into a wrapper
//     at-rule around the imported tree. `source(none)`, which both corpus repositories use, is
//     accepted and ignored because it selects content files rather than affecting the theme. Any
//     other modifier is an error rather than a silent drop; see resolveImports.
//   - `@config` and `@plugin`. Both load JavaScript, and cohere runs none at lint time. A
//     JavaScript config that defined theme values would land in the theme upstream and not here,
//     which is a real gap and is why loadStylesheetTheme returns the directive it skipped rather
//     than swallowing it.
//   - `@utility`, `@custom-variant`, `@variant`, `@apply`, `@tailwind`. These are sibling tasks
//     (#4f04x54, #1cahbv8) and none of them contributes a theme entry.
//   - Keyframes collected from inside `@theme`. See the note in theme.go: cohere emits no CSS.
package tailwind

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/system-inc/cohere/internal/lint/ecmascript/text"
)

// StylesheetResolver maps an `@import` specifier to a file path.
//
// base is the directory of the importing file, so a relative specifier resolves against the file it
// appeared in rather than against the process's working directory. Returning an error stops the
// load; a resolver that wants to ignore an unresolvable import must return a path to an empty file
// rather than an error, so that "ignored" is a decision the caller made rather than a failure the
// loader swallowed.
type StylesheetResolver func(specifier, base string) (path string, err error)

// NodeStylesheetResolver resolves `@import` the way a bundler does, given the root of an installed
// tailwindcss package.
//
// `tailwindcss` and `tailwindcss/theme` name files inside that package; everything else is relative
// to the importing file. A specifier with no `.css` extension gets one, matching the engine, which
// appends no extension itself and relies on the host's resolver to do it.
func NodeStylesheetResolver(tailwindPackageRoot string) StylesheetResolver {
	return func(specifier, base string) (string, error) {
		var path string
		switch {
		case specifier == "tailwindcss":
			path = filepath.Join(tailwindPackageRoot, "index.css")
		case strings.HasPrefix(specifier, "tailwindcss/"):
			path = filepath.Join(tailwindPackageRoot, strings.TrimPrefix(specifier, "tailwindcss/"))
		default:
			path = filepath.Join(base, specifier)
		}
		if !strings.HasSuffix(path, ".css") {
			path += ".css"
		}
		return path, nil
	}
}

// SkippedDirective records a directive the loader recognized and did not act on.
//
// This exists so that a gap is reported rather than inferred. `@config` can, upstream, contribute
// theme values through a JavaScript config, and a loader that silently ignored it would produce a
// theme that is quietly short by however many keys that config declared. Callers that care can
// assert the list is empty; the test in theme_test.go asserts exactly which directives the corpus
// produces, so a repository that starts using one is a failing test rather than a wrong answer.
type SkippedDirective struct {
	// Name is the at-rule name, including its `@`.
	Name string
	// Params is the at-rule's parameters.
	Params string
	// Path is the stylesheet it appeared in.
	Path string
}

// LoadThemeFromFile builds a Theme from a stylesheet on disk and its whole `@import` graph.
//
// The returned SkippedDirective slice names every `@config` and `@plugin` encountered. See its doc
// comment for why they are reported rather than ignored.
func LoadThemeFromFile(path string, resolve StylesheetResolver) (*Theme, []SkippedDirective, error) {
	loader := &themeLoader{
		theme:    NewTheme(),
		resolve:  resolve,
		visiting: make(map[string]bool),
	}
	if err := loader.loadFile(path); err != nil {
		return nil, nil, err
	}
	return loader.theme, loader.skipped, nil
}

// themeLoader carries the state of one load: the theme being filled, the resolver, and the cycle
// guard.
type themeLoader struct {
	theme   *Theme
	resolve StylesheetResolver
	skipped []SkippedDirective

	// visiting is the set of files currently on the import stack, so a cycle is an error rather
	// than a stack overflow. It is not a "seen" set: a file imported twice from different places is
	// processed twice, which matches the engine, and matters because the second pass can overwrite
	// values the first one set.
	visiting map[string]bool
}

func (loader *themeLoader) loadFile(path string) error {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolve %s: %w", path, err)
	}

	if loader.visiting[absolutePath] {
		return fmt.Errorf("circular @import: %s imports itself", absolutePath)
	}
	loader.visiting[absolutePath] = true
	defer delete(loader.visiting, absolutePath)

	content, err := os.ReadFile(absolutePath)
	if err != nil {
		return fmt.Errorf("read %s: %w", absolutePath, err)
	}

	nodes, err := ParseCSS(string(content))
	if err != nil {
		return fmt.Errorf("parse %s: %w", absolutePath, err)
	}

	return loader.ingest(nodes, absolutePath)
}

// ingest walks a parsed stylesheet, following imports and recording theme entries.
//
// The walk is explicit rather than a call to Walk because `@theme` needs its subtree handled by
// ingestThemeBlock rather than descended into, and because an `@import` has to be loaded at the
// point it appears: a theme is order-sensitive, and an imported file's entries must land between
// the entries written before and after the `@import` line. Walk's visitor cannot express "load
// another file here, then continue" without the loader reentering itself mid-traversal, which is
// what this does directly and legibly.
func (loader *themeLoader) ingest(nodes []*Node, path string) error {
	for _, node := range nodes {
		switch {
		case node.Kind == KindAtRule && node.Name == "@import":
			if err := loader.followImport(node, path); err != nil {
				return err
			}

		case node.Kind == KindAtRule && node.Name == "@theme":
			if err := loader.ingestThemeBlock(node, path); err != nil {
				return err
			}

		case node.Kind == KindAtRule && (node.Name == "@config" || node.Name == "@plugin"):
			loader.skipped = append(loader.skipped, SkippedDirective{
				Name:   node.Name,
				Params: node.Params,
				Path:   path,
			})

		case node.IsContainer():
			// `@layer theme { @theme default { ... } }` is how the framework's own index.css is
			// written, so descending into container at-rules is required rather than defensive.
			if err := loader.ingest(node.Nodes, path); err != nil {
				return err
			}
		}
	}
	return nil
}

// followImport resolves one `@import` and loads the file it names.
//
// The specifier is the first space-separated segment of the params, with its quotes stripped.
// Anything after it is a modifier, and every modifier other than `source(...)` changes what the
// imported file means, so an unrecognized one is an error rather than a silently dropped qualifier.
// That choice is the opposite of permissive on purpose: a `theme(reference)` import quietly treated
// as a plain one produces a theme whose entries all lack ThemeOptionReference, which is a wrong
// answer that looks exactly like a right one.
func (loader *themeLoader) followImport(node *Node, path string) error {
	parts := segment(text.TrimWhitespace(node.Params), ' ')
	if len(parts) == 0 || parts[0] == "" {
		return fmt.Errorf("%s: @import with no specifier", path)
	}

	specifier := strings.Trim(parts[0], `"'`)

	for _, modifier := range parts[1:] {
		modifier = text.TrimWhitespace(modifier)
		if modifier == "" {
			continue
		}
		// `source(...)` selects the files Tailwind scans for class names. It contributes no theme
		// entry, so it is accepted and ignored rather than refused.
		if strings.HasPrefix(modifier, "source(") {
			continue
		}
		return fmt.Errorf(
			"%s: unsupported @import modifier %q on %q; see the boundary note in themeloader.go",
			path, modifier, specifier,
		)
	}

	resolved, err := loader.resolve(specifier, filepath.Dir(path))
	if err != nil {
		return fmt.Errorf("%s: resolve @import %q: %w", path, specifier, err)
	}
	return loader.loadFile(resolved)
}

// ingestThemeBlock records every custom property in one `@theme` block.
//
// Upstream walks the block's whole subtree rather than only its direct children, so a declaration
// nested inside a rule within `@theme` is still recorded. This uses Walk for that, skipping
// `@keyframes` subtrees the way upstream does, because a `@keyframes` body is full of declarations
// that are not theme entries.
//
// Upstream throws when the block holds anything other than a custom property or `@keyframes`. That
// is reproduced: a stylesheet the engine rejects must not quietly produce a theme here, since any
// theme built from input the engine refuses is one no repository can actually be using.
func (loader *themeLoader) ingestThemeBlock(node *Node, path string) error {
	options, prefix := parseThemeOptions(node.Params)

	if prefix != "" {
		if !isValidThemePrefix(prefix) {
			return fmt.Errorf(
				"%s: the prefix %q is invalid. Prefixes must be lowercase ASCII letters (a-z) only",
				path, prefix,
			)
		}
		loader.theme.Prefix = prefix
	}

	var walkErr error
	Walk(node.Nodes, func(child *Node) WalkAction {
		if child.Kind == KindAtRule && child.Name == "@keyframes" {
			// Upstream collects these to re-emit; cohere emits no CSS. See theme.go.
			return WalkSkip
		}

		if child.Kind == KindComment {
			return WalkContinue
		}

		if child.Kind == KindDeclaration && strings.HasPrefix(child.Property, "--") {
			// The property is unescaped on the way in, so `--color-a\/b` is stored as
			// `--color-a/b`. GetOptions and MarkUsedVariable unescape their arguments to match.
			if err := loader.theme.Add(unescapeCSSIdentifier(child.Property), child.Value, options); err != nil {
				walkErr = fmt.Errorf("%s: %w", path, err)
				return WalkStop
			}
			return WalkContinue
		}

		// A rule inside `@theme` is not itself an error upstream; its children are walked and each
		// is judged on its own. Only a non-custom-property declaration or a comment-free foreign
		// node is rejected.
		if child.IsContainer() {
			return WalkContinue
		}

		walkErr = fmt.Errorf(
			"%s: `@theme` blocks must only contain custom properties or `@keyframes`, found %s %q",
			path, child.Kind, child.Property,
		)
		return WalkStop
	})

	return walkErr
}

// parseThemeOptions reads an `@theme` block's params, and is upstream's `parseThemeOptions`.
//
// The options are a space-separated set of bare words plus an optional `prefix(ident)`. An
// unrecognized word is ignored, matching upstream, which tests for each known option and falls
// through silently otherwise.
func parseThemeOptions(params string) (ThemeOptions, string) {
	options := ThemeOptionNone
	prefix := ""

	for _, option := range segment(text.TrimWhitespace(params), ' ') {
		switch {
		case option == "reference":
			options |= ThemeOptionReference
		case option == "inline":
			options |= ThemeOptionInline
		case option == "default":
			options |= ThemeOptionDefault
		case option == "static":
			options |= ThemeOptionStatic
		case strings.HasPrefix(option, "prefix(") && strings.HasSuffix(option, ")"):
			prefix = option[len("prefix(") : len(option)-1]
		}
	}

	return options, prefix
}

// isValidThemePrefix is upstream's `IS_VALID_PREFIX`, `/^[a-z]+$/`.
func isValidThemePrefix(prefix string) bool {
	if prefix == "" {
		return false
	}
	for index := 0; index < len(prefix); index++ {
		if prefix[index] < 'a' || prefix[index] > 'z' {
			return false
		}
	}
	return true
}
