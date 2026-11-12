# Scoped-tail measurement evidence

`summary.json` contains all per-pass counts/times, both exact tail findings, the
paired child resource records, and the census unit-clock totals.
`unit-ledger.json.gz` retains all 4,566 unit records with their pass ledgers.
`provenance.json` identifies the parent, compiler source base, pinned dependency,
input hashes, actual runtime overlays, binary, toolchain and resource limits.
`validation.json` retains the final checks' own child exits. The mutant logs and
positive fixture stream retain the main-source guard's observations.

The 81 resolved implementation sources are partitioned by `groups.json`.
Every child loads the same tsc entry and complete source graph and retains the
stock independent refusal scans and source prepasses. `select.py` adds only a
scratch file-selection condition to the first-error unit driver. No production
selection behavior changes. Each body's IR and recorded sites remain the unit's
own product. Generic call-witness metadata and checker caches are per child;
there is no fabricated aggregate relation trailer or global cache size.

At most four census children ran concurrently. Each uses `run.py`, which records
and propagates its child's actual exit (including signals), imposes a deadline
of at most 1,700 seconds, and records child peak RSS. All twelve final children
exited 0 without reaching that deadline. Completion means the measurement
finished; the resulting findings still include compiler stops.

From the repository root with its pinned dependencies available, prepare the
same measurement-only entry tool outside tests:

```sh
latent_repo=$(pwd)
latent_scratch=/path/to/new-scratch
latent_input=/path/to/new-adapted
bash stage3/apply.sh "$latent_input"
python3 stage3/census/latent/make_overlay.py "$latent_repo" "$latent_scratch/overlay"
python3 stage3/meter/entry_overlay.py "$latent_scratch/overlay" "$latent_scratch/entry"
python3 stage3/census/latent/evidence/scoped-tail/select.py "$latent_repo" "$latent_scratch/entry"
go build -buildvcs=false -overlay="$latent_scratch/entry/overlay.json" -o "$latent_scratch/entry-census" ./stage3/census/latent/tool
```

For each group 0 through 5, run once with `latent_tail=0` and once with
`latent_tail=1`; use a fresh destination for every invocation. Allow each child
to finish before reusing its destination. The two variants use the same binary.

```sh
latent_group=0
latent_tail=1
latent_destination="$latent_scratch/run-$latent_group-$latent_tail"
latent_group_files=$(python3 - "$latent_input" "$latent_group" <<'PY'
import json
from pathlib import Path
import sys
manifest = json.loads(Path('stage3/census/latent/evidence/scoped-tail/groups.json').read_text())
group = next(group for group in manifest['groups'] if group['id'] == int(sys.argv[2]))
print('\n'.join(str(Path(sys.argv[1]).resolve() / name) for name in group['files']))
PY
)
python3 stage3/census/latent/evidence/scoped-tail/run.py "$latent_destination" env \
  LATENT_FULL=0 LATENT_SPECULATIVE=0 LATENT_ASSERT_NO_OUTPUT=1 \
  "LATENT_TAIL=$latent_tail" "LATENT_ONLY_FILE=$latent_group_files" \
  "LATENT_TIMES=$latent_destination/units.jsonl" \
  "$latent_scratch/entry-census" "$latent_input/src/tsc/tsc.ts" "$latent_destination/census.jsonl"
```

The source guard can be replayed with `guard.py REPOSITORY OVERLAY`, under the
same bounded `run.py` wrapper. It compiles the real lowering source and the
single overlay guard, excluding production-only tests whose entry APIs the
measurement overlay disables. The case is
`TestLatentScopedTailRecordStopOwnsSnapshotLoom03dg`. The automatic overlay-suite
integration case is `TestLatentScopedTailEntryFindingLoom03dg`.

The full-mode fixture also produces the same `checkRecordGuardedReads` tail
finding. The default first-error fixture stream is retained as `fixture.jsonl`.
The implementation and measured result are described in
[SCOPED_TAIL.md](../../SCOPED_TAIL.md).
