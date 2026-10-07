Built: no-throw-literal, no-useless-backreference and prefer-arrow-callback in separate .a files, with a dedicated raw/read symbol question.
Commits: previous twelve ports pushed at b651321e; claim ac619f54 pushed before code; implementation a11e7a79 pushed on codex/typeaware-wave-23.
Commands and outputs: dedicated gate PASS 103.894s; bridge PASS 74.544s; checker PASS 0.142s; filtered Node oracle PASS 16.650s; vet, formatting and whitespace logs empty.
Mutants: three rule mutants and the symbol-origin mutant exit 0 with empty stderr and fail Go bytes; question guard fails its checker assertion; released registry, seven foundation mutants and the Node byte mutant caught.
Not covered: full repository test gate, emitted-JavaScript execution of these checker-linked rules, malformed/error-recovery source, every compiler configuration, suppression/edit application or cohere CLI support for .a inputs.

The previous twelve claimed rules, tests and reports were already pushed before
this selection. Public fetch succeeded and the initial push retry confirmed the
branch was up to date. Fetch explicitly requested every origin head, because the
configured default refspec includes only main. Selection checked 347 fetched
origin refs, the same 26 production ports on main and codex/tsgo-c-library, and
126 claimed ranked rules across 33 distinct Markdown claim blobs. These were the
first three of 46 remaining entries in VOLUME_REPORT's combined descending count
ranking, with lexical ties. All original counts are zero. Already-claimed
candidates were skipped, including no-obj-calls, no-object-constructor and
no-promise-executor-return. No conflicting claim was found for these three.
[selection.json](validation-wave23-behavior/selection.json) preserves exact refs,
claim texts, exclusions and the remaining ranking. The claim push completed
before any implementation was written. No further rules were claimed.

Main remains ef3d907ecdc4c771b016f7d9c52372def057a340 and the bridge branch remains
5afbdb83da2ed7ad9815657cd3f6ececd5294bf6. Cohere remains
715ba94f3608a6500086b1076ce5cb7e51b836db, typescript-go remains
8d550c837c90bd1805b047b7eeccc2baac2d5e7a, and the external TypeScript v6.0.3
corpus remains 050880ce59e30b356b686bd3144efe24f875ebc8. Work continues on the
same branch, originally based on 0d540f4. No submodule pin changed.

[no_throw_literal.a](no_throw_literal.a) reproduces couldBeError's exact syntax
arms, including the asymmetric logical/assignment/comma cases and conditional
branches. Parenthesized values retain Go's judgment. The global-undefined test
uses declaration-file flags, accepting unresolved and declaration-less globals
and declining bindings declared in any source file. The finding covers the whole
throw statement and preserves both object and undef messages.

[no_useless_backreference.a](no_useless_backreference.a) scans regex structure in
[backreference_structure.a](backreference_structure.a). It preserves captures,
branch paths, lookaround direction, the precedence of all five diagnoses,
duplicate named-group resolution and message suffixes. Character classes and
pattern escapes obey the production scanner's flag-dependent widths. The same
limited pattern-validation decisions as Go reject unbalanced groups/classes,
unresolved Unicode references and unmatched Unicode quantifier braces. Findings
cover the whole literal or constructor/call expression.
[backreference_tracking.a](backreference_tracking.a) follows global RegExp and
global-object member paths through aliases, assignments, wrappers, conditionals,
logical/comma expressions, parameter defaults and object destructuring. It keeps
cycle guards and duplicate traversal results. Constant pattern and flag folding
follows same-file const or effectively constant bindings, plus strings,
templates, literals and string addition; unknown values decline.

[prefer_arrow_callback.a](prefer_arrow_callback.a) tracks this, super, new.target
and implicit arguments per function owner, preserving arrow inheritance. Named
self-references use raw symbol identity, so a shadowing declaration is distinct.
Both configurable flags and their defaults match Go. Callback detection looks
through Go's wrapper/bind shapes; the fixer separately checks adjacency. Repairs
preserve exact byte offsets, order and text, comment retention, bind removal and
required parentheses. Duplicate parameters, this parameters, unbound this,
async line breaks and unsafe bind/comment shapes retain the production declines.

The initial node-symbol-details question panicked in the compiler corpus while
rendering a computed declaration-parent name. The new question in
[callback_symbol_facts.go](../../../bridge/tsgo/checker/callback_symbol_facts.go)
and [callback_symbol_facts.a](callback_symbol_facts.a) avoids rendering names.
Its response is the existing version/mode header, raw opaque symbol ID,
declaration count, and one IsDeclarationFile flag per declaration. The optional
`read` variant returns the shorthand assignment's value symbol or local export
target; otherwise it returns GetSymbolAtLocation. Decisions stay in Adamic.
Unknown questions forward to the previous handler. Registration changes exactly
one line of bridge/tsgo/checker/facts.go. The private checker test compares 16
identifiers in both variants against direct checker facts, including computed
names, source and declaration-file origins, unresolved names, shorthand/export
targets and malformed-question rejection. No shared harness, registration
generator or protected compiler file changed.

The worker-owned .a runner uses Adamic's parser and the native C-linked bridge.
The independent Go oracle loads the same roots/configuration and runs unmodified
production cohere rules, with no bridge import. Complete sorted diagnostic bytes
include finding spans, IDs, messages, fix counts and ordered repairs, suggestion
counts and suggestion serialization. These production rules emit no suggestions;
that zero count is compared, not omitted. Default and all three nondefault arrow
option combinations run normally and under ASan/UBSan/LSan.

| Population | Roots | Findings | Fix records | Identical bytes |
| --- | ---: | ---: | ---: | ---: |
| Default controls | 326 | 200 | 128 | 75,027 |
| allowNamedFunctions true | 326 | 190 | 107 | 72,223 |
| allowUnboundThis false | 326 | 210 | 128 | 77,661 |
| Both flags changed | 326 | 199 | 107 | 74,594 |
| Compiler src/compiler | 77 | 0 | 0 | 5,318 |
| Frozen repository corpus | 287 | 0 | 0 | 18,485 |

The default controls include 22 throw findings (19 object, 3 undef), 117 regex
findings (22 nested, 14 disjunctive, 55 forward, 15 backward, 11 negative-lookaround),
and 61 arrow findings. They combine extracted production fixture tables with
additional shadows, alias/default/destructuring/cycle cases, global writes,
constant patterns/flags, computed property names, Unicode and CRLF. Of 328
submitted controls, two extracted entries containing legacy octal string escapes
fail Go parse validation; [excluded-controls.json](validation-wave23-behavior/excluded-controls.json)
preserves them. They are excluded before comparison and no error-recovery
coverage is claimed. Neither corpus has parse-filter exclusions. Both zero-finding
corpus results are independently verified; the positive controls hold the rule
judgments, fixes and message variants.

| Mutant | Evidence that catches it |
| --- | --- |
| Undefined throw uses the object message | Successful normal exit, empty stderr; Go byte mismatch at 56,021 |
| Backward backreference becomes forward | Successful normal exit, empty stderr; Go byte mismatch at 3,864 |
| Arrow insertion gains one trailing space | Successful normal exit, empty stderr; Go repair-byte mismatch at 334 |
| Every symbol declaration reported as a declaration file | Successful normal exit, empty stderr; Go byte mismatch at 39,245 |
| Callback question validation removed | Compiles; checker test fails with accepted malformed question |
| Released program retained in the handle registry | Mutant exits 0 with empty stderr; healthy probe must exit 70 with invalid-or-released-handle panic |

No rule-mutant build failure or crash counts as a kill. The bridge foundation
separately catches input/output length mutations with ASan, removed C-buffer
frees and heap-instead-of-region allocation with LeakSanitizer, retained released
entries with stale-handle assertions, source-file instead of node lookup with the
Go byte oracle at byte 6, and a removed link opt-in guard with the required refusal.
It compares 1,600 positions across four compiler files and 54,982 bytes under
ASan/UBSan/LSan, and checks 100 C ABI queries, output survival after release,
zero/stale handles and distinct second handles. Sanitizers instrument native/C
memory, not the Go heap. The filtered Node gate selects six fixtures, compares
native, Node and emitted JavaScript behavior, and catches its deliberate one-byte
oracle mismatch. That is compiler regression evidence, not emitted-JavaScript
execution of these checker-linked lint rules.

Three quiet alternating whole-process rounds, with no builds or tests active,
include program load, parsing, decisions, rendering and teardown. Complete finding
streams are compared on every round:

| Corpus | Native median | Go median | Native / Go |
| --- | ---: | ---: | ---: |
| Compiler | 2.084 s | 0.356 s | 5.85 |
| Repository | 0.260 s | 0.128 s | 2.02 |

Native query counts are 449 and 25 respectively. Native is slower on this machine;
these observations do not establish why. Build and sanitizer times are excluded.
[measurements.json](validation-wave23-behavior/measurements.json) retains every
round and phase log.

The successful setup from this session was reused: Go 1.27.1 ready 0s;
clang 20.1.8 ready 0s; Node 24.19.0 ready 0s; submodules 0s; cache warm 83s;
setup done 83s. The original setup log is preserved. nproc was rechecked: 5.
Build and test commands sourced /workspace/adamic-tools/env.sh. Every test output went to
a log file, never a pipe. Exact gates:

```sh
ADAMIC_WAVE23_BEHAVIOR_ARTIFACTS=/workspace/wave-23/behavior-final \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-23/typescript \
ADAMIC_WAVE23_COMPILER_MANIFEST=/workspace/wave-23/compiler.manifest \
ADAMIC_WAVE23_REPOSITORY_MANIFEST=/workspace/wave-23/repository.manifest \
go test ./stage1/cohere/typeaware -run '^TestWave23BehaviorAgreementAndMutants$' -count=1 -timeout=15m -v > /workspace/wave-23/behavior-final/gate.log 2>&1

ADAMIC_TSGO_CORPUS=/workspace/wave-23/typescript go test ./bridge/tsgo/... -count=1 -timeout=15m -v > /workspace/wave-23/behavior-final-bridge.log 2>&1
go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|functions|closures|sorting)\.a$' -count=1 -timeout=10m -v > /workspace/wave-23/behavior-node.log 2>&1
go test -overlay /workspace/wave-23/callback-guard-overlay.json ./bridge/tsgo/checker -run '^TestCallbackSymbolFacts$' -count=1 -v > /workspace/wave-23/behavior-callback-guard-mutant.log 2>&1
# The guard mutant must fail its checker assertion. It returned exit 1.
go vet ./... > /workspace/wave-23/behavior-vet.log 2>&1
gofmt -l cmd internal bridge/tsgo stage1/cohere/typeaware > /workspace/wave-23/behavior-gofmt.log 2>&1
git diff --check > /workspace/wave-23/behavior-diff.log 2>&1

python3 stage1/cohere/typeaware/validation-wave23-constructor/measure.py /workspace/wave-23/behavior-final/native /workspace/wave-23/behavior-final/wave23-behavior-oracle /workspace/wave-23/typescript /workspace/wave-23/behavior-bench /workspace/adamic > /workspace/wave-23/behavior-bench.log 2>&1
```

[validation-wave23-behavior](validation-wave23-behavior) preserves full compressed
outputs, matching hashes, root/config hashes, portable manifests, fixture sources,
selection provenance, mutants and logs. Fixture filenames are .a. The pinned
cohere CLI's .a extension refusal remains recorded in earlier reports; no new CLI
lint pass is claimed. Shared emitted-JavaScript/profile integration remains owned
by codex/lint-harness-dot-a. No PR was opened.
