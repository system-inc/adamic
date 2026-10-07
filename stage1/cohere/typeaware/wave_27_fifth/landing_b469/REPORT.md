Rebased twelve completed default-option ports onto lint-area b46914832, containing main c7991b900.
Tested rule inputs are unchanged; rebased code 580970fcdb980c81d4ff885ed0d6b773bde7923a.
Four batches match Go again on controls and both corpora, normally and under sanitizers; registry tests PASS.
Thirteen rule mutants and released/retained-handle checks rerun; five fact mutants retain prior b84 evidence.
No unclaimed rules remain; dynamic RegExp, parked React analysis, shared certification and full external gate remain unresolved.

# Scope and commands

Origin/area/stage1-lint advanced only stage1/cohere/lint with legacy syntax-rule
registration migration. Compiler, bridge, cohere submodule, module configuration,
TypeScript parser and owned type-aware rule sources are unchanged. recheck.py
asserts that their Git diff from b84a9d931 to the fetched area tip is empty before
running. Reused binaries and checker archives are those freshly built and fully
validated on b84a9d931, recorded in ../landing_b84/REPORT.md. They were not rebuilt
for this source-identical rebase. All 26 owned patches rebased cleanly. No shared
or protected source was edited and main/area branches were not pushed.

Shells source /workspace/adamic-tools/env.sh. Exact commands:

```
git rebase origin/area/stage1-lint
go run ./cmd/lint-registry
go test -v -count=1 ./stage1/cohere/lint/registry
python3 stage1/cohere/typeaware/wave_27_fifth/landing_b469/recheck.py
```

Subprocess output goes directly into retained stdout/stderr files; the runner's
summary is wave-27-b469-recheck.log. All comparisons PASS. Shared registry tests
PASS in 0.116s, including invalid-descriptor and .a module cases. These tests do
not establish shared registration or emitted-JavaScript certification for the
owned rules, which still use their typed drivers. Compiler and repository inputs
are supplied, not skipped. Original setup timing remains ready 0s/warm cache
116s/total 116s, nproc=5. No full repository or 17-check external suite was run;
no skipped correctness test is claimed passing and no check was weakened.

# Fresh agreement and mutants

Complete findings, ranges, descriptions, fixes and ordered suggestion edits agree
with production Go. Controls: first 17177 bytes/39 findings, next 62201/101,
third 225707/437, fifth 64739/108. Frozen 287-file repository and 77 compiler-file
corpora match 18485 and 5934 bytes, zero findings, for every batch. Normal and
sanitized streams match with empty native stderr. Existing parser exclusions
remain 34 third-batch candidates and 24 fifth-batch candidates, not passing cases.

Thirteen existing rule mutants rerun with exit 0 and empty stderr; only Go byte
comparison catches ORM 57, Serializable 5466, Verify 8264, pending output 80,
race 52234, blocking 581, Effect regex 28395, rest 54, Hook message 4962,
regex arity 112016, symbol 39039, typeof 46485, await 33546. The five bridge fact
mutants were not rebuilt/rerun here because the bridge/compiler inputs are
unchanged and those mutant binaries had been removed by the original runner.
Their b84 gate evidence remains retained at first differences 56/1318/12452/
42069/61584. Released raw-type, wave queries, binding queries and four checker
links again panic exactly 70. Retained-registry mutants again exit 0 and therefore
fail the required refusal contract. No new rule logic was written this turn.

# Fresh timing and remaining scope

Three alternating whole-process samples for fifth batch retain identical output:
repository native 557.468ms / Go 237.125ms (2.351x);
compiler native 3542.208ms / Go 747.244ms (4.740x).

Runtime additionalHooks remains blocked by nonconstant RegExp lowering, unchanged
since b84. The isolated new RegExp(pattern, 'u') helper is present. Three React
HIR/SSA/capture/post-dominance claims remain parked. Older nine ports await handed-
node migration. Shared discovery/emitted-JavaScript certification, nondefault
options, inherited gates, full repository gate and full 17-check external suite
remain untested. Prior checker/vet/predicate proof results remain those of b84;
those packages were not retested here.

Audit: 644 origin refs, 197 ranked rules, 33 Markdown claims, zero unclaimed rules.
No next claim is taken. The pre-push fetch confirms main c7991b900 and area
b46914832 remain current, and this branch contains both. selection.json retains
the full audit evidence.
