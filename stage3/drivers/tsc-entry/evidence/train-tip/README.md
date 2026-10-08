Rebuilt the tsc entry with train 1e81b051 in native split modes 0 and 1.
Both optional compiler merges conflicted and were aborted; exact pins and conflict files are recorded.
First stop: builder.ts:1246:69 TS2345, Path | undefined passed to string; sixteen total stops agree between split modes.
Local Node witnesses and evidence checks are observations; the train rehearsal remains pending.
No native tsc binary, native --tiny result, or train-gate pass is claimed.

The topic branch is codex/stage3-tsc-entry-train-tip, based directly on
origin/cloud/land-train-4-views-v3-787cea7a-6c28ada8 at
1e81b051c52eee40d47c44f4a5f68b821e606708. The compiler stays at that pin.

Optional compositions

project-references-source at 88fe8de47cba6928539f1b35d4fd84c1ad06f34d conflicts in
internal/load/load.go, internal/load/source_fs.go and internal/oracle/counts.md.
stricter-options-next at 8f32e51e8fc41b8f1177453213ca5453ce764486 conflicts in twelve
files: see provenance.json and logs/merge-stricter-options.txt. Both attempts were
aborted without resolving conflicts or retaining partial changes. The second
attempt was independent, against the unchanged train. Neither branch is included.
Its source-reference and checked-options behavior therefore remains untested here.

The train first rejects all builds in checking. Split C emission and linking are
not reached. Agreement of failed split attempts is not proof of either backend.
The sixteen-stop sequence consists of the pristine first stop and fifteen more
conditional stops, after body-only throws in an uncommitted disposable tree.
Source and adaptation code are unchanged in this topic.

Stops and owners

Locations below refer back to the original adapted source, with UTF-16
correspondence replayed through fifteen replacements. stops.json also retains
the current scratch positions and complete first diagnostic messages.

C means compiler checked strictness, with codex/stricter-options-next as the
requested integration candidate. That is an owner assessment, not a claim that
the conflicted branch clears a site. S means scratch continuation/inference;
those messages are absent from pristine diagnostics and are not established
production defects. Even a message present in the pristine diagnostic stream
does not prove that the altered program preserves runtime behavior.

| Stop | Original location | Code | Owner | Replaced body |
|---:|---|---|---|---|
| 1 | src/compiler/builder.ts:1246:69 | TS2345 | C | `<anonymous>` |
| 2 | src/compiler/builder.ts:1258:65 | TS2488 | C | getBuildInfo |
| 3 | src/compiler/builder.ts:2273:9 | TS2375 | C | createBuilderProgramUsingIncrementalBuildInfo |
| 4 | src/compiler/builder.ts:395:118 | TS2345 | C | `<anonymous>` |
| 5 | src/compiler/builder.ts:991:17 | TS2345 | C | `<anonymous>` |
| 6 | src/compiler/checker.ts:10022:63 | TS2345 | C | `<anonymous>` |
| 7 | src/compiler/checker.ts:10026:97 | TS2345 | C | `<anonymous>` |
| 8 | src/compiler/checker.ts:10033:67 | TS18048 | C | `<anonymous>` |
| 9 | src/compiler/checker.ts:10034:53 | TS2345 | S | serializeModule |
| 10 | src/compiler/checker.ts:10319:83 | TS2345 | C | `<anonymous>` |
| 11 | src/compiler/checker.ts:10320:51 | TS2345 | C | serializeAsClass |
| 12 | src/compiler/checker.ts:10941:33 | TS2379 | C | serializePropertySymbol |
| 13 | src/compiler/checker.ts:12034:30 | TS18048 | C | getFlowTypeFromCommonJSExport |
| 14 | src/compiler/checker.ts:12859:13 | TS2322 | S | getTypeOfAlias |
| 15 | src/compiler/checker.ts:13986:47 | TS2345 | C | getResolvedMembersOrExportsOfSymbol |
| 16 | src/compiler/checker.ts:14643:66 | TS2345 | C | limit reached |

The first five stops are in builder.ts; the next eleven are in checker.ts.
The two closure files outside src/compiler are separately recorded in
outside-compiler.json: src/tsc/tsc.ts and src/tsc/_namespaces/ts.ts. Neither
produces a stop in this sequence. The independent stock compiler API confirms
81 reached files, and the untouched entry runs on Node as Version 6.0.3.

Witnesses

The existing probes/01 through probes/15 are rerun with this compiler. The
new witnesses/16-symbol-array.a has its in-place TS2345 a-check header. All sixteen
run on Node with exit 0 and empty stderr. Fourteen reproduce the exact first
message. Stops 3 and 12 use smaller present-undefined object witnesses: same
TS2375/TS2379 cause and code, different displayed object shapes. Their exactness
flags are false. Each native witness fails checking; none is a native runtime pass.

Build and source provenance

Normal stage3/apply.sh failed during the large upstream test-baseline checkout.
The retry uses sparse-git.sh only as a scratch git-clone shim. It retains all src,
scripts, package manifests, and the API baseline required by adaptation 40, and
runs the repository apply pipeline unchanged. Its patch table reports 79 files,
5,247 lines added and 5,213 removed, with no baseline deletions. Raw apply and
failed-checkout logs are retained as gzip. The source closure is pinned by hashes.

cloud/setup.sh exited 1 after its last timing, clang ready at 0.425s. Reported
timings were Node 0.048s, Go 0.049s, submodules 0.140s, markdown 0.141s; nproc is 5.
Its go-list diagnostic log is empty. Sourcing /workspace/adamic-tools/env.sh and
building ./cmd/adamic separately in the runner succeeded. This is a targeted
build workaround, not a successful complete setup or gate.

Reproduction

Use a fresh scratch directory, Node 24.19.0 and the environment from setup.
To reproduce the sparse run, install sparse-git.sh as /tmp/tsc-train-tools/git
and prepend that directory to PATH for the existing run.sh. Then:

```sh
bash stage3/drivers/tsc-entry/run.sh /tmp/tsc-train-build-sparse > /tmp/tsc-train-run-sparse.txt 2>&1
mkdir -p /tmp/tsc-train-disposable
cp -a /tmp/tsc-train-build-sparse/adapted/src /tmp/tsc-train-disposable/src
export SCANNER_TYPESCRIPT=/home/agent/.cache/adamic-stage3/api/node_modules/typescript/lib/typescript.js
python3 stage3/drivers/tsc-entry/continue.py /tmp/tsc-train-disposable/src /tmp/tsc-train-build-sparse/adamic /tmp/tsc-train-progress --limit 16 > /tmp/tsc-train-progress.txt 2>&1
python3 stage3/drivers/tsc-entry/evidence/train-tip/verify.py --source-tree /tmp/tsc-train-build-sparse/adapted > /tmp/tsc-train-verify.txt 2>&1
python3 stage3/drivers/tsc-entry/evidence/train-tip/mutants.py > /tmp/tsc-train-mutants.txt 2>&1
```

collect.py packages those scratch streams, original coordinate mappings and
compact replacement receipts. It also expects the sixteen witness streams
under /tmp/tsc-train-witnesses and /tmp/tsc-train-16-*; all of the measured streams
are already retained under logs/. The modified source tree is never committed.

Local checks

The parser fixture retains an outside-compiler diagnostic path and coordinates;
the UTF-16 fixture measures four code units for a plus a supplementary character
plus b. The scalar-count mutant fails it. Seven receipt mutants are rejected:
pass status, dropped stop, split byte, Node byte, owner, dropped outside file and
source hash. The complete verifier reads all 32 failed build streams and all
sixteen Node/native witness pairs. The unchanged Gate.aCheck accepts the sole
new .a witness against its measured TS2345 header. Local artifact consistency and fixture checks
do not promote the pending rehearsal to a pass.

No complete Go gate or train rehearsal gate was run. No native executable was
produced, so native --tiny diagnostics and Node/native runtime comparison could
not run. Results after body replacement are exploratory and conditional.
