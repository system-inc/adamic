import json,pathlib,re
p=pathlib.Path(__file__).resolve().parent
mutants=json.loads((p/'mutants.json').read_text());boundary=json.loads((p/'cleared-boundary-mutants.json').read_text())
assert len(mutants)==16 and all(r['caught'] for r in mutants)
assert len(boundary)==2 and all(r['caught'] for r in boundary)
measured=json.loads((p/'cleared-final-leaf-seconds.json').read_text())
assert all(r['seconds']<60 for r in measured)
text="""
Final validation after runtime clearance

Runtime explicitly cleared proposed-opaque-diagnostic.patch, including both conditions: the local kind rewrite is only after all matching returns in adamic_view_untagged_plain_select, immediately before its count-zero failure call; the generic selector's found word is exactly adamic_view_union_kind_name(value->kind), with no payload inspection. With zero contracts the generic selector cannot match. The patch applied exactly as cleared. No expected text changed.

The opaque-pointer control is not new to this base: 1bf89669871d2c3ae50ebb50b661c7f202c8baf1 introduced it, and ac19a4e6b0c4c5986bbc0d74a7578e68ab40c306 enabled sanitized execution, both before 2c410041's parent. Header and control are unchanged between that parent and dbdf977f5b. The new generic payload read in D was an existing regression, not a rebase API change.

The final cold full-reader selection passed all 52 packages but exposed five D leaves over budget: TestFX7TextScalarStatic 66.86 s, TestLoomViewsDTupleUnion 68.30 s, TestFX7Depth65 68.52 s, TestLoomViewsDTupleObject 68.70 s, TestLoomViewsDWideCallable 69.75 s. These combined three-backend leaves were split into top-level native-release, Sanitized and JavaScript controls before further verification. Every original comparison, Node check and expected diagnostic was retained. The original names now select native release. Catchers for the three JavaScript mutants use their JavaScript leaves. The new tuple-match control's combined two-backend version also took 60.91 s and was split into Release and Sanitized leaves, first passing in 5.62 / 5.72 s. Only the revised grain was subsequently run.

The unchanged existing TestViewMixedUnionUnknownAndUnavailable passed every release and ASan scenario, including missing-adapter and missing-contract controls. Its cold run took 60.84 s with concurrent compiler work. That existing over-budget whole leaf was not rerun; only its focused ASan subcases were used for the reverse mutant. The final focused selection does not select it. This is disclosed as a budget limitation, not hidden as a sub-60 observation.

Final commands, all bounded, with direct log output:

* timeout 180 go test ./internal/native -run '^(TestViewMixedUnionUnknownAndUnavailable|TestFX7TupleSelectionBeforeDiagnosticKind)$' -count=1 -v -timeout 90s: PASS, cleared-selector.log. This predates the new control's grain split.
* timeout 360 go test $(cat reader-packages-test-imports.txt) -run '^(TestFX7|TestLoomViewsD|TestCallableContract|TestViewCheckedCastMetadata$|TestRuntimeSetsAreEachUsed$)' -count=1 -v -timeout 90s: PASS for all 52 packages, cleared-readers.log; the five grain observations above caused the split.
* timeout 150 go test ./internal/flow -run '^TestFlowCorpus(UnitsCoverEveryProgram|SetupIsShared|Remainder)$' -count=1 -v -timeout 80s: PASS, 0.02 / 0.19 / 10.50 s, cleared-flow-corpus.log.
* timeout 180 go test ./internal/lower ./internal/native ./internal/javascript -run '^(TestView|TestUntaggedView|TestReadiness|TestInheritance.*Views|TestFunctionValueUnionViewsStayNotYet$|TestCallableNamespace|TestAView)' -skip '^TestViewMixedUnionUnknownAndUnavailable$' -count=1 -v -timeout 90s: PASS, cleared-related-regressions.log.
* timeout 300 python3 cleared-boundary-mutants.py: both mutants caught, each test bounded by timeout 150 and -timeout 55s. Moving the rewrite before the matching loop changes the empty tuple's successful member-1 result into exit 70 with found array. Restoring generic payload inspection reproduces ASan stack-buffer-overflow. No compile or timeout failure is credited.
* timeout 840 python3 run-mutants.py: all 16 source-report mutants caught independently, restored in finally; each test bounded by timeout 120 and -timeout 55s. Diagnostic mutants target their relocated untagged classification, not the removed mixed code.
* cleared-final-checks.sh: timeout 240 go test ./internal/javascript ./internal/lower ./internal/native ./internal/oracle -run '^(TestFX7|TestLoomViewsD|TestCallableContract|TestViewCheckedCastMetadata$|TestRuntimeSetsAreEachUsed$)' -count=1 -v -timeout 55s: PASS, all revised leaves below 60 s, cleared-final-focused.log. P09, P67 and P72 each still check native release, sanitized native and JavaScript against unchanged shared diagnostic text.
* Same final script: timeout 150 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -v -timeout 90s, timeout 150 go vet ./internal/ir ./internal/javascript ./internal/lower ./internal/native ./internal/oracle, timeout 150 go build -o /tmp/loom-r-views-d ./cmd/adamic, timeout 360 python3 replay.py: all PASS on restored final sources. Their exact logs are cleared-final-*.log.
* final-evidence.py: all 23 final probes retain the source commit's exact Node/native/JavaScript results. Native artifact hashes are recorded and binaries moved outside the repository. All final focused and reader-guard leaves are under 60 s.

Counts remain the single Linux clang 20.1.8 regeneration documented above: 18 new rows, zero modified/removed existing rows. The cleared change only moves failure-word classification, with no counted allocation or fixture changes; the matching-order control is inline C rather than an oracle fixture, and backend splitting adds no fixture. No second complete counts regeneration was warranted. No row was hand-edited.

runtime-lines-versus-source.md lists every final changed runtime line against 2c410041. The full object.c differences remain inherited main work; the only additional runtime changes are the exact cleared -5 mixed / +6 untagged relocation. No other runtime file changed. There were no conflicts or conflict resolutions. No IR map or runtime feature set was added.

Final mutant catchers:

| Mutant | Catcher | Command seconds |
|---|---|---:|
"""
for r in mutants:text+=f"| {r['name']} | {r['test']} | {r['seconds']:.2f} |\n"
for r in boundary:text+=f"| {r['name']} | {r['test']} | {r['seconds']:.2f} |\n"
text+='\nFinal passing test measurements (setup inside each test included):\n\n| Test | Seconds |\n|---|---:|\n'
for r in measured:text+=f"| {r['test']} | {r['seconds']:.2f} |\n"
text+='\nThe final branch is one commit on dbdf977f5b, pushed only to compiler/loom-r-views-fx-d. No ahra command, full gate, PR or Loom submission was run. The cold budget exceptions above and the five earlier existing count-update exceptions remain disclosed. The split preserves coverage rather than deleting over-budget assertions.\n'
report=p/'REPORT.md';s=report.read_text();assert 'Final validation after runtime clearance' not in s;report.write_text(s+text)
print('Final report appended: 18 mutant kills, complete final measurements and runtime-line references')
