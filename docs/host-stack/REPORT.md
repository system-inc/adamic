Built: embedded pinned Node declarations and their standalone loader on library/host-1.
Commit: host-1 has parent c0fd8945b872e57a341ccee5973e3f4d1b75c34b; the handoff gives its exact tip.
Gate: final go build ./..., go vet ./internal/load, and all 20 loader tests passed.
Mutants: declaration byte corruption failed integrity; existsSync inversion differed from Node only in stdout.
Not covered: later host units, c2, project configuration, macOS, and the whole repository test suite.

Task: #n49nnje

The fetched main is c0fd8945b872e57a341ccee5973e3f4d1b75c34b. The area tip is
b2c7dfabed73494055f2cf05c1acd99b7779cb4c. The excluded c2 tip is
281079757e9a16e94af2bf8d7a9de75404a6d41d. The survey used the requested
three-dot file list, then compared the host files with current main. Main already
contains fs file core and Buffer/SHA-256 runtime work. A three-dot list alone is
therefore not a list of implementations that are absent from current main.

The proposed stack, with each branch based on its predecessor, is:

1. library/host-1: embed the unchanged published @types/node 25.3.3 and
   undici-types 7.18.2 packages, remove the stage3/api installation dependency,
   and test pin integrity, foreign directories, official signatures, and caller
   package isolation. This is the only branch built in this turn.
2. library/host-2: remaining fs file corrections and exact host options, including
   null/undefined write positions, Buffer arguments, validation and error
   observations. Keep main's rm present-undefined ERR_INVALID_ARG_TYPE behavior.
   Adapt shared expression/refusal glue only for these operations. Keep the
   existing Date getTime/valueOf path; omit the area's dateStringMethod dispatch.
3. library/host-3: directories, Dirent/Stats method dispatch, symlinks, realpath,
   and path operations. Carry NodeHostCall IR/backend support, the directory/path
   runtime, and their oracle witnesses. Shared flow/freshness, expression,
   exception and runtime-header hooks must cover only these host operations.
4. library/host-4: process and performance, including exit status, environment,
   cwd/output barriers, OS observations and supported performance methods.
   Carry process IR/backends and narrowly scoped prelude, startup, cleanup and
   program-exit glue. Async/nextTick execution and timers remain excluded.
5. library/host-5: remaining Buffer/crypto corrections, private-layout guards,
   crypto namespace presence and the proven absent-crypto fallback. Main already
   has Buffer encodings and SHA-256, so retain those implementations and carry
   only the host delta and its witnesses. Do not carry Uint16Array or records.
6. library/host-6: explicit WASI refusals, host-specific WASI adaptations and
   native/WASI agreement tests, followed by host status measurement. Carry only
   fixtures/status inputs that the selected tests actually read. Regenerate
   measurements on this stack; do not copy historical logs or acceptance output.

This ordering separates the file-option corrections from directory receiver
layout changes. Directory Stats methods can then use NodeHostCall without making
host-2 depend on a later branch. Buffer/crypto representation guards can reuse
main's byte-array runtime. The later units still need individual compilation,
mutants and timing gates; this plan is not a claim they already pass. No c2 symbol
is required by host-1. Later host pieces will stop at a concrete missing c2 symbol
rather than importing its implementation.

The first branch copies the loader, embedded packages, manifest, integrity test
and loader tests from the area tip. Its only shared compiler change is the Node
filesystem wrapper, bundled root and index in internal/load/load.go. The area's
nodePrelude replacement was adjusted to main's existing string console signature.
LoadDisplayed and the existing source filesystem remain intact. Project-loader,
Set/prelude restructuring and regex changes were not imported. The declarations
include all 135 published files, 56,660 vendored lines, including package notices
and the version-specific declaration files. The three published test tarballs are
retained; unused registry JSON metadata was removed. THIRD_PARTY_NOTICES.md adds
only the two embedded-package notices.

The final branch adds no standalone Adamic programs or oracle registrations.
Consequently counts.md and flow/fresh generated units have no final changes.
An earlier attempt copied two area portability witnesses. Main lacks their
path.join implementation, so they were narrowed to main's existsSync. Their WASI
comparisons passed, but their foreign-directory native build workers failed with
"buildcache: no adamic go.mod above the working directory" after loading/lowering.
Changing internal/buildcache is excluded. Those workers and fixtures were deferred;
the source-tip loader tests still check real foreign working directories and an
unchanged published @types/node 24.0.0 caller package.

During that attempt, review/test-split-flow/generate.py was run rather than merging
generated files. It also picked up 21 getter fixtures already on main. All of that
generated output was reverted when the new standalone fixtures were deferred.
The required TestCountsAreRecorded -update-counts command was attempted with
-timeout 90s and hit the package timeout. It did not publish a table. Its longest
completed individual test was 30.57s; the timeout listed active child tests at
0s, 0s, 1s, 7s and 12s. No count unit was rerun. This exploratory command was red,
and is not claimed as a passing full-oracle gate. There is no count-changing
runtime/compiler edit or fixture addition in the final loader unit.

Final verification, after restoring the mutants, used:

    export ADAMIC_GOCACHE_OFF=1 GOCACHEPROG=''
    source /workspace/adamic-tools/env.sh
    timeout 180 go build ./...
    timeout 120 go vet ./internal/load
    timeout 150 go test -json -count=1 -timeout 90s ./internal/load
    timeout 30 gofmt -l internal/load
    timeout 30 git diff --check

All exited 0. Build, vet, formatting and diff checks printed no diagnostics.
All 20 top-level loader tests passed (49 completed test/subtest records). The
longest final test was TestNodeLibraryUsesPinnedDeclarations, 2.92s. The earlier
focused loader run's longest was 3.93s. A separate bounded-batch run also passed
all 20 tests; its longest was 2.00s. No final gate test exceeded 60s.

An uncommitted /tmp/host-first-probe/main.a imports existsSync under an alias and
uses a type-only Stats import. Node, sanitized native with no sanitizer report,
and wasm32-wasi each printed exactly "exists=false\n" and exited 0. cmp confirmed
both compiled outputs matched Node. The standalone compiler was then run from
that foreign directory with "types main.a" and resolved Size to { size: number; }.
Native compilation was invoked from the repository to respect main's buildcache
contract; foreign-directory native builds themselves remain outside this unit.

Mutants actually run, with source restored unconditionally:

- Replace fs.d.ts's final newline with a space. This is valid declaration syntax.
  TestEmbeddedNodeTypesIntegrity exited 1 with
  "embedded byte mismatch: @types/node/fs.d.ts" against the independently pinned
  published archive. This proves the loader's byte oracle, not a Node execution
  comparison: declarations have no runtime behavior. It was caught in both the
  initial and corrected temporary runner.
- In main's existing adamic_fs_file_exists, replace return exists with
  return !exists. Rebuild the compiler so its embedded runtime really changes,
  then compile the same uncommitted probe with --sanitize. Build and execution
  succeeded, exited 0 and produced no sanitizer stderr, but native printed
  "exists=true\n" while Node printed "exists=false\n". Only stdout comparison
  catches this mutation. Host-1 adds no runtime function family; this mutation
  proves the new loader reaches the existing host operation under its real
  declaration identity. An earlier temporary runner selected the deferred
  portability test and ran no tests, so that attempt was not a valid kill;
  the corrected CLI comparison above caught it.

Setup and machine observations:

- nproc: 5; cpu.max: 400000 100000; memory: 17.6 GB.
- Go 1.27.1, native clang 20.1.8, Node v24.19.0, WASI SDK 27.
- Initial timeout 600 bash cloud/setup.sh: Node ready 0.079s, Go 0.105s,
  clang 0.610s, markdown dependencies 1.374s, submodules 299.988s,
  shared cache 307.790s; exited 124 during cache warming.
- Retry timeout 600 bash cloud/setup.sh --wasi-sdk passed: Go 0.046s,
  Node 0.063s, submodules 0.132s, markdown 0.173s, clang 0.401s,
  WASI SDK and link probe 21.352s, shared cache 33.492s,
  go build 539.099s, cache warm 539.273s, done 539.377s.
- The environment file was /workspace/adamic-tools/env.sh. Early build/vet/test
  attempts were stopped or timed out during cold compilation; those stopped
  attempts had not executed tests. Final gates used the local Go cache.

Logs remain in /tmp: host-setup.log, host-setup-wasi.log,
host-final-build.log, host-final-vet.log, host-final-load.jsonl,
host-final-format.log, host-final-diff-check.log, host-gate-summary.json,
host-mutants-final.log, host-mutant-embedded-byte-final.log,
host-mutant-runtime-build-0.log, host-mutant-runtime-build-1.log,
host-mutant-node-comparison.log, host-probe-node.log, host-probe-native.log,
host-probe-wasi.log, host-probe-foreign-types.log, host-counts.jsonl,
host-portable-final.jsonl and host-wasi.jsonl.

No changes to go.mod, go.sum, internal/buildcache, toolchain source, c2, regex,
Object/Array/JSON/Date/Error implementations or parallel-test infrastructure are
included. No old stage3 logs or acceptance evidence are included. No PR or Loom
submission is made. Ahra submits after reviewing the branch and this plan.
