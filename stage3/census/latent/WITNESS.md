# Deferred generic witnesses

Witness mode forecasts what lowering encounters in a deferred generic body at
a real call site. It is intended for bodies the census never entered.

1. Enable it with `LATENT_WITNESS=1`. The default census is unchanged.
2. Write the separate JSONL stream with `LATENT_WITNESS_OUT`. Each row identifies
   the deferred declaration and reports the selected call site and its
   checker-printed type arguments in binder order. Witness records stay out
   of the census findings, units, timings and meter counts.
3. Use the first concrete call in program order and count calls and skipped
   non-concrete calls. A row without a concrete call can report
   `no-concrete-witness`. Constraint-based witnesses are permitted only for
   declarations called exclusively from outside the census and require the
   `synthetic` label.
4. A clean lowering is a `forecast` with no stop. It contributes no coverage.
5. Lower every witness on a snapshot of lowering state.

Build the guarded overlay as described in the [census README](README.md), then run:

```sh
LATENT_WITNESS=1 LATENT_WITNESS_OUT=/tmp/witness.jsonl \
  /tmp/latent-census /path/to/sources /tmp/latent.jsonl
```
