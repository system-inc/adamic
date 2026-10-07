Built the native react/jsx-no-constructed-context-values analysis, including memo dependencies, source callees and escape checks.
Validated implementation commit 47e48bb178ebd0020cba80346d8cafef9259c6cb; evidence commit follows on codex/typeaware-wave-20.
Checks: 171 controls/97 findings, both frozen corpora, sanitizer and three released-query checks PASS; formatting and configured lint clean.
Mutants: construction kind, missing list, render escape, helper escape and factory identity caught only by Go bytes; retained handle caught by required panic.
Not covered: shared checker-context registration, own emitted JavaScript, all possible Go source shapes and the full repository gate.

All three fifth-batch JSX analyses now exist and have independent native/Go
comparison gates. The constructed-context rule consumes only its handed opening
or self-closing node; the driver selects those named ast.Kind listeners once.
The rule never reads or refetches its entry to decide relevance. Child and
ancestor kinds describe the semantic shapes being analyzed. The existing
rule.json declaration remains JsxOpeningElement and JsxSelfClosingElement.

The construction half matches all 41 upstream controls and additional kinds,
component bindings/classes/wrappers, declaration ordering, cycle limits,
assertions, member usage, trivia and Unicode. It preserves Go's construction
anchor, usage location, exact messages, empty fixes and empty suggestions.
Module-scope and enclosing-function values keep their identity; the nearest
function decides the component test. The Unicode uppercase predicate is a
RegExp literal with the Lu property, matching Go's unicode.IsUpper category.
There are no Go regex matchers in this rule to hand-roll.

The memo half translates Go's separate, stricter analysis: a recomputing memo
must also produce a fresh identity. Primitive types compare by value; only
render-scoped const declarations and function declarations are followed.
Resolved source callees are inspected in their own files; async and generator
callees produce fresh promises and iterators. Unknown values remain unknown,
not proven fresh. Alternative results retain both their unknown paths and
holders, so cache stores, aliases, closures and method receivers can silence
an otherwise fresh value. Return traversal excludes nested functions/classes.
The active set and depth limit preserve Go's cycle behavior. Foreign source
views are cached. Missing lists, nested memos, stable factory results and all
literal production memo/escape table cases are held by complete output bytes.
Foreign helper, cached helper, async helper and ambient context witnesses were
added, as were CRLF and Unicode source positions. Go is invoked through an
independent loader and unmodified production Run, with no native implementation
or bridge import in its oracle.

No new bridge questions or shared files were needed. MemoView uses raw
binding-origin, type-shape and resolved-callee questions already on this branch.
Token positions are read from the source view; findings convert native UTF-16
positions to Go byte positions at the reporting boundary. The released probe
calls that same MemoView query wrapper after releasing the program. Each of the
three questions panics 70 with invalid or released checker handle. Omitting the
release compiles and exits 0 with empty stderr for all three, and the required
panic check catches it. This is one retained-handle source mutant tested on
three questions, not three independent registry mutants.

The semantic mutants compile, exit 0 with empty stderr, then differ only in
independent Go finding bytes:

- Construction kind: object relabeled array.
- Missing list: a one-argument memo fails to identify its missing dependency list (case-161).
- Render escape: a render fallback stored for a later render is falsely reported (case-140).
- Helper escape: a caching helper's fresh result is falsely reported (case-136).
- Factory identity: a memo returning an existing object is falsely reported (case-146).

Both frozen corpora pass normal and ASan/UBSan/LSan comparisons: repository 287
roots/18,485 output bytes; TypeScript compiler 77 roots/5,241 output bytes.
Both have zero findings for this rule; the 97 positive control findings supply
semantic coverage. The ordinary and sanitizer control stderr is empty.
Observed whole-process native/Go median seconds: repository .365240/.266588;
compiler 2.022580/.522749. These are observations, not an isolated speed claim.
The unmodified Go production tests pass in .204s. The cohere formatter is
idempotent on all ten new .a modules and configured cohere lint reports findings
0 on those modules. The full repository and shared lint gates were not run;
no compiler or shared harness source changed. Prior eleven-rule landing gates
remain recorded at validated source 26f1d659f and pushed evidence 7c056571b.

The shared RuleContext still lacks a checker program/lease, checker config and
root manifest. Shared registration of these three JSX analyses cannot be
completed inside owned rule directories against that API. They use the existing
private typed runner, with no zero-finding placeholder shared implementation.
This is the remaining integration blocker; construction and memo source are now
implemented. These three rules are not parked under the HIR exception. The
three earlier React compiler claims remain parked for their documented blockers.

Initial experiments are preserved: constructor initialization proof refusals,
an incomplete memo refusal probe, and two syntax errors during style edits.
All were corrected within owned files and the final source revalidated.
Classes were split into their own files; scanner state is mutated by its owning
source view. No shared compiler workaround was added. Earlier experiment
artifacts are not included in the final PASS counts. Completed obsolete scratch
ELF artifacts totaling 286,481,910 bytes were reclaimed, keeping source and logs.

Fresh setup: Go 1.27.1 ready 0s, clang 20.1.8 ready 0s, Node 24.19.0 ready 0s,
submodules 0s, cache 63s, total 63s, nproc 5 with a four-core quota.
Cloud-environment runtime skill confirmed enforced network policy and current
runtime readiness after resume; no secret values were read. Fetched main
39638d9e2 and integrated lint d65a8f931 remain ancestors of the implementation.
No rebase was necessary in this turn. Only the own branch is pushed.

Commands, with /workspace/adamic-tools/env.sh sourced and output to files:

```sh
bash cloud/setup.sh > /tmp/wave20-constructed-setup.log 2>&1
python3 stage1/cohere/typeaware/react-jsx-no-constructed-context-values/verify.py --artifacts /workspace/wave20-validation/constructed-green --compiler /workspace/wave20-validation/area-third/adamic --checker /workspace/wave20-validation/area-third/checker.a --sanitized-checker /workspace/wave20-validation/area-third/checker-asan.a > /tmp/wave20-constructed-green.log 2>&1
python3 stage1/cohere/typeaware/react-jsx-no-constructed-context-values/mutants.py --artifacts /workspace/wave20-validation/constructed-mutants --controls /workspace/wave20-validation/constructed-expanded --compiler /workspace/wave20-validation/area-third/adamic --checker /workspace/wave20-validation/area-third/checker.a > /tmp/wave20-constructed-mutants.log 2>&1
(cd cohere && go test ./internal/lint/rules/react -run '^TestJsxNoConstructedContextValues' -count=1 -v > /tmp/wave20-constructed-production.log 2>&1)
/workspace/wave20-validation/constructed-lint/lint /workspace/adamic/tsconfig.json /workspace/wave20-validation/constructed-lint/manifest > /tmp/wave20-constructed-lint-final.log 2> /tmp/wave20-constructed-lint-final.stderr
```

The formatter overlay command and its logs are archived alongside complete
stdout/stderr, fixture records and source hashes. Final exact gate outputs:

```text
PASS 171 controls 97 findings; complete Go/native/sanitizer bytes
KILLED construction-kind mutant: compiled, exit 0, empty stderr, Go bytes differ
PASS repository 18485 complete corpus bytes normal/sanitized
PASS compiler 5241 complete corpus bytes normal/sanitized
PASS released MemoView binding-origin query panics 70
PASS released MemoView type-shape query panics 70
PASS released MemoView resolved-callee query panics 70
KILLED retained-handle mutant binding-origin compiled, exit 0; required panic 70 catches it
KILLED retained-handle mutant type-shape compiled, exit 0; required panic 70 catches it
KILLED retained-handle mutant resolved-callee compiled, exit 0; required panic 70 catches it
TIMING repository {'native': 0.3652399529964896, 'go': 0.26658795299954363}
TIMING compiler {'native': 2.0225803690009343, 'go': 0.52274923600271}
PASS complete analysis controls, both corpora, semantic mutant, sanitizer and released question; shared checker integration pending
KILLED missing-list case-161 compiled, exit 0, empty stderr; only Go finding bytes differ
KILLED render-escape case-140 compiled, exit 0, empty stderr; only Go finding bytes differ
KILLED helper-escape case-136 compiled, exit 0, empty stderr; only Go finding bytes differ
KILLED factory-identity case-146 compiled, exit 0, empty stderr; only Go finding bytes differ
PASS four compiling memo/escape mutants
```

## No unclaimed work remains

After pushing all three JSX source analyses, fetched all origin heads again.
Main 39638d9e2 and lint d65a8f931 are unchanged and remain ancestors. The audit
inspected 626 origin refs and 33 distinct Markdown claim blobs. The by-volume
inventory has 197 rules: 172 claimed names and 25 baseline checker-dependent
ports. No unported, unclaimed entry remains. No new claim was made.

An initial boundary regex incorrectly treated the colon after a rule name as
part of the name. Manual inspection of wave-01.md caught all three apparent
candidates before any claim: nexus/correctness-no-implicit-return,
@typescript-eslint/no-deprecated, and no-else-return. Their list entries end in
colons before the finding counts. Corrected matching retains slash and hyphen
boundaries (so core names are not confused with namespaced rules) and admits
colon punctuation. All three are skipped as already claimed. Compressed full
selection evidence is beside the oracle outputs. Only the own branch was pushed.
