Rebased: codex/lint-helpers-from-lint-wave1-12 onto current main f8013f0baac41ddc340d76f83bddde38536a8f07; no new claims or production helper changes.
Commits: previous pushed helper tip b32ed264df65f493965eecb7d4d7038b73543738; rebased implementation tip before this evidence commit a55c1d2d93d837b86e21c93dbf92f1e491d61bf6.
Checks: owned helper suite PASS 115.493s, inherited foundation PASS 152.967s, comments PASS 200.106s; vet clean; filtered uncached comparison oracle PASS 6.894s.
Mutants: all eight owned controls and inherited helper/comment controls caught; actual Go comparisons retain exact supported-domain and parser-adapter boundaries.
Uncovered: exact nextBuildCount still refuses bigint return lowering; the sibling rule branch is not default-integration green; no new helper or whole-rule parity is claimed.

Main f8013f0b changes lowering and map/set runtime code. The branch was rebased before validation and preserves main's compiler files exactly. The ownership claim and both delivered .a helper implementations remain unchanged. Their 40,221 rows / 1,164,929 observation bytes match actual Go, source Node, emitted JavaScript and ASan/UBSan native, with all six semantic mutations caught by comparison. Unsupported segment inputs retain their exact refusal observations and a compiling guard mutant; the compiling number-counter approximation disagrees with Go at row 5 on all three backends. The exact counter itself is not delivered.

The inherited helpers and comments suites also pass, including their existing mutants and explicit parser gaps. Their existing Node/native adapter scope is not reclassified as whole-rule or emitted-JavaScript parity. f801-oracles.log preserves complete output for all three packages.

Commands, after sourcing /workspace/adamic-tools/env.sh:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/wave12 ./stage1/cohere/lint/helpers ./stage1/cohere/lint/helpers/comments -count=1 -v -timeout=20m > /tmp/wave12-f801-helpers.log 2>&1
go vet ./stage1/cohere/lint/helpers/... > /tmp/wave12-f801-helpers-vet.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' -count=1 -v -timeout=10m > /tmp/wave12-f801-comparison-oracle.log 2>&1
```

Setup: ADAMIC_TOOLS=/workspace/adamic-tools bash cloud/setup.sh passed. Go, clang, Node and submodules ready at 0s; cache warm 179s; done 179s, nproc 5, four-core quota, 17.6 GB. Cohere remains 715ba94f3608a6500086b1076ce5cb7e51b836db. The filtered oracle bypassed caches with native and Node hits 0, misses 1.

Both delivered helpers still remove two dependencies from each of the six Tailwind consumers named in README.md; zero final rule blockers are removed alone. No new claim, main push or area/ push is made. The rule branch's shared profile/registration/supplied-node boundaries still prevent passing the landing cap, so selecting more helpers remains out of scope. No complete repository gate, missing consumer repository fixtures, raw invalid UTF-8, arbitrary non-ASCII segment separators or fresh throughput result is claimed.
