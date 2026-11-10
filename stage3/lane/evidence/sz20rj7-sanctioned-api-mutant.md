# Sanctioned API freshness and proof mutant

Task sz20rj7. Base: main bdbc73864f82dbc0c51434decf0ffe96d4782767.
Pinned TypeScript: 050880ce59e30b356b686bd3144efe24f875ebc8.

The canonical Python generator wraps the Node reconstruction, adds the exact
baseline diff, and serializes the committed manifest. A full real-tree generation
into a temporary directory took 9.779 seconds (child exit 0), including the
initial pinned-object reads while setup was indexing Git objects. Its complete
bytes equaled main's committed sanction, including every proof hash. Generation
requires the pinned source Git objects and API dependencies, not an adapted build.

The new parallel Go wrapper invokes the ordinary lane Python runner. It runs
one canonical generation against the real adaptation tree into temporary output
and compares all bytes. Failure lists the first eight differing JSON keys and
the command to regenerate. No Go or compiler subprocess is invoked by this test.
The existing runner preserves the Python child's exit status in a failure.

Command for each focused run:

```sh
go test ./stage3/lane -run '^TestLaneSanctionedApiFreshnessPython$' -count=1 -v
```

Cold wrapper: exit 0, 14.57 seconds, while setup was building.
Warm restored wrapper: exit 0, 9.31 seconds.
The same wrapper passed under the complete lane package load in 7.87 seconds.
A loaded probe pinned the test and one busy-loop competitor to CPU 0 while setup
and product preparation also ran: exit 0, 16.80 seconds. The competitor was
terminated and reaped in a finally block. These are wrapper times from Go's
uncached test events, excluding Go compilation.

The mutant replaced exactly one byte at zero-based offset 2 in
`stage3/adapt/40-explicit-any/class-rules.json`: ASCII space became tab. This
preserves valid JSON and the proof's meaning while changing its hash. The
committed sanction was untouched. A finally block restored the original proof.

Red output, child exit 1, wrapper 6.01 seconds:

```text
AssertionError: sanctioned-api.json differs from full tree generation; first differing keys: $.proofs[1](stage3/adapt/40-explicit-any/class-rules.json).sha256
Regenerate from the repository root:
NODE_PATH="$PWD/stage3/api/node_modules" python3 stage3/lane/generate-api.py "$ADAMIC_TYPESCRIPT_SOURCE"
--- FAIL: TestLaneSanctionedApiFreshnessPython (6.01s)
```

Restored green output, child exit 0:

```text
--- PASS: TestLaneSanctionedApiFreshnessPython (9.31s)
PASS
ok  github.com/system-inc/adamic/stage3/lane 9.339s
```

An earlier warm attempt overlapped the intentional mutant and was discarded;
the warm number above is the separate run after restoration.

The new test reads the sanction, generators, normalizer, source pin, adaptation
proofs and dynamically discovered sanctions. None of the added or edited files
is in the shard runtime list. The manifest check exited 0:

```sh
python3 -B stage3/lane/generate_shards.py --check
```

The runtime hash, shard plan, shard manifest, generated Go declarations and
committed sanction remain unchanged. No manifest regeneration was needed.

## Complete package run

Setup and both pinned dependency installs exited 0. The complete command was:

```sh
go test ./stage3/lane/... -count=1 -timeout=90m -json
```

It exited 1. The lane package took 2469.610 seconds; the adapted-tree subpackage
passed in 961.443 seconds. The new freshness wrapper passed in 7.87 seconds,
the Python suite census passed, and the manifest check passed. Eight existing
shard tests timed out at their unchanged 50-second child limit: file units
081, 082, 094, 095, 096, 279, 280 and 281. All other tests passed; none skipped.
The complete logs are `/tmp/sz20rj7-whole-lane.jsonl` and
`/tmp/sz20rj7-whole-lane.stderr`.

A serial rerun was interrupted by an environment transition after six shards
had passed. That incomplete result is not counted as a successful rerun.

The complete serial retry exited 0 in 47.541 seconds. All eight previously
timed-out shards passed, individually taking 4.98 to 6.69 seconds. Its logs are
`/tmp/sz20rj7-failed-serial-complete.jsonl` and the matching stderr file.
This supports load sensitivity; the original full-run failures remain recorded
above. A checkpoint is published before an additional complete run with lower
parallelism; that later run does not change the implementation.
