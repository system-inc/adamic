# Wave 16, next three rules

The claimed follow-up implementations were finished and their final validation
pushed in 828d4a57 before selection resumed. The next claim was pushed in
be3ab457 before writing code: nexus/security-no-interpolated-shell-command,
nexus/security-no-interpolated-sql-string and no-alert. All three have zero
findings in the combined compiler/repository ranking. A fresh all-origin-head
fetch excluded ports on origin/main and origin/codex/tsgo-c-library and claims
on every origin head. Additional whole-stage1/cohere searches of those two heads
confirmed these three were inventory/count references only, not ports. The scan
covered 197 ranked rules and 33 Markdown claim documents; 70 candidates remained.
No rules were skipped or claimed beyond this set.

Each port is its own .a file. Adamic owns call selection, declaration ancestry
matching, const expansion, literal safety, SQL scanning, heredoc scanning,
escape-chain recognition, shadowing decisions, report spans and messages.
The independent Go overlay oracle calls the unmodified pinned production rules;
it imports no bridge implementation. Canonical output includes every finding,
complete fix and suggestion fields and file heading, sorted with duplicates
preserved. None of these three production rules offers a fix or suggestion.

## Isolated facts and the exact integration gap

Two new questions, text-type-facts and value-declaration, each have their own
Go implementation and Adamic decoder file. The former supplies checker type
identities, flags, intrinsic-error status, literal values, constraints, union/
intersection parts and template-literal texts/holes. The latter returns the
symbol's value declaration, not every binding declaration. They register in
this worker's isolated additional_questions.go map and contain no lint verdict.
The shell port reuses the earlier raw signature-ancestry question. No-alert uses
the existing resolved-name question and needs no new compiler operation.

Shared registration and shared test harness files remain untouched. The
rule-specific suite uses a Go build overlay containing exactly the three-line
hook in wave16_followup_registration.patch. The overlay is explicit in the test
and benchmark provenance. It proves the native implementations, not integration
of the shared tree. An ordinary archive without that hook exits 70 with:

```
adamic: panic: unsupported checker question: signature-ancestry
```

Its stdout contains the initial declaration-root file header only, with no
completed finding stream. The actual stdout/stderr and build logs are preserved.
Integrating the supplied hook is the sole shared-dispatch dependency for these
ports; all seven question registrations already live in their own files.
No-alert itself has no new-question dependency. No shared registration generator,
shared test harness, protected lowerer, emitter or oracle files were edited.

## Final commands and observations

Every command sourced /workspace/adamic-tools/env.sh; every test wrote directly
to a log file. Go/cohere/TypeScript pins and the 77-file compiler and frozen
287-file repository populations are unchanged from the original wave 16.

```
ADAMIC_WAVE16_THIRD_ARTIFACTS=/workspace/wave16-artifacts/third-final \
ADAMIC_WAVE16_COMPILER_MANIFEST=/workspace/wave16-artifacts/compiler.manifest \
ADAMIC_WAVE16_REPOSITORY_MANIFEST=/workspace/wave16-artifacts/repository.manifest \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave16-corpus/typescript \
go test ./stage1/cohere/typeaware -run '^TestWave16ThirdAgreementAndMutants$' \
-v -count=1 -timeout=20m > /workspace/wave16-artifacts/third-final-test.log 2>&1
```

PASS, 93.450 seconds. Both normal and ASan/UBSan/LeakSanitizer comparisons
produce identical complete bytes, with empty native stderr:

| Population | Findings | Identical bytes |
| --- | ---: | ---: |
| Controls, 19 subjects plus helper and declaration root | 50 | 28607 |
| Repository, 287 roots | 0 | 18485 |
| TypeScript compiler, 77 roots | 0 | 5780 |

Controls cover global and shadowed alert calls, parentheses and optional chains,
renamed/namespace child-process signatures, promisified exec, literal and opaque
commands, const and conditional expansion, argument arrays, shell-option types,
quoted/unquoted/tab-stripping/multiple heredocs, malformed openers, SQL strings,
comments, prose, fragments, unfinished strings, bare comparisons, double escaping,
string literals, branded/template/constrained types, tagged templates, Unicode
and CRLF. Positive controls are required for each rule.

Each rule mutant compiled and completed with exit 0 and empty stderr. Only the
independent Go diagnostic bytes caught it:

| Mutant | Change | First differing byte | Findings |
| --- | --- | ---: | ---: |
| shell-heredoc | Ignore the quoted-body exemption | 8250 | 53 |
| sql-string | Treat string/string-mapping parts as safe | 11707 | 31 |
| alert-range | Add one to the report end | 116 | 50 |

The unchanged finding count of the range mutant is deliberate. A count-only
comparison cannot catch it. Querying a released checker program again panics 70
with invalid or released checker handle; retaining the program in the registry
makes the mutant exit 0 and the required refusal catches it.

```
go test ./bridge/tsgo/... -v -count=1 -timeout=15m \
 > /workspace/wave16-artifacts/third-bridge-test.log 2>&1
# PASS bridge 89.395s, checker 0.174s.
python3 /workspace/wave16-artifacts/third-question-mutants.py \
 > /workspace/wave16-artifacts/third-question-mutants.log 2>&1
# Six compiling mutants caught by their intended direct-fact/refusal assertions.
go vet ./bridge/tsgo/... ./stage1/cohere/typeaware \
 > /workspace/wave16-artifacts/third-vet.log 2>&1
# Empty log.
gofmt -l bridge/tsgo/checker/*.go stage1/cohere/typeaware/wave16_third_test.go \
 > /workspace/wave16-artifacts/third-gofmt.log
# Empty log.
```

Direct raw-question tests compare every serialized identity, constituent,
constraint, literal, template text/hole and value-declaration position with direct
checker operations. They also hold wrong-kind and non-UTF-8 refusal paths.
The six mutants change type flags, literal kind, declaration end, declaration-kind
refusal, literal UTF-8 refusal and template-text UTF-8 refusal. Each test compiles
and fails on the intended assertion, not a build error. Foundation checks again
pass 100 C ABI queries and 162 independent positions / 3261 bytes under
sanitizers, plus all seven established length, stale-handle, position, linking,
output-free and region-ownership mutants. Full logs are committed.

Initial failures are preserved rather than called passes: keyword-array generation
produced invalid source, stage 0 refused a never[] argument, and the SQL walk
panicked before checking the root's parent. Typed empty arrays and checking the
parent index before reading it fixed those refusals. The byte oracle then caught
a missed promisified exec call: its enclosing namespace is four ancestor edges,
not five. Correcting that raw ancestry lookup produced the 50 matching controls.
The final formatted run includes all fixes and all mutants above.

## Quiet native versus Go timing

Three alternating rounds ran after all builds/tests finished. Every round checks
the complete output SHA-256; all native/Go streams for a population agree.
The native executable links the explicit registration overlay. Median seconds:

| Corpus | Native load | Native run | Native process | Go load | Go run | Go process |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Compiler | 0.255443 | 3.519011 | 3.809196 | 0.254321 | 1.169763 | 1.431048 |
| Repository | 0.062887 | 0.332491 | 0.400486 | 0.059257 | 0.109414 | 0.183141 |

Native whole process is 2.66 times slower on the compiler and 2.19 times slower
on the repository. With zero findings, these measure traversal/checking overhead,
not positive-finding throughput. Raw nanoseconds, query counts, hashes, benchmark
script and timing stderr are in validation-wave16-third. No performance parity
is claimed. nproc is still 5 with the same four-core quota; original setup timing
and toolchain versions are in WAVE16_REPORT.md. No submodule pins changed.

## Remaining limits

The shared hook is still required for ordinary security-rule execution. The
shared .a/profile/emitted-JavaScript harness integration is owned by the other
worker, and was not modified here. No full repository go-test gate, emitted-JS
comparison or complete upstream fixture/options matrix is claimed. The new text
question explicitly refuses checker literal/template strings containing invalid
UTF-8 (including lone-surrogate literals), rather than silently changing them.
Both refusal paths are tested and have compiling mutants. These inputs are outside
the measured corpora and controls. All normal supported inputs measured above
are held to Go byte for byte. Every new Adamic source is .a.
