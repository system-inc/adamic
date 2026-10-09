Built test-only lowering guards toward task #jhxsq14, private/public statics, readiness messages, and namespace bindings.
Commits: see the delivery branch log; the implementation and evidence contain only tests and review files.
Focused commands: bounded internal/lower leaves pass; exact replay commands and exits are in probe-results.json and restored.jsonl.
Mutants: M06 fails Node agreement, M13 fails the Node panic message, M12 fails the exact named-binding assertion; all seven empty-answer probes fail.
Limit: M12's Node comparison survives because the parser-factory JavaScript is byte-identical under that mutant; no behavioral kill is claimed for M12.

The branch started at main 68db8ddd145281a62655452496bdc32ef848bdf3 and fast-forwarded to c0a7667baadaf161d6bb9066e0838b32774caa6c before the final focused run. The evidence was fetched from the two requested branches and taken from exact audit commits 9a19cbc6 and 1915b058. Their audit bases were 8171b3173bdb and f91994f01970. All three original mutant diffs apply unchanged on current main, with no hunk-offset adaptation or semantic rebase needed. The final replay preserves git apply --check --verbose output for every diff.

M06's source and lowered JavaScript both run with `node --disable-warning=ExperimentalWarning oracle/node.mjs`. Source Node prints `29 11 11` and exits 0. The first two values are the private static and public static on Box, and the third is the inherited public static on Child. Under M06, the lowered program exits 70 without stdout, with `Cannot read private member field from an object whose class did not declare it`. The initial same-class-only witness survived, so inherited public access was added to expose the misplaced private brand check.

M13's definite-assignment check owns an Adamic panic, which source JavaScript does not insert. The generated program is run through Node and must exit 70 with exactly `adamic: panic: read before assignment: field 'n' in new Box().n`. The original M13 changes this to `field 'n' in n` and fails the assertion. This checks the observable message, not an IR snapshot or readiness count.

M12 is caught by TestParserFactoryBindingHoisting immediately after declareModule: exactly one NamespaceVar named factoryCreateNodeArray and one named factoryCreateNumericLiteral must exist. The body must also be nonempty. The separate source-versus-lowered Node program prints its two factory binding values, `11 22`, and compares stdout and exit status. That behavioral test passes under M12. A temporary evidence test additionally lowered the original namespaces_parser_factory.a at a stable path, before and under M12: both JavaScript artifacts were 10,320 bytes, SHA-256 d13c3e58a27feb38584cfbc5ab2338e7a93933145f5e0e9cb003b16ec3a045e7; cmp exited 0. Its logs are M12-generated-base.log and M12-generated-mutant.log. The temporary Go test was removed before committing. Observation: that artifact cannot produce a behavioral difference on Node. Inference: for these global bindings, changing ir.Assign to ir.Declare is erased by the JavaScript emitter's global-assignment path. The amendment's requested behavioral M12 kill was therefore not achieved; the original brief's named lowering-binding guard was retained instead of claiming the Node comparison caught it.

compiler/lower-agree was absent from `git ls-remote origin refs/heads/compiler/lower-agree` when the runner was implemented. staticsAgreeWithNode is a minimal local runner following class_static_guard_test.go, with a ten-second bound on each Node command.

The five u030 rows now have floors: the two repair tests and sound neighbors require an executable body containing console writes, while each rejecting signature test first requires a supported optional generic object signature to be recognized as ir.Object. The namespace initialized-read proof also requires executable output, and the ambient-host initialization test first checks that a runtime namespace read before initialization is rejected. The parser-factory binding and body floors separately reject both declareModule and namespaceBody empty output. Every exact probe now fails by a test assertion, with no compile-error kill. See probe-results.json for each command, failed leaf, exit 1, and wall seconds. Every compiler diff is restored after its run.

Setup used `export GOPROXY='https://proxy.golang.org|direct'` followed by `timeout 240 bash cloud/setup.sh`. Timing lines: Node ready 0.031s; Go ready 0.033s; clang ready 0.228s; Markdown dependencies ready 1.395s; submodules ready 17.407s; Go build ready 211.380s; cache warm 211.473s; done 211.502s. nproc=5, cgroup quota 4 CPUs. Setup printed `/workspace/adamic-tools/env.sh`; the initially attempted `/opt/adamic-tools/env.sh` did not exist, so subsequent shells source the printed path. Existing host tests also needed their pinned Node types: `timeout 120 npm ci --prefix stage3/api` exited 0, without any tracked dependency change. The first witnesses used console's unsupported numeric/multiple arguments; those checker failures were corrected to template-string output before mutant validation.

Final focused command (output in restored.jsonl):

```sh
source /workspace/adamic-tools/env.sh
timeout 100 go test -json ./internal/lower \
  -run 'Test(PrivateAndPublicStaticsAgreeWithNode|NamespaceFactoryBindingsAgreeWithNode|ReadinessErrorIncludesReceiverExpression|ClassWrongOutput(KeysRepair|PrivateRepair)|DefiniteAssignmentSoundNeighbors|ClockGenericReturnsT01Rejects(Null|Index)BeforeBody|ModuleNamespaceInitializedReadProof|ParserFactoryBindingHoisting|NamespaceAmbientHostInitialization)$' \
  -count=1 -timeout 90s
```

Replay command: `timeout 360 python3 review/compiler/statics-guards/replay.py`, after sourcing the toolchain. The runner bounds every go test with a 100-second subprocess timeout and `-timeout 90s`, and bounds patch application. Only targeted leaves run, never the whole lower package or repository gate. No oracle fixture or counts.md change was made. The new tests exercise Node, not native emission or sanitizers.

Final top-level leaf timings, all below 60 seconds:

| Test | Seconds |
| --- | ---: |
| TestModuleNamespaceInitializedReadProof | 0.03 |
| TestClassWrongOutputKeysRepair | 0.1 |
| TestReadinessErrorIncludesReceiverExpression | 0.14 |
| TestNamespaceFactoryBindingsAgreeWithNode | 0.22 |
| TestClockGenericReturnsT01RejectsIndexBeforeBody | 0.13 |
| TestPrivateAndPublicStaticsAgreeWithNode | 0.28 |
| TestParserFactoryBindingHoisting | 0.09 |
| TestClockGenericReturnsT01RejectsNullBeforeBody | 0.12 |
| TestClassWrongOutputPrivateRepair | 0.11 |
| TestDefiniteAssignmentSoundNeighbors | 0.26 |
| TestNamespaceAmbientHostInitialization | 1.18 |
