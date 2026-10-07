Rebased twelve completed default-option ports onto lint-area b84a9d931, containing main c7991b900.
Tested code 5783a00e49897cb4af60c0005a949c61a5051c45; publication stays on codex/typeaware-wave-27.
Four owned Go oracle gates, sanitizers, handles, regressions, vet and focused lowering tests PASS.
Thirteen rule mutants and five fact mutants exit 0 and are caught by Go bytes.
No unclaimed rules remain; dynamic RegExp, three parked React claims and shared certification remain unresolved.

# Landing and validation

Origin/area/stage1-lint is b84a9d9314b65d3d0261ee017e233287b4f071da;
main is c7991b900362796aefd111474e65eb5398e91953. Incoming changes add predicate
and proven-relation lowering and record runtime support. All 25 owned patches
rebased without conflict; the incoming shared changes are retained. No shared or
protected source was edited. Only the owned wave branch is pushed.

Rebuilt compiler with go build -o /workspace/wave-27-scratch/f801-third/adamic
./cmd/adamic; go run ./cmd/lint-registry validates integrated descriptors. Shells
source /workspace/adamic-tools/env.sh, TMPDIR=/tmp/adamic-gate. Original setup
passed ready 0s/warm cache 116s/total 116s; nproc=5. Current artifacts reuse the
owned scratch paths and frozen manifests documented in ../landing_c019/REPORT.md.
Only named obsolete scratch checker archives were removed to make space.

The exact tests, with output written directly to retained logs:

```
go test -v -count=1 -timeout 15m ./stage1/cohere/typeaware -run '^TestWave27AgreementAndMutants$'
python3 stage1/cohere/typeaware/wave_27_next/validate.py
python3 stage1/cohere/typeaware/wave_27_third/validate.py
python3 stage1/cohere/typeaware/wave_27_fifth/validate.py
go test -v -count=1 -timeout 15m ./bridge/tsgo/checker ./stage1/cohere/typeaware -run '^(TestCoverageCheckerQuestions|TestFactEncoding|TestShapeAndNameFacts|TestExactIndexMatchesCompilerNodes|TestAdamicRootKeepsConfigDeclarations|TestFactsDecoderGuards|TestInspectRequestRefusals|TestPinnedTypeFlags)$'
go vet ./bridge/tsgo/checker ./stage1/cohere/typeaware
go test -v -count=1 -timeout 15m ./internal/lower -run '^(TestProvenRelationsRefuse|TestProvenRelationsErase|TestUnprovenPredicateReturnsAreRefused|TestPredicateBodiesAreProven)$'
```

First gate PASS in 186.988s; three Python gates print PASS. Checker/type-aware
regressions PASS in 0.656s/97.113s, focused lowering tests PASS in 3.789s;
vet exits zero with empty output. Owned gates have compiler and repository inputs
supplied, not skipped. The full repository gate and the full newly announced
17-check external correctness suite were not run; no skipped test is claimed
passing, and no check was relaxed or removed.

# Agreement, mutants and handles

Full finding ranges, ids, descriptions, fixes and ordered suggestions/edits match
independent production Go cohere. First controls: 105 files/39 findings/17177
bytes; next 93/101/62201; third 605/437/225707; fifth 186/108/64739. Third retains
34 independent-Go-parser exclusions from 639 candidates and fifth 24 from 210,
not passing cases. Every batch matches the 287-file repository corpus (18485
bytes) and 77 compiler files (5934 bytes), zero findings. Controls and both corpora
pass normally and with Go ASan archives/native --sanitize; sanitizer stderr is
empty. Fifth independently checks 351 internal kind constants and three named
rule.json listeners. Its owned driver caches and hands current nodes to listeners.

Normally exiting rule mutants have empty stderr; only complete Go bytes catch:
ORM nullable inversion 57; Serializable optional inversion 5466; Verify array
inversion 8264; pending output 80; race ownership 52234; blocking early return
581; Effect regex 28395; rest declaration 54; Hook message 4962; regex arity
112016; symbol ambient 39039; typeof legal string 46485; await contract bypass
33546. Numbers are first differing bytes. Five fact mutants exit 0 and differ at
56 (declaration file), 1318 (CFG), 12452 (never flags), 42069 (type-only imports),
61584 (declaration kind). Released queries reject stale handles with exact panic
70; retained-registry mutants exit 0 and fail required-refusal checks. Focused
regressions include nineteen wire guards, pinned flags and wrong-kind/unknown-
question guard mutants. Exact observations are retained in the logs.

# Timing and limits

Three alternating whole-process samples keep identical streams. Other gates ran
concurrently, so these are contended measurements, not speedup claims:
fifth repository native 674.449ms / Go 241.973ms (2.787x);
fifth compiler native 3604.900ms / Go 471.636ms (7.643x).

The rebuilt compiler still exits 1 at additional_hooks_pattern.a:4:23: stage 0
can't lower RegExp with a nonconstant pattern yet. The isolated constructor is
new RegExp(pattern, 'u'), but runtime additionalHooks integration remains blocked
without shared lowering changes. Three React HIR/SSA/capture/post-dominance claims
remain parked. Older nine rules await handed-node migration. Owned shared-registry
and emitted-JavaScript certification, nondefault options, inherited port gates,
full repository gate and full external correctness suite remain untested.

Origin audit: 636 refs, 197 ranked rules, 33 Markdown claim records, zero unclaimed
rules. No new claim is taken. selection.json retains complete exclusion evidence.
