Rebased twelve completed default-option ports onto lint-area runtime tip d65a8f931.
Tested code 886a024dbc2e5d4f0951818d8f3f0b72a9ad7c7d; publication stays on codex/typeaware-wave-27.
Four owned oracle gates, sanitizer corpora, handles, regressions, vet and focused runtime tests PASS.
Thirteen rule mutants and five fact mutants exit 0 and are caught by Go bytes.
No unclaimed rules remain; dynamic RegExp and three parked React analysis claims remain blocked.

# Landing base and commands

The requested area base advanced from 7481e0324 to
d65a8f931c98655936ae04c6899f38f14862b73e with release-path, string identity and
reverse-search optimizations. All 24 owned patches rebased without conflict.
This includes harness 41eb6eab2 and main 39638d9e. No shared or protected sources
were edited; incoming runtime changes were preserved. Only the owned branch is
pushed. go build -o /workspace/wave-27-scratch/f801-third/adamic ./cmd/adamic
rebuilt the compiler; go run ./cmd/lint-registry validated the integrated registry.

All shells source /workspace/adamic-tools/env.sh and use TMPDIR=/tmp/adamic-gate.
The manifests, compiler-source pin and artifact environments are the same as
../landing_c019/REPORT.md. Original setup was ready 0s/warm cache 116s/total 116s;
nproc=5. Tests write directly to the retained wave-27-d65-*.log files.

```
go test -v -count=1 -timeout 15m ./stage1/cohere/typeaware -run '^TestWave27AgreementAndMutants$'
python3 stage1/cohere/typeaware/wave_27_next/validate.py
python3 stage1/cohere/typeaware/wave_27_third/validate.py
python3 stage1/cohere/typeaware/wave_27_fifth/validate.py
go test -v -count=1 -timeout 15m ./bridge/tsgo/checker ./stage1/cohere/typeaware -run '^(TestCoverageCheckerQuestions|TestFactEncoding|TestShapeAndNameFacts|TestExactIndexMatchesCompilerNodes|TestAdamicRootKeepsConfigDeclarations|TestFactsDecoderGuards|TestInspectRequestRefusals|TestPinnedTypeFlags)$'
go vet ./bridge/tsgo/checker ./stage1/cohere/typeaware
go test -v -count=1 -timeout 15m ./internal/native ./internal/oracle -run '^(TestRuntimeReleasePaths|TestRuntimeStringEquality|TestRuntimeLastIndexOfMatchesNode)$'
```

First gate PASS in 188.329s. Three Python validators print PASS. Focused bridge
and type-aware regressions PASS in 0.243s and 104.578s; vet has empty output.
Native runtime tests PASS in 27.845s; reverse-search Node oracle PASS in 20.901s,
with 758 identical bytes across Node, emitted JavaScript, release and sanitizer
native. The runtime oracle uses its normal cache: native hits=1/misses=2, Node
hits=0/misses=2. Incoming runtime mutation proofs were not rerun here; the owned
lint mutation proofs below were all rerun with the new compiler/runtime.

# Exact rule agreement and mutants

All four gates compare complete findings, ranges, ids, descriptions, automatic
fixes and ordered suggestions/edits against production Go cohere. Controls:
first 105 files/39 findings/17177 bytes; next 93/101/62201; third 605/437/225707;
fifth 186/108/64739. Third retains 34 Go-parser exclusions from 639 candidates,
and fifth 24 from 210; those are not passing cases. Every batch matches the
287-file repository corpus (18485 bytes) and 77 TypeScript compiler files (5934
bytes), with zero findings. Normal and Go-ASan/native--sanitize comparisons pass
on controls and both corpora; sanitizer stderr is empty. Fifth checks 351 numeric
compatibility constants and three named rule.json declarations against Go.

All rule mutants compile and exit 0 with empty stderr, caught only by Go byte
comparison: ORM nullable 57; Serializable optional 5466; Verify array 8264;
pending output 80; race ownership 52234; blocking early return 581; Effect regex
28395; rest declaration 54; Hook message 4962; regex arity 112016; symbol ambient
39039; typeof legal string 46485; await contract bypass 33546. Each number is the
first differing byte. Five normally exiting fact mutants are caught at bytes 56
(declaration file), 1318 (CFG), 12452 (never flags), 42069 (type-only import),
61584 (declaration kind). Released queries reject stale handles with exact panic
70; retained-registry mutants exit 0 and fail the required panic check. Focused
regressions also cover nineteen wire guards, wrong-kind/unknown-question guard
mutants and pinned flags. Detailed observations are in the retained logs.

# Timing and remaining scope

Alternating whole-process samples retain identical streams; other independent
gates ran concurrently, so the following are contended observations:
fifth repository native 751.958ms / Go 244.330ms (3.078x);
fifth compiler native 3107.256ms / Go 611.896ms (5.078x).

The newly rebuilt compiler still rejects additional_hooks_pattern.a:4:23 with
exit 1: stage 0 can't lower RegExp with a nonconstant pattern yet. The isolated
new RegExp(pattern, 'u') helper is present, but runtime additionalHooks integration
cannot be completed within rule directories. Three React HIR/SSA/capture/post-
dominance claims remain parked. Older nine rules await handed-node migration;
newest three use cached handed nodes and named listener metadata. Owned rules'
shared-registry discovery and emitted-JavaScript certification, nondefault
options, inherited port gates and full repository gate remain untested. The
focused runtime JavaScript comparison does not certify lint emitted JavaScript.

Audit: 605 origin refs, 197 ranked rules, 33 Markdown claims, zero unclaimed rules.
No next claim is taken. selection.json records the complete exclusion evidence.
