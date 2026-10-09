package structure

import (
	"crypto/sha256"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

const (
	themeFilePath   = "/repository/libraries/structure/source/components/buttons/ButtonTheme.ts"
	libraryFilePath = "/repository/libraries/structure/source/components/buttons/UseButton.tsx"
	projectFilePath = "/repository/app/things/UseButton.tsx"
)

// buttonThemeSource is the theme every fixture below resolves against. Two suffixes, so a rule
// keyed to one of them cannot pass the whole pair.
const buttonThemeSource = "export interface ButtonVariantsInterface {\n" +
	"    Ghost: string;\n" +
	"    Outline: string;\n" +
	"}\n\n" +
	"export interface ButtonSizesInterface {\n" +
	"    Small: string;\n" +
	"    Base: string;\n" +
	"}\n"

func TestBoundaryNoProjectThemeValueFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		// A self-closing element, which is the dominant form in this codebase and the one a port
		// checking only JsxOpeningElement would silently skip.
		{"a self-closing element", "export const a = <Button variant=\"Emphasized\" />;\n"},
		// An element with children, which is the other kind carrying attributes.
		{"an element with children", "export const a = <Button variant=\"Emphasized\">go</Button>;\n"},
		// The second suffix, so the mapping is exercised beyond `variant`.
		{"a size value", "export const a = <Button size=\"Huge\" />;\n"},
		// Aliases render another component's theme and take that component's values.
		{"an aliased component", "export const a = <AnimatedButton variant=\"Emphasized\" />;\n"},
		{"a second alias to the same theme", "export const a = <TipButton variant=\"Emphasized\" />;\n"},
		// Casing matters: the values are interface keys, compared exactly.
		{"a value differing only in case", "export const a = <Button variant=\"ghost\" />;\n"},
		{"an empty value", "export const a = <Button variant=\"\" />;\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.RunTypedFiles(t, BoundaryNoProjectThemeValue, map[string]string{
				themeFilePath:   buttonThemeSource,
				libraryFilePath: testCase.sourceText,
			}, libraryFilePath), "forbiddenThemeValue")
		})
	}
}

func TestBoundaryNoProjectThemeValueStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		fileName   string
		sourceText string
	}{
		// Values the theme defines. These are the population the rule runs over and must not touch:
		// 458 attributes on the real tree resolve to a theme and every one of them is allowed.
		{"a defined variant", libraryFilePath, "export const a = <Button variant=\"Ghost\" />;\n"},
		{"the other defined variant", libraryFilePath, "export const a = <Button variant=\"Outline\" />;\n"},
		{"a defined size", libraryFilePath, "export const a = <Button size=\"Small\" />;\n"},
		{"a defined value on an alias", libraryFilePath, "export const a = <AnimatedButton variant=\"Ghost\" />;\n"},
		// A defined value spelled with a character reference. ESLint's parser decodes a JSX attribute
		// string, and the twin reads that decoded value, so this is Ghost to both (#5f61r5b).
		{"a defined value spelled with a character reference", libraryFilePath, "export const a = <Button variant=\"Gh&#111;st\" />;\n"},
		// The boundary itself. The same undefined value outside the library is a project extending
		// the library, which is the system working. This fixture is the one that pins the gate, and
		// it has to carry a value that would otherwise report.
		{"an undefined value outside the library", projectFilePath, "export const a = <Button variant=\"Emphasized\" />;\n"},
		// A component with no theme. Nothing defines its values, so the rule has no standing.
		{"a component with no theme", libraryFilePath, "export const a = <Unthemed variant=\"Emphasized\" />;\n"},
		// A property that is not a theme property, carrying a value no theme defines.
		{"a non-theme property", libraryFilePath, "export const a = <Button title=\"Emphasized\" />;\n"},
		{"a property merely containing a theme word", libraryFilePath, "export const a = <Button variantName=\"Emphasized\" />;\n"},
		// An expression rather than a literal. The value cannot be read without evaluating it, and
		// guessing is how a boundary rule starts reporting on correct code.
		{"an expression value", libraryFilePath, "export const a = <Button variant={chosen} />;\n"},
		{"a template value", libraryFilePath, "export const a = <Button variant={`Gh${x}`} />;\n"},
		{"a shorthand attribute with no value", libraryFilePath, "export const a = <Button variant />;\n"},
		// A qualified element name is not a bare identifier and names no component in the cache.
		{"a qualified element name", libraryFilePath, "export const a = <Foo.Button variant=\"Emphasized\" />;\n"},
		// A lowercase intrinsic element, which is a DOM tag rather than a component.
		{"an intrinsic element", libraryFilePath, "export const a = <button variant=\"Emphasized\" />;\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTypedFiles(t, BoundaryNoProjectThemeValue, map[string]string{
				themeFilePath:     buttonThemeSource,
				testCase.fileName: testCase.sourceText,
			}, testCase.fileName))
		})
	}
}

// TestBoundaryNoProjectThemeValueReadsTheProgramRatherThanTheFilesystem is the guard for the one
// place this rule diverges from the original.
//
// The original discovers theme files by walking the filesystem from `process.cwd()`. This reads
// them out of the program. The difference is invisible to every fixture above, because those pass
// whether the cache came from disk or from the graph, so it needs its own test: a theme that exists
// only as an in-memory program file, on a path nothing has ever written to disk.
//
// If the rule were still walking the filesystem, no theme would be found, the cache would be empty,
// and every one of these would report nothing.
func TestBoundaryNoProjectThemeValueReadsTheProgramRatherThanTheFilesystem(t *testing.T) {
	t.Parallel()

	const inMemoryThemePath = "/repository/libraries/structure/source/components/cards/CardTheme.ts"
	const inMemoryTheme = "export interface CardVariantsInterface {\n    Flat: string;\n}\n"

	t.Run("a value the in-memory theme defines", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectClean(t, rule_testing.RunTypedFiles(t, BoundaryNoProjectThemeValue, map[string]string{
			inMemoryThemePath: inMemoryTheme,
			libraryFilePath:   "export const a = <Card variant=\"Flat\" />;\n",
		}, libraryFilePath))
	})

	t.Run("a value it does not define", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectFindings(t, rule_testing.RunTypedFiles(t, BoundaryNoProjectThemeValue, map[string]string{
			inMemoryThemePath: inMemoryTheme,
			libraryFilePath:   "export const a = <Card variant=\"Raised\" />;\n",
		}, libraryFilePath), "forbiddenThemeValue")
	})

	// The exported-only check. A theme interface the file keeps to itself is not the library's
	// published surface, and the original only looks inside an ExportNamedDeclaration. Without this
	// fixture nothing pins that: every other theme in the suite is exported, so dropping the check
	// changes none of them. Here the only definition of Card's variants is unexported, so a rule
	// that read it would find "Raised" undefined and report, and a rule that correctly ignores it
	// finds no theme for Card at all and declines.
	t.Run("an unexported theme interface defines nothing", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectClean(t, rule_testing.RunTypedFiles(t, BoundaryNoProjectThemeValue, map[string]string{
			"/repository/libraries/structure/source/components/cards/CardTheme.ts": "interface CardVariantsInterface {\n    Flat: string;\n}\nexport const unused = 1;\n",
			libraryFilePath: "export const a = <Card variant=\"Raised\" />;\n",
		}, libraryFilePath))
	})

	// Two theme files sharing a component prefix, which the tree actually contains: there are two
	// CalendarTheme.ts files, and the second declares a suffix the first does not.
	//
	// Different suffixes accumulate. `themes[prefix]` is created once and each suffix is written
	// into it, which is what the original's `Object.assign(existing, suffixMap)` does. Verified by
	// executing the original's merge on both shapes rather than by reading it, because the reading
	// admits two interpretations and only one of them is what `Object.assign` performs.
	t.Run("two theme files sharing a prefix both contribute", func(t *testing.T) {
		t.Parallel()
		files := map[string]string{
			"/repository/libraries/structure/source/components/calendars/CalendarTheme.ts":     "export interface CalendarVariantsInterface {\n    A: string;\n}\n",
			"/repository/libraries/structure/source/components/time/calendar/CalendarTheme.ts": "export interface CalendarSizesInterface {\n    Base: string;\n}\n",
			libraryFilePath: "export const a = <Calendar variant=\"A\" size=\"Base\" />;\n",
		}
		rule_testing.ExpectClean(t, rule_testing.RunTypedFiles(t, BoundaryNoProjectThemeValue, files, libraryFilePath))
	})

	// The same pair, asking about values neither file defines. Without this the case above passes
	// even if the whole Calendar entry were missing from the cache, because a rule that finds no
	// theme declines silently. This is the half that tells a merge apart from a loss.
	t.Run("both contributed suffixes are actually enforced", func(t *testing.T) {
		t.Parallel()
		for _, attribute := range []string{"variant=\"Nope\"", "size=\"Nope\""} {
			files := map[string]string{
				"/repository/libraries/structure/source/components/calendars/CalendarTheme.ts":     "export interface CalendarVariantsInterface {\n    A: string;\n}\n",
				"/repository/libraries/structure/source/components/time/calendar/CalendarTheme.ts": "export interface CalendarSizesInterface {\n    Base: string;\n}\n",
				libraryFilePath: "export const a = <Calendar " + attribute + " />;\n",
			}
			rule_testing.ExpectFindings(t, rule_testing.RunTypedFiles(t, BoundaryNoProjectThemeValue, files, libraryFilePath),
				"forbiddenThemeValue")
		}
	})

	// The filename pattern, which is the one piece of the original's discovery that is reproduced
	// rather than replaced. A hook named useTheme.ts also ends in Theme.ts and is not a theme; the
	// leading capital is what separates them. On the real tree that file exists and carries no theme
	// interface, so this fixture is what keeps the distinction from being harmless-by-accident.
	t.Run("a lowercase theme-suffixed file is not a theme", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectClean(t, rule_testing.RunTypedFiles(t, BoundaryNoProjectThemeValue, map[string]string{
			"/repository/libraries/structure/source/theme/hooks/useTheme.ts": "export interface CardVariantsInterface {\n    Flat: string;\n}\n",
			libraryFilePath: "export const a = <Card variant=\"Raised\" />;\n",
		}, libraryFilePath))
	})
}

// The cache is keyed on the program pointer, and both directions have to be pinned.
//
// This is the fixture the existing corpus cannot supply. The rule finds zero on the real tree, so a
// cache returning a stale map, an empty map, or a map from a different program all look exactly like
// today's correct-but-slow behavior: silence. An empty theme map disables the rule entirely, which
// this file already names as the hazard, so a cache bug does not fail loudly, it stops linting.
//
// Two runs with different programs must not share an answer, and that is the miss. Two calls within
// one run must reach the same map, and that is the hit.
func TestBoundaryNoProjectThemeValueCachesPerProgram(t *testing.T) {
	t.Parallel()

	// A theme naming `Raised`, so `variant="Flat"` is a violation and `variant="Raised"` is not.
	firstProgram := map[string]string{
		"/repository/libraries/structure/source/components/CardTheme.ts": "export interface CardVariantsInterface {\n    Raised: string;\n}\n",
		libraryFilePath: "export const a = <Card variant=\"Flat\" />;\n",
	}

	// A different program whose theme names `Flat` instead. If the cache answered from the first
	// program, this file would report, because `Flat` is absent from that theme.
	secondProgram := map[string]string{
		"/repository/libraries/structure/source/components/CardTheme.ts": "export interface CardVariantsInterface {\n    Flat: string;\n}\n",
		libraryFilePath: "export const a = <Card variant=\"Flat\" />;\n",
	}

	t.Run("the first program reports against its own theme", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectFindings(t,
			rule_testing.RunTypedFiles(t, BoundaryNoProjectThemeValue, firstProgram, libraryFilePath),
			"forbiddenThemeValue")
	})

	// The miss. A stale cache keyed on nothing would answer with the first program's theme and
	// report here, so this failing is what a key-less memo looks like.
	t.Run("a second program is not answered from the first", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectClean(t,
			rule_testing.RunTypedFiles(t, BoundaryNoProjectThemeValue, secondProgram, libraryFilePath))
	})

	// The hit. Re-running the first program after the second must recompute rather than keep the
	// second's answer, which is the same bug in the other direction and is what a cache that
	// replaces without comparing the key would do.
	t.Run("returning to the first program reports again", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectFindings(t,
			rule_testing.RunTypedFiles(t, BoundaryNoProjectThemeValue, firstProgram, libraryFilePath),
			"forbiddenThemeValue")
	})
}

// themeFingerprintOf is the rule's program fingerprint over a fixture, read through the Program the walk hands
// the rule, viewed under its own reads.
func themeFingerprintOf(t *testing.T, files map[string]string) [sha256.Size]byte {
	t.Helper()
	var fingerprint [sha256.Size]byte
	probe := BoundaryNoProjectThemeValue
	probe.Run = func(ctx rule.Context, options any) rule.Listeners {
		fingerprint = BoundaryNoProjectThemeValue.ProgramFingerprint(ctx.Program, options)
		return nil
	}
	rule_testing.RunTypedFiles(t, probe, files, libraryFilePath)
	if fingerprint == ([sha256.Size]byte{}) {
		t.Fatal("the probe never ran, so no fingerprint was read")
	}
	return fingerprint
}

// The fingerprint is the whole of what a file's verdict reads beyond the file (#kdee854), proven both ways:
// every change to the theme map moves it, a value's order included since the message lists them in order, and
// an edit that cannot change any verdict leaves it, so the findings cache replays across it.
func TestBoundaryNoProjectThemeValueFingerprintsExactlyTheThemeMap(t *testing.T) {
	t.Parallel()
	base := map[string]string{
		themeFilePath:   buttonThemeSource,
		libraryFilePath: "export const a = <Button variant=\"Ghost\" />;\n",
		projectFilePath: "export const b = 1;\n",
	}
	with := func(path string, text string) map[string]string {
		files := map[string]string{}
		for name, contents := range base {
			files[name] = contents
		}
		files[path] = text
		return files
	}
	original := themeFingerprintOf(t, base)

	moves := map[string]map[string]string{
		"a value added":   with(themeFilePath, strings.Replace(buttonThemeSource, "Outline: string;", "Outline: string;\n    Emphasized: string;", 1)),
		"a value renamed": with(themeFilePath, strings.Replace(buttonThemeSource, "Ghost", "Spectral", 1)),
		"two values swapped": with(themeFilePath, strings.Replace(buttonThemeSource, "Ghost: string;\n    Outline: string;",
			"Outline: string;\n    Ghost: string;", 1)),
		"a theme file added": with("/repository/libraries/structure/source/components/cards/CardTheme.ts",
			"export interface CardVariantsInterface {\n    Flat: string;\n}\n"),
	}
	for name, files := range moves {
		if themeFingerprintOf(t, files) == original {
			t.Errorf("%s left the fingerprint unchanged, so a stale finding would replay", name)
		}
	}

	holds := map[string]map[string]string{
		"a library file edited":       with(libraryFilePath, "export const a = <Button variant=\"Outline\" />;\n"),
		"a project file edited":       with(projectFilePath, "export const b = 2;\n"),
		"a comment in the theme file": with(themeFilePath, "// The button's looks.\n"+buttonThemeSource),
		"an unexported interface beside the theme": with(themeFilePath,
			buttonThemeSource+"\ninterface ButtonKindsInterface {\n    Secret: string;\n}\n"),
	}
	for name, files := range holds {
		if themeFingerprintOf(t, files) != original {
			t.Errorf("%s moved the fingerprint, so files whose verdicts cannot change were walked again", name)
		}
	}
}
