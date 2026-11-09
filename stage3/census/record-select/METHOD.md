# Selection measurement method

The compiler revision is f0c358f2caa605525a58dad49fabb38bd92418a9. Its merge base with main is fd8b74b9bd4fa5d8012dfcc0c0252c3f1964cd4d. Adaptations come from that base, without merging Memory into main. TypeScript source is 050880ce59e30b356b686bd3144efe24f875ebc8.

The throwaway test is compiled in package lower. The scratch loader retains diagnostics, as the latent census does, but parses adapted src/compiler/tsconfig.json and replaces its root list with src/tsc/tsc.ts. It retains the production source filesystem, Adamic prelude, bundled libraries, single-threaded checker, and Node-library loading path. Production Load remains strict in the repository; the permissive loader exists only in the throwaway worktree.

Every non-declaration source file in the entry-root compiler program is attempted as a group entry, with tsc first. Each attempt calls lowerCheckedWithCycleTypes with fresh lowering state and no program-region plan. Its moduleOrder determines the group. It does not remove failing statements, invent closures, or infer captures from syntax. The original production prepasses and all lowering stages preceding findCycles run in their original order. Instrumentation records the last phase, entry into findCycles, entry into graphTypes, completion of selection, and each selected Tarjan component. Post-selection acyclic cache adoption is not counted as another selected component.

The observer is called inside the existing selected-component branch. It does not implement a second SCC algorithm. Seed membership is the existing seeds map. For a slot seed, the observer repeats slotsOf to recover its original first refusal and the write formatted by writtenAt. It records write kind, field, source position, function and freshness reason. Captured cells carry their declaration position. Record membership uses the actual recordElement predicate, not the static TypeScript approximation. Literal attribution matches inferred or contextual identities to the component record type; ESMModuleOrder from the tsc entry supplies zero-based load indices.

Each group runs a second time with the final throughRecord stop condition forced false. All seed and component selection code is identical. Promise and earlier record-slot refusals stay enabled. The compiler does not emit or run native code. Selection entry without completion is not counted as a complete empty selection. A blocked group never contributes a zero to a whole-program exposure claim.

## Replay the scratch test

Start with a clean worktree at the compiler pin and an adapted tree made by the merge-base main's apply.sh. Initialize the pinned cohere submodules and run the setup and npm commands from the task. Keep all logs in files. In the compiler worktree, apply scratch.patch and copy record_select_census_test.go.txt to internal/lower/record_select_census_test.go. No scratch production file or live Go test belongs in the report commit.

The measured invocation was equivalent to:

```sh
source /workspace/adamic-tools/env.sh
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/record-select-typescript
export RECORD_SELECT_ENTRY=/workspace/record-select-adapted/src/tsc/tsc.ts
export RECORD_SELECT_CONFIG=/workspace/record-select-adapted/src/compiler/tsconfig.json
export RECORD_SELECT_OUT=/tmp/record-select.jsonl
export RECORD_SELECT_FIXTURE_OUT=/tmp/record-select-fixture.json
go test -c -o /tmp/record-select.test ./internal/lower > /tmp/record-select-build-final.log 2>&1
rc=$?; echo "child_exit=$rc"; test "$rc" -eq 0 || exit "$rc"
/tmp/record-select.test -test.run '^TestRecordSelect(Census|ObserverFixture)$' -test.v -test.timeout 100m > /tmp/record-select-run.log 2>&1
rc=$?; echo "child_exit=$rc"; test "$rc" -eq 0 || exit "$rc"
```

The actual invocation used a Python parent that watched the run log, terminated the entire child process group after 1,800 seconds without log growth, waited for the child, and saved its own return code in run-exit.json. Stop and record a silent run rather than silently waiting beyond that limit. The Go test flushes and syncs every header and group row. Its JSONL first row becomes dump.json's header, and the remaining rows become groups. The separate fixture output becomes fixture.json. Run `python3 validate.py` beside those JSON files to compare all normal/control pairs and verify the fixture.
