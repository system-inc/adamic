# latent.jsonl reader audit

Task 8dbaea5. Inventory: `git grep -n latent.jsonl`, excluding evidence and
run folders. The complete inventory is saved in readers.txt.

- `stage3/meter/twice-daily.sh` uses the branch's latent/make_overlay.py and
  compiles the replaced latent driver, so compiler-directory streams already
  contain the relation trailer. There is no stage3/meter/make_overlay.py.
  `entry_overlay.py` replaces that driver with entry.go.txt; it previously
  discarded lower.Latent's returned report. The entry driver now writes and
  syncs the original header before lowering and writes a relabeled trailer
  after normal completion. Blocked first-error entries do not attempt lowering
  and retain their single blocked header.
- `stage3/meter/report.py`: compiler and entry summaries validate a single final
  trailer, exclude it from source coverage, and count its program findings in
  the same unique-site totals and reason tables as file findings. JSON exposes
  completion and both observed endpoints; Markdown explicitly warns when the
  trailer is absent. Root-order accounting also retains program relation stops,
  so they cannot grant whole-program lowering credit.
- `stage3/meter/report_test.py`, `entry_test.py`: fixture writers and readers;
  regression cases cover completed and missing trailers, program NotYet counts,
  invalid trailer placement, and root-order credit. Real entry probes verify
  trailer emission and relabeling. `compiler_test.py` uses synthetic streams
  and checks gzip products, retaining legacy missing-trailer coverage.
- `cloud/meter-daily/submit.sh` does not parse latent JSONL. Its only literal
  reference is an explanatory comment about Git ignores. It copies the entire
  run, compresses logs, and commits allowed artifacts. No schema change needed.
- `stage3/adapt/40-explicit-any/split-previous-133.py`: derives its source-tree
  prefix only from file rows; findings still include the trailer.
- `stage3/adapt/70-readonly-views/survey.cjs`: reads findings from every row after
  the header, then filters exact variance reasons. A relation trailer has a
  findings array, so it is already handled and unrelated reasons are excluded.
- `stage3/census/tsc-closure/test_fixture.py` and rerun-2026-10-09/reproduce.sh
  produce streams and invoke measure.py. The latter's hidden.read_rows now
  uses the latent stream adapter, excluding trailers from file/unit coverage.
  measure.py also counts the adapter's program findings in its all/unattributed
  totals. Historical no-trailer streams remain readable.
- `stage3/census/latent/README.md`, `stage3/census/tsc-closure/README.md`,
  `stage3/meter/README.md`, adapt/70 and adapt/99-cli-perf-hooks-import README
  references document invocations or retained artifacts, without parsing.
- adapt/32-indexed-reads-program/proof/directory-exists/{census-delta,results}.json
  and adapt/99-correctness-axolotl-j/{census-delta,prior-proof-results}.json are
  saved result/path/hash inventories, not executable readers.
- `stage3/meter/.gitignore` is artifact filtering, not a reader.

The latent census's own Python readers already use stream_records.census_records
on this branch. This change does not rework their historical reporting models.
