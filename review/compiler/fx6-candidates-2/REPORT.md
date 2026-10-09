# Second fx6 gate candidate

Base: 9aff8dbf25a0dd99f8e71d0d636bf835912c4a9d (compiler/fx6-candidates). Merged routing d587b09e, valid programs fb1d7339, then main cf735d9fba9e38de6368575e5630e44375a86eaf.

Resolution preserves checked reads by receiver type identity and member. Object destructuring calls readViewMember, the same boundary as dot and finite keys; tuple destructuring retains receiver/type/location metadata on its existing path. Fixed p35-p38, p51, p53 remain active, and p49 retains its ruled contract stop. p70 remains an active located NotYet with its scalar length workaround.

The initial tracking ref for candidates was stale at 339c86b3. After explicitly fetching the requested branches, the newly created local branch was reset to the required 9aff8dbf before performing the recorded merges. No remote history was rewritten.

The old blanket literal-element refusal is replaced by checked routing, as requested. The old refusal revert is historical and cannot be applied to a function that routing removed; its current proof is element-bypass. Unrouted element forms retain finiteElementRead's located NotYet guards for unsupported key, prototype/method, optional and storage representations. Enumeration, spread and tuple binding support are not claimed by this merge.

All five current mutants are independently restored in finally blocks, and fail on behavior rather than a build error:
- element-bypass: TestCheckedViewElementP05 catches exit 0 instead of the ruled exit 70. Native prints a tiny run-specific number and true; unchecked JavaScript prints NaN and false. No sanitizer failure.
- destructure-bypass: TestCheckedViewDestructuredUnion catches true and exit 0 instead of the membership stop, across all three backends.
- revert-receiver: TestFX6P37 catches JavaScript stdout 3 instead of Node 3 followed by number 5.
- revert-destructure-type-id: TestFX6P53 catches empty JavaScript stdout instead of Node n5. This reverts the field type identity in its new shared helper location.
- revert-p70: TestTupleObjectViewRefused catches nil instead of the located NotYet.

Runner commands: timeout 600 bash -c 'python3 review/compiler/fx6-candidates-2/run-routing-mutants.py && python3 review/compiler/fx6-candidates-2/run-valid-mutants.py', after sourcing /workspace/adamic-tools/env.sh. Individual child commands use -timeout 90s and subprocess limits. Every diff and failure log is archived here; production sources were restored before green checks.

Setup: GOPROXY=https://proxy.golang.org|direct timeout 240 bash cloud/setup.sh. Go 0.030s, Node 0.029s, submodules 0.099s, clang 0.238s, build cache 45.627s, total 45.654s. nproc=5; cgroup cpu.max=400000 100000 (four CPU quota). Toolchain sourced from /workspace/adamic-tools/env.sh.

Validation commands, output in the named log files:
- timeout 300 go test ./internal/lower -count=1 -timeout 240s (lower.log): PASS 78.375s.
- timeout 120 go test ./internal/ir -run TestCallTargetReaders -count=1 -timeout 90s (readers.log): PASS 22.520s.
- timeout 360 go test ./internal/oracle -run 'TestCountsAreRecorded|TestCheckedView(Element|LiteralTypedKey|Destructured)|TestFX6|TestNativeAgreesWithNode/internal/lower/testdata/tuple_object_views/p70_copy.a|TestReviewProgramsAgreeWithNode/fxspptb_oct9_views_p(0[1567]|35|36|37|38|49|51|53)_|TestReviewProgramsRefuse/fxspptb_oct9_views_p70|TestReviewProgramsNoLooseFiles' -count=1 -v -timeout 5m -args -update-counts (oracle-counts.log): PASS 93.783s; counts.md refreshed with no diff.
- Integration lane command from the repository root, after commit: git fetch -q origin main devtools/fast-gate cloud/merge-tree && git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 - (lane.log).
- git diff --check: PASS.

No new test leaves were added. No full oracle package or full repository gate was run; integration owns the gate. This combines the item 132 routing, items 134/136 valid view fixes, and item 137 narrow refusal toward the requested valid-view unit. Full tuple/object view interoperability remains outstanding. No PR opened.

Restored focused fixtures: timeout 120 go test ./internal/lower -run "^TestTupleObjectView|^TestFX6|^TestView.*Element" -count=1 -v -timeout 90s: PASS 0.992s; all leaves under one second, archived in focused.log.
