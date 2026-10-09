package main

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/system-inc/cohere/internal/edit"
	"github.com/system-inc/cohere/internal/format/formatfiles"
	"github.com/system-inc/cohere/internal/format/formatoptions"
	"github.com/system-inc/cohere/internal/format/native"
	"github.com/system-inc/cohere/internal/format/printing"
)

// fakeEngine stands in for the real formatter so the seam can be tested before the engine exists.
//
// It is not a mock of convenience: every outcome the adapter must distinguish is reachable through
// it, which is what makes these fixtures a test of the mapping rather than a test of goja. The real
// engine drops in behind the same interface and these keep meaning what they mean.
type fakeEngine struct {
	handled        []string
	format         func(fileName string, text string) (string, error)
	enumerate      func(root string) (formatfiles.Enumeration, error)
	options        func(fileName string) (string, error)
	askedFor       []string
	enumeratedRoot string

	// askedForMutex guards askedFor, since the nested drift check formats several files at once.
	askedForMutex sync.Mutex
}

func (e *fakeEngine) Handles(fileName string) bool {
	for _, extension := range e.handled {
		if strings.HasSuffix(fileName, extension) {
			return true
		}
	}
	return false
}

// Enumerate lets the fake stand in for the real engine on the enumeration path too, so the wiring
// and the coverage line can be tested without a goja runtime or a real tree.
//
// It returns whatever the fixture set, including the account of what the walk removed, because the
// numbers in that account are the thing under test rather than the walk itself.
func (e *fakeEngine) Enumerate(root string) (formatfiles.Enumeration, error) {
	e.enumeratedRoot = root
	if e.enumerate != nil {
		return e.enumerate(root)
	}
	return formatfiles.Enumeration{Root: root}, nil
}

// OptionsFingerprint is one fixed configuration, unless the fixture says otherwise.
func (e *fakeEngine) OptionsFingerprint(fileName string) (string, error) {
	if e.options != nil {
		return e.options(fileName)
	}
	return "fake options", nil
}

func (e *fakeEngine) Format(fileName string, text string) (string, error) {
	e.askedForMutex.Lock()
	e.askedFor = append(e.askedFor, fileName)
	e.askedForMutex.Unlock()
	if e.format != nil {
		return e.format(fileName, text)
	}
	return text, nil
}

// prettierLike is the engine as @system_cohere_format describes it: the extensions it handles today,
// and a stand-in transform that collapses runs of spaces so a formatted result is observable.
func prettierLike() *fakeEngine {
	return &fakeEngine{
		handled: []string{".ts", ".tsx", ".js", ".jsx", ".mjs", ".md", ".css", ".json"},
		format: func(_ string, text string) (string, error) {
			for strings.Contains(text, "  ") {
				text = strings.ReplaceAll(text, "  ", " ")
			}
			return text, nil
		},
	}
}

// A file type the formatter does not handle must be skipped with its reason named, never silently
// returned unchanged.
//
// Returning the input unchanged would work and would be wrong: it is indistinguishable from
// "already correctly formatted", so a formatter that handles nothing would report a whole tree as
// clean. That is the ambiguity `edit.ErrSkipped` exists to remove.
func TestAnUnhandledFileTypeIsSkippedWithItsReason(t *testing.T) {
	t.Parallel()
	transform := formatTransform(prettierLike())

	for _, fileName := range []string{"Makefile", "notes.txt", "query.graphql", "script.sh"} {
		t.Run(fileName, func(t *testing.T) {
			t.Parallel()
			_, err := transform(fileName, "anything at all\n", nil)
			if !errors.Is(err, edit.ErrSkipped) {
				t.Fatalf("expected a skip for %s, got %v", fileName, err)
			}
			if !strings.Contains(err.Error(), "not a file type the formatter handles") {
				t.Fatalf("the skip did not say why: %v", err)
			}
		})
	}
}

// The skip reason names the extension rather than the file, so a run over four hundred markdown
// files reports one grouped reason instead of four hundred distinct ones.
func TestSkipReasonsGroupByExtension(t *testing.T) {
	t.Parallel()
	transform := formatTransform(prettierLike())

	_, firstError := transform("docs/one.txt", "x\n", nil)
	_, secondError := transform("notes/two.txt", "y\n", nil)
	if firstError == nil || secondError == nil {
		t.Fatalf("expected both files to skip, got %v and %v", firstError, secondError)
	}

	first, second := firstError.Error(), secondError.Error()
	if first != second {
		t.Fatalf("two files of one type produced different skip reasons, so the summary cannot group them:\n  %q\n  %q", first, second)
	}
	if !strings.Contains(first, ".txt") {
		t.Fatalf("the reason does not name the extension: %q", first)
	}
}

// Every file type the formatter does handle must not be skipped. A list that silently excluded a
// real type would leave that type unformatted forever, reported as a clean run.
func TestHandledFileTypesAreNotSkipped(t *testing.T) {
	t.Parallel()
	transform := formatTransform(prettierLike())

	for _, fileName := range []string{"a.ts", "b.tsx", "c.js", "d.jsx", "e.mjs", "f.md", "g.css", "h.json"} {
		t.Run(fileName, func(t *testing.T) {
			t.Parallel()
			_, err := transform(fileName, "const a = 1;\n", nil)
			if errors.Is(err, edit.ErrSkipped) {
				t.Fatalf("%s was skipped but should be formatted: %v", fileName, err)
			}
		})
	}
}

// A file that does not parse is a skip rather than a failure. The engine already refuses to fix a
// file in that state and says so, and the types phase reports it in a form a reader can act on, so
// a second complaint from the formatter adds noise rather than information.
func TestAnUnparseableFileIsSkippedNotFailed(t *testing.T) {
	t.Parallel()
	_, err := formatTransform(parseFailingEngine())("broken.ts", "function alpha( {\n", nil)

	if !errors.Is(err, edit.ErrSkipped) {
		t.Fatalf("an unparseable file should skip, not fail: %v", err)
	}
	if !strings.Contains(err.Error(), "does not parse") {
		t.Fatalf("the skip did not say why: %v", err)
	}
}

// A well-formed file formats and comes back as text, and the transform is idempotent on its own
// output.
//
// Idempotence is the property the engine's convergence loop assumes and cannot enforce. A formatter
// that alternated between two valid outputs would make every run rewrite every file, so a tree
// would never reach a steady state and every commit would carry churn nobody authored.
func TestFormattingIsIdempotent(t *testing.T) {
	t.Parallel()
	transform := formatTransform(prettierLike())
	source := "const   alpha    =   1;\nfunction beta(  ) {   return alpha   }\n"

	first, err := transform("sample.ts", source, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	second, err := transform("sample.ts", first, nil)
	if err != nil {
		t.Fatalf("unexpected error on the second pass: %v", err)
	}

	if first != second {
		t.Fatalf("the formatter is not idempotent:\n  pass 1 %q\n  pass 2 %q", first, second)
	}
}

// Formatting must not change what the code means, which at this layer means the result still
// parses. The engine checks this too, and checking it here as well is deliberate: the engine's
// guard proves the pipeline is safe, this one proves the transform itself is not the thing
// generating garbage.
func TestFormattedOutputStillParses(t *testing.T) {
	t.Parallel()
	formatted, err := formatTransform(prettierLike())("sample.ts", "const   alpha  =  1;\nexport   { alpha };\n", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parses, reason := edit.Parses("sample.ts", formatted); !parses {
		t.Fatalf("the formatter produced text that does not parse: %s (%q)", reason, formatted)
	}
}

// extensionOf must name something a reader recognizes even for the awkward paths, since its output
// goes straight into the summary line.
func TestExtensionOfNamesTheType(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		fileName string
		want     string
	}{
		{"a.ts", ".ts"},
		{"deep/nested/path/b.tsx", ".tsx"},
		{"Makefile", "a file with no extension"},
		{"weird.dir/noext", "a file with no extension"},
		{"archive.tar.gz", ".gz"},
	} {
		if got := extensionOf(testCase.fileName); got != testCase.want {
			t.Fatalf("extensionOf(%q) = %q, want %q", testCase.fileName, got, testCase.want)
		}
	}
}

// parseFailingEngine reports a parse failure the way a real formatter does, as an ordinary error
// with a message rather than a distinct type.
func parseFailingEngine() *fakeEngine {
	return &fakeEngine{
		handled: []string{".ts", ".tsx", ".js", ".md"},
		format: func(fileName string, _ string) (string, error) {
			return "", fmt.Errorf("prettier: %s does not parse", fileName)
		},
	}
}

// A genuine format failure must stay a failure, not be downgraded to a skip.
//
// The two mean opposite things to a reader: a skip says the formatter chose not to look, a failure
// says it looked and broke. Downgrading would hide a broken formatter behind the same number as a
// file type nobody formats, which is the coverage-line failure wearing a different costume.
func TestAFormatFailureIsNotDowngradedToASkip(t *testing.T) {
	t.Parallel()
	engine := &fakeEngine{
		handled: []string{".ts"},
		format: func(_ string, _ string) (string, error) {
			return "", errors.New("the formatter ran out of memory")
		},
	}

	_, err := formatTransform(engine)("sample.ts", "const a = 1;\n", nil)

	if err == nil {
		t.Fatalf("a failing formatter reported success")
	}
	if errors.Is(err, edit.ErrSkipped) {
		t.Fatalf("a real failure was downgraded to a skip: %v", err)
	}
	if !strings.Contains(err.Error(), "ran out of memory") {
		t.Fatalf("the underlying error was lost: %v", err)
	}
}

// The engine must not be asked to format a file it does not handle. Asking would mean running the
// formatter to learn it should not have run, which on a 3,416-file tree is the difference between a
// scoped pass and a full one.
func TestAnUnhandledFileIsNeverHandedToTheEngine(t *testing.T) {
	t.Parallel()
	engine := prettierLike()
	transform := formatTransform(engine)

	if _, err := transform("Makefile", "all:\n", nil); !errors.Is(err, edit.ErrSkipped) {
		t.Fatalf("expected a skip, got %v", err)
	}
	if len(engine.askedFor) != 0 {
		t.Fatalf("the engine was asked to format a file it does not handle: %v", engine.askedFor)
	}

	if _, err := transform("a.ts", "const a = 1;\n", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(engine.askedFor) != 1 || engine.askedFor[0] != "a.ts" {
		t.Fatalf("the engine was not asked about the file it does handle: %v", engine.askedFor)
	}
}

// A run that did not ask to format is a skip with a reason, never a crash and never a silent pass.
//
// A nil engine reaching this code means formatting was not requested, and that must show up in the
// coverage line rather than as a tree that reports itself perfectly formatted. The reason says so,
// not that no formatter is configured: a plain --no-fix run on ahra, which has a format block, printed
// `11 not formatted (11 no formatter is configured)`.
func TestAnUnrequestedFormatSkipsWithAReason(t *testing.T) {
	t.Parallel()
	_, err := formatTransform(nil)("a.ts", "const a = 1;\n", nil)

	if !errors.Is(err, edit.ErrSkipped) {
		t.Fatalf("a run that did not ask to format should skip, got %v", err)
	}
	if !strings.HasSuffix(err.Error(), ": not requested") {
		t.Fatalf("the skip did not say formatting was not requested: %v", err)
	}
}

// isUnparseable must discriminate, not merely detect. A matcher that said yes to everything would
// turn every real failure into a skip, which is the downgrade the fixture above forbids.
func TestUnparseableDetectionDiscriminates(t *testing.T) {
	t.Parallel()
	for _, message := range []string{
		"prettier: a.ts does not parse",
		"SyntaxError: Unexpected token (3:1)",
		"parse error at line 4",
	} {
		if !isUnparseable(errors.New(message)) {
			t.Fatalf("failed to recognize a parse failure: %q", message)
		}
	}

	for _, message := range []string{
		"the formatter ran out of memory",
		"connection refused",
		"the goja runtime panicked",
	} {
		if isUnparseable(errors.New(message)) {
			t.Fatalf("a real failure was misread as a parse failure: %q", message)
		}
	}
}

// The real error strings the goja engine produces must be recognized as parse failures.
//
// These four are measured rather than invented: @system_cohere_format ran four malformations
// through the fork and reported what came back. They are Prettier's own errors surfacing through
// goja unchanged, so the phrasing belongs to upstream and can move, but the `SyntaxError:` prefix is
// the stable part and is what the matcher keys on.
//
// Pinned here because a matcher tested only against strings I wrote myself is a matcher tested
// against my imagination. The one that matters is the one the engine actually emits.
func TestTheEnginesRealParseErrorsAreRecognized(t *testing.T) {
	t.Parallel()
	for _, message := range []string{
		"SyntaxError: Property assignment expected. (1:12)",
		"SyntaxError: Expression expected. (1:12)",
		"SyntaxError: Unexpected token. A constructor, method, accessor, or property was expected.",
		"SyntaxError: JSX element 'span' has no corresponding closing tag. (1:7)",
	} {
		if !isUnparseable(errors.New(message)) {
			t.Fatalf("a real parse failure was not recognized, so it would be reported as a broken formatter: %q", message)
		}
	}
}

// The native printers' parse failures must be recognized too, for the reason the goja ones above
// are pinned: the matcher has to know the errors the engine actually emits, and switching engines
// changes who emits them. Measured by formatting malformed source, never written by hand.
func TestTheNativeEnginesParseErrorsAreRecognized(t *testing.T) {
	t.Parallel()
	engine, err := configuredFormatter(true)
	if err != nil {
		t.Fatal(err)
	}
	for fileName, malformed := range map[string]string{
		"a.ts":  "const a = ;\n",
		"a.tsx": "const a = <span>;\n",
		"a.js":  "function (\n",
		// Their parsers word failures their own way ("json: line 1: ...", graphql-js's "Syntax Error: ..."),
		// so only printing.Syntax's mark tells the phase they are parse failures.
		"a.json":       "{ \"a\": }\n",
		"package.json": "{ \"a\": 1, // comments are not allowed in json-stringify\n}\n",
		"a.graphql":    "query { a(b: ) }\n",
	} {
		_, formatError := engine.Format(fileName, malformed)
		if formatError == nil {
			t.Errorf("%s: the native engine formatted malformed source without complaint", fileName)
			continue
		}
		if !isUnparseable(formatError) {
			t.Errorf("%s: a parse failure would be reported as a broken formatter: %v", fileName, formatError)
		}
	}
}

// With formatting off there is no formatter, and saying so is the coverage line's job, not an error.
func TestFormattingOffConfiguresNoFormatter(t *testing.T) {
	t.Parallel()
	engine, err := configuredFormatter(false)
	if err != nil || engine != nil {
		t.Fatalf("formatting off configured %v, %v; want no formatter and no error", engine, err)
	}
}

// A parse failure a native printer marks is a skip whatever its wording, and marking is the only way
// an unfamiliar message gets there: the same words unmarked are still a broken formatter.
func TestAMarkedSyntaxErrorIsUnparseableWhateverItSays(t *testing.T) {
	t.Parallel()
	plain := errors.New("json: line 1: a wording no matcher has seen")
	if isUnparseable(plain) {
		t.Fatal("an unmarked, unfamiliar message was read as a parse failure, so the marker proves nothing")
	}
	if !isUnparseable(fmt.Errorf("formatting a.json: %w", printing.Syntax(plain))) {
		t.Fatal("a marked parse failure, wrapped once more, was not recognized")
	}
}

// nonIdempotentUnifiChain is real input: modules/unifi/UnifiFleetCommandLineInterface.ts as it stood
// before ahra 7872fceb, cut to the expression that does not settle. One pass of Prettier breaks the
// `(…).filter(…).map(…)` chain across lines, and a second pass joins it again; the native printer
// matches both passes. Formatted once, it fails the `--no-fix` check on the text the write produced.
const nonIdempotentUnifiChain = `export function report(networkResults: PromiseSettledResult<Read>[], networkConsoles: Console[]): string {
    return JSON.stringify(
        {
            networks: networkResults.map((result, index) =>
                result.status === 'fulfilled'
                    ? {
                          console: networkConsoles[index]?.name,
                          blockedPorts: result.value.snapshot.devices.flatMap((device) =>
                              (device.port_table ?? []).filter(unifiIsPortBlocked).map((port) => ({ device: unifiDeviceName(device), port: port.port_idx, state: port.stp_state })),
                          ),
                      }
                    : { console: networkConsoles[index]?.name, error: String(result.reason) },
            ),
        },
        null,
        4,
    );
}
`

// The transform formats to a fixpoint, so what it writes is what the next run would leave alone.
func TestTheTransformFormatsARealNonIdempotentInputToItsFixpoint(t *testing.T) {
	t.Parallel()
	// ahra's options, named rather than resolved: resolving from this test's directory finds no
	// config and prints at Prettier's default width, where this input settles in one pass.
	printer := native.Formatter{Options: formatoptions.Default()}
	engine := &fakeEngine{handled: []string{".ts"}, format: printer.Format}

	// The premise: one pass really does not settle this input. Without it the case below would pass
	// on a printer that happened to be idempotent here and prove nothing about the loop.
	once, err := engine.Format("Report.ts", nonIdempotentUnifiChain)
	if err != nil {
		t.Fatal(err)
	}
	twice, err := engine.Format("Report.ts", once)
	if err != nil {
		t.Fatal(err)
	}
	if once == twice {
		t.Fatal("one pass settled the Unifi chain, so this fixture no longer reproduces a non-idempotent input")
	}

	formatted, err := formatTransform(engine)("Report.ts", nonIdempotentUnifiChain, nil)
	if err != nil {
		t.Fatal(err)
	}
	if formatted != twice {
		t.Fatalf("the transform stopped short of the fixpoint:\n--- transform\n%s--- two passes\n%s", formatted, twice)
	}
	if again, _ := formatTransform(engine)("Report.ts", formatted, nil); again != formatted {
		t.Fatalf("the transform's own output is rewritten by the next run:\n%s", again)
	}
}

// The bound: a file still changing on the last pass is a failure naming it, never a skip and never
// the last pass written as though it were formatted. The stability check: a file that settles on the
// second pass comes back settled, and an already formatted one costs a single pass.
func TestTheFormatFixpointIsBoundedAndStopsWhenStable(t *testing.T) {
	t.Parallel()
	growing := &fakeEngine{
		handled: []string{".ts"},
		format:  func(_ string, text string) (string, error) { return text + "x", nil },
	}
	_, err := formatTransform(growing)("Drifting.ts", "const a = 1;\n", nil)
	if err == nil || errors.Is(err, edit.ErrSkipped) {
		t.Fatalf("a file still changing at the bound was not a failure: %v", err)
	}
	if !strings.Contains(err.Error(), "Drifting.ts") || !strings.Contains(err.Error(), "not idempotent") {
		t.Fatalf("the failure does not name the file and why: %v", err)
	}
	if len(growing.askedFor) != formatPassLimit {
		t.Fatalf("formatted %d times, want the bound of %d", len(growing.askedFor), formatPassLimit)
	}

	settling := &fakeEngine{
		handled: []string{".ts"},
		format: func(_ string, text string) (string, error) {
			switch text {
			case "a\n":
				return "b\n", nil
			default:
				return "c\n", nil
			}
		},
	}
	if formatted, err := formatTransform(settling)("Settling.ts", "a\n", nil); err != nil || formatted != "c\n" {
		t.Fatalf("a file that settles on the second pass came back %q, %v; want \"c\\n\"", formatted, err)
	}

	stable := prettierLike()
	if _, err := formatTransform(stable)("Formatted.ts", "const a = 1;\n", nil); err != nil {
		t.Fatal(err)
	}
	if len(stable.askedFor) != 1 {
		t.Fatalf("an already formatted file was formatted %d times, want 1", len(stable.askedFor))
	}
}
