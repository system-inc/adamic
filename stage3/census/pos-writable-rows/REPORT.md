# Pos-writable rows — #aryr7ea

**Stopped before the compiler-file census: the unmodified full-mode audit fails at main's tip.**
No compiler or adaptation changes were made. No compiler file piece was started.
`rows.jsonl.gz` is an empty checkpoint archive, **not a measured count of zero**.
The requested site rows and target/callee counts have not been obtained.

## Stop evidence

The README's overlay and tool build succeeded. The unmodified legacy `audit.py`
passed, including its no-IR, disabled-production-loader, attribution and continuation
checks. The unmodified `full_audit.py` failed at its baseline assertion (line 20):

```text
expected: {'a string as a condition', 'a number as a condition', 'var'}
actual:   {'var'}
AssertionError: {'var'}
```

The same failure occurred twice, including a fresh evidence directory. This does
not establish that the compiler is wrong: the fixture may be stale. It does mean
the supplied full-mode continuation proof does not pass on these exact bytes.
I stopped instead of altering the compiler, adapters, overlay or audit expectations.
The audit had generated its mutant observations, but failed before checking them;
no full-mode mutant validation is claimed.

`full-audit.log.txt`, `legacy-audit.log.txt`, `audit-probe.a.txt` and `preflight.json`
retain the failure, fixture, commands, timestamps, elapsed seconds and binary hash.
The complete original audit streams remain under the scratch path below.
The collector refuses to run when this saved preflight has a failing full audit.

## Provenance and preparation

- Main tip fetched and measured setup built from: `c41391ea3346727231dfa3472d061826e26f6aad`.
- Fresh `git worktree add --detach` under `/workspace/scratch/aryr7ea/tree`; only
  after creation, attached to the new output branch `typescript/pos-writable-rows`.
  All branch changes are confined to `stage3/census/pos-writable-rows/`.
- `git submodule update --init --recursive --depth 1` completed. Cohere:
  `5360753e19e4417675a90e92a1823bea66b13918`; nested TypeScript:
  `d92d9bfee114c80be2c375d72edae966176e3a4f`.
- Upstream TypeScript v6.0.3, `050880ce59e30b356b686bd3144efe24f875ebc8`,
  cloned under scratch and named by `ADAMIC_TYPESCRIPT_SOURCE`.
- `npm ci` in this worktree's `stage3/api`; `stage3/apply.sh
  /workspace/scratch/aryr7ea/adapted` completed successfully.
- `make_overlay.py`, `gofmt` and `go build -buildvcs=false -overlay=...` followed
  `stage3/census/latent/README.md`. Tool: `/workspace/scratch/aryr7ea/latent-census`.
- `TMPDIR`, `GOTMPDIR`, `GOCACHE`, `GOMODCACHE`, npm and stage3 caches were all
  under `/workspace/scratch/aryr7ea`. Go 1.27.1; Node v24.19.0.
- Setup/build took place approximately 16:24–16:40 UTC on 2026-11-16 (observed
  interval, not individually timed phases). The timed audit confirmations started
  16:42:04 UTC: legacy 0.355 seconds, full 0.299 seconds, exit statuses 0 and 1.

Build and adaptation logs are `/workspace/scratch/aryr7ea/{apply,overlay,build}.log`.
Audit streams are in `full-audit/` and `full-audit-confirm/` under that scratch root.
`manifest.json` inventories all **79** adapted compiler `.ts` files and their SHA256s.
No source or compiler changes outside the output directory appear in the branch diff.

## Meter comparison and family definition

The saved daily meter `20261115T131735Z.PG4H0R`, main
`7816a8d85d60db6ffb618132eaf8b17c91ef8866`, reports **550** pos-writable sites
across **173** exact reasons in `main/report.json` → `latent_lowering.per_reason`.
The exact Node-to-TextRange reason alone contributes **1** site; the requested
family is broader than that single source/target pair.

`internal/lower/invariance.go` constructs the reason, and the meter preserves
that text. Selection is `kind == "Refused"` with this literal substring in `reason`:

```text
, whose readonly field pos becomes writable:
```

For example:

```text
a value of type Node seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace
```

The collector retains exact reason/text, adapted-relative `where`, file/line/col,
attempting `unit`, `unit_name`, `unit_kind`, and `multiplicity`. It deduplicates by
`(file, line, col, reason, unit)` and separately counts meter keys
`(kind, where, reason, text)`, which omit the attempting unit. Columns are one-based
UTF-16 columns, as printed by the census. Paths are relative to the adapted root.

There is **no measured total or delta against 550**. An empty archive cannot be
compared numerically with the meter. The pins also differ, and no paired run has
been performed to attribute any future delta.

| Target type | Distinct sites |
| --- | --- |
| TextRange | not measured |
| SourceMapRange \| undefined | not measured |
| TextRange \| undefined | not measured |
| Mutable&lt;Node&gt; | not measured |
| others | not measured |

No callee or receiving-slot tally was obtained. The findings schema does not name
those slots for this refusal; attempting-unit names must not be presented as callees.

## Coverage and limits

Planned execution: one exact absolute `LATENT_ONLY_FILE` per process over the
complete adapted `src/compiler` directory, up to three processes at a time, with
`LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 CENSUS_MODULE_FORMAT=CommonJSBundle` and
speculative mode unset. The collector has a 6600-second per-process cutoff and
checkpoints each completed piece. **None of these compiler-file runs occurred.**
`pieces.json` is empty. Failed compiler pieces: none; all 79 remain unattempted.
The failed preflight is separate from that file ledger.

**A full-mode per-file census is a floor where whole-program facts differ.**
Isolated state, cross-file record facts and census-wide generic instantiation
witnesses can differ. Failed constructs can hide children. This is a measurement
ledger, not proof of exhaustive blockers, successful compilation or runtime semantics.

The branch was initially pushed with Git. When HTTPS credentials became unavailable,
checkpoints were published through the GitHub API with expected-head checks and
`force=false`; the uploaded tree hashes were checked against the local trees.
