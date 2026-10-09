package structure

import (
	"crypto/sha256"
	"maps"
	"slices"
	"strings"
	"sync"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/jsx"
	"github.com/system-inc/cohere/internal/lint/ecmascript/module"
	"github.com/system-inc/cohere/internal/lint/ecmascript/text"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/types/sourcename"
	"github.com/system-inc/cohere/policy"
)

// boundaryNoProjectThemeValueText is the rule's message, whose wording lives in
// `policy/messages/boundary-no-project-theme-value.json`.
var boundaryNoProjectThemeValueText = policy.MessageOf("structure/boundary-no-project-theme-value", "forbiddenThemeValue")

// messageForbiddenThemeValue names the value, the component, and what is allowed instead.
//
// The allowed list is the whole finding. "This value is not allowed" sends the author hunting for
// the theme file; naming the values it can be is the edit.
func messageForbiddenThemeValue(component string, propertyName string, value string, allowed []string) rule.Message {
	return rule.Message{
		Id: boundaryNoProjectThemeValueText.Id,
		Description: boundaryNoProjectThemeValueText.Render(map[string]string{
			"component":    component,
			"propertyName": propertyName,
			"value":        value,
			"allowed":      strings.Join(allowed, ", "),
		}),
	}
}

// themePropertySuffixes maps a JSX property name to the interface suffix that defines its values:
// a `variant` property is checked against `ButtonVariantsInterface`.
var themePropertySuffixes = map[string]string{
	"variant":  "Variants",
	"size":     "Sizes",
	"kind":     "Kinds",
	"position": "Positions",
	"side":     "Sides",
}

// componentAliases names components that render another component's theme.
//
// `AnimatedButton` has no `AnimatedButtonVariantsInterface`; it wraps `Button` and takes Button's
// values. Without this the rule finds no theme for it and declines, which is silent rather than
// wrong but leaves the component unguarded.
var componentAliases = map[string]string{
	"AnimatedButton": "Button",
	"TipButton":      "Button",
	"Tip":            "Popover",
}

// interfaceNameSuffix is the trailing word every theme interface in the tree carries.
const interfaceNameSuffix = "Interface"

// BoundaryNoProjectThemeValue reports a theme value used in library code that the library does not
// define.
//
//	valid:   <Button variant="Ghost" />        inside libraries/structure, and Ghost is in the theme
//	valid:   <Button variant="Emphasized" />   outside libraries/structure, where projects extend
//	invalid: <Button variant="Emphasized" />   inside libraries/structure
//
// Ported from `structure/boundary-no-project-theme-value`.
//
// # The cache is built from the type graph rather than from the filesystem
//
// This is a deliberate divergence from the original and the only one in this rule.
//
// The original discovers its theme files by walking `process.cwd() + "libraries/structure/source"`
// with `readdirSync`, parsing each match with a second `@typescript-eslint/parser` instance, and
// caching the result in a module-level variable. It does that because ESLint hands a rule one file
// at a time and gives it no way to see the others. Cohere has the whole program, so this reads the
// same files out of `ctx.Program` instead.
//
// Two defects are avoided rather than ported, and both are ones this codebase has already paid for:
//
//   - `process.cwd()` makes the verdict depend on where the linter was invoked from. That is the
//     invisible dependency `import-require-path-alias` had, fixed there by taking the root as a
//     required option rather than reading it from the process.
//   - The original wraps its parse in `catch { return {} }`, so an unparseable theme file yields an
//     empty cache and the rule goes quiet. Quiet is indistinguishable from clean, which is the
//     failure this whole tool exists to remove.
//
// **The two mechanisms were measured against each other rather than assumed equivalent.** The
// filesystem walk finds 28 files; the program holds 29 whose names end in `Theme.ts`. The extra one
// is `source/theme/hooks/useTheme.ts`, which the original's `^[A-Z]\w*Theme\.ts$` pattern correctly
// rejects for its lowercase first letter. It exports one hook and no theme interface, so it
// contributes nothing either way and the two mechanisms see the same 28 contributing files. The
// pattern is reproduced here anyway, so the difference stays at zero rather than at "currently
// harmless".
//
// # Only inside the library
//
// A project using its own values is the system working. The boundary is one-directional: the
// library may not reach into the project, and the project may extend the library freely.
var BoundaryNoProjectThemeValue = rule.Rule{
	Name: "structure/boundary-no-project-theme-value",

	// Every theme file in the program.
	ProgramReads: rule.ReadsOtherFiles,
	// A file's verdict reads its own bytes, its own path, and the theme map, so the theme map is the whole
	// fingerprint (#kdee854). See themeFingerprint.
	ProgramFingerprint: themeFingerprint,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil || !FileContextFor(ctx.SourceFile.FileName().AsString()).IsInLibrariesStructure {
			return nil
		}
		if ctx.Program == nil {
			return nil
		}

		themes := themeValuesForProgram(ctx.Program)
		if len(themes) == 0 {
			// No themes found means the rule cannot decide anything, and reporting nothing would be
			// indistinguishable from a clean tree. Declining is the honest answer; the coverage
			// line's "listened to no files" note is what makes the silence visible.
			return nil
		}

		return rule.Listeners{
			ast.KindJsxAttribute: func(node *ast.Node) {
				attribute := node.AsJsxAttribute()
				name := attribute.Name()
				if name == nil || name.Kind != ast.KindIdentifier {
					return
				}
				suffix, isThemeProperty := themePropertySuffixes[name.Text()]
				if !isThemeProperty {
					return
				}

				// A string literal only. `variant={expression}` cannot be checked without evaluating
				// it, and guessing is how a boundary rule starts reporting on correct code.
				if attribute.Initializer == nil || attribute.Initializer.Kind != ast.KindStringLiteral {
					return
				}
				// Decoded as ESLint's parser decodes it, which is the value the ESLint twin reads.
				value := text.UnescapeStringLiteralText(attribute.Initializer.Text())

				componentName := jsxElementNameOf(node)
				if componentName == "" {
					return
				}
				componentTheme, found := themes[componentName]
				if !found {
					if aliased, isAlias := componentAliases[componentName]; isAlias {
						componentTheme, found = themes[aliased]
					}
				}
				if !found {
					return
				}

				allowed := componentTheme[suffix]
				if len(allowed) == 0 {
					return
				}
				for _, allowedValue := range allowed {
					if allowedValue == value {
						return
					}
				}
				ctx.ReportNode(attribute.Initializer,
					messageForbiddenThemeValue(componentName, name.Text(), value, allowed))
			},
		}
	},
}

// jsxElementNameOf returns the component name of the element an attribute belongs to.
//
// Two element kinds carry attributes, and that is a real difference from the original rather than a
// detail. ESTree has one `JSXOpeningElement` for both forms; typescript-go distinguishes a
// self-closing element from an opening one. A port checking a single kind would silently skip every
// `<Button variant="X" />`, which is the dominant form in this codebase.
//
// A qualified name (`<Foo.Bar />`) returns empty. It is not a bare identifier, so it names no
// component in the theme cache, and the original's `JSXIdentifier` check excludes it the same way.
func jsxElementNameOf(attribute *ast.Node) string {
	attributes := attribute.Parent
	if attributes == nil || attributes.Kind != ast.KindJsxAttributes {
		return ""
	}
	element := attributes.Parent
	if element == nil {
		return ""
	}

	// The two-kind split is `jsx.ElementParts`, which answers nil for anything that is not an
	// element and so subsumes the default arm this used to spell out. What stays local is the walk
	// upward from an attribute, which the shelf does not do and which only this rule wants.
	tagName, _ := jsx.ElementParts(element)
	if tagName == nil || tagName.Kind != ast.KindIdentifier {
		return ""
	}
	return tagName.Text()
}

// themeCache holds the theme map for one program, so the scan happens once per run.
//
// Keyed on the program pointer rather than on nothing, which is what makes this safe to reuse across
// trees. The original memoized in a module-level variable with no key, correct for a process that
// lints once and exits and wrong for anything that lints twice: the second tree would read the
// first's themes. Pointer identity gives a miss on a new program for free.
//
// A package variable rather than the file cache, and that is a design decision rather than
// convenience. `rule.FileCache` is per-file by construction and its own comment says sharing an
// entry across files would be a bug, which is why it carries no mutex. Run-scoped state does not
// belong there. It cannot live at registration either: the program is built after rules register,
// so there is nothing to key on at that point.
//
// The mutex sits at the only access point and the key is immutable for the run, so this is safe
// under the parallel walk by construction rather than by discipline.
var themeCache struct {
	sync.Mutex
	program rule.ProgramIdentity
	values  map[string]map[string][]string
}

// themeValuesForProgram returns the theme map for this run, computing it at most once.
//
// The rule reads this on every file and does almost no per-node work: measured at 2329ms of setup
// against 0.54ms of listening across 1,862 files, which was 64.5% of all rule time for a rule that
// reports nothing on this tree. The scan was being redone per file, so the cost was the file count
// rather than the work.
func themeValuesForProgram(program rule.Program) map[string]map[string][]string {
	themeCache.Lock()
	defer themeCache.Unlock()

	if themeCache.program == program.Identity() && themeCache.values != nil {
		return themeCache.values
	}

	values := themeValuesFromProgram(program)
	themeCache.program = program.Identity()
	themeCache.values = values
	return values
}

// themeFingerprint is the rule's program fingerprint: the theme map, every component, suffix and value, in the
// order a message lists them. Components and suffixes are sorted, since a map has no order; values keep theirs,
// since the message names them in it. Nothing else outside a file reaches its verdict, so an edit anywhere but a
// theme interface's exported keys leaves this alone and the file's findings replay (#kdee854). The rule takes no
// options, so none reach it.
func themeFingerprint(program rule.Program, _ any) [sha256.Size]byte {
	themes := themeValuesForProgram(program)
	hash := sha256.New()
	for _, component := range slices.Sorted(maps.Keys(themes)) {
		hash.Write([]byte(component))
		hash.Write([]byte{0})
		for _, suffix := range slices.Sorted(maps.Keys(themes[component])) {
			hash.Write([]byte(suffix))
			hash.Write([]byte{1})
			for _, value := range themes[component][suffix] {
				hash.Write([]byte(value))
				hash.Write([]byte{2})
			}
			hash.Write([]byte{3})
		}
		hash.Write([]byte{4})
	}
	var fingerprint [sha256.Size]byte
	copy(fingerprint[:], hash.Sum(nil))
	return fingerprint
}

// themeValuesFromProgram collects every theme interface's keys, by component and by suffix.
//
// Call `themeValuesForProgram` rather than this: an uncached call rescans every source file in the
// program and this rule runs on 1,862 of them.
func themeValuesFromProgram(program rule.Program) map[string]map[string][]string {
	themes := map[string]map[string][]string{}

	for _, sourceFile := range program.SourceFiles() {
		if sourceFile == nil || !isThemeFileName(sourceFile.FileName().AsString()) {
			continue
		}
		sourceFile.AsNode().ForEachChild(func(statement *ast.Node) bool {
			if statement.Kind != ast.KindInterfaceDeclaration {
				return false
			}
			declaration := statement.AsInterfaceDeclaration()
			// Exported only, matching the original, which looks for an interface inside an
			// ExportNamedDeclaration. A theme interface the file keeps to itself is not the
			// library's published surface.
			if !module.IsExportedByName(declaration.Modifiers()) {
				return false
			}
			name := declaration.Name()
			if name == nil || name.Kind != ast.KindIdentifier {
				return false
			}

			for _, suffix := range themePropertySuffixes {
				componentPrefix, matched := componentPrefixFromInterfaceName(name.Text(), suffix)
				if !matched {
					continue
				}
				keys := interfacePropertyNames(declaration)
				if len(keys) == 0 {
					continue
				}
				// Accumulate rather than replace, which matters because the tree has two
				// `CalendarTheme.ts` files and one declares a suffix the other does not. Creating
				// the inner map only when absent means a second file contributing `Sizes` does not
				// discard the first file's `Variants`.
				//
				// A same-prefix, same-suffix collision still keeps the last file, and that is the
				// original's behavior rather than a shortfall: its `Object.assign(existing,
				// suffixMap)` copies suffix keys onto the existing map, so distinct suffixes
				// accumulate and a repeated one overwrites. Established by executing the original's
				// merge on both shapes, not by reading it, because the reading admits two
				// interpretations and only one is what `Object.assign` does.
				//
				// Worth knowing rather than fixing: two files declaring the same suffix for one
				// component is a duplicate the library should not have, and silently unioning them
				// would hide that instead of reporting it. Whether the duplicate calendars should
				// both exist is a question for the library rather than for its linter.
				if themes[componentPrefix] == nil {
					themes[componentPrefix] = map[string][]string{}
				}
				themes[componentPrefix][suffix] = keys
			}
			return false
		})
	}
	return themes
}

// isThemeFileName reproduces the original's `^[A-Z]\w*Theme\.ts$` on the base name, scoped to the
// library's source directory.
//
// The capital is load-bearing rather than cosmetic: `useTheme.ts` also ends in `Theme.ts` and is a
// hook rather than a theme. It carries no theme interface, so including it would change nothing
// today, which is exactly why the pattern is reproduced rather than relaxed to a suffix test.
func isThemeFileName(fileName string) bool {
	normalizedPath := strings.ReplaceAll(fileName, `\`, "/")
	if !strings.Contains(normalizedPath, "/libraries/structure/source/") {
		return false
	}
	baseName := normalizedPath[strings.LastIndex(normalizedPath, "/")+1:]
	// An Adamic `XTheme.a` is named as the `XTheme.ts` it is (#kwt1htp).
	if !strings.HasSuffix(sourcename.TreatedAs(baseName), "Theme.ts") {
		return false
	}
	if baseName == "" || baseName[0] < 'A' || baseName[0] > 'Z' {
		return false
	}
	// `\w*` between the capital and `Theme`: letters, digits, underscore.
	for index := 1; index < len(sourcename.TreatedAs(baseName))-len("Theme.ts"); index++ {
		character := baseName[index]
		switch {
		case character >= 'a' && character <= 'z':
		case character >= 'A' && character <= 'Z':
		case character >= '0' && character <= '9':
		case character == '_':
		default:
			return false
		}
	}
	return true
}

// componentPrefixFromInterfaceName turns `ButtonVariantsInterface` into `Button` for the `Variants`
// suffix.
//
// Both spellings resolve, with and without the trailing `Interface`. The original's comment records
// why: matching only the bare form left the cache empty, because every theme in the tree uses the
// suffixed spelling, and an empty cache disabled the rule silently.
//
// An empty prefix is rejected, which is what keeps a bare `VariantsInterface` from registering a
// component with no name and swallowing every unmatched component into it.
func componentPrefixFromInterfaceName(interfaceName string, suffix string) (string, bool) {
	remainder := strings.TrimSuffix(interfaceName, interfaceNameSuffix)
	if !strings.HasSuffix(remainder, suffix) {
		return "", false
	}
	componentPrefix := strings.TrimSuffix(remainder, suffix)
	if componentPrefix == "" {
		return "", false
	}
	return componentPrefix, true
}

// interfacePropertyNames lists an interface's property names, in declaration order.
func interfacePropertyNames(declaration *ast.InterfaceDeclaration) []string {
	if declaration.Members == nil {
		return nil
	}
	var names []string
	for _, member := range declaration.Members.Nodes {
		if member.Kind != ast.KindPropertySignature {
			continue
		}
		name := member.AsPropertySignatureDeclaration().Name()
		if name == nil {
			continue
		}
		switch name.Kind {
		case ast.KindIdentifier, ast.KindStringLiteral:
			names = append(names, name.Text())
		}
	}
	return names
}
