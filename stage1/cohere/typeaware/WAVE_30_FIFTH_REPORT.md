Built: parked the prior three React analysis claims and ported independently testable components for the next three JSX rules; full JSX rule ports are blocked.
Commits: parking and claim 1f12ffd47 was pushed before code; native components and probes 1d023cc6c; current main base remains f8013f0b.
Commands and outputs: component gate PASS 12.875s; prerequisite probe PASS 9.001s; vet PASS; setup 25s, nproc 5; no full new-rule corpus comparison is claimed.
Mutants: qualified fragment object, custom-element name exclusion and function-declaration remedy all exit 0 and are caught only by production-Go byte comparison.
Not covered: complete findings/spans, fixes/suggestions and compiler/repository parity for these three rules, complete binding/memo-stability analyses, shared visitor integration or full-rule timings.

Ahra explicitly authorized parking react-hooks/preserve-manual-memoization,
react-hooks/purity and react-hooks/refs when native React high-level IR,
single-assignment or capture analysis blocks them. These claims now say PARKED
and name native React HIR/SSA, reactive scopes/memoization and capture/render
analysis as their blockers. Their previously tested components remain pushed.
Analysis work is tracked on #dnv6f2c and JSX integration on area/stage1-lint.
No main or area/ branch is a push target.

The unit's only pushed branch remains rebased on current main f8013f0b. Its
existing ten gates, all 21 mutants, bridge and fifteen filtered Node fixtures
already passed on this unchanged base in 79fe15332. New work is isolated to
owned JSX component directories and private wave-30 test files; no existing
runtime, parser, bridge, shared harness or registration generator changed.

Selection snapshot validation-wave-30-fifth/selection.json covers all 529 origin
refs. The ranking is independently reconstructed from the compiler-all.counts
and repository-all.counts linked by VOLUME_REPORT.md and matches its 197 entries.
Twenty-five ranked rules are already ported on main/c-library. Markdown claim
documents mention 154 ranked rules; the union leaves 18 available candidates.
Inventory JSON stored under claims/ is explicitly excluded as evidence of
ownership: wave-26-refresh.json lists all 197 rules but claims none merely by
listing them. The initial scan treated those inventory mentions as claims; the
corrected scan distinguishes inventory from reservation and also checks the
selected short names in all actual claim documents.

The first three available candidates are react/jsx-fragments,
react/jsx-no-constructed-context-values and react/jsx-no-undef. They do not import
the parked native React HIR/SSA analysis pipeline. Their claim was committed and
pushed before any implementation. No additional rules are claimed.

Native components, all authored as .a:

- JSX fragments: qualified React.Fragment matching checks numeric identifier
  kinds and exact identifier text on the handed FragmentNameNode. It does not
  implement bare-name import/destructuring alias resolution or whole findings.
- JSX no-undef: component-name classification reproduces the production
  isComponentName predicate for empty, ASCII/lowercase, underscore/dollar,
  non-ASCII and dash-containing names on the handed ComponentNameNode. It does
  not implement symbol scope, file-local declaration or allowGlobals judgments.
- Constructed context values: all ten construction categories, named/unnamed
  messages and useMemo/useCallback remedies operate on a handed ConstructionNode.
  ASCII quoted names reproduce Go escaping in the exercised cases. Non-ASCII
  named messages explicitly panic NotYet rather than silently formatting them
  incorrectly. Construction detection and memo-input stability are unimplemented.

Each own directory contains rule.json with the full rule's kinds and corresponding
pinned numericKinds. Fragments listen to JsxFragment 289, JsxElement 285 and
JsxSelfClosingElement 286; the other two listen to JsxOpeningElement 287 and
JsxSelfClosingElement 286. Metadata says partial components only and identifies
the shared blocker. It is not registered with the absent shared registry, and
these components are not represented as completed visit implementations. All
component functions use the fact object supplied by their caller; none retrieves
parser nodes, dispatches other rules or compares syntax-kind strings. Names and
construction labels appearing in messages are text, not syntax dispatch.

The independent oracle is an overlay test in the unmodified production Go React
package. It calls jsxFragmentsNameIsFragment on factory-built property accesses,
isComponentName on 1034 name controls, and jsxNoConstructedContextValuesMessage
on 140 construction/name controls. No bridge or copied Go lint judgments supply
the expected output. The 48 qualified-fragment controls produce 1215 bytes,
component names 12037 bytes and construction messages 46522 bytes. Native,
ASan/UBSan/LSan native and emitted JavaScript on Node match all bytes. Numeric
metadata is checked against the pinned Go AST constants.

Each component has a successful-exit native mutant with empty stderr:

| Component | Mutation | First differing byte |
| --- | --- | ---: |
| qualified fragments | React object name changed to Other | 98 |
| component names | dash exclusion changed to underscore | 22 |
| construction messages | function-declaration code changed to class-expression code | 13844 |

Only Go comparison catches those differences. They prove these components,
not the three complete source rules. A separate unsupported-name probe proves
non-ASCII named message input refuses with native panic 70 and the exact NotYet
text. There are no new checker handles/questions, so the new components have no
released-handle surface; existing bridge safety proofs remain in the landing report.

The whole-rule prerequisite probe uses the independent production Go registry
and actual TSX source controls. All three Go rules produce findings, zero fixes
and zero suggestions. The unchanged native parser rejects all three controls:

| Rule | Native failure |
| --- | --- |
| jsx-fragments | expected GreaterThanToken, got SlashToken at 30 |
| jsx-no-constructed-context-values | expected GreaterThanToken, got Identifier at 42 |
| jsx-no-undef | expected GreaterThanToken, got SlashToken at 23 |

This passing probe establishes the shared prerequisite failure; it does not
claim Go/native source agreement. Main's parser cannot deliver JSX nodes and
still has string-valued ParseNode.kind. The shared numeric handed-node driver,
rule.json registration and finding-model integration are also not on this main
base. Producing whole-rule findings over the requested corpora is blocked until
those interfaces land. Shared files were not edited to work around their owners.
Native/Go whole-rule times are consequently unavailable; component gate durations
are validation costs, not rule-performance benchmarks.

Commands, each writing exclusively to its named log:

- ADAMIC_TOOLS=/workspace/adamic-tools bash cloud/setup.sh. Timing lines Go 0s,
  clang 0s, Node 0s, submodules 1s, cache warm 25s, total 25s; nproc 5, quota 4 cores.
- After source /workspace/adamic-tools/env.sh, ADAMIC_WAVE30_JSX_COMPONENTS=/workspace/wave-30-fifth-components
  go test ./stage1/cohere/typeaware -run '^TestWave30JsxComponents$' -count=1 -timeout 15m -v.
- ADAMIC_WAVE30_JSX_PREREQUISITES=/workspace/wave-30-fifth-prerequisites
  go test ./stage1/cohere/typeaware -run '^TestWave30JsxPrerequisites$' -count=1 -timeout 15m -v.
- go vet ./stage1/cohere/typeaware; git diff --check.

The new two tests pass; the prior ten gates were not needlessly repeated because
none of their code or compiler inputs changed. No full repository gate, full
oracle matrix, full new-rule corpus comparison or whole-rule mutant is claimed.
Full test logs and exact gzip stdout/stderr streams with SHA-256 indexes are
preserved beside the selection snapshot. Test-source TSX suffixes are symlinks to
.a source controls; no new authored Adamic .ts file was introduced.
