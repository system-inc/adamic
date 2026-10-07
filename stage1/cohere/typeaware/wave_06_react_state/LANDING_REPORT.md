Built: rebased wave 06 onto current origin/main and revalidated twelve complete ports plus three partial React reporters.
Commits: main e8ba3d5d81de4d3773c723914fccd4c76248b965; rebased implementation b8e4be0fada4e084b69016ff15aa9a7480f0da67; evidence is the commit adding this report.
Checks: four full suites PASS against Go on controls and both corpora, ASan/UBSan and released handles; bridge and uncached filtered Node oracle PASS.
Mutants: all twelve rule mutants, four checker-projection mutants, released-registry mutant and six partial React reporting/refusal mutants caught again.
Not covered: full root gate and full React analyses; native React HIR/SSA, capture translation and memoization passes remain missing. No new claims.

# Landing refresh, 2026-10-07

Only `codex/typeaware-wave-06` was pushed by this worker. Its eleven commits
above the former main ancestor rebased without conflicts. Main advanced from
`e011f8f6` to `e8ba3d5d` during the first validation, adding compiler
call-target routing and devirtualization. A second rebase and complete repeat
of the wave's checks holds this report to the latter tip. No implementation,
shared-harness or protected compiler edits were needed.

Twelve completed ports remain separate from three unfinished React source
analyses. Rebase readiness does not turn partial reporters into completed
rules. No additional rules were claimed.

| Suite | Controls | Findings | Control bytes | Compiler bytes | Repository bytes |
| --- | ---: | ---: | ---: | ---: | ---: |
| Original nominal, duplicate and return | 166 | 129 | 36604 | 22,060 | 18,485 |
| Nexus child-process, response-status and independent-await | 84 | 42 | 28696 | 5,318 | 18,485 |
| Constructor rules | 109 | 51 | 23338 | 5,318 | 18,485 |
| Throw, backreference and arrow callback | 210 | 136 | 51544 | 5,318 | 18,485 |

These compare complete canonical finding, fix and suggestion fields, preserving
ordering, duplicates and spans. Normal and sanitized native agree with Go on
controls and both corpora. Every sanitizer run has empty stderr. The original
suite reports 53 findings across the 77 compiler roots; the other suites report
zero. All four suites report zero across the 287 repository roots.

Root manifests are frozen, but repository contents follow main. The old two
nominal findings disappear on both Go and native after rebasing. Control stream
lengths change with the scratch paths. Controls, corpus input hashes and exact
streams are retained, so neither change is concealed as a count-only pass.

All mutants were rerun:

- Original: reversed nominal symbol equality, duplicate span end+1 and useless
  return fix end+1 each compile, exit zero and lose Go byte equality. Removing
  async promised-type unwrapping also loses equality. Keeping the released
  registry entry fails the required panic-70 contract.
- Nexus: three rule diagnostic-ID mutations exit zero and lose Go equality;
  substituting the call node for its resolved declaration also loses equality.
- Constructors: disabled shadow rejection, callee span replaced by new-expression
  span, and removed parenthesis skipping each exit zero and lose equality.
  Flipping the declaration-file fact also loses equality.
- Callbacks: inverted throw classification, inverted forward-backreference
  ordering and corrupted arrow-fix text each exit zero and lose equality.
  Flipping the declaration-file fact also loses equality.
- Partial React: three reporter-ID mutants lose Go bytes; three removed
  analysis-refusal mutants fail required panic 70. These are reporting/refusal
  mutants, not full-rule analysis mutants.

Released-handle probes for both new raw questions pass. The full bridge
regression and its ownership, bounds, release and link-guard mutants pass.
The uncached filtered Node oracle passes nine source/native/emitted-JavaScript
fixture comparisons plus the one-byte mutation check. Root vet, formatting and
whitespace checks pass. The full root test matrix was not run.

The original suite passes in 290.610s, complete bridge in 96.168s, checker
package in 0.207s and filtered uncached Node oracle in 2.235s. Bridge mutants
are individually accounted for:

| Mutant | What caught it |
| --- | --- |
| Input length advanced one byte | ASan heap-buffer-overflow |
| Output length advanced one byte | ASan heap-buffer-overflow |
| Released handle kept live | Stale-handle assertion |
| Type queried at source-file position | Independent Go mismatch at byte 6 |
| Link opt-in guard removed | Unlinked-command refusal check |
| C output free removed | LeakSanitizer |
| Region result allocated on heap | LeakSanitizer |

The Node oracle's appended dedication byte is caught by stdout disagreement.

React validation remains reporting-only: four actual production Go findings
match 2,367 bytes of normal and sanitized native reporter output. Go supplies the ranges, so
this proves neither native verdicts nor native span discovery. All three
analyzers explicitly refuse. The unchanged shared parser still exits 70 on the
owned JSX probe. Its exact message is saved in the evidence.

The refreshed search across 431 fetched origin refs finds no native uses of
`ControlDominators`, `AsCompilationUnit`, `UnconditionalBlocks` or
`ForFunctionWithoutManualMemoization`. This is a scoped API search, not proof
that every possible implementation spelling was checked. Isolated JSX parser
work elsewhere does not supply React HIR/SSA, capture translation, memoization
inlining, control dominance or compilation-unit lowering. Those prerequisites
still block the three analyses described in REPORT.md. Claims remain reserved.

# Native against Go

Final timings use sequential fresh processes after validation jobs finish,
existing built artifacts and warm filesystem caches. Complete output is
compared again for every row. These are single observations of each whole
three-rule suite, including loading, parsing, queries, judging and formatting.
Native is slower in every observation.

| Suite | Population | Native wall | Go wall | Native / Go |
| --- | --- | ---: | ---: | ---: |
| original | compiler | 24.876467s | 3.392108s | 7.334 |
| original | repository | 0.948017s | 0.292696s | 3.239 |
| next | compiler | 5.143037s | 1.752858s | 2.934 |
| next | repository | 0.665913s | 0.263535s | 2.527 |
| constructors | compiler | 1.911880s | 0.389938s | 4.903 |
| constructors | repository | 0.322687s | 0.163813s | 1.970 |
| callbacks | compiler | 2.984326s | 0.489428s | 6.098 |
| callbacks | repository | 0.458268s | 0.155198s | 2.953 |

React has no comparable lint benchmark while native analysis refuses.

# Commands and artifacts

All test output goes directly to log files. Full compressed stdout/stderr,
continuation commands and exits, controls, input hashes, origin SHAs, setup
and isolated timing records are in [landing_evidence](landing_evidence).
Run from the repository, sourcing `/workspace/adamic-tools/env.sh` and setting
`TMPDIR=/workspace`:

```sh
ADAMIC_WAVE06_CORPORA=1 ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-06-typescript ADAMIC_WAVE06_ARTIFACTS=/workspace/wave-06-landing2-original go test ./stage1/cohere/typeaware -run '^TestWave06(AgreementAndMutants|PinnedFlags)$' -count=1 -timeout=30m -v > /tmp/wave-06-landing2-original.log 2>&1
python3 stage1/cohere/typeaware/wave_06_next/validate.py --scratch /workspace/wave-06-landing2-next --compiler /workspace/wave-06-typescript > /tmp/wave-06-landing2-next.log 2>&1
python3 stage1/cohere/typeaware/wave_06_constructors/validate.py --scratch /workspace/wave-06-landing2-constructors --compiler /workspace/wave-06-typescript > /tmp/wave-06-landing2-constructors.log 2>&1
python3 stage1/cohere/typeaware/wave_06_callbacks/validate.py --scratch /workspace/wave-06-landing2-callbacks --compiler /workspace/wave-06-typescript > /tmp/wave-06-landing2-callbacks.log 2>&1
python3 stage1/cohere/typeaware/wave_06_react_state/validate_partial.py --scratch /workspace/wave-06-landing2-react > /tmp/wave-06-landing2-react.log 2>&1
go test ./bridge/tsgo/... -count=1 -timeout=15m -v > /tmp/wave-06-landing2-bridge.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures)\.a$' -count=1 -timeout=10m -v > /tmp/wave-06-landing2-node.log 2>&1
go vet ./... > /tmp/wave-06-landing2-vet.log 2>&1
gofmt -l cmd internal bridge/tsgo stage1/cohere/typeaware > /tmp/wave-06-landing2-format.log
git diff --check > /tmp/wave-06-landing2-diff.log 2>&1
```

`bash cloud/setup.sh` succeeds: Go 0s, clang 0s, Node 0s, submodules 0s,
cache warm 105s, done 105s. `nproc` is 5, CPU quota 4, memory 17.6 GB.
Go 1.27.1, clang 20.1.8, Node 24.19.0 and both submodule pins are unchanged.
No PR is opened. Only owned rule-directory evidence is newly written.
