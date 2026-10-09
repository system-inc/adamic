Built runtime-store corpus classification, closed push-gap checks, restored scanner offset pushes, and registered self-compare fixtures for task #wj4pmt1.
Base: 379dbfb3; delivery commits are reported with the pushed HEAD.
All assigned checks pass; all 68 fresh remainder programs pass after installing pinned Node API types.
Six independent mutant checks exited 1 at their intended assertions; mutants.json records each exact command and catcher.
Not covered: other gate families, whole packages, the full gate, readiness/stack semantic changes, and the full scanner corpus.

The conservative classification assumes namespace containers lack a checker holder type and builtin Error initializers synthesize their string prefix without a source receiver. It recognizes site-zero record stores only on IR locals marked NamespaceObject, and string name/message stores only in the initializer paired with a BuiltinError class. Those stores still enter ProveWrites. All other missing sites and unknown operations still fail the census. Removing a real source assignment's site in step21_saved_error.a fails it.

Both unchanged push probes agree with source Node. Markdown also checks native, the JavaScript backend and leaks. Scanner's push probe checks sanitized native. The restored two-value supplementary-character push is checked against the upstream Go scanner on source Node and sanitized native, including a following token's byte offsets. The self-compare unit retains the existing recorded Node, native/sanitizer/leak and stage0 assertions for all three fixtures.

This supports the lowering chain's step 21 Error corpus and the closed multi-argument push work, with selectable fixture coverage under step 79. It changes no compiler semantics or pinned readiness/stack outcomes.

Commands, all tests with -count=1 -timeout 90s and an outer timeout 90s:

```text
go test ./internal/fresh -run 'TestFreshCorpusRemainder/../oracle/testdata/(placeholder_nonnull_namespace|step21_error_subclasses|step21_saved_error)' -v
PASS: four leaves, 0.03s each; package 0.144s
go test ./internal/fresh -run '^TestFreshCorpusRemainder$' -v
PASS: 68 programs, package 2.065s
go test ./stage1/cohere/markdownblocks -run '^TestParserRepresentationProbes$/^gaps$/^10_multiple_push.ts$' -v
PASS: leaf 0.33s, package 0.340s
go test ./stage1/typescript/scanner -run '^TestGapStandsWhereGapsMdSays$' -v
PASS: 13.67s with product preparation; restored final run 2.71s, package 2.715s
go test ./stage3/fixtures -run '^(TestFixtureDirectoriesHaveTopLevelTests|TestFixturesSelfCompare)$' -v
PASS: new unit including preparation 14.942s; restored final package 1.016s
Directory guard: 13 directories, 182 fixtures, including self-compare's three fixtures.
go test ./internal/ir -run '^TestCallTargetReaders$'
PASS: package 21.965s
python3 review/compiler/after-chain-fix-fresh/run-mutants.py
PASS: six caught, exit 0
```

Mutants and observations:

| Mutant | Catcher |
| --- | --- |
| Remove source assignment sites | fresh step21_saved_error: a write lowering didn't record, function 1 |
| Drop second push argument | markdown native byte comparison, first difference at 1, lengths 2/4 |
| Drop second push argument | scanner native push prints a rather than Node's ab |
| Supplementary end offset is bytes + 3 | scanner port Identifier 6 9 versus Go Identifier 6 10 |
| Remove self-compare top-level unit | directory guard: self-compare has no top-level test |
| Compare closure to itself with !== | self-compare recorded Node true versus observed false |

Setup: GOPROXY was exported as https://proxy.golang.org|direct before cloud/setup.sh. Initial setup hit its outer 90s limit in Go cache warming. Timing lines: Go 0.073s, Node 0.082s, clang 0.482s, markdown dependencies 1.237s (npm step 1.046s), submodules 17.064s. nproc: 5. /opt/adamic-tools/env.sh does not exist here; /workspace/adamic-tools/env.sh was generated and sourced in every test shell. Focused cache warming completed through bounded test builds. The complete remainder initially found the missing @types/node 25.3.3 dependency; npm ci --prefix stage3/api installed the pinned lockfile, then the remainder passed. api-setup.log records it.

The first focused fresh selector selected zero leaves because Go splits test names at slashes; it was corrected before any passing coverage claim. The first mutant runner expected a different markdown diagnostic phrase; the real comparison had failed, and the corrected complete runner was rerun successfully. Initial lane-script retrieval used FETCH_HEAD instead of the explicit remote tracking ref and found no file; explicit fetching of the two integration refs corrected it.

No permanent fixtures were added or changed, so no new counts row or counts update is needed. Only existing scanner source changed to remove the closed gap workaround. Evidence and full test outputs are in this directory; mutant source is .go.txt or .patch, never compilable Go. Final lane-check output is lane.log. No PR was opened.
