# Parallel files scout

Status: prepared source, blocked at stage 0. This branch is not green for native
integration. No ownership check was weakened and no native/compiler source changed.

Branch `lint-scout/parallel-files` starts at lint
`9156bf5c579a44d687c9955d13e44f9ad8bbb6f8` and merges runtime
`cdfa22555589194e6f3133d986dd110061aafaa9`, without rebasing. Merge commit
`bdb6292c5e714b971a2fcbbd9ba8f8771debec64` retains both checker-replay and
parallelMap exports in the Node oracle shim.

The manifest creates file jobs, runs `parallelMap`, and prints returned buffers
in input order after joining. Each job opens and releases its own checker program.
Fixes still run inside that job with the existing ten-pass reparse budget.
Shards retain the original case numbers. The single-file path uses the same renderer.

## Observations

`TestParallelFilesSourceControl` passed against the original serial driver source:
154 cases, 1,074 findings, 154 fixed outputs, 529,109 bytes in the final run.
Byte lengths and SHA256 values include temporary corpus paths, so different test
invocations can have different lengths and hashes.
The reversed-result source mutant is caught by exact byte comparison. This is a
Node source observation, not proof of native result publication.

`TestParallelFilesRegexControl` passed against Go cohere: all three owned shouting
witnesses, 1,816 bytes in the final run. Local regex factories preserve the translated literals.

The agreement test uses this same 154-row corpus at N=1, 2 and NumCPU (5 here).
All three native observations remain **unmeasured**: compilation refuses before
any N runs. `TestParallelFilesThreadSanitizer` stops at that same refusal, before
clang or TSan executes. Neither test skips or treats refusal as green.

Shortest reproducer for @system_adamic_runtime, also in `parallel_files_refusal.a`:

```typescript
import { parallelMap } from 'adamic';
if(true) {
    const jobs: { readonly row: string }[] = [{ row: 'file' }];
    console.log(parallelMap(jobs, (job) => job.row).join(','));
}
```

Exact stage-0 reason:

```text
Adamic 0.1 refuses cannot move jobs: the owner must be a local declared in the call's block; return it through the results
```

Command: `go run ./cmd/adamic c /tmp/parallel-files-refusal.a`.
It prints the reason at line 4, column 29 and exits 1. The real driver refuses
at `main.ts:239:33`. Further sharing diagnostics may remain hidden behind this
first refusal; this scout does not claim they are resolved.

## State audit

The static import walk reached 288 source modules. Per-file objects and mutable
storage found in the driver and its walk:

- Row fields, path/source, options, replay text/error and row-header hash.
- Parser scanner/path/modes, nodes, roots, recoveredJsx, diagnostics and listContexts.
- Both scanner instances: text, pos, start, fullStart, kind, value, flags and errors.
- Settings text and values Map, including option arrays.
- Linter parser/scanner/source/options/checker, findings, skipped, rejected,
  unconverged, junkRows, parents and root.
- RuleContext findings/skipped, parent table, options and checker; lazy lineStarts
  and literalEndsByStart caches. Every rule set and rule instance is constructed
  for that context, with its context pointer confined to the file.
- Fix-engine current source, proposals, accepted/disjoint edits, pieces, last
  proposals and pass counter, plus fresh parser/scanner/linter instances per pass.
- Findings, extra fixes, suggestions and edit arrays.
- Rendering output array, byte offsets, line starts, byte count, range-search and
  repair/suggestion temporaries; final immutable FileResult count/output.
- Checker program handle, parser/path, offsets, refusals, recorded frames, replay
  Frames text/cursor, record flag, root, current rule and declared reads.
- Rule-local helper arenas, node/signature tables, maps, sets, scopes, comment
  lists and traversal stacks; temporary regex instances and match cursors.

Rule instance fields beyond their context pointers, found by declaration scan:

| Rule | Fields |
|---|---|
| max-classes-per-file | maximum, ignoreExpressions, count |
| max-depth | maximum, disabled |
| max-nested-callbacks | maximum, checkConstructorCallCallbacks, depths |
| next-no-sync-scripts | tagName, attributes |
| nexus-consistency-no-abbreviated-identifier | bindings |
| nexus-consistency-no-stuttering-name | names |
| no-dupe-else-if | arena |
| no-empty-pattern | allowObjectPatternsAsParameters |
| no-prototype-builtins | nodes |
| no-redundant-type-constituents | shared, port |
| no-unused-labels | labels, used |
| no-warning-comments | anchors |
| operator-assignment | comments |
| prefer-template | reported |
| react-jsx-curly-brace-presence | props, children, elements |
| react-require-optimization | nodes, reports |
| typescript-eslint-consistent-type-definitions | style |
| typescript-init-declarations | mode, ignoreForLoopInit |
| typescript-no-wrapper-object-types | declaredTypes |

Module storage found:

- Scanner `keywords` and `punctuators` Maps, Unicode `folds` Map: initialized at
  module load, only read during walks. Their mutable declared types could still
  trigger a sharing refusal after the first blocker is fixed; no assertion of a
  compiler sharing proof is made here.
- `noEdits`, `noSuggestions`, message strings, Unicode/policy/kind tables and
  pass-budget constants: no writes during a walk. `junkRows` is an immutable flag.
- Shouting patterns formerly shared RegExp lastIndex state. All fifteen are now
  factories: digitOrUnderscore, vowel, uppercaseToken, fencedBlock,
  inlineBackticks, doubleQuoted, singleQuotedToken, singleQuotedCapitalPhrase,
  notNewline, jsDocGutter, blockCommentGutter, exampleTag, anyJsDocTag,
  commandLine, structureCommandStarter. Literal text and flags are unchanged;
  each matcher belongs to its caller. Pagination regexes were already local.
- The external checker bridge's `programs.next` and `programs.live` handle map
  are guarded by `programs.Mutex`, including create/query/inspect/release. Its
  optional process-wide profile file, before/inspect/parts counters are accessed
  under that same mutex. Per-program Compiler/typeIDs/typesByID/exactRanges now
  belong to a single file job. The bridge still serializes checker operations.
- Node oracle `utf8Text`/`utf8Encoded` caches are used by its sequential witness;
  they are not native walk storage. Go test build caches are harness storage.

## Scout measurements and limits

Native single-core instructions versus serial and native default-worker wall
measurements are unavailable: the parallel driver has no emitted C or binary.
No speedup, instruction reduction or TSan-clean claim is made. Neither perf nor
valgrind was installed on this box. These are blockers, not zero measurements.

Target benchmark settings: same manifest and selected rules for both drivers;
ADAMIC_THREADS=1 for instruction comparison, unset for default wall. The host
reports nproc=5 and cgroup cpu.max=400000 100000 (four-core quota). One-minute
load at the blocked test inspection was 1.55; after the final source control it
was 0.85; after Go witness compilation it was 3.65. These loads accompany
validation observations, not performance figures. Native test build flags use
native.Options{}; the prepared race lane uses ThreadSanitize:true,
TSAN_OPTIONS=halt_on_error=1:history_size=4:report_atomic_races=1,
ADAMIC_TSAN_PERTURB=1 and ADAMIC_THREADS=5. The Go archive itself is not TSan
instrumented by this lane; any future bridge race validation needs its own check.

Setup with GOPROXY=https://proxy.golang.org|direct completed in 188.948 seconds:
Go ready 0.042s, clang ready 0.261s, dependencies ready 0.992s, submodules ready
4.252s, go build ready 188.727s. Go 1.27.1, clang 20.1.8, Node 24.19.0.

Commands (all test output went to files):

```text
go run ./cmd/lint-registry
go run ./cmd/adamic c stage1/cohere/lint/main.ts
  exit 1, ownership refusal before C emission
go test ./stage1/cohere/lint -run '^TestParallelFiles' -count=1 -timeout 15m -v
  exit 1: native agreement and TSan compilation refused; source and Go controls pass
  final package run 6.351s, source 1.57s, Go regex control 3.05s
go test ./stage1/cohere/lint -run '^TestParallelFiles(RegexControl|SourceControl)$' -count=1 -timeout 15m -v
  exit 0: source control 1.65s, Go regex control 48.70s, package 50.380s
git diff --check
  exit 0
```

Full gate, live typed-project agreement, emitted-JavaScript comparison, native
ordering mutants and native race mutants were not run. Native refusal is reported
rather than edited around. Push this scout once; integration must not merge it
until the runtime owner resolves the blocker and all native checks run green.
