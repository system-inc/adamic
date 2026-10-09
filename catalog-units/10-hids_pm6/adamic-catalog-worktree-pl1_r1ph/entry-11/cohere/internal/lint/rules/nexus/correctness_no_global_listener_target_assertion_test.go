package nexus

import (
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const correctnessNoGlobalListenerTargetAssertionFile = "/repository/source/CorrectnessNoGlobalListenerTargetAssertion.ts"

// correctnessNoGlobalListenerTargetAssertionPrelude brings in lib.dom, as ahra's `lib` does, and
// declares the names the cases use.
var correctnessNoGlobalListenerTargetAssertionPrelude = strings.Join([]string{
	`/// <reference lib="dom" />`,
	"declare function use(value: unknown): void;",
	"declare const textarea: HTMLTextAreaElement;",
	"declare const textareaReference: { current: HTMLTextAreaElement | null };",
	"",
}, "\n")

func correctnessNoGlobalListenerTargetAssertionSource(lines ...string) string {
	return correctnessNoGlobalListenerTargetAssertionPrelude + strings.Join(lines, "\n") + "\n"
}

func correctnessNoGlobalListenerTargetAssertionRun(t *testing.T, sourceText string) rule_testing.Result {
	t.Helper()
	return rule_testing.RunTypedFiles(t, CorrectnessNoGlobalListenerTargetAssertion, map[string]string{
		correctnessNoGlobalListenerTargetAssertionFile: sourceText,
	}, correctnessNoGlobalListenerTargetAssertionFile)
}

// correctnessNoGlobalListenerTargetAssertionExpect asserts the findings in source order by the text
// each points at, and that none carries a fix.
func correctnessNoGlobalListenerTargetAssertionExpect(t *testing.T, result rule_testing.Result, want []string) {
	t.Helper()
	// A silent case goes through the shared assertion that the rule stays quiet.
	if len(want) == 0 {
		rule_testing.ExpectClean(t, result)
		return
	}
	diagnostics := append(result.Diagnostics[:0:0], result.Diagnostics...)
	sort.SliceStable(diagnostics, func(left int, right int) bool {
		return diagnostics[left].Range.Pos() < diagnostics[right].Range.Pos()
	})
	wantIds := make([]string, len(want))
	for index := range want {
		wantIds[index] = correctnessNoGlobalListenerTargetAssertionId
	}
	rule_testing.ExpectFindings(t, result, wantIds...)
	text := result.SourceFile.Text()
	for index, diagnostic := range diagnostics {
		if span := text[diagnostic.Range.Pos():diagnostic.Range.End()]; span != want[index] {
			t.Fatalf("finding %d is %q, want %q", index, span, want[index])
		}
		if len(diagnostic.Fixes) != 0 || len(diagnostic.Suggestions) != 0 {
			t.Fatalf("expected no fix and no suggestion, got %d and %d", len(diagnostic.Fixes), len(diagnostic.Suggestions))
		}
	}
}

// correctnessNoGlobalListenerTargetAssertionCode models the Tab handler in Structure's
// `source/components/code/Code.tsx` as the own-history pass found it, outside React: a `keydown`
// listener on `document` that treats whatever had focus as the editor's textarea. `fixed` is the
// shape of the fix, the handler on the textarea itself reading the textarea's own caret.
func correctnessNoGlobalListenerTargetAssertionCode(fixed bool) string {
	receiver := "document"
	target := "event.target as HTMLTextAreaElement"
	if fixed {
		receiver = "textarea"
		target = "textarea"
	}
	return correctnessNoGlobalListenerTargetAssertionSource(
		"export function attachTabHandler(code: string, setCode?: (code: string) => void) {",
		"    function handleKeyDown(event: KeyboardEvent) {",
		"        if(event.key !== 'Tab') return;",
		"        event.preventDefault();",
		"        const target = "+target+";",
		"        const start = target.selectionStart;",
		"        setCode?.(code.substring(0, target.selectionStart) + '    ' + code.substring(target.selectionEnd));",
		"        target.selectionStart = target.selectionEnd = start + 2;",
		"    }",
		"    "+receiver+".addEventListener('keydown', handleKeyDown);",
		"    return function() {",
		"        "+receiver+".removeEventListener('keydown', handleKeyDown);",
		"    };",
		"}",
	)
}

// The real site, before the fix.
func TestCorrectnessNoGlobalListenerTargetAssertionFiresOnCode(t *testing.T) {
	t.Parallel()

	result := correctnessNoGlobalListenerTargetAssertionRun(t, correctnessNoGlobalListenerTargetAssertionCode(false))
	correctnessNoGlobalListenerTargetAssertionExpect(t, result, []string{"event.target as HTMLTextAreaElement"})
}

// The real site, after the fix: the listener is on the textarea.
func TestCorrectnessNoGlobalListenerTargetAssertionStaysSilentOnFixedCode(t *testing.T) {
	t.Parallel()

	result := correctnessNoGlobalListenerTargetAssertionRun(t, correctnessNoGlobalListenerTargetAssertionCode(true))
	correctnessNoGlobalListenerTargetAssertionExpect(t, result, nil)
}

// Every handler form, both receivers, both assertion spellings, the wrappers around the read, a
// nested callback, and a union of specific types.
func TestCorrectnessNoGlobalListenerTargetAssertionFiresOnEveryHandlerForm(t *testing.T) {
	t.Parallel()

	result := correctnessNoGlobalListenerTargetAssertionRun(t, correctnessNoGlobalListenerTargetAssertionSource(
		"window.addEventListener('click', (event) => { (event.target as HTMLInputElement).select(); });",
		"document.addEventListener('input', function(event) { use((<HTMLInputElement>event.target).value); });",
		"const handleChange = function(event: Event) { use((event.target! as HTMLSelectElement).value); };",
		"document.addEventListener('change', handleChange);",
		"const handlePaste = (event: ClipboardEvent) => { use(((event.target) as unknown as HTMLTextAreaElement).value); };",
		"(document).addEventListener('paste', handlePaste, { capture: true });",
		"document.addEventListener('keyup', function(event) {",
		"    requestAnimationFrame(function() { use((event.target as HTMLInputElement | HTMLTextAreaElement | null)?.value); });",
		"});",
		"window.addEventListener('pointerdown', function(event) { use((event.target as SVGPathElement).getTotalLength()); });",
	))
	correctnessNoGlobalListenerTargetAssertionExpect(t, result, []string{
		"event.target as HTMLInputElement",
		"<HTMLInputElement>event.target",
		"event.target! as HTMLSelectElement",
		"(event.target) as unknown as HTMLTextAreaElement",
		"event.target as HTMLInputElement | HTMLTextAreaElement | null",
		"event.target as SVGPathElement",
	})
}

// The nearest legitimate shapes: general element types, an `instanceof` guard the checker sees,
// another listener receiver, a local named `document`, `currentTarget`, a nested handler's own event,
// and a non-element assertion.
func TestCorrectnessNoGlobalListenerTargetAssertionStaysSilentOnOtherShapes(t *testing.T) {
	t.Parallel()

	result := correctnessNoGlobalListenerTargetAssertionRun(t, correctnessNoGlobalListenerTargetAssertionSource(
		"document.addEventListener('click', function(event) { const target = event.target as HTMLElement; use(target.closest('[data-menu]')); });",
		"document.addEventListener('click', function(event) { use((event.target as Element).tagName); use(event.target as Node); use(event.target as SVGElement); });",
		"document.addEventListener('keydown', function(event) {",
		"    if(!(event.target instanceof HTMLTextAreaElement)) return;",
		"    use((event.target as HTMLTextAreaElement).selectionStart);",
		"});",
		"textarea.addEventListener('keydown', function(event) { use((event.target as HTMLTextAreaElement).selectionStart); });",
		"textareaReference.current?.addEventListener('keydown', function(event) { use((event.target as HTMLTextAreaElement).value); });",
		"document.body.addEventListener('click', function(event) { use((event.target as HTMLButtonElement).form); });",
		"function scoped(document: HTMLFormElement) {",
		"    document.addEventListener('submit', function(event) { use((event.target as HTMLFormElement).action); });",
		"}",
		"document.addEventListener('click', function(event) { use((event.currentTarget as Document).title); });",
		"document.addEventListener('click', function(event) {",
		"    use(event.type);",
		"    textarea.addEventListener('focus', function(event) { use((event.target as HTMLTextAreaElement).value); });",
		"});",
		"let handleLater = function(event: Event) { use((event.target as HTMLInputElement).value); };",
		"document.addEventListener('input', handleLater);",
		"handleLater = function(event: Event) { use(event.type); };",
		"window.addEventListener('message', function(event) { use(event.source as Window); use(event.target as Window); });",
		"declare function attach(handler: (event: Event) => void): void;",
		"attach(function(event) { use((event.target as HTMLInputElement).value); });",
		"document.addEventListener('click', function(event: { target: unknown }) { use((event.target as HTMLInputElement).value); });",
		"window.addEventListener('resize', use);",
	))
	correctnessNoGlobalListenerTargetAssertionExpect(t, result, nil)
}

// A project declaration of `document` or of an element type is not lib.dom's.
func TestCorrectnessNoGlobalListenerTargetAssertionStaysSilentOnLocalDeclarations(t *testing.T) {
	t.Parallel()

	result := correctnessNoGlobalListenerTargetAssertionRun(t, correctnessNoGlobalListenerTargetAssertionSource(
		"interface CodeEditorElement extends HTMLElement { caret: number }",
		"document.addEventListener('keydown', function(event) { use((event.target as CodeEditorElement).caret); });",
		"function local() {",
		"    const window = { addEventListener(type: string, handler: (event: Event) => void) { use(type); use(handler); } };",
		"    window.addEventListener('click', function(event) { use((event.target as HTMLInputElement).value); });",
		"}",
	))
	correctnessNoGlobalListenerTargetAssertionExpect(t, result, nil)
}

// A project declaration merged into lib.dom's `addEventListener` makes it something the rule has not
// read, so the whole file's `document` listeners are declined.
func TestCorrectnessNoGlobalListenerTargetAssertionStaysSilentOnAMergedAddEventListener(t *testing.T) {
	t.Parallel()

	result := correctnessNoGlobalListenerTargetAssertionRun(t, correctnessNoGlobalListenerTargetAssertionSource(
		"declare global {",
		"    interface Document { addEventListener(type: 'editor-ready', listener: (event: CustomEvent) => void): void }",
		"}",
		"document.addEventListener('keydown', function(event) { use((event.target as HTMLTextAreaElement).selectionStart); });",
	))
	correctnessNoGlobalListenerTargetAssertionExpect(t, result, nil)
}
