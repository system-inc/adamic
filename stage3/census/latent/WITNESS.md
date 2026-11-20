# Deferred generic witnesses

Witness mode forecasts what lowering encounters in a deferred generic body at
a real call site. It is intended for bodies the census never entered.

1. Enable it with `LATENT_WITNESS=1`. The default census is unchanged.
2. The separate JSONL stream defaults to the census output path plus
   `.witness.jsonl`. Set `LATENT_WITNESS_OUT` to override it. Each row identifies
   the deferred declaration and reports the selected call site and its
   checker-printed type arguments in binder order. Witness records stay out
   of the census findings, units, timings and meter counts.
3. Use the first concrete call in program order and count calls and skipped
   non-concrete calls. A call belongs to its nearest enclosing census unit. A row
   without a concrete call can report `no-concrete-witness`. Constraint-based witnesses
   are permitted only for declarations called exclusively from outside the
   census and require the
   `synthetic` label. This implementation emits `no-concrete-witness` for
   these declarations.
4. A clean lowering is a `forecast` with no stop. It contributes no coverage.
5. Lower every witness on a snapshot of lowering state.

Build the guarded overlay as described in the [census README](README.md), then run:

```sh
LATENT_WITNESS=1 LATENT_WITNESS_OUT=/tmp/witness.jsonl \
  /tmp/latent-census /path/to/sources /tmp/latent.jsonl
```

Use a fresh witness output path. The directory and meter entry drivers accept
the same witness settings. The witness pass runs after the census fixes its
reach flags. Bodyless signatures have no forecast row.
The witness pass uses the checker mapper even for a binder absent from the
parameter and result types.

## Fixture tests

Run these commands from the repository root. Build the baseline binary at
`/tmp/latent-census-baseline` with the same overlay build steps from a
pre-witness checkout, omitting `--witness-tests`. One such checkout is `aab8592a885c9cbc6cd15a13d747582e8f8910c8`.

```sh
source /opt/adamic-block/environment.sh  # on a Loom box; elsewhere, any shell with Adamic's Go, clang and Node on PATH
python3 stage3/census/latent/make_overlay.py "$PWD" /tmp/latent-witness-overlay --witness-tests
gofmt -w /tmp/latent-witness-overlay/*.go
go build -buildvcs=false -overlay=/tmp/latent-witness-overlay/overlay.json \
  -o /tmp/latent-census ./stage3/census/latent/tool

witness_suite=stage3/census/latent/witness
python3 "${witness_suite}_test.py" /tmp/latent-census /tmp/latent-census-baseline
go test -overlay=/tmp/latent-witness-overlay/overlay.json ./internal/lower \
  -run '^TestLatentWitnessSnapshotPreservesLaterCallResult$' -count=1
```

The Python suite uses the prebuilt binaries. The Go test uses the generated
overlay and calls lowering in-process. Go tests may not run `go build`, so
these tests currently run by hand.
They join the gate with the census overlay's built-ahead product, which is
in Loom's line (#hcxsh1k, chg_sfjxpratk7rnq176x0a6stjqa5).

The tail-on fixture checks that scoped passes executed and compares the census
bytes after removing their `pass_ledger[].wall_seconds` readings. The ordinary
fixture compares the raw census bytes.

### Meter entry fixture

Build the meter entry binary from the same scratch census overlay, then run
the fixture with the prebuilt entry binary. For a byte comparison with main,
build the entry baseline from the pre-witness checkout in the same way and set
`ENTRY_CENSUS_BASELINE_BINARY`.

```sh
python3 stage3/meter/entry_overlay.py /tmp/latent-witness-overlay /tmp/latent-entry-overlay
gofmt -w /tmp/latent-entry-overlay/*.go
go build -buildvcs=false -overlay=/tmp/latent-entry-overlay/overlay.json \
  -o /tmp/latent-entry-census ./stage3/census/latent/tool
entry_suite=stage3/meter/entry
ENTRY_CENSUS_BINARY=/tmp/latent-entry-census python3 "${entry_suite}_test.py" EntryWitnessTests
```
