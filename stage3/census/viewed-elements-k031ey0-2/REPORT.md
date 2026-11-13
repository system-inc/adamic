# Viewed-elements entry census with probe seam — k031ey0-2

## Answer

**No.** Head does not show `every=true` with a non-empty trip anywhere that main shows `every=false` in this measurement. Both sides enter with `every=false` and leave with `every=true`, and both trips are empty. Thus the new nesting rule supplies no observed differential switch on this tsc program. Main already turns `every` on. This is a measurement of this program and selected census unit, not a general equivalence claim about the two implementations.

## Seam and reachability

The exact identical probe-only edit on both sides is saved in `seam.diff`, `main-seam.diff` and `head-seam.diff` (byte-identical):

```diff
 		base := fresh()
+		base.elementKey = &viewElementKey{}
```

This matches production's `key := &viewElementKey{}` construction in `internal/lower/lower.go`; no additional field initializers are required. The seam is only in the scratch measurement template. The evidence commit contains the diff as data; it does not change that template or production code.

**The seam was reached on both sides:** each complete stderr contains exactly one viewed-elements line. Neither side is missing a line. Both census processes exited **0** without timeout. Both loaded **81** source files through the same adapted `src/tsc/tsc.ts`, with **zero checker diagnostics**, and attempted the single `KindExportDeclaration` unit of the same 101-byte file `src/compiler/_namespaces/ts.performance.ts`, with no findings. The selected file limits units; the collector receives the entry's whole-program module order.

## Counts by class

| every-on-entry | every | trip non-empty | main | head |
|---|---|---|---:|---:|
| false | false | false | 0 | 0 |
| false | false | true | 0 | 0 |
| false | true | false | 1 | 1 |
| false | true | true | 0 | 0 |
| true | false | false | 0 | 0 |
| true | false | true | 0 | 0 |
| true | true | false | 0 | 0 |
| true | true | true | 0 | 0 |

Totals: **main 1; head 1**. `count.py` collects every line beginning `viewed-elements:` from the complete stderr files, validates its shape and build label, and counts the requested classes; `counts.json` contains the results and all matching lines.

## Verbatim lines and pairing

There are **zero head lines with non-empty trip**, so there are none of that class to quote. For completeness, the only head line and its main counterpart are quoted word for word below, paired by the single selected file's registration lowering:

Head:
```text
viewed-elements: build=head every-on-entry=false every=true viewed=1 trip=""
```

Main:
```text
viewed-elements: build=main every-on-entry=false every=true viewed=180 trip=""
```

The different viewed counts do not imply a different final every flag. Head's collector stops further `hold` traversal once `key.every` becomes true; main does not have that short-circuit. This log does not identify which existing non-nesting cause first set every. The empty head trip shows no recorded nesting-threshold trip in this run.

## Commands and evidence

`PREPARATION.md` records the original pins, reused adapted tree, exact seam application and invocation. `measure.sh` contains the exact executed commands for regenerating each side's latent and entry overlays, separately building each binary from its own compiler sources, and running each one. Full mode was used with `LATENT_ASSERT_NO_OUTPUT=1`, `ADAMIC_VIEWED_ELEMENTS_LOG=1` and the exact same absolute `LATENT_ONLY_FILE` as before. Go package builds use `-p=1`. All builds and runs were sequential, completing in about two minutes; the wrappers print and save each census child's own exit code.

Both full stderr logs, stdout, exit codes, JSONL records, build logs, binary SHA-256 files, seam diffs, counts and exact commands are included. No stderr was piped through head or tail, truncated or reduced. `SHA256SUMS` records all evidence file hashes. The adapted tree was reused unchanged from the first measurement, whose exact preparation commands and logs remain in commit 06b9fe01617af0f43a4a48d86fe755595a516c08.

This branch contains one commit on c07d3b5204082b9d862091c9e53ce241ab65670d, adding only `stage3/census/viewed-elements-k031ey0-2/`. The seam is probe-only and never for landing.
