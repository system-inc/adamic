# Stage1 JSX parser and lint integration

Branch `codex/stage1-jsx-lint`, from origin/main
`d090af531216ddd3c25a0dede6b82d7c0a6edf76`. No compiler hot files changed.
The batch-8 runner and tests are brought from published commit
`4189abd3490757e8abe13722ceb365c451293e92` because main did not contain them.
Its original default five-rule runner remains separate. There is no PR.

Implementation commit: `e715ef4a2f898230af63c40195dea6586a557899`. The following
report/evidence commit contains the frozen-source logs and SHA-256 manifest.

## Missing behavior and implementation

The existing scanner already implemented JSX text scanning, including
`LessThanSlashToken`, whitespace classification and raw text. The parser never
entered that mode. It interpreted `<` only as a type assertion and built none
of the JSX node family. The scanner also lacked JSX name extension and quoted
attribute-value mode.

The new `jsx.ts` descent follows typescript-go's JSX parser. It builds all
thirteen AST kinds: elements, self-closing/opening/closing elements, fragments
and their opening/closing nodes, attributes and their container, spread
attributes, expressions, text and namespaced names. Whitespace-only text is a
`JsxText` node with a semantic flag, not a separate AST kind, exactly as in Go.

The scanner extends already scanned identifiers with dash name parts. Attribute
strings preserve backslashes, entities and line endings; they do not use
JavaScript escape cooking. JSX text scanning is entered after `>` and after
child-expression `}`; a closing tag resumes either JSX text or ordinary scanning
according to its parent context. Names include keywords, namespaces, dotted
members and `this`. TypeScript JSX accepts type arguments, including nesting and
trailing commas. JavaScript JSX does not consume type arguments.

Dashes are extended in standalone and namespace names, but not after a member
dot. The Go scanner also converts hash-prefixed tag/attribute names into ordinary
identifiers; those accepted forms are preserved. Private names after a member
dot and Unicode escapes in JSX names are explicitly rejected, as Go diagnoses
them. This does not reproduce Go's diagnostic messages or recovered invalid tree.

File extension selects the same TS/TSX/JS/JSX language variant as the Go oracle.
JS, JSX, MJS and CJS use JSX mode; TS and TSX differ. JSX generic arrows require
a comma, a default or a real constraint. This prevents a JSX tag from being
misclassified as a generic arrow. The existing ordinary TS assertion and generic
arrow behavior remains covered by its inherited tests.

Nodes retain numeric child indexes with no owning cycles. Their preorder answer
retains Go's child order, kind, UTF-8 byte position/end, cooked identifier and
literal text, raw template text, optional/literal flags and existing declaration
semantics. New JSX metadata includes child/attribute counts, text whitespace
flags and type-argument count/trailing-comma state. Node-list source ranges,
general parser context/error flags and binder fields are outside the existing
slice's canonical protocol.

The original cohere fixture `<></* valid *//>` is malformed JSX even though it
is a clean rule case. Go recovers its missing closing-fragment delimiter without
consuming the next token, then inserts a zero-width operand at EOF. The port
matches this recovered tree. `--jsx-recovery` permits Go's diagnosed JSX fixture
trees only in the explicit upstream replay; ordinary generated inputs still
require no Go parse diagnostics. This is not general recovery or diagnostic
parity. Other unsupported malformed JSX stops explicitly.

## Direct lint traversal

The batch-8 runner now always calls `parser.file()`. The Go-boundary injection
path and the fourteen-case skip ledger are removed. The JSX comment-text listener
visits real `JsxText` nodes. Parent indexes give the React call listeners normal
ancestry inside JSX expressions. All five selected React listeners are wired
through this tree; this unit does not add cohere's other unported JSX rules.
Published batches 4, 5 and 6 were inspected and had no additional JSX listeners.

All 714 original batch-8 source/rule/file/options combinations are replayed.
That includes all 54 JSX-bearing cases: forty formerly injected listener cases
and fourteen formerly excluded no-find-dom-node/no-is-mounted cases. Original
Go assertions run before capture. Canonical lint comparison retains human
findings, byte ranges, message IDs, all automatic edits, every suggestion's ID,
description and ordered edits, and complete converged fixed sources.

## Tests and mutants

The independent typescript-go oracle uses a Go build overlay and no modified
parser or rule. The cohere pin is
`715ba94f3608a6500086b1076ce5cb7e51b836db`; TypeScript 6.0.3 compiler sources are
pinned at `050880ce59e30b356b686bd3144efe24f875ebc8`.

`TestJsxNode` checks every kind in the Go-provided JSX AST inventory. The 348
generated files cover name/attribute/child combinations, nested fragments,
spread attributes/children, empty expressions and comments, quoted multiline
attributes, Unicode and CRLF positions, embedded regex/templates, JSX-valued
attributes, generics and arrow disambiguation. `TestJsxNative` compares them
under ASan/UBSan with leak checking and separately in release. The original 54
cohere JSX fixture trees are also compared in full by `TestJsxLintTrees`.

Every mutant has a matching original control first. Eleven mutants must compile,
exit successfully with no stderr on Node and sanitized native, then disagree
with Go's tree. Three more remove a Go-backed refusal: their original programs
exit 70 identically on Node/native, Go diagnoses the input, and each permissive
mutant must compile and exit cleanly with an AST. The fourteen new mutants are:

| Mutation                                   | Comparison that catches it |
| ------------------------------------------ | -------------------------- |
| Append a character to text payload         | Exact JsxText text         |
| Lose whitespace-only flag                  | JsxText semantic field     |
| Emit QualifiedName for namespace           | Node kind                  |
| Emit opening element for self-closing      | Node kind                  |
| Lose trailing type-argument comma          | Type-argument metadata     |
| Increment attribute count                  | JsxAttributes list count   |
| Emit parentheses for JSX expression        | Node kind                  |
| Reverse element children                   | Preorder child order       |
| Shift text start by one                    | UTF-8 byte range           |
| Append a character to raw attribute string | Exact StringLiteral text   |
| Append a character to JSX name             | Exact Identifier text      |
| Extend dashes after a member dot           | Go-backed rejection        |
| Accept a private member name               | Go-backed rejection        |
| Accept a Unicode escape in a JSX name      | Go-backed rejection        |

Six negative inputs cover the dashed member, private member and escaped tag,
member, namespace and attribute names. Accepted hash-prefixed tag, namespace
and attribute names belong to the positive tree corpus.

The twelve batch-8 family/repair mutants and inherited parser/scanner/lint
mutants are also rerun. The extra repair mutants change only a second suggestion
ID and an automatic edit payload. The external uncached one-byte oracle check
demonstrates fresh native and Node observations.

Each batch-8 mutant compiles and finishes cleanly on Node and sanitized native;
the exact Go finding/fix/suggestion comparison kills it after its positive control:

| Rule family                                | Mutation                             | Field that catches it                    |
| ------------------------------------------ | ------------------------------------ | ---------------------------------------- |
| no-octal-escape                            | Disable octal-width report           | Missing finding                          |
| no-unexpected-multiline                    | Invert line comparison               | Missing finding                          |
| no-unused-private-class-members            | Invert read test                     | Missing finding                          |
| no-useless-constructor                     | Skip every constructor body          | Missing finding                          |
| prefer-template                            | Invert concatenation test            | Missing finding                          |
| react/forward-ref-uses-ref                 | Invert parameter-count test          | Missing finding                          |
| react/jsx-no-comment-textnodes             | Invert line-start test               | Missing finding from a real JsxText node |
| react/no-find-dom-node                     | Invert callee text test              | Missing finding                          |
| react/no-is-mounted                        | Invert member text test              | Missing finding                          |
| react/no-redundant-should-component-update | Invert pure-base test                | Missing finding                          |
| react/forward-ref-uses-ref                 | Change second suggestion ID only     | Suggestion ID                            |
| prefer-template                            | Add a space to automatic replacement | Edit payload and fixed source            |

The earlier complete parser run also killed precedence corruption, lost optional
chain flags, parentheses misclassified as arrows, for-of changed to for-in,
lost type-only import phase and keyof changed to readonly by exact tree bytes.
Its two zero-contribution counters retained identical AST bytes but failed
Go's selected-tree count (18) and whole-tree count (12). The final scanner run
killed `!=` changed to `==`, an accepted invalid decimal separator and a skipped
regex rescan by exact token/diagnostic bytes. The final inherited five-rule lint
gate also exercises suppressed duplicate cases, reported empty function bodies
and suggestions applied as automatic fixes, caught by its Go canonical output.
The retained logs record every compiled Node/native observation and first diff.

## Commands and observations

Toolchain: `bash cloud/setup.sh`, followed in every toolchain shell by
`source /workspace/adamic-tools/env.sh`. Setup printed Go 0s, clang 0s, Node 0s,
submodules 0s, build-cache warm 125s, total 125s. `nproc` is 5; CPU quota is
four CPUs, reported memory 17.6 GB; Intel Xeon Platinum 8573C. Go 1.27.1,
clang 20.1.8, Node 24.19.0.
Tests write their output to files, never pipelines.

```sh
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/lint-batch2/typescript-6.0.3
TIMEFORMAT='wall %3R seconds user %3U seconds sys %3S seconds'
{ time go test ./stage1/typescript/parser -count=1 -v -timeout 30m; } > /tmp/jsx-parser-final.log 2>&1
{ time go test ./stage1/typescript/parser -run '^TestJsx(NameBoundaryRejections|MemberNameRejection|Node|Native)$' -count=1 -v -timeout 30m; } > /tmp/jsx-boundary-check.log 2>&1
{ time go test ./stage1/typescript/parser -run '^TestJsx(ScannerMutants|Mutants)$' -count=1 -v -timeout 30m; } > /tmp/jsx-mutants-final.log 2>&1
{ time go test ./stage1/typescript/scanner -count=1 -v -timeout 30m; } > /tmp/jsx-scanner-complete.log 2>&1
{ time ADAMIC_LINT_BENCH=1 go test ./stage1/cohere/lint -count=1 -v -timeout 30m; } > /tmp/jsx-lint-complete.log 2>&1
go vet ./... > /tmp/jsx-vet-complete.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run TestTheOracleCatchesOneByte -count=1 -v -timeout 10m > /tmp/jsx-external-oracle.log 2>&1
gofmt -l cmd internal stage1/typescript/parser stage1/typescript/scanner stage1/cohere/lint > /tmp/jsx-gofmt-complete.log 2>&1
/workspace/scratch/lint-batch2/cohere --no-fix --no-cache stage1/typescript/parser/jsx.ts stage1/typescript/parser/parser.ts stage1/typescript/parser/lookahead.ts stage1/typescript/scanner/scanner.ts stage1/cohere/lint/batch8/*.ts > /tmp/jsx-source-complete.log 2>&1
```

Raw outputs are retained under [jsx_evidence](jsx_evidence/). The source manifest
fingerprints all 28 parser, scanner and batch-8 TS files, and its verification
log records every entry as `OK`. A second manifest fingerprints all 26 Go
verification files. Logs named `probe` or `before` document earlier observations
and failures; they are not substituted for final-source checks.

The complete parser package passed before the final JSX name-guard corrections:
Go reported 434.153s, wall 440.316s (7m20s). This includes all 77 compiler files
held as whole trees (44,766,682 bytes), 1,676 generated expressions, 68 generated
whole files, the 42-kind type inventory and eight inherited parser mutants.
It is retained as `jsx_evidence/parser-full-before-name-guards.log`; it is not
claimed as a final frozen-source complete parser gate. After those corrections,
all JSX tests were rerun in the two filtered commands above. Their union checks
348 clean-source trees (277,842 bytes), six refusals and all fourteen new mutants
on the final implementation.

The final boundary/tree command passed in 113.819s wall (Go 102.114s). The final
eleven tree/scanner mutant command passed in 124.834s wall (Go 123.003s).
The three refusal mutants are in the boundary/tree command, making fourteen.
All mutant binaries compile and finish cleanly; their semantic comparisons,
rather than compiler errors or sanitizer crashes, catch the mutations. The final
complete scanner package passed in 53.759s wall (Go 51.350s): 77 compiler files,
122 stage1 files, 18,236 generated inputs and its three inherited scanner mutants.
Repository-wide vet and gofmt exited zero with empty output. The exact changed
TS source lint/format/type gate exited zero: 276 rules, 26 checked, 100% ready.
The uncached external one-byte oracle passed in 3.712s with one native cache miss
and one Node cache miss; neither observation was served from cache.

The final complete lint package passed: Go 397.964s, wall 401.852s (6m42s),
user 581.427s, system 62.117s. It includes all twelve batch-8 mutants, all original
JSX tests and the inherited five-rule suite with its three mutants. No completed
package gate exceeded ten minutes. The earlier complete parser gate was 7m20s;
the final scanner gate was 54s and the two final JSX parser commands were 1m54s
and 2m05s. The complete lint log is `jsx_evidence/lint-final.log`.

The final batch-8 comparisons retain 12,274,406 identical finding/fix/suggestion
bytes across all 714 upstream cases. The 199 source files produce 12,369,785
identical bytes in sanitized and release builds; fourteen boundary/position
cases produce 12,303 bytes in both builds. All 54 original cohere JSX trees
match on Go, Node and sanitized native (45,527 bytes); running all ten rules
on those sources also matches in release (25,037 finding/fix/suggestion bytes).
The final lint C artifact is 2,440,147 bytes; its ordinary sanitized build's
clang step took 17.112s and its release clang step took 3.738s.

## Findings per second

Release binaries use normal `-O2`, without sanitizers. Each best-of-five sample
is a fresh process after canonical output comparison. Timing includes startup,
reading, scanning, parsing and visitors; it excludes compilation, rendering and
fix application. JSX rounds rotate Go/native/Node order. All count samples must
match Go, and the JSX count must be positive. Full fixes/suggestions are checked
separately before count-only timing.

| Workload                              | Findings | Go                     | Native                 | Node                 |
| ------------------------------------- | -------- | ---------------------- | ---------------------- | -------------------- |
| Original JSX fixtures, 54 files       | 27       | 3,513.23/s (0.007685s) | 5,185.83/s (0.005206s) | 173.59/s (0.155541s) |
| Compiler, 77 files, ten batch-8 rules | 161      | 465.98/s (0.345508s)   | 69.64/s (2.311740s)    | 127.27/s (1.265069s) |

The small JSX workload is sensitive to startup and file-cache effects. Native
is faster on that sample but substantially slower on the compiler workload;
these observations do not establish that native cohere is faster overall.

The inherited five-rule compiler baseline separately counted 451 findings:
Go 1,490.74/s (0.302535s), native 298.29/s (1.511974s), Node 402.88/s
(1.119428s), also best of five fresh processes.

## Limits and development observations

No type checker, binder, JSX emit/React transform, or new unported JSX lint rule
is claimed. The parser reads JSX syntax for analysis; compiling arbitrary JSX
programs as Adamic application code remains a different compiler unit. This unit
does not implement full malformed-source recovery, parser diagnostic equivalence,
or all TypeScript conformance fixtures. Ordinary trees and fixes are never
normalized after disagreement. All original batch-8 fixtures are included.

Early probes corrected the oracle's JsxText accessor and a generated raw `>`
probe that Go diagnoses. The latter became diagnostic-free entity text in the
generated corpus. Original diagnosed upstream inputs remain unchanged. Source
lint caught an alias and a fallthrough warning; the preliminary gate was stopped
before fixing them. A final name-boundary probe caught permissive dotted dash
names. Its first negative corpus incorrectly assumed Go rejects hash-prefixed
namespace and attribute names; the Go result disproved that assumption. The
scanner and positive corpus now preserve those forms. The earlier full parser
gate is retained as earlier evidence; final JSX and scanner gates validate the
subsequent name-boundary changes.
Automatic approval review rejected two bulk edit commands; explicit context
patches applied the authorized changes without losing existing code.

The full repository test suite is not claimed. The earlier complete parser
package and final JSX subset are distinguished above; final complete scanner
and lint packages, repository-wide vet and the named external oracle are run.


## Variant gate diagnosis, October 7, 2026

Built: matched release profiles and a gate-off mutant on codex/jsx-variant-gate, from a8a62d62ca49db7415e14c3887dd305022b17309. No scanner or parser implementation change.
Commits: a8a62d6 is the measured implementation; this evidence is a separate report commit, not a parse-speed fix or an achieved variant-gate unit.
Commands and outputs: 77 compiler trees, 348 generated JSX inputs and 54 upstream JSX fixtures match Go; the baseline is 8,997,591,885 Ir.
Mutant: evaluating JSX arrow lookahead unconditionally adds 386,910,853 Ir while preserving every compiler/JSX tree byte; the zero-entry assertion catches it.
Not covered: reproduction of the reported 10.00G run, a new once-at-entry implementation, the 8.96G threshold, or a full repository/native/oracle package gate.

### Finding that blocks the proposed change

At this requested base, Scanner.scan has no language-variant test and calls none
of scanJsx, scanJsxIdentifier or scanJsxAttributeValue. Those methods are entered
only from JSX descent. Parser's constructor already computes the variant once
from the filename. Its assignment method short-circuits JSX arrow speculation
with `!this.jsx || jsxArrowAhead(this.scanner)`; unary descent enters JSX only
for LessThanToken with this.jsx true. All five JSX entry counters below are
zero on the pinned 77 TypeScript compiler files. Adding a scanner-entry flag
cannot remove JSX scanning work that is already absent.

The Go implementation also stores the variant at initialization, but its scanner
checks it in the less-than token branch (scanner.go, LanguageVariantJSX), and its
parser checks it during arrow disambiguation. Selecting a variant once does not
mean Go contains no token-dependent variant checks. Go uses JSX variants for
.js, .jsx, .mjs, .cjs and .tsx, not just .jsx/.tsx. Restricting the existing
extension selection to the latter two would break the inherited JSX corpus.

The requested 12% increase is not reproduced with either available parse-only
driver. Under a matched compiler/runtime, the source-only JSX delta is
10,563,753 Ir, or 0.1175%. Of that delta, string-equality self costs rise
5,458,593 Ir; the rest is 5,105,160 Ir across generated code and runtime work.
There is no measured JSX scanner execution to attribute the delta to.
No counterfactual claim is made about an unidentified 10.00G artifact.
The driver/command for that run was requested before editing production source.

### Exact native build and measurement configuration

Every native instruction number in this section uses clang 20.1.8 and the
following release command: -O2 -g, sanitizers off, -ffp-contract=off,
-fno-optimize-sibling-calls, no LTO, no ADAMIC_COUNT and no entry probes.
The warning flags are also retained exactly. All variants use a8a62d6's stage-0
compiler and runtime; no runtime-area or codex/parse-speed fix is imported.

```sh
source /workspace/adamic-tools/env.sh
go build -o /workspace/scratch/jsx-variant-gate/adamic ./cmd/adamic > /tmp/jsx-variant-compiler-build.log 2>&1
/workspace/scratch/jsx-variant-gate/adamic c /workspace/scratch/parse-speed/batch8/parse.a > /workspace/scratch/jsx-variant-gate/before.c 2> /tmp/jsx-variant-before-emit.log
clang -std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -O2 -g -I internal/native/runtime /workspace/scratch/jsx-variant-gate/before.c internal/native/runtime/*.c -lm -o /workspace/scratch/jsx-variant-gate/before > /tmp/jsx-variant-before-clang.log 2>&1
VALGRIND_LIB=/workspace/scratch/parse-speed/valgrind/usr/libexec/valgrind /workspace/scratch/parse-speed/valgrind/usr/bin/valgrind --tool=callgrind --callgrind-out-file=/workspace/scratch/jsx-variant-gate/before.callgrind /workspace/scratch/jsx-variant-gate/before --manifest /workspace/scratch/parse-speed/compiler.txt --count > /workspace/scratch/jsx-variant-gate/before.stdout 2> /workspace/scratch/jsx-variant-gate/before.stderr
```

The accepted prior parse-only driver is batch8 4189abd's driver with visit
removed, previously reconstructed for codex/parse-speed. It includes file reads,
parsing, parent mapping, line tables and cleanup; count output is `0` because
rule traversal is absent. Its imports are absolute references to the checked-out
parser/scanner. The second driver copies this branch's batch8 .ts files to
scratch .a files, rewrites their imports and removes only visit(context, root).
The pre-JSX control copies parser/scanner .ts files from e715ef4's parent to
scratch .a files and rewrites the accepted driver's imports to that tree. All
retained application modules are copied too, so their Parser types refer to the
same isolated tree. No production .ts source is edited during these controls.

| Variant, with the exact release flags above | Ir |
| --- | ---: |
| Pre-JSX sources, matched compiler/runtime and accepted driver | 8,987,028,132 |
| a8a62d6 sources, accepted driver | 8,997,591,885 |
| a8a62d6 sources, updated batch8 driver | 8,997,570,749 |
| Existing JSX lookahead gate disabled, accepted driver | 9,384,502,738 |

Even the matched pre-JSX control exceeds 8.96G by 27,028,132 Ir. Neither a
restored 8.96G count nor a 10.00G baseline is claimed. All four count stdout
files contain exactly `0\n`. Callgrind emitted its brk-segment limitation notice
on these large processes; each completed successfully and produced a profile
whose disjoint self costs reconcile exactly to its summary.

### Gate-off mutant and independent entry counters

The scratch mutant changes only Parser.assignment:

```text
before:
if((!this.jsx || jsxArrowAhead(this.scanner)) && arrowAhead(this.scanner, allowReturn)) {
mutant:
const jsxArrow = jsxArrowAhead(this.scanner);
if((!this.jsx || jsxArrow) && arrowAhead(this.scanner, allowReturn)) {
```

The result is still ignored outside JSX, but its speculation executes. This
keeps syntax/trees unchanged and isolates the gate's performance value.
Sanitized native matches Go on 77 compiler trees (44,766,682 bytes) and 348 JSX
inputs (277,842 bytes); no stderr or sanitizer/leak error occurs. The instruction
check detects the 386,910,853 Ir rise, separately from those passing tree checks.

Separate generated-C entry probes increment an unsigned counter at each of the
five named function definitions and print counters after normal main cleanup.
Their O2 builds use the same warning/runtime flags, but are never used as Ir
measurements. The five-function zero-entry assertion passes on the control and
fails for the mutant's 189,810 lookahead calls. Other entries remain zero:

| Entry on the 77 files | Control calls | Gate-off mutant calls |
| --- | ---: | ---: |
| jsxArrowAhead | 0 | 189,810 |
| Scanner.scanJsxIdentifier | 0 | 0 |
| Scanner.scanJsxAttributeValue | 0 | 0 |
| Scanner.scanJsx | 0 | 0 |
| Parser.jsxParser | 0 | 0 |

### Correctness controls and retained evidence

```text
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 go test ./stage1/typescript/parser -run '^Test(WholeCompilerAgrees|JsxNode|JsxNative)$' -count=1 -v -timeout 30m
PASS, 48.896s: 77 files / 44,766,682 identical tree bytes; 348 JSX inputs / 277,842 bytes on Node, sanitized and release native.
go test ./stage1/cohere/lint -run '^TestJsxLintTrees$' -count=1 -v -timeout 30m
PASS, 49.134s: all 54 original cohere JSX trees / 45,527 identical bytes.
go test -overlay=/workspace/scratch/jsx-variant-gate/test-overlay.json ./stage1/typescript/parser -run '^TestJsxVariantGateOffEvidence$' -count=1 -v -timeout 30m
PASS, 12.228s: gate-off sanitized native, 77 compiler and 348 JSX files, identical Go tree bytes.
go vet ./stage1/typescript/parser ./stage1/typescript/scanner
exit 0, empty output.
```

The overlay adds a scratch-only Go test which uses the existing compilerManifest,
jsxManifest, goOracle, execute and difference helpers. It compares the scratch
mutant's parser/main.a binary with the independent Go oracle in whole-tree mode.
That binary uses the warning/runtime flags above with -O1 -g,
-fsanitize=address,undefined and -fno-sanitize-recover=all; Linux LeakSanitizer
is enabled by the sanitizer runtime. No verification helper is weakened.

Setup: Go/clang/Node/submodules ready at 0s, cache warm and done at 81s;
nproc 5, cpu.max 400000 100000, memory 17.6 GB. Every build/test shell sources
/workspace/adamic-tools/env.sh. Compiler corpus pin remains
050880ce59e30b356b686bd3144efe24f875ebc8; cohere remains 715ba94.

Full raw profiles, generated C, mutant copies, counter outputs and overlay test
are under /workspace/scratch/jsx-variant-gate. Test logs are
/tmp/jsx-variant-base-parity.log, /tmp/jsx-variant-gate-off-parity.log and
/tmp/jsx-variant-upstream-parity.log. Entry assertions are in
/workspace/scratch/jsx-variant-gate/gate-check.log. SHA-256 identities:

```text
compiler.txt 212a132782f06d00d57521cc80c7d0cad9e1896e9be0356af1524457f9800e3e
accepted parse.a 9aa97951a5d2d9ac956badc8e8c0306716814096a8e86312b08395fa81a6e5ea
before.c f0740c932a267902b7763cd64dc2fe035c14b1cd00e302826dba47602a57685c
current-driver.c 941a241816cf5590094477a894d588ba547767aaba6d6c486cfd165c8328259f
pre.c acb1f7bec3bc0552ff0cdd1b39b2dbbb41fca9ff32dd02747433011f7340c743
gate-off.c 8907f3a95d26954582ab63140aedfdf4f4d5c905df76753a612d03667f4b7c53
```

This branch records diagnosis only. It does not touch codex/parse-speed, runtime,
emission, lint visitors or any parser/scanner implementation file. A new source
fix needs the mismatching baseline artifact or another measured mechanism;
adding a redundant flag is not evidence of the requested speedup.
