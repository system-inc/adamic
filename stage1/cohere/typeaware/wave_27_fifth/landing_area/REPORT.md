Rebased the twelve completed default-option ports onto origin/area/stage1-lint 7481e0324.
Tested code eefcea8d5276b0b9d87acda3969c408e670bf424; only codex/typeaware-wave-27 is published.
Four owned Go comparison gates, sanitizer corpora, released handles, focused regressions and vet PASS.
Thirteen normally exiting rule mutants and five fact mutants are caught by Go byte comparison.
No unclaimed rules remain; runtime additionalHooks and three parked React claims remain blocked.

# Integration base

The user requested the harness landing at 50a5f105. Current origin/area/stage1-lint
is 7481e0324e34a2537aafa9db7eeacda50405611b, including that landing, harness
41eb6eab2b6de45ede0a40250765be295ee25fbd and current main
39638d9e278d38bb5aeae887f46d55a70e47aaad. All 23 owned patches rebased cleanly.
Incoming shared changes were preserved. No shared or protected source was edited
this turn. Main and area branches were not pushed.

The compiler was rebuilt with go build -o
/workspace/wave-27-scratch/f801-third/adamic ./cmd/adamic. go run ./cmd/lint-registry
validates fifteen discovered shared rules. go test -v -count=1
./stage1/cohere/lint/registry passes in 0.080s, including deterministic generation,
invalid-descriptor rejection, duplicate adapters and .a entry modules. Wave-27
rules remain in their owned typed drivers; this does not certify shared-registry
discovery or emitted-JavaScript agreement for these twelve ports. Registry output
is ignored and uncommitted.

# Owned checks and exact outputs

Shells source /workspace/adamic-tools/env.sh and use TMPDIR=/tmp/adamic-gate.
The frozen manifests, source pin and artifact environment paths are unchanged
from ../landing_c019/REPORT.md. Original cloud setup passed: ready 0s, warm cache
116s, total 116s; nproc=5. Every test writes output directly to a log file.
Only named obsolete reproducible scratch checker archives were removed for space;
current artifacts, source witnesses and evidence are retained.

```
go test -v -count=1 -timeout 15m ./stage1/cohere/typeaware -run '^TestWave27AgreementAndMutants$'
python3 stage1/cohere/typeaware/wave_27_next/validate.py
python3 stage1/cohere/typeaware/wave_27_third/validate.py
python3 stage1/cohere/typeaware/wave_27_fifth/validate.py
go test -v -count=1 -timeout 15m ./bridge/tsgo/checker ./stage1/cohere/typeaware -run '^(TestCoverageCheckerQuestions|TestFactEncoding|TestShapeAndNameFacts|TestExactIndexMatchesCompilerNodes|TestAdamicRootKeepsConfigDeclarations|TestFactsDecoderGuards|TestInspectRequestRefusals|TestPinnedTypeFlags)$'
go vet ./bridge/tsgo/checker ./stage1/cohere/typeaware
```

First gate PASS in 168.128s. Three Python gates print PASS. Focused regressions
PASS in 0.250s/73.229s; vet exits 0, empty output. Retained logs name every mutant,
refusal, corpus and timing observation.

Each control suite matches full finding ranges, ids, messages, fixes and ordered
suggestions/edits against independent production Go cohere. First: 105 controls,
39 findings, 17177 identical bytes. Next: 93 controls, 101 findings, 62201 bytes.
Third: 605 parseable of 639 candidates, 437 findings, 225707 bytes. Fifth: 186
parseable of 210 candidates, 108 findings, 64739 bytes. Independent Go parser
exclusions (34 and 24) are retained, not passing cases. Each batch matches the
287-file frozen repository corpus (18485 bytes) and 77 compiler files (5934 bytes),
both zero findings. Normal and Go-ASan/native--sanitize comparisons pass for
controls and both corpora; sanitizer stderr is empty. Fifth checks 351 internal
kind constants against Go and three named rule.json declarations. The owned
driver caches and hands the current node to those three listeners.

Rule mutants compile and exit 0 with empty stderr; only Go bytes catch them:
ORM nullable inversion (57), Serializable optional inversion (5466), Verify array
inversion (8264), pending-output inversion (80), race ownership inversion (52234),
blocking early-return inversion (581), Effect->EffectNever regex (28395), rest
declaration inversion (54), Hook message inversion (4962), regex arity (112016),
symbol ambient inversion (39039), typeof legal-string replacement (46485), await
contract bypass (33546). Parentheses give first differing output bytes.
Five normally exiting fact mutants are caught at 56 (declaration file), 1318
(CFG reachability), 12452 (never flags), 42069 (type-only imports), 61584
(declaration kind). Released queries return exact panic 70; every owned retained
registry mutant exits 0 and fails the required refusal assertion. Focused wire
regressions exercise the nineteen malformed-wire guards and wrong-kind/unknown-
question mutants, plus pinned type flags and checker queries.

# Native time and limits

Three alternating whole-process samples per engine retain identical streams.
Independent batch gates ran concurrently, so these are contended observations:
fifth repository native 627.383ms / Go 219.621ms (2.857x);
fifth compiler native 3744.965ms / Go 549.415ms (6.816x).

The freshly rebuilt compiler still exits 1 at additional_hooks_pattern.a:4:23:
"stage 0 can't lower RegExp with a nonconstant pattern yet". The isolated helper
preserves new RegExp(pattern, 'u'); runtime additionalHooks integration remains
unfinished. The fixed Effect regex uses the shared translation; no hand-rolled
matcher replaces it. Three React HIR/SSA/capture/post-dominance claims remain
parked per user instructions. The older nine ports still await handed-node
migration. Shared discovery/emitted-JavaScript comparison for owned ports, other
nondefault options, inherited port gates and full repository gate were not run.

Fresh selection scans 597 origin refs, 197 ranked rules, 33 Markdown claim records:
zero unclaimed rules. No new claim is made. The final fetch confirms the same
main and area tips; this branch contains both. See selection.json.
