Built: rebased the existing nine rule implementations, class listener declarations and numeric rule.json manifests onto current main f8013f0b; eight rules are green and no-object-constructor retains three parser refusals. No new claims.
Commits: rebased tip 8a6ed0ee replaces pushed 26d42f05; rule code 1ef2790c, 647430cb and 816d1f99; numeric class declarations 2790b146; JSON declarations 8a6ed0ee. All claim commits remain before their code.
Commands and outputs: original trio gate PASS 301.222s; full bridge PASS 236.804s; filtered Node oracle PASS 69.812s; metadata guard PASS 0.374s; checker PASS 0.718s; vet PASS; normal and sanitized Nexus 129/129 and core 430/433 controls agree.
Mutants: nine compiling rule mutants and two export-fact mutants are caught only by full byte comparison; seven foundation mutants and released-registry mutant are caught by their dedicated checks; all 18 in-memory class/JSON metadata mutations are rejected by the independent production-listener comparison.
Not covered: three shared-parser controls, shared numeric-node/descriptor-loader/supplied-node dispatch, full repository Go gate and emitted-JavaScript rule parity. The whole branch is not claimed fully green or landing-ready, and no more rules were claimed.

## Landing unit and scope

The only pushed branch owned by this unit is codex/typeaware-wave-13. Main moved
from e8ba3d5d to f8013f0baac41ddc340d76f83bddde38536a8f07 with lowering and Map/Set
runtime changes. A clean 18-commit rebase brought in those changes, then all
native suites, semantic rule mutants and sanitizer comparisons were rebuilt on
that compiler. The normal and sanitized C checker archives were also rebuilt.
The independent production Go oracles for the later trios are unchanged binaries:
cohere/TypeScript pins and production Go rule sources did not change. The original
trio's Go oracle is rebuilt by its full gate.

This unit changed no rule logic, parser, compiler, registration generator or
shared harness. The compiler/runtime changes are inherited from main. The new
edits are only this owned rule directory's report and evidence. An exact
force-with-lease, pinned to remote 26d42f053df66ee63b3c6384fc88733f40c1ddca,
publishes the user-requested rebase without overwriting a concurrent remote edit.
Publication targets only codex/typeaware-wave-13, never main or area/. No PR.

## Current native agreement and refusals

| Population | Findings | Equal bytes, normal / ASan |
| --- | ---: | ---: |
| Original trio controls, 51 roots | 32 | 18957 / 18957 |
| Nexus controls, 129 programs / 254 roots | 166 | 120897 / 120897 |
| Supported core controls, 430 programs | 289 with 171 suggestions | 162621 / 162621 |
| Original compiler, 77 roots | 4 | 7087 / 7087 |
| Original repository, 287 roots | 1 | 19144 / 19144 |
| Each later trio compiler, 77 roots | 0 | 4933 / 4933 |
| Each later trio repository, 287 roots | 0 | 18485 / 18485 |

These are complete finding/fix/suggestion bytes, including ordering, spans,
identifiers, descriptions and replacements. The original control-byte count
changes with the scratch path prefix on both sides. Normal/sanitized per-control
JSON results are identical. Both corpus manifests are the original frozen
populations; no zero-finding result is taken alone as proof of rule behavior.
Positive controls and compiling semantic mutants prevent that failure mode.
No additional mismatch or sanitizer finding appeared.

The core validate.py command still reports cases 433 failures 3 and exits 1.
Go accepts these upstream controls; native refuses parsing before rule execution:

| Case | Go bytes / exit | Native and ASan bytes / exit | Refusal |
| --- | --- | --- | --- |
| 53d1c0a9ffc2fefa | 699 / 0 | 0 / 70 | JSX .ts: GreaterThanToken/SlashToken at 5 |
| aff6ced2ca61fe89 | 699 / 0 | 0 / 70 | JSX .ts: GreaterThanToken/Identifier at 7 |
| defcb4c4ce6a2921 | 611 / 0 | 0 / 70 | yield label: expected semicolon at 37 |

These same three failures are preserved in both modes. Ahra's shared-file
restriction prevents repairing the shared parser in this unit. No source
rewriting, relaxed fixture or fake empty finding stream was introduced.

Numeric listenerKinds fields and rule.json kinds for all nine rules remain
checked against production Go listener ASTs and actual pinned parser enum values.
The shared parser still exposes kind: string; no numeric node/descriptor loader
or supplied-node callback has landed. Legacy string relevance checks and scans
therefore remain in the old run() path. Their migration is explicitly blocked,
not claimed complete. Main's Diagnostic did not change and this message named
no batch-8 finding-model SHA, so no finding-model branch rebase was attempted.

## Every repeated mutant and lifetime check

All nine rule mutants compile, exit 0 with empty stderr, and are caught only by
complete output comparison with production Go:

| Mutant | First differing byte |
| --- | ---: |
| unassigned | 1345 |
| caught | 4136 |
| exports | 14959 |
| process-state | 134 |
| race-handle | 118 |
| blocking-order | 1690 |
| namespace-alias | 72 |
| literal-parentheses | 587 |
| executor-span | 101 |

The compiling export-chain-flags and export-module-lookup mutations also exit
0 with empty stderr; byte comparisons catch them at 14677 and 17114. Retaining
a released registry handle changes native exit 70 to 0 and fails the lifetime
assertion. Unmutated original, source-context and read-symbol released-handle
probes all exit 70 with invalid or released checker handle, no sanitizer report.

The full bridge gate re-proves seven foundation checks: input and output length
off by one each trigger ASan heap-buffer-overflow; retaining a released handle
fails the stale-handle assertion; asking for a type at a source-file position
differs at byte 6; removing link opt-in fails the refusal check; omitting the C
output free and allocating a region result on the heap each trigger LeakSanitizer.
The unmutated bridge compares 1600 positions across four compiler files, 54982
identical bytes under ASan/UBSan/LSan. Its C ABI also tests 100 queries, Unicode,
output surviving release, stale/zero handles and a distinct second handle.

Both listener metadata tests ran again: nine source-declaration mutations and
nine JSON wrong-kind mutations are rejected against the production listener
sets. These are in-memory metadata variants, not compiling semantic mutants.
The other five raw-fact semantic mutants and seven question-guard variants were
not repeated: their Go sources are unchanged, and prior evidence remains in the
original reports. Fresh baseline controls still prove their current native
transmission; dedicated Go checker tests pass on this base.

## Toolchain, commands and native/Go timings

bash cloud/setup.sh succeeds. Its timing lines report Go 1s, clang 1s, Node 1s,
submodules 1s, cache warm 217s, total 217s. nproc is 5; cpu.max is 400000/100000,
memory 17.6 GB. Sourced /workspace/adamic-tools/env.sh in every build/test shell.
All test output was written directly to log files, never piped.

```
ADAMIC_WAVE13_STAGE0=/workspace/wave13-f801-adamic ADAMIC_WAVE13_ARCHIVE=/workspace/wave13-f801-checker.a ADAMIC_WAVE13_ARTIFACTS=/workspace/wave13-f801-original ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave13-corpus ADAMIC_WAVE13_COMPILER_MANIFEST=/workspace/wave-13-compiler.manifest ADAMIC_WAVE13_REPOSITORY_MANIFEST=/workspace/wave-13-repository.manifest go test ./stage1/cohere/typeaware -run '^TestWave13AgreementAndMutants$' -count=1 -v -timeout=30m
ADAMIC_TSGO_CORPUS=/workspace/wave13-corpus go test ./bridge/tsgo/... -count=1 -timeout=15m -v
go test ./stage1/cohere/typeaware/wave_13_more -count=1 -v
go test ./internal/oracle -run 'TestTheOracleCatchesOneByte|TestNativeAgreesWithNode/internal/oracle/testdata/(closures|method_closures|generic_functions|regions|regions_throw|library_map_set_iterator_exhausted|library_map_set_iterator_number_hash|047cb0d_n_arrayindex|047cb0d_n_coalesce|047cb0d_n_element|047cb0d_n_optional|047cb0d_n_conditional)\.a$' -count=1 -timeout=10m -v
go vet ./...
```

Rebuilt stage0 with go build ./cmd/adamic and C archives with go build [-asan]
-buildmode=c-archive ./bridge/tsgo/archive. Rebuilt both continuation suite.a
files through adamic build --tsgo, adding --sanitize for sanitizer runs.
Ran each directory's validate.py/corpora.py in both modes and mutants.py with
fresh native builds. Existing rule reports record their positional arguments.
The filtered Node fixtures exercise new-main Map/Set and narrowed-element
behavior plus closure/generic/region behavior; the full repository gate was not
run. Own-rule emitted-JavaScript comparison is still unclaimed.

After all builds and tests finished, three alternating full-output runs per
corpus retained equal hashes. These medians include checker loading:

| Suite / corpus | Native seconds | Go seconds | Native / Go |
| --- | ---: | ---: | ---: |
| original / compiler | 6.885545 | 2.116492 | 3.25x |
| original / repository | 0.385626 | 0.126106 | 3.06x |
| next / compiler | 2.891683 | 0.413175 | 7.00x |
| next / repository | 0.615413 | 0.168479 | 3.65x |
| more / compiler | 3.392854 | 0.444457 | 7.63x |
| more / repository | 0.441683 | 0.119360 | 3.70x |

Native is slower on every measured population. No timing claim uses concurrent
gate runs. Formatting and source-only whitespace checks pass. Exact old oracle
stdout with significant trailing suggestion spaces remains preserved.
validation/landing-f801 retains complete control/corpus results, fresh gate and
mutant logs, setup/rebase logs, raw-stream archive, lifetimes and isolated
measurements; sha256.json checks top-level artifacts. This report supersedes
previous main-base and timing claims for the current landing attempt; older
reports remain historical evidence. No additional rules were claimed.
