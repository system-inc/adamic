package nexus

import (
	"crypto/sha256"
	"regexp"
	"slices"
	"strings"
	"sync"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/imports"
	"github.com/system-inc/cohere/internal/lint/ecmascript/property"
	"github.com/system-inc/cohere/internal/lint/ecmascript/text"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/types/sourcename"
	"github.com/system-inc/cohere/policy"
)

// minimumLengthToFlagAsUntranslated is how long a string must be before matching English is taken
// as evidence it was never translated.
//
// Short strings are genuinely the same in many languages. "OK", "Email", and a bare "%" are correct
// translations of themselves, and flagging them teaches translators that the rule cries wolf, which
// costs more than the handful of real findings it would add.
const minimumLengthToFlagAsUntranslated = 4

// untranslatedValuePreviewLength is how much of a value a finding quotes before eliding.
const untranslatedValuePreviewLength = 40

// localizationNoUntranslatedValueMissingTranslationText is the rule's `missingTranslation` message,
// whose wording lives in `policy/messages/localization-no-untranslated-value.json`.
var localizationNoUntranslatedValueMissingTranslationText = policy.MessageOf("nexus/localization-no-untranslated-value", "missingTranslation")

func messageMissingTranslation(key string, locale string) rule.Message {
	return rule.Message{
		Id: "missingTranslation",
		Description: localizationNoUntranslatedValueMissingTranslationText.Render(map[string]string{
			"key":    key,
			"locale": locale,
		}),
	}
}

// localizationNoUntranslatedValueAmbiguousEnglishSiblingText is the rule's `ambiguousEnglishSibling`
// message, whose wording lives in `policy/messages/localization-no-untranslated-value.json`.
var localizationNoUntranslatedValueAmbiguousEnglishSiblingText = policy.MessageOf("nexus/localization-no-untranslated-value", "ambiguousEnglishSibling")

func messageAmbiguousEnglishSibling(first string, second string) rule.Message {
	return rule.Message{
		Id: "ambiguousEnglishSibling",
		Description: localizationNoUntranslatedValueAmbiguousEnglishSiblingText.Render(map[string]string{
			"first":  first,
			"second": second,
		}),
	}
}

// localizationNoUntranslatedValueMissingEnglishSiblingText is the rule's `missingEnglishSibling`
// message, whose wording lives in `policy/messages/localization-no-untranslated-value.json`.
var localizationNoUntranslatedValueMissingEnglishSiblingText = policy.MessageOf("nexus/localization-no-untranslated-value", "missingEnglishSibling")

func messageMissingEnglishSibling(expected string, alternative string) rule.Message {
	return rule.Message{
		Id: "missingEnglishSibling",
		Description: localizationNoUntranslatedValueMissingEnglishSiblingText.Render(map[string]string{
			"expected":    expected,
			"alternative": alternative,
		}),
	}
}

// localizationNoUntranslatedValueIdenticalToSourceText is the rule's `identicalToSource` message,
// whose wording lives in `policy/messages/localization-no-untranslated-value.json`.
var localizationNoUntranslatedValueIdenticalToSourceText = policy.MessageOf("nexus/localization-no-untranslated-value", "identicalToSource")

func messageIdenticalToSource(key string, locale string, value string) rule.Message {
	preview := value
	if len(preview) > untranslatedValuePreviewLength {
		preview = preview[:untranslatedValuePreviewLength] + "..."
	}
	return rule.Message{
		Id: "identicalToSource",
		Description: localizationNoUntranslatedValueIdenticalToSourceText.Render(map[string]string{
			"key":     key,
			"preview": preview,
			"locale":  locale,
		}),
	}
}

// LocalizationNoUntranslatedValue requires a locale file to hold real translations rather than empty
// strings or copied English.
//
//	valid:   export default { Greeting: 'Hola' }        beside an en.ts saying 'Hello'
//	invalid: export default { Greeting: '' }            English has content
//	invalid: export default { Greeting: 'Hello' }       identical to English
//
// Both defects are invisible in the ordinary places a team would look. An empty value renders a
// blank rather than raising, and a copied English string is indistinguishable from a finished
// translation to every tool that counts keys, so a locale can report full coverage while showing a
// reader English. That is the whole reason this is a lint rule and not a spreadsheet review.
//
// The comparison is against the sibling en.ts, read from the program rather than from disk. The
// TypeScript original reads the file with `fs.readFileSync`, parses it with a second parser
// instance, and memoizes the result in a module-level cache keyed by directory. Every part of that
// is machinery we do not need: en.ts is a .ts file in the same directory, so it is already in the
// program and already parsed. Asking the program for it costs a map lookup, it cannot go stale
// against the file being linted, and it does not hold parsed ASTs alive for the process lifetime.
//
// An Adamic translation set's English is en.a, and a set moving to Adamic file by file can hold either,
// so the sibling is looked up with the file's own extension first, then the other; both present is
// reported, since which is English would be a guess (#kwt1htp).
//
// A file with neither English sibling in the program gets one finding at its start, naming the
// siblings it looked for, and no per-key comparison. Declining it silently made a set whose en.ts
// was missing or misnamed read as clean, which is the one result this rule exists to rule out
// (#techtr1); reporting every key instead would be a wall of findings that says nothing about the
// translations, since the comparison basis is what is missing.
//
// No fix. The repair is a translation, which a rule cannot write.
var LocalizationNoUntranslatedValue = rule.Rule{
	Name: "nexus/localization-no-untranslated-value",

	// The English translation table, en.ts, which this file does not import, and the current directory the
	// fingerprint names those tables against.
	ProgramReads: rule.ReadsCompilerOptions | rule.ReadsOtherFiles,
	// A file's verdict reads its own bytes and path and its directory's English tables, so every English
	// table in the program is the fingerprint (#31ffaaa). See localizationNoUntranslatedValueFingerprint.
	ProgramFingerprint: localizationNoUntranslatedValueFingerprint,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		localeCode, translationsDirectory, isTranslationFile := translationFileInformation(ctx.SourceFile)
		if !isTranslationFile {
			// Deciding on the path alone, before any node is visited, is what keeps this rule free
			// on the overwhelming majority of files that are not locale data.
			return nil
		}

		if ctx.Program == nil {
			return nil
		}
		englishSourceFile, both := englishSibling(ctx.Program, translationsDirectory, ctx.SourceFile.FileName().AsString())
		fileStart := ctx.SourceFile.AsNode().Loc.WithEnd(ctx.SourceFile.AsNode().Loc.Pos())
		if both != nil {
			ctx.ReportRange(fileStart, messageAmbiguousEnglishSibling(both[0], both[1]))
			return nil
		}
		if englishSourceFile == nil {
			lookedFor := englishSiblingBaseNames(ctx.SourceFile.FileName().AsString())
			ctx.ReportRange(fileStart, messageMissingEnglishSibling(lookedFor[0], lookedFor[1]))
			return nil
		}
		englishValues := map[string]string{}
		if englishObject := exportedObjectLiteral(englishSourceFile); englishObject != nil {
			collectStringValues(englishObject, "", englishValues)
		}
		if len(englishValues) == 0 {
			return nil
		}

		return rule.Listeners{
			ast.KindExportAssignment: func(node *ast.Node) {
				assignment := node.AsExportAssignment()
				// `export = thing` is a different statement wearing the same node kind, and it is
				// not the shape a translation file uses.
				if assignment == nil || assignment.IsExportEquals || assignment.Expression == nil {
					return
				}

				objectLiteral := resolveObjectLiteral(assignment.Expression, ctx.SourceFile)
				if objectLiteral == nil {
					return
				}

				reportUntranslated(ctx, objectLiteral, "", englishValues, localeCode)
			},
		}
	},
}

// reportUntranslated walks a locale object against the English values, reporting as it goes.
//
// Nested objects recurse with a dotted prefix, so a key path in a finding reads the same way it is
// written in the file.
func reportUntranslated(
	ctx rule.Context,
	objectLiteral *ast.Node,
	prefix string,
	englishValues map[string]string,
	localeCode string,
) {
	forEachStringProperty(objectLiteral, prefix, func(key string, value string, valueNode *ast.Node) {
		englishValue, hasEnglishValue := englishValues[key]
		// A key English does not have is not this rule's business. It may be locale-specific, or it
		// may be stale, and either way the comparison this rule performs has no basis.
		if !hasEnglishValue {
			return
		}

		if text.TrimWhitespace(value) == "" && text.TrimWhitespace(englishValue) != "" {
			ctx.ReportNode(valueNode, messageMissingTranslation(key, localeCode))
			return
		}

		if value == englishValue && len(englishValue) >= minimumLengthToFlagAsUntranslated {
			ctx.ReportNode(valueNode, messageIdenticalToSource(key, localeCode, englishValue))
		}
	}, func(key string, nested *ast.Node) {
		reportUntranslated(ctx, nested, key, englishValues, localeCode)
	})
}

// collectStringValues flattens an object literal into dotted key paths, which is how the English
// side is read.
func collectStringValues(objectLiteral *ast.Node, prefix string, into map[string]string) {
	forEachStringProperty(objectLiteral, prefix, func(key string, value string, valueNode *ast.Node) {
		into[key] = value
	}, func(key string, nested *ast.Node) {
		collectStringValues(nested, key, into)
	})
}

// forEachStringProperty visits one object literal, handing string properties to onString and nested
// object properties to onObject, both with the dotted key path already built.
//
// One traversal serves both sides of the comparison. Writing the walk twice is how the English
// reader and the locale reader drift into disagreeing about what counts as a key, which would show
// up as findings on keys that do exist.
func forEachStringProperty(
	objectLiteral *ast.Node,
	prefix string,
	onString func(key string, value string, valueNode *ast.Node),
	onObject func(key string, nested *ast.Node),
) {
	literal := objectLiteral.AsObjectLiteralExpression()
	if literal == nil || literal.Properties == nil {
		return
	}

	for _, property := range literal.Properties.Nodes {
		// Shorthand properties, spreads, methods, and getters carry no literal string to compare.
		// The original filters to `Property` for the same reason.
		if property.Kind != ast.KindPropertyAssignment {
			continue
		}
		assignment := property.AsPropertyAssignment()
		if assignment == nil || assignment.Initializer == nil {
			continue
		}

		keyName, hasKeyName := propertyKeyName(assignment.Name())
		if !hasKeyName {
			continue
		}
		fullKey := keyName
		if prefix != "" {
			fullKey = prefix + "." + keyName
		}

		initializer := assignment.Initializer
		switch {
		case ast.IsStringLiteralLike(initializer):
			onString(fullKey, initializer.Text(), initializer)
		case initializer.Kind == ast.KindObjectLiteralExpression:
			onObject(fullKey, initializer)
		}
	}
}

// propertyKeyName reads the key of a property, whether it was written bare or quoted.
//
// A computed key is skipped, which is why `property.Computed` is absent from the accept set: its
// value is not known from the syntax, so there is no key path to compare against English. The
// remaining four kinds are exactly what this rule read before adoption.
func propertyKeyName(keyNode *ast.Node) (string, bool) {
	return property.Name(keyNode, property.Named|property.Quoted|property.Templated|property.Numeric)
}

// exportedObjectLiteral finds the object a file default-exports, or nil.
func exportedObjectLiteral(sourceFile *ast.SourceFile) *ast.Node {
	if sourceFile == nil {
		return nil
	}
	for _, statement := range sourceFile.Statements.Nodes {
		if statement.Kind != ast.KindExportAssignment {
			continue
		}
		assignment := statement.AsExportAssignment()
		if assignment == nil || assignment.IsExportEquals || assignment.Expression == nil {
			continue
		}
		if objectLiteral := resolveObjectLiteral(assignment.Expression, sourceFile); objectLiteral != nil {
			return objectLiteral
		}
	}
	return nil
}

// resolveObjectLiteral unwraps an exported expression down to the object literal underneath.
//
// Four shapes reach here, and all four appear in real translation files:
//
//	export default { ... }
//	export default { ... } satisfies TranslationsType
//	export default { ... } as TranslationsType
//	const Spanish = { ... } as const; export default Spanish
//
// The last one exists because some locale files need their literal-string types preserved through an
// intermediate binding, so a slot validator can compare placeholders against the contract, before
// the widening assertion at export. Following the identifier is what keeps those files linted
// rather than silently skipped, and a file this rule silently skips is a file whose translations
// nobody is checking.
func resolveObjectLiteral(expression *ast.Node, sourceFile *ast.SourceFile) *ast.Node {
	current := unwrapTypeAssertions(expression)
	if current == nil {
		return nil
	}
	if current.Kind == ast.KindObjectLiteralExpression {
		return current
	}

	if ast.IsIdentifier(current) && sourceFile != nil {
		return topLevelObjectLiteralBinding(current.Text(), sourceFile)
	}
	return nil
}

// topLevelObjectLiteralBinding finds `const name = { ... }` among a file's top-level statements.
//
// Only the top level is searched, matching the original. A binding declared inside a block is not
// something an export at the top level could have referred to.
func topLevelObjectLiteralBinding(name string, sourceFile *ast.SourceFile) *ast.Node {
	for _, statement := range sourceFile.Statements.Nodes {
		if statement.Kind != ast.KindVariableStatement {
			continue
		}
		variableStatement := statement.AsVariableStatement()
		if variableStatement == nil || variableStatement.DeclarationList == nil {
			continue
		}
		declarationList := variableStatement.DeclarationList.AsVariableDeclarationList()
		if declarationList == nil || declarationList.Declarations == nil {
			continue
		}
		for _, declaration := range declarationList.Declarations.Nodes {
			variableDeclaration := declaration.AsVariableDeclaration()
			if variableDeclaration == nil || variableDeclaration.Initializer == nil {
				continue
			}
			declarationName := variableDeclaration.Name()
			if declarationName == nil || !ast.IsIdentifier(declarationName) || declarationName.Text() != name {
				continue
			}
			initializer := unwrapTypeAssertions(variableDeclaration.Initializer)
			if initializer != nil && initializer.Kind == ast.KindObjectLiteralExpression {
				return initializer
			}
		}
	}
	return nil
}

// unwrapTypeAssertions strips `as T` and `satisfies T` wrappers, which are type-only and leave the
// value underneath unchanged.
//
// Repeatedly, because `{ ... } as const satisfies TranslationsType` wears both.
func unwrapTypeAssertions(expression *ast.Node) *ast.Node {
	current := expression
	for current != nil {
		switch current.Kind {
		case ast.KindAsExpression:
			asExpression := current.AsAsExpression()
			if asExpression == nil {
				return current
			}
			current = asExpression.Expression
		case ast.KindSatisfiesExpression:
			satisfiesExpression := current.AsSatisfiesExpression()
			if satisfiesExpression == nil {
				return current
			}
			current = satisfiesExpression.Expression
		case ast.KindParenthesizedExpression:
			parenthesized := current.AsParenthesizedExpression()
			if parenthesized == nil {
				return current
			}
			current = parenthesized.Expression
		default:
			return current
		}
	}
	return nil
}

// localeCodeStem is a translation file's name without its extension: a two- or three-letter language
// code, then any region or script subtags (`zh-TW`, `pt-BR`, `zh_Hant`).
//
// Recognizing locale data by its own name, rather than by excluding names that are not, is what keeps a
// helper that happens to sit in a translations directory out of the rule (#xt9hkse). The exclusion list
// this replaced named index.ts, locales.ts and the *Translations, *TranslationsType and *Interface type
// files, and a file it did not name was read as a locale: Structure's translations/TranslationTemplate.ts,
// a type brand, drew missingEnglishSibling once a missing sibling reported. None of the old exclusions
// can match this pattern. Measured against all 35 locales in the four consumers' sets, it keeps every one.
var localeCodeStem = regexp.MustCompile(`^[a-z]{2,3}(?:[-_][A-Za-z0-9]+)*$`)

// translationFileInformation decides whether a file is locale data this rule should read, returning
// its locale code and the directory its en.ts sits in.
//
// A locale file is one whose stem is a locale code, other than en: en.ts is the basis of the
// comparison, so every value in it is identical to English by definition and the rule would report
// the entire file.
//
// The directory match is on a whole path segment. Matching a bare substring would pull in a
// directory named "translations-archive" and, worse, would silently start linting whatever a future
// directory named that way holds.
func translationFileInformation(sourceFile *ast.SourceFile) (string, string, bool) {
	// An Adamic `.a` translation file is read as the `.ts` it is (#kwt1htp).
	fileName := imports.NormalizedFileName(sourceFile)
	if !strings.HasSuffix(sourcename.TreatedAs(fileName), ".ts") {
		return "", "", false
	}

	lastSlash := strings.LastIndex(fileName, "/")
	if lastSlash < 0 {
		return "", "", false
	}
	directory := fileName[:lastSlash]
	baseName := fileName[lastSlash+1:]

	localeCode := strings.TrimSuffix(sourcename.TreatedAs(baseName), ".ts")
	if localeCode == "en" || !localeCodeStem.MatchString(localeCode) {
		return "", "", false
	}

	if !isTranslationsDirectory(directory) {
		return "", "", false
	}

	return localeCode, directory, true
}

// isTranslationsDirectory is whether a directory holds a translation set: one with a `translations` or
// `_translations` path segment. The rule and its fingerprint both ask it, so the English tables the
// fingerprint hashes are exactly the ones a locale file can read.
//
// A whole path segment rather than a bare substring. Matching the substring would pull in a directory
// named "translations-archive" and, worse, would silently start linting whatever a future directory named
// that way holds.
func isTranslationsDirectory(directory string) bool {
	return imports.HasPathSegment(directory, "_translations") || imports.HasPathSegment(directory, "translations")
}

// localizationNoUntranslatedValueFingerprints holds the fingerprint for one program. The walk asks for it
// on every worker that serves a selection, and the answer scans every file the program holds, so it is
// made once per program.
var localizationNoUntranslatedValueFingerprints struct {
	sync.Mutex
	program     rule.ProgramIdentity
	fingerprint [sha256.Size]byte
	made        bool
}

// localizationNoUntranslatedValueFingerprint is the rule's program fingerprint (#31ffaaa): every English
// table in the program, en.ts and en.a in a translations directory, each named relative to the current
// directory and hashed by its text, sorted by name.
//
// A locale file's verdict reads, beyond its own bytes and path, which English tables sit in its directory
// and what they hold. Both present report ambiguousEnglishSibling with their names, neither present
// reports missingEnglishSibling, and one is read for its values. One value has to serve every file, and
// each file reads its own directory's tables, so every table any locale file could read is in it.
// Presence is in it as much as content: adding, deleting or renaming either name moves it. The text
// stands in for a version, since the rule reads the text, so an edit to a comment in en.ts moves it too.
// Nothing else outside the file reaches a verdict: recognizing locale data reads only the file's own name.
// The rule takes no options, so none reach it.
func localizationNoUntranslatedValueFingerprint(program rule.Program, _ any) [sha256.Size]byte {
	cache := &localizationNoUntranslatedValueFingerprints
	cache.Lock()
	defer cache.Unlock()
	if cache.made && cache.program == program.Identity() {
		return cache.fingerprint
	}

	type englishTable struct {
		name string
		text [sha256.Size]byte
	}
	var tables []englishTable
	for _, sourceFile := range program.SourceFiles() {
		fileName := imports.NormalizedFileName(sourceFile)
		lastSlash := strings.LastIndex(fileName, "/")
		if lastSlash < 0 {
			continue
		}
		if baseName := fileName[lastSlash+1:]; baseName != "en.ts" && baseName != "en"+sourcename.AdamicExtension {
			continue
		}
		if !isTranslationsDirectory(fileName[:lastSlash]) {
			continue
		}
		tables = append(tables, englishTable{name: rule.FingerprintPath(program, fileName), text: sha256.Sum256([]byte(sourceFile.Text()))})
	}
	slices.SortFunc(tables, func(first englishTable, second englishTable) int { return strings.Compare(first.name, second.name) })

	hash := sha256.New()
	for _, table := range tables {
		hash.Write([]byte(table.name))
		hash.Write([]byte{0})
		hash.Write(table.text[:])
	}
	copy(cache.fingerprint[:], hash.Sum(nil))
	cache.program, cache.made = program.Identity(), true
	return cache.fingerprint
}

// englishSibling finds a translation file's English table: en with the file's own extension first,
// then the other one, since a set moving to Adamic file by file holds `.ts` and `.a` side by side
// (#kwt1htp). When both en.ts and en.a are in the program, which is English is a guess, so both names
// come back for the rule to report and the file is not compared at all. Neither found returns nil.
func englishSibling(program rule.Program, directory string, fileName string) (*ast.SourceFile, []string) {
	var found *ast.SourceFile
	var names []string
	for _, baseName := range englishSiblingBaseNames(fileName) {
		name := directory + "/" + baseName
		if sourceFile := program.GetSourceFile(name); sourceFile != nil {
			if found == nil {
				found = sourceFile
			}
			names = append(names, name)
		}
	}
	if len(names) > 1 {
		return nil, names
	}
	return found, nil
}

// englishSiblingBaseNames is the English table's base names in the order they are looked up: the
// linted file's own extension first, then the other (#kwt1htp).
func englishSiblingBaseNames(fileName string) []string {
	if sourcename.IsAdamic(fileName) {
		return []string{"en" + sourcename.AdamicExtension, "en.ts"}
	}
	return []string{"en.ts", "en" + sourcename.AdamicExtension}
}
