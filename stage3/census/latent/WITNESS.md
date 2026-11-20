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

Use a fresh witness output path. The witness pass runs after the directory
census fixes its reach flags. Bodyless signatures have no forecast row.
The witness pass uses the checker mapper even for a binder absent from the
parameter and result types.

## Fixture tests

Run these commands from the repository root. Build the baseline binary at
`/tmp/latent-census-baseline` with the same overlay build steps from a
pre-witness checkout, omitting `--witness-tests`. One such checkout is `529b6fdb427a7fa7ccd9bdf58d4347e72e49f16c`.

```sh
source /workspace/adamic-tools/env.sh
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
parked on Loom's side (#hcxsh1k, chg_9ezc1c6q).
