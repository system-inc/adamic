Follow-up on codex/stage1-estree: all 92 output mismatches are fixed.
Acceptance disagreements fall from 984 to five proved raw-input refusals.
15,647 files, 549,618,641 bytes per build, match Go on all three port builds.
All 11,717 Go refusals reject on native; the full gate and 22 driver mutants pass.
UTF-8 loss and parser stalls have separate pushed commits; shared code is unedited.

Baseline: 27,369 frozen files, 14,699 identical, 861 refused Go answers, 123
accepted Go refusals, 92 output mismatches and 11,594 both refused. Refusal
includes Go's caught canonical-serialization/location panics, not just syntax
errors. Saved inventory and original answers are reused; these counts concern
that frozen corpus, not a new inventory of the added implementation files.

Setup: `bash cloud/setup.sh > /tmp/estree-followup-setup.log 2>&1` passed.
Timing lines: Go ready 0s; clang ready 1s; Node ready 1s; submodules ready 1s;
build cache warm 19s; done 19s on 5 processors. `nproc` = 5. Versions:
Go 1.27.1, clang 20.1.8, Node 24.19.0. Main and Go pins unchanged.

### JSX checkpoint

The local parser is a copy of main 5d4c801's parser, with imports redirected to
its unchanged scanner/statements/grammar/lookahead dependencies. JSX parsing
and conversion live inside the claimed ESTree directory. Shared parser,
compiler and runtime are untouched. Go's parse.go selects JSX for .tsx/.jsx and
TSX for .js/.mjs/.cjs; matching both was necessary to rescue all 523 files.

Source audit command:
`node --disable-warning=ExperimentalWarning testdata/corpus.mjs
/tmp/estree-baseline-go.jsonl /tmp/estree-jsx-all-corpus.jsonl
/tmp/estree-jsx-all-identical > /tmp/estree-jsx-all-audit.log 2>&1`.
It records 15,222 identical, 338 refused Go answers, 123 accepted Go refusals,
92 different outputs and 11,594 both refused: 461 acceptance disagreements.
No old matching file regressed, and no new wrong output/acceptance appeared.
The audit-only progress guard still bounds the unported recovery stalls.

Nine generated JSX files match 11,077 canonical bytes on source Node,
sanitized native and emitted JS, held to Go. Pinned original libraries prove
empty generic argument rejection, exact ampersand spelling differences after
postprocessing and the existing CRLF normalization difference. All other raw
and postprocessed JSX fields are compared. The extra selfClosing mutant
finishes normally on Node/native and is caught at line 23; the three original
port mutants are rerun at this checkpoint.

Retained logs under validation/followup/ provide commands and observed outputs.

Green commands: `ADAMIC_ESTREE_LIBRARY=/workspace/scratch/estree-library
 go test -count=1 -v ./stage1/cohere/estree > /tmp/estree-jsx-green-final.log 2>&1`
passed in 173.050s; all original checks, three original mutants and the extra
JSX mutant pass. `ADAMIC_ESTREE_CORPUS=/tmp/estree-jsx-rescued go test -count=1
-v ./stage1/cohere/estree -run '^TestRepositoryAgreement$' >
/tmp/estree-jsx-all-rescued.log 2>&1` passed in 24.955s: all 523 rescued files,
6,131,940 canonical bytes on source Node, ASan/UBSan native and emitted JS.
The initial library comparison failed on the documented deltas; the final gate
contains their exact proving checks. Cohere checks 21/21 sources, 100% ready.

### Decorated export checkpoint

JSX commit fe07100524fe56da82b15d8d8ff50358ff85c92d was pushed. The next
largest output group had 47 files whose first difference was an export wrapper.
Ignoring decorator children when reading modifiers and starting the wrapper at
the export keyword resolves 46 files; the remaining file has another difference.
The source audit is now 15,268 identical, 338 refused Go answers, 123 accepted
Go refusals, 46 different and 11,594 both refused. No old match regressed.

`go test -count=1 -v ./stage1/cohere/estree -run '^TestDecoratedExport' >
/tmp/estree-exports-gate.log 2>&1`: PASS 39.380s, six generated files and
successful wrapper-range mutant caught on source Node and sanitized native.
`ADAMIC_ESTREE_LIBRARY=/workspace/scratch/estree-library
ADAMIC_ESTREE_CORPUS=/tmp/estree-exports-rescued go test -count=1 -v
./stage1/cohere/estree -run '^TestRepositoryAgreement$|^TestDecoratedExportLibraries$'
> /tmp/estree-exports-rescued.log 2>&1`: PASS 20.277s, all 46 files and 251,078
bytes on all three port builds, six raw and postprocessed library agreements.
The source audit command is the JSX command with exports output paths.
Cohere reports 21/21, 100 percent ready. Logs and full dispositions are retained.

### Parser stall checkpoint

Export checkpoint 068218be92c9caa19e9ce981f178abf599371632 was pushed.
The local Parser.next now bounds repeated identical scanner positions at 32.
The audit injection has been removed entirely, so source census protection is
now the same parser used by the real driver and native build. The census stays
15,268 identical / 338 refused Go answers / 123 accepted Go refusals / 46
mismatches / 11,594 both refused. Exactly 13 real progress refusals remain:
seven Go-accepted and six Go-refused files. This fixes termination, not Go's
recovered AST or diagnostic equivalence for the seven accepted files.

`go test -count=1 -v ./stage1/cohere/estree -run
'^TestBoundedPortParser$|^TestPortStallControl$|^TestGeneratedAgreement$' >
/tmp/estree-stalls-gate.log 2>&1`: PASS 59.127s; 78 generated matches, three
minimal EOF stalls, and a disabled-guard control caught by the 500ms deadline
on source Node and sanitized native. That control is a liveness mutant, not
a successful wrong-output mutant. `go test -count=1 -v ./stage1/cohere/estree
-run '^TestBoundedPortParser$' > /tmp/estree-stalls-recorded.log 2>&1`: PASS
21.871s, all 13 original recorded stalls plus the three minimal cases explicitly
refuse with empty stdout before 2s on source Node, sanitized native and emitted
JS. All 13 inputs are retained under validation/followup/stalls/ as .input files.
`gaps/portParserRecovery.ts` is the smallest local-parser proving program.
The older shared-parser timeout test remains a dependency-gap observation.

### UTF-8 and cooked surrogate checkpoint

Stall checkpoint 040fafcb7bf544bdc62fc3bf27c2f2c4ab39d512 was pushed.
Canonical serialization now matches Go's replacement decoding of an unpaired
cooked WTF-8 surrogate: its three bytes become three U+FFFD characters. Paired
UTF-16 surrogates retain their existing representation. This rescues all 32
cooked-surrogate refusals. Source census: 15,300 identical, 306 refused Go
answers, 123 accepted Go refusals, 46 mismatches, 11,594 both refused: 429
acceptance disagreements. No earlier match regressed or new disagreement arose.

`ADAMIC_ESTREE_LIBRARY=/workspace/scratch/estree-library go test -count=1 -v
./stage1/cohere/estree -run '^TestCooked|^TestLossyInput|^TestRawInputGap$' >
/tmp/estree-utf8-gate.log 2>&1`: PASS 79.103s. Eight generated files, 3,606
bytes match Go in all three port builds; the one-versus-three replacement
mutant completes and is caught on Node/native at line 8. The original pinned
library retains unpaired UTF-16, proved independently. `ADAMIC_ESTREE_CORPUS=
/tmp/estree-utf8-rescued go test -count=1 -v ./stage1/cohere/estree
-run '^TestRepositoryAgreement$' > /tmp/estree-utf8-rescued.log 2>&1`: PASS
28.853s, all 32 repository files and 15,119,046 bytes on all three port builds.
Cohere passes after removing the newly unused panic import.

Raw input loss is not fixed: equal-size files with bytes F0 90 80 and EF BF BD
still return the same text and size from the documented readTextFile API. Go
has different comment values for them. TestRawInputGap proves this on source
Node, native and emitted JS. TestLossyInputRefusal verifies both explicitly
refuse with no output. Disabling the input guard completes on Node/native with
wrong Go bytes at line 17, caught by comparison. This is an API information
boundary, not a hypothetical risk or an inference that size can recover bytes.
No raw-byte API is exposed by this toolchain; a runtime-team raw reader is
required for original malformed-byte parity. Valid literal U+FFFD is still
conservatively refused because it cannot be distinguished from loss. The
minimal programs are gaps/rawInput.ts and gaps/cookedSurrogate.ts.

### Decorated expressions and await/yield checkpoint

UTF-8 checkpoint 94db1f81d68e535486292375412c879652a1e3db was pushed.
Decorated class expressions are parsed with their decorator children. Await and
yield use Go's same-line identifier/keyword/numeric/bigint/string lookahead,
rather than only an identifier. The source audit rescues 64 files: 15,364
identical, 242 refused Go answers, 122 accepted Go refusals, 46 wrong outputs,
11,595 both refused. Acceptance disagreements fall to 364. The initial expansion
newly accepted one orphan decorator; the explicit parent check fixes that and
one previous incorrect acceptance. The final audit has no match regression.

`go test -count=1 -v ./stage1/cohere/estree -run '^TestRecoveredExpression|
^TestUnattachedDecorator$' > /tmp/estree-expressions-gate-final.log 2>&1`:
PASS 58.495s, seven generated files and 8,717 matching bytes, line-break mutant
caught on Node/native at line 302, orphan decorator refusals in all three
builds. `go test -count=1 -v ./stage1/cohere/estree
-run '^TestUnattachedDecoratorControl$' > /tmp/estree-expressions-control.log
2>&1`: PASS 19.599s, both samples refused by Go; disabling the parent guard
accepts both on Node/native and fails the acceptance predicate.
`ADAMIC_ESTREE_CORPUS=/tmp/estree-expressions-rescued go test -count=1 -v
./stage1/cohere/estree -run '^TestRepositoryAgreement$' >
/tmp/estree-expressions-rescued.log 2>&1`: PASS 20.632s, all 64 files,
2,142,447 bytes on all three port builds. Cohere is green. The first mutant
anchor was not unique; that failed attempt is not credited as a mutant.

### Recovery checkpoint

Expression checkpoint e1f0958 was pushed. The local parser now adapts the recovery
work on origin/codex/parser-recovery at b85afdde325a8881a54a3bdb100ea19decd4d966,
while retaining this unit's JSX, decorated expressions and progress guards. Its
statements, grammar, lookahead, lexical messages, recovery and spelling helpers
are copied into the ESTree territory. Only scanner and node-model dependencies
remain shared. This is a local adaptation, not a claim that the recovery branch
has complete diagnostic parity. The original recovery trial regressed matches;
those regressions were fixed before promotion.

The final source census is 15,554 identical / 98 refused Go answers / 77 accepted
Go refusals / zero different / 11,640 both refused. This resolves all 92 baseline
output mismatches, with no old matching file regressing or newly incorrect
acceptance. First-modifier accessibility, private and bigint property names,
parameter modifiers, static-block contexts, object cover defaults, empty generic
lists, first heritage clauses, external-module await, export-default async calls,
namespace ASI, recovered index signatures and conditional arrow speculation are
held to Go. Colon detection now scans the next token instead of looking inside
intervening JSDoc text. Type-list construction has its own helper and scanner.

`ADAMIC_ESTREE_LIBRARY=/workspace/scratch/estree-library
ADAMIC_ESTREE_CORPUS=/tmp/estree-recovery-rescued go test -count=1 -v
./stage1/cohere/estree -run '^TestRecover|^TestBoundedPortParser$|
^TestPortStallControl$|^TestUnattachedDecorator|^TestRepositoryAgreement$'`
ran the 190 newly matching frozen files, 15,512,882 bytes on all three builds,
in 34.95s for TestRepositoryAgreement. Other passing checks included sixteen
new generated files, 23,734 bytes, the await line-break mutant, orphan-decorator
acceptance control, and all thirteen recorded stalls. The first library assertion
incorrectly expected primitive implements to be rejected: the original library
accepts it and agrees in both raw and postprocessed trees. That assertion was
corrected and the focused gate rerun; retained failed output is not called green.

The three new wrong-output mutants finish normally on Node and sanitized native:
first accessibility forced public (line 96), empty generic end decremented
(line 419), and module await reparse removed (line 627). The disabled class-member
progress guard reaches the 500ms deadline on both builds. EOF recovery now emits
parser diagnostics; every original recorded stall follows Go's acceptance
outcome, recovering a byte-identical tree where Go accepts. The older shared
parser remains unchanged and its separate gap still times out.

`TestRecoveryLibraryGaps` proves Go accepts f<>(); and class C<> {}, while the
pinned original library refuses. Four valid minimal programs match raw and
postprocessed library trees, including primitive implements and module await.
`gaps/emptyGenerics.ts` and `gaps/portParserRecovery.ts` are retained proving
programs. The former goes through the actual driver; the latter now proves the
actual driver's diagnostic boundary rather than calling a parser without
checking its diagnostics. Full corpus records and final green logs are retained
under validation/followup/. The later full gate is reported below.

Final focused recovery gate: `ADAMIC_ESTREE_LIBRARY=/workspace/scratch/estree-library
 go test -count=1 -v ./stage1/cohere/estree -run
'^TestRecoveredGrammar$|^TestRecoveryMutants$|^TestRecoveryLibraryGaps$|
^TestBoundedPortParser$|^TestPortStallControl$' > /tmp/estree-recovery-green.log 2>&1`
passed in 152.496s. `cohere --no-cache stage1/cohere/estree/*.ts` passed,
29/29 Adamic-ready sources. The final actual-driver source audit remains
15,554 / 98 / 77 / zero / 11,640 on all 27,369 frozen records.

### Acceptance checkpoint

Recovery checkpoint 4a8deeb was pushed. The next source census is 15,635 identical,
17 refused Go answers, three accepted Go refusals, zero different and 11,714 both
refused: 20 acceptance disagreements. No previously matching file regressed.
The largest remaining recoverable causes were declaration modifiers, bare yield,
object modifiers and Go's heritage-list recovery. Local structural validation
also follows Go for exponentiation, super, optional new expressions, untagged
invalid escapes, private type names and prefix update operands. Import assertions
produce the pinned Go diagnostic instead of being silently accepted.

Catch initializers are not type annotations; implements can name a class when
it does not start a heritage clause. Optional tagged-template markers are excluded
from type arguments, and a recovered newline throw gets Go's empty identifier.
Pure modifier classification has its own helper. One callback capturing this
was explicitly refused by Adamic's cycle checker and was replaced with a loop.
No shared implementation files were changed.

The actual-driver source audit covers all 27,369 frozen records. The 81 newly
matching files pass TestRepositoryAgreement on source Node, sanitized native
and emitted JS: 474,925 canonical bytes. Eleven generated cases match another
8,800 bytes. Eight Go-refused minimal programs explicitly refuse with empty stdout
before the two-second deadline on all three builds.

Three controls are run on source Node and sanitized native: catch initializer
misclassified as a type annotation (wrong output, line 182), class-name recovery
suppressed (wrong output), and syntax validation disabled (incorrect acceptance
of ++await 42). The first class-name mutant was behaviorally inert and survived;
it is not credited. The corrected mutant must produce wrong output and finish
normally. Final gate output and final source census are retained below.

Final acceptance gate: `go test -count=1 -v ./stage1/cohere/estree -run
'^TestAcceptance' > /tmp/estree-acceptance-final-green.log 2>&1` passes in
115.498s. Corrected class-name mutant differs at line 151 on both builds.
`ADAMIC_ESTREE_CORPUS=/tmp/estree-acceptance-rescued go test -count=1 -v
./stage1/cohere/estree -run '^TestRepositoryAgreement$|^TestRecoveredGrammar$'
passes in 46.223s. Cohere passes with 31/31 Adamic-ready sources.

### Remaining grammar checkpoint

Acceptance checkpoint 8e88b7b was pushed. The new actual-driver census is 15,644
identical / eight refused Go answers / zero accepted Go refusals / zero different /
11,717 both refused. No old matching file regresses. Five remaining refusals are
the documented raw-input limitation; three are deep binary-expression stack
limits, addressed in the next checkpoint.

Generic openings now split << in type contexts while preserving ordinary shifts.
Go permits initializers and accessor bodies in type members, and trailing members
in mapped types; conversion preserves the same omission of mapped members.
Type-query generics and type predicates respect line breaks. Assertion predicates
accept this. Binary as/satisfies parsing stops when erasure would change operator
precedence. Leading reference pragmas use Go's attribute parsing and do not inspect
strings, block comments or comments after the first code token.

The native interface-based binary helper explicitly panicked; it was discarded.
Direct parser calls plus pure precedence helpers avoid that path. Scanner-only
speculation moves to the existing lookahead helper. A malformed generated accessor
case exposed an erroneous semicolon requirement after a body; it was fixed before
the final gate. Failed trial logs are not credited as green.

`ADAMIC_ESTREE_LIBRARY=/workspace/scratch/estree-library
ADAMIC_ESTREE_CORPUS=/tmp/estree-syntax-rescued go test -count=1 -v
./stage1/cohere/estree -run '^TestSyntax|^TestRepositoryAgreement$|
^TestInterfaceTypeMethodGap$' > /tmp/estree-syntax-final-checkpoint.log 2>&1`
passes in 143.574s. Nine recovered files match 146,391 bytes in all three builds;
nine generated cases match 15,547 bytes. Four original-library cases agree both
raw and postprocessed. Three explicit Go refusals give empty stdout on all builds.
The mapped-constraint mutant differs at line 324 on Node/native; disabled erasure
precedence and reference checks accept Go-refused inputs on both builds.

The new 13-line interfaceTypeMethod program proves a separate compiler gap even
with both arguments supplied. Node/emitted JS print 1; sanitized native explicitly
panics. Removing concrete defaults is the control: all builds print 1. Compiler
files remain unedited. TestTypeMemberLibraryGap separately passes in 0.954s and
proves the original library refuses an interface initializer accepted by Go.
Cohere passes, 33/33 Adamic-ready sources. Logs and the frozen census are retained.

### Deep traversal checkpoint

Grammar checkpoint cc1ba3b was pushed. The actual source audit now has 15,647
identical / five refused Go answers / zero accepted Go refusals / zero different /
11,717 both refused on all 27,369 frozen records. All 92 original output mismatches
and 979 of 984 original acceptance disagreements are resolved. No old matching
record regresses. The five remaining records hit the documented raw-input guard.

The remaining three stack-limited files are binderBinaryExpressionStress variants.
An explicit ordinary-binary left spine avoids recursive conversion; preorder and
postorder work lists avoid recursive postprocessing; canonical dump frames preserve
all original own-field and child order. The helper receives a direct callback,
without an interface call to a defaulted method. Shared code remains untouched.

The scratch native trial checked all three recovered frozen files, 5,333,598 bytes,
in all three port builds (28.796s). The promoted driver passes three generated
4096-operand numeric/string/logical cases, another 2,551,820 canonical bytes.
`go test -count=1 -v ./stage1/cohere/estree -run
'^TestDeep|^TestThreePortMutants$' > /tmp/estree-deep-green.log 2>&1`
passes in 168.531s. New binary-operator, postorder-alias and dump-field-order mutants
finish normally and differ at lines 25, 89 and 2 on source Node and sanitized
native. The original computed-member, logical-rebalance and merged-JSDoc mutants
also finish normally and differ at lines 3810, 3543 and 4661 on both builds.
Cohere passes with 34/34 Adamic-ready sources.

The complete matching-corpus check and complete native refusal check passed
in final validation. These frozen-corpus counts are not a new inventory of the
implementation added in this thread. Deep-source agreement does not claim that
every arbitrary nesting shape is free from all recursive parser limits.

### Frozen corpus matrix

| Disposition | Baseline | Final source audit |
| --- | ---: | ---: |
| Identical Go answer | 14,699 | 15,647 |
| Port refuses Go answer | 861 | 5 |
| Port accepts Go refusal | 123 | 0 |
| Different output | 92 | 0 |
| Both refuse | 11,594 | 11,717 |
| Total | 27,369 | 27,369 |

This is 948 newly matching files: 856 formerly refused and 92 formerly different.
All 123 former incorrect acceptances now refuse. Acceptance disagreements shrink
by 979, from 984 to five. The remaining frozen files are parseReplacementCharacter.js
and .ts, regexInvalidUtf8WithUnicodeFlag.js and .ts, and the selector unit's nodes.ts.
Their full paths and dispositions are in deep-corpus.jsonl.gz. Some contain valid
literal U+FFFD, intentionally covered by the conservative input guard.

The complete corpus matrix passes on the promoted code. TestRepositoryAgreement
checks all 15,647 source-Node-identical Go answers, 549,618,641 canonical bytes per
build, on source Node, sanitized native and emitted JS (616.08s). All 11,717 frozen
Go refusals explicitly refuse with empty stdout on sanitized native, before the
per-file two-second deadline (227.99s). Compiler-bug panics, sanitizer failures,
timeouts and successful acceptance would fail the native refusal check.
Source Node independently audits all 27,369 records. Emitted-JS refusals are checked
on the focused/generated cases; the complete negative corpus is checked on native
and source Node, not by spawning 11,717 emitted-JS processes.

### Pushed implementation commits

- `fe07100524fe56da82b15d8d8ff50358ff85c92d`: Port JSX parsing and conversion at the Go source boundary.
- `068218be92c9caa19e9ce981f178abf599371632`: Preserve decorated export modifiers and wrapper ranges.
- `040fafcb7bf544bdc62fc3bf27c2f2c4ab39d512`: Bound ESTree parser stalls in the actual driver.
- `94db1f81d68e535486292375412c879652a1e3db`: Match Go cooked WTF-8 and prove the lossy input boundary.
- `e1f095831b5f0e703c1e8e3a92d79bd07befccd0`: Parse decorated expressions and match await lookahead.
- `4a8deeb87dad865bac1c6c378cff51438e0e92c3`: Match Go recovered syntax and generic list boundaries.
- `8e88b7bdddc10a72a803adc92a8d1070c8e149b7`: Match Go acceptance for modifiers and recovered expressions.
- `cc1ba3bc7f8dfb2085dc3e4fa79f3e4489e5ef8a`: Complete Go grammar and acceptance recovery.
- `5623b0657b261c1b904729656d25e24e1fcb7c9b`: Handle deep binary trees without traversal recursion.

The first deep-checkpoint push failed on GitHub while writing/renaming a temporary
pack file. Retrying the same fast-forward push succeeded. No force push was used.
The complete final package gate and its timing are recorded below.
Throughput figures in README remain historical; no new speed claim is made here.
Coverage does not include every arbitrary source nesting shape or the complete
JavaScript printer. The raw-input boundary is unresolved, with a proving program;
compiler/runtime files and the shared TypeScript parser remain unedited.

### Final green gate

`source /workspace/adamic-tools/env.sh` followed by
`ADAMIC_ESTREE_LIBRARY=/workspace/scratch/estree-library
ADAMIC_ESTREE_CORPUS=/tmp/estree-final go test -timeout 30m -count=1 -v
./stage1/cohere/estree > /tmp/estree-final-gate.log 2>&1`
passes in 1783.266s. No correctness check is skipped. TestThroughput is intentionally
skipped because this follow-up is correctness work; earlier measurements are
historical. The whole repository's go test ./... was not run. The touched package
ran completely, including full positive corpus, full native refusals, all focused
library comparisons and gap programs.

`go test -count=1 -v ./internal/oracle -run '^TestTheOracleCatchesOneByte$'
> /tmp/estree-final-oracle.log 2>&1` passes in 0.101s (native/Node cache hits).
`go vet ./stage1/cohere/estree > /tmp/estree-final-vet.log 2>&1` exits zero,
with empty output. Final `cohere --no-cache stage1/cohere/estree/*.ts` plus the
three new minimal gap files passes: 40 checked, 34/34 Adamic-ready. `git diff
--check` passes. The deepBinary proving program prints 788941 on source Node.
These outputs are retained under validation/followup/final-*.log.

The final package gate runs 22 driver mutants on source Node and sanitized native:
17 successful wrong-byte outputs, four successful incorrect acceptances and one
500ms liveness timeout. They are:

| Mutant | What catches it |
| --- | --- |
| Computed member flag reversed | Canonical byte comparison |
| Logical rebalance disabled | Canonical byte comparison |
| Merged JSDoc separator changed | Canonical byte comparison |
| JSX selfClosing cleared | Canonical byte comparison |
| Decorated export range includes prefix | Canonical byte comparison |
| Await/yield line-break rule removed | Canonical byte comparison |
| First accessibility forced public | Canonical byte comparison |
| Empty generic end decremented | Canonical byte comparison |
| External-module await reparse removed | Canonical byte comparison |
| Catch initializer treated as annotation | Canonical byte comparison |
| Class implements name suppressed | Canonical byte comparison |
| Mapped constraint removed | Canonical byte comparison |
| Binary operator forced minus | Canonical byte comparison |
| Postorder child alias ignored | Canonical byte comparison |
| Canonical own-property traversal reversed | Canonical byte comparison |
| Cooked unpaired surrogate gets one replacement | Canonical byte comparison |
| Raw-input guard disabled | Canonical byte comparison |
| Syntax validation disabled | Go refusal versus accepted Program |
| Unattached-decorator guard disabled | Go refusal versus accepted Program |
| Erasure precedence check disabled | Go refusal versus accepted Program |
| Reference-pragma check disabled | Go refusal versus accepted Program |
| Class-member progress guard disabled | 500ms deadline on both builds |

Every byte/acceptance mutant finishes normally; crashes and refusals are not
credited as wrong-output byte mutants. The earlier scan-progress-disabled control
is also recorded in its own checkpoint. The default-free interface-method control
is separate compiler-gap evidence: all three builds print 1 when the concrete
defaults are removed. Two initially inert/nonunique mutation attempts are described
above and are not credited.

UTF-8 commit 94db1f81 repairs cooked WTF-8 and proves raw-input information loss.
Stall commit 040fafcb places the guard in the actual driver. Later recovery makes
Go-accepted stall programs byte-identical and leaves Go-refused programs explicit.
Five raw-input refusals remain: the current text API cannot represent both original
byte sequences faithfully. No compiler/runtime changes were made to bypass that
boundary. The original library's narrower grammar and normalization differences
remain observed pin-specific gaps, not assertions about the TypeScript spec.
