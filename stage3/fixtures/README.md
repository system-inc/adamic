# Stage 3 fixtures

Each bucket owns its .a programs and status.json. The schema is the shared Stage
3 fixture contract: file, tsc source spans, census reason, recorded Node stdout,
stderr and exit, and stage0 outcome plus full diagnostic text. Diagnostics use
repository-relative paths, omit the CLI's adamic prefix and final newline, and
preserve diagnostic content. TypeScript attribution is in NOTICE.

This unit seeds runner/ with NotYet, Refused, and matching Compiles observations.
Run the parallel fixture gate from the repository root:

```sh
source /workspace/adamic-tools/env.sh
go test ./stage3/fixtures -count=1 -timeout 10m -v > /tmp/stage3-fixtures.log 2>&1
```

Every `*/status.json` is read, with buckets and fixtures run in parallel. Each
fixture has named `node`, `stage0`, and, when it currently compiles, `native`
checks. Erasable fixtures run afresh from source through the current oracle/node.mjs;
stdout, stderr, and exit must equal the recorded observation. stage0 checks the
outcome and complete diagnostic. A change says `gap changed: update status.json
and check the native output`. Newly compiling gaps run natively even when their
record still says NotYet or Refused, so a gap closure cannot hide wrong output.

Native checks build the oracle test binary once and invoke its small exported
`TestStage3FixtureHook`. The hook reuses `lowered`, `nativelyUncached` and
`leaksUncached`: the same ASan/UBSan flags, bounded process execution and platform
leak checks as internal/oracle. No helper implementation was copied and no
oracle_test.go or production compiler file was changed. Observations are uncached.
The hook transports bytes as base64 JSON and is dormant in ordinary oracle runs.
Node and native execute with the same repository-root working directory.

```sh
go test ./stage3/fixtures -count=1 -timeout 10m -args -update > /tmp/stage3-fixtures-update.log 2>&1
```

`-update` refreshes only currently compiling fixtures whose native output equals
current Node and whose recorded Node observation is still correct. It replaces
only the stage0 JSON value; all bytes outside that value, including node,
provenance, extra fields, ordering and whitespace, remain unchanged. Noncompiling
fixtures cannot be refreshed automatically because they have no matching native
output to prove the change. Miscompiles and stale Node expectations fail and are
never updated. Each changed status file is replaced atomically after its parallel
fixtures finish, and a concurrent external change aborts the write.

For audits, `-args -fixtures /tmp/copied-fixtures` selects a scratch fixture root.
The test still uses the repository's compiler, runtime and Node runner. The
three requested mutants and additional diagnostic/update guards were run on
scratch copies; see ../meter/REPORT.md. The native mutant uses a Go overlay of
runtime/string_build_impl.h, changing only "true" to "truf".

The enums and namespaces buckets select Node's independent transform mode,
`stripTypeScriptTypes(source, { mode: 'transform' })`, as on codex/flag-enums and
codex/namespaces-tsc. The harness derives a temporary transformed runner from the
current source oracle, preserves its runtime URL and import hooks, and leaves
oracle/node.mjs unchanged. Ordinary erasable fixtures continue to run that
original runner. This transformation is Node's implementation, not Adamic output.

`file` may name a slash-separated relative path such as `01_call_time/main.a`.
Absolute paths, `.`, `..`, empty components, backslashes and drive prefixes are
rejected before joining the path with its bucket. Cycles therefore keep their
entry modules and support files together.

A status entry may declare `"platform": "linux"` (using Go's GOOS names). An
omitted platform runs everywhere. A different platform skips the fixture before
Node, checker or native execution, with the declared and current platforms in
the printed reason. Linux is the gate of record. Only host/09_realpath.a,
host/14_getCurrentDirectory.a and host/24_useCaseSensitiveFileNames.a carry this
field because their recorded filesystem observations are Linux-specific.
Those three metadata additions are the only edits to another worker's bucket;
recorded Node and stage0 values are unchanged.


## Landing gate on Linux

Base: origin/land/stage3 at d9fc3df3f734ae50ab803c7a210f401f192deb21.
The supplied tree has 11 status buckets, including runner, and 173 fixtures.
All entries were matched against final Go JSON test events; every fixture passed,
including 13 enums, 12 namespaces, 10 directory entries and the three Linux-only
host entries. Six fixtures also ran natively through ASan/UBSan/leak checks.

| Bucket | Pass | Fail | Skip |
| --- | ---: | ---: | ---: |
| assertions | 20 | 0 | 0 |
| cycles | 10 | 0 | 0 |
| enums | 13 | 0 | 0 |
| host | 25 | 0 | 0 |
| namespaces | 12 | 0 | 0 |
| nested-functions | 11 | 0 | 0 |
| objects | 25 | 0 | 0 |
| predicates | 11 | 0 | 0 |
| records | 19 | 0 | 0 |
| runner | 3 | 0 | 0 |
| taste | 24 | 0 | 0 |
| **Total** | **173** | **0** | **0** |

Command: `go test ./stage3/fixtures -run 'TestFixtures$|TestFixturePaths$'
-count=1 -timeout 30m -json > /tmp/stage3-landing-fixtures.jsonl
2> /tmp/stage3-landing-fixtures.stderr`. Exit 0, package time 41.717s.
The twelve path cases and `go vet ./stage3/fixtures` pass. A foreign-platform
scratch fixture prints `fixture records platform darwin; current platform is
linux` and skips before attempting to run its deliberately missing source.
macOS execution was not performed; Linux is the gate of record.

Three scratch Go-overlay mutants independently prove the new checks: disabling
transform selection fails the enums' recorded Node byte comparisons; ignoring
path components makes TestFixturePaths reject accepted traversal/absolute paths;
replacing the platform skip with a log executes the missing foreign fixture and
fails. Normal skip probe exit 0, each mutant exit 1. Audit summary is
/tmp/stage3-landing-audit.log. Recorded Node and stage0 fields were never refreshed
or weakened. Counts are also saved in
../meter/runs/20261007T025701Z.runner-landing/fixtures.json.
