Rebased all twelve completed wave-27 default-option ports onto current main b8fb957aa.
Tested code 591107297f9a0478451f0c24a8652317113014e0; publication stays on codex/typeaware-wave-27.
All four owned oracle gates, sanitizer corpora, handles, focused regressions and vet PASS.
Thirteen rule mutants and five fact mutants exit 0; complete Go output catches every one.
No unclaimed ranked rules remain; runtime additionalHooks and three parked React analysis claims remain blocked.

# Current-main landing

Main b8fb957aa839a9e8cb0b54279dd9864fa317bd30 adds inherited static-field reads.
All 22 owned patches rebased without conflicts. No shared or protected files were
edited during this landing. The compiler was rebuilt with go build -o
/workspace/wave-27-scratch/f801-third/adamic ./cmd/adamic. Shells source
/workspace/adamic-tools/env.sh and set TMPDIR=/tmp/adamic-gate. Original setup
passed with ready 0s, warm cache 116s and total 116s; nproc was 5.

The following ran with the same frozen manifests and artifact environments
specified in ../landing_c019/REPORT.md. Output was captured directly to the
retained wave-27-b8-*.log files, never piped:

```
go test -v -count=1 -timeout 15m ./stage1/cohere/typeaware -run '^TestWave27AgreementAndMutants$'
python3 stage1/cohere/typeaware/wave_27_next/validate.py
python3 stage1/cohere/typeaware/wave_27_third/validate.py
python3 stage1/cohere/typeaware/wave_27_fifth/validate.py
go test -v -count=1 -timeout 15m ./bridge/tsgo/checker ./stage1/cohere/typeaware -run '^(TestCoverageCheckerQuestions|TestFactEncoding|TestShapeAndNameFacts|TestExactIndexMatchesCompilerNodes|TestAdamicRootKeepsConfigDeclarations|TestFactsDecoderGuards|TestInspectRequestRefusals|TestPinnedTypeFlags)$'
go vet ./bridge/tsgo/checker ./stage1/cohere/typeaware
```

The first gate PASS took 137.979s. All three Python validators print PASS.
Focused regressions PASS in 0.250s and 59.252s; vet exits 0 with empty output.
Controls match full finding ranges, messages, fixes and ordered suggestions:
first 105 files/39 findings/17177 bytes; next 93/101/62201; third
605/437/225707; fifth 186/108/64739. Third excludes 34 of 639 candidates and
fifth excludes 24 of 210 by independent Go parsing, not as passing cases.
Each batch also matches the frozen 287-file repository corpus (18485 bytes)
and 77 TypeScript compiler files (5934 bytes), both with zero findings.
All these comparisons pass normally and with Go ASan archives/native --sanitize;
sanitizer stderr is empty. Fifth independently checks 351 internal enum constants
and three named rule.json declarations. Historical numeric metadata descriptions
are superseded by NAMED_KINDS_REPORT.md and the current descriptors.

# Mutants and refusal checks

Every rule mutant compiles, exits 0 and has empty stderr. Only the Go byte
comparison catches it: ORM nullable inversion 57; Serializable optional inversion
5466; Verify array inversion 8264; pending-output inversion 80; race ownership
inversion 52234; blocking early-return inversion 581; Effect->EffectNever regex
28395; rest declaration inversion 54; Hook message inversion 4962; regex arity
112016; symbol ambient inversion 39039; typeof legal-string replacement 46485;
await contract bypass 33546. Numbers are first differing output bytes.
Five normally exiting fact mutants are caught by Go at bytes 56 (declaration file),
1318 (CFG reachability), 12452 (never flags), 42069 (type-only imports), 61584
(declaration kind). Released raw-type, wave questions, binding declarations and
all four checker-link queries reject stale handles with exact panic 70.
Each batch's retained-registry mutants exit 0 and fail required-refusal checks.
Malformed-wire, pinned flags and wrong-kind/unknown-question guard regressions
also pass; guard mutants exit 0 and fail the required panic assertions.

# Time and remaining scope

Whole-process samples were collected while independent batch gates ran, so they
are contended observations, not quiet performance claims. Fifth medians:
repository native 530.916ms / Go 199.992ms (2.655x);
compiler native 3603.205ms / Go 472.789ms (7.621x).

The freshly rebuilt compiler still exits 1 with: stage 0 can't lower RegExp with
a nonconstant pattern yet, at additional_hooks_pattern.a:4:23. The required
new RegExp(pattern, 'u') helper is present; runtime additionalHooks integration
cannot be completed without shared lowering work. No hand-rolled matcher is used.
Three React HIR/SSA/capture/post-dominance claims remain parked as documented.
Shared registry integration, emitted-JavaScript comparison, other nondefault
options and full repository gate remain untested. Earlier nine rules still await
handed-node migration; the newest three already take cached numeric wrappers and
have named metadata. Inherited port gates were not rerun.

Origin audit after fetch: 584 refs, 197 ranked rules, 33 Markdown claim records,
zero unclaimed entries. selection.json retains the complete exclusion evidence.
The pre-push main check still reports b8fb957aa. No further claims are taken.
