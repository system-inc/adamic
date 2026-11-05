"""Render a stopped census without converting missing observations into zeroes."""
import collections
import hashlib
import json
import traceback


def write(out, tree, pair, streams, load, base_commit, head_commit):
    evidence = out / 'evidence'
    assert all(rows[0]['latent_mode'] == 'first-error' for rows in streams.values())
    assert streams['base'][0] == streams['head'][0]
    assert len(streams['head']) == 1, 'unexpected head records: review completeness instead'
    try:
        pair.compare(streams['base'], streams['head'], tree,
                     out / 'scratch/base/stage3/api/node_modules/typescript')
    except ValueError as error:
        pairing_error = str(error)
        (evidence / 'rerun-head-pair-error.log').write_text(traceback.format_exc())
    else:
        raise AssertionError('pair accepted the missing head file records')
    base_sites = {tuple(f[k] for k in ('kind', 'where', 'reason', 'text'))
                  for row in streams['base'][1:] for f in row['findings']}
    counts = collections.Counter(key[0] for key in base_sites)
    kinds = ('NotYet', 'Refused', 'panic', 'error')
    timers = {label: load(label, 'times') for label in streams}
    sums = {label: {key: sum(t[key] for t in rows)
                    for key in ('wall_seconds', 'cpu_seconds')}
            for label, rows in timers.items()}
    source_hashes = {str(p.relative_to(tree)): hashlib.sha256(p.read_bytes()).hexdigest()
                     for p in sorted((tree / 'src').rglob('*')) if p.is_file()}
    assert source_hashes == json.loads((evidence / 'source-manifest.json').read_text())
    stock = json.loads((evidence / 'stock-units.json').read_text())
    expected = sum(not u['deferred_generic'] for row in stock for u in row['units'])
    interruption = json.loads((evidence / 'rerun-head-interruption.json').read_text())
    foreground_run = json.loads((evidence / 'rerun-head-run.json').read_text())
    result = dict(schema_version=1, status='incomplete',
        reason='The foreground head rerun was interrupted before buffered file findings were emitted; its compiler was found defunct and no supervisor exit record survived.',
        base_commit=base_commit, head_commit=head_commit, latent_mode='first-error', head_evidence_prefix='rerun-head', foreground_run=foreground_run, prior_interrupted_attempt=json.loads((evidence / 'interruption.json').read_text()),
        typescript_pin=json.loads((out / 'scratch/base/stage3/source.json').read_text()),
        box=json.loads((evidence / 'box.json').read_text()),
        cgroup=json.loads((evidence / 'cgroup.json').read_text()),
        count_definition='unique (kind, where, reason, text), deduplicated across attempts',
        counts={'base': {kind: counts[kind] for kind in kinds}, 'head': None},
        any_member_findings=None, any_member_unit_first_stops=None,
        regressions=None, new_panics=None, unit_first_stop_delta=None,
        pairing_succeeded=False, pairing_error=pairing_error,
        coverage={label: {'file_records': len(rows)-1,
                          'completed_timers': len(timers[label]),
                          'expected_file_records': len(stock),
                          'expected_timed_attempts': expected,
                          'complete': label == 'base'} for label, rows in streams.items()},
        unit_time_sums=sums,
        unit_time_scope={'base': 'complete', 'head': 'partial; completed timers only; active/interrupted unit excluded'},
        interruption=interruption)
    (out / 'delta.json').write_text(json.dumps(result, indent=2) + '\n')
    (out / 'pair-report.md').write_text('# Pair not available\n\nMain stage3/census/pair rejected the incomplete head stream: `' + pairing_error + '`. No paired deltas or regression counts are asserted.\n')
    lines = ['# Paired any-members census (tmdhz1d): stopped', '',
        '**Any-member findings and first stops: unknown. This run cannot establish 0.**', '',
        f'The head findings list is unavailable. The foreground rerun completed {len(timers["head"]):,} of {expected:,} timed attempts before interruption. Its compiler was found defunct, and only the measurement header survived. No head member location, owner unit, full reason, or first-stop status can be recovered from unit timers.', '',
        '## Regressions', '',
        '**Unknown:** previously finding-free units or sites that stop on head, new panics, and units whose first lowering stop changed cannot be assessed. No empty regression list or zero-panic claim is made for head.', '',
        f'Main stage3/census/pair was attempted on these saved streams and refused them with `{pairing_error}`. Its traceback is in `evidence/rerun-head-pair-error.log.gz`. `delta.json` uses null for unavailable answers.', '',
        '## Pins and box', '',
        f'Base: `origin/main` at `{base_commit}`. Head: `optimizer/any-members-3` at `{head_commit}`, one commit directly on base.', '',
        'TypeScript v6.0.3, upstream commit `050880ce59e30b356b686bd3144efe24f875ebc8`.', '',
        f'Box: cloud host `{result["box"]["hostname"]}`; {result["box"]["nproc"]} available processors; CPU quota 400000/100000 us (4 cores); cgroup memory limit 17179869184 bytes (16 GiB). Host memory is `{result["box"]["total_memory"]}`. Go 1.27.1 linux/amd64; Node v24.19.0.', '',
        '## Unique finding kinds', '',
        'Meter definition: unique `(kind, where, reason, text)` across attempts. The head header is not a finding-free census.', '',
        '| Kind | Base, complete | Head |', '| --- | ---: | --- |']
    for kind in kinds:
        lines.append(f'| {kind} | {counts[kind]} | unknown |')
    lines += ['', '## Unit lowering time on this box', '',
        'No base was rerun. This table uses the retained complete base and the latest foreground head attempt on the same box. Head is interrupted and partial. These sums are not comparable full-run totals. Timers include unit registration, synchronous dependency lowering and recovery; they exclude checker work, file scans and other shared setup. The active interrupted unit has no completed timer.', '',
        f'| Sum | Base, {len(timers["base"]):,} completed attempts | Foreground head rerun, {len(timers["head"]):,} completed attempts only |',
        '| --- | ---: | ---: |']
    for key in ('wall_seconds', 'cpu_seconds'):
        lines.append(f'| {key} | {sums["base"][key]:.6f} | {sums["head"][key]:.6f} |')
    lines += ['', '## Method and stopping evidence', '',
        'Read `typescript/kind-predicate-census` (`33dca393`) REPORT.md and compare.py, and the meter launcher and README. Built one adapted tree with b0efce6d stage3/apply.sh, its lockfile-pinned API packages and the upstream pin above. Both census tools were separately built from their exact compiler commits with recorded cohere/checker pins, main make_overlay.py, gofmt, and `go build -buildvcs=false -overlay=... ./stage3/census/latent/tool`. The adaptation and census inputs do not differ between the compiler pins. The head also changes interface-downcast test fixtures, outside those inputs.', '',
        'The retained base finished before either head attempt. The foreground rerun reused the verified head binary and the same adapted tree: all 746 source hashes matched the manifest before launch and after interruption. Both used this invocation:', '',
        '```sh', 'LATENT_TIMES=times.jsonl LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 latent-census ADAPTED/src/compiler latent.jsonl', '```', '',
        'No base rerun, full mode, entry mode, backend, replay, or subset run was performed. Base started at 16:43:29 UTC and finished at 18:49:21 UTC (7,552 seconds), exiting 0 with all 79 compiler files and 4,056 timed attempts. Its attempted-unit identities independently match the stock TypeScript top-level catalog.', '',
        f'Foreground head rerun started at {foreground_run["start_utc"]}. Its last completed timer was written at {interruption["last_timer_mtime_utc"]} for `src/compiler/utilities.ts:4407:1`. The interrupted current unit is not identified by that previous-unit timer.', '',
        'The foreground supervisor awaited the compiler and recorded timer progress every 30 seconds; it used no detached or background shell job. No other work was performed while the census ran. The tool wait was then reported as aborted and the environment reannounced as starting. On reconnection, head PID 11440 was defunct (state Z, parent PID 1). No supervisor exit code or end timestamp was saved. The termination cause is not established. Process snapshots, progress logs, timestamps and memory counters are retained under the rerun-head prefix.', '',
        'The findings files stayed at their headers while timers advanced. This revision deliberately buffers every file record in `Latent` until all callers finish, to finalize deferred-generic instantiation witnesses (`stage3/census/latent/lower.go.txt:213`, the Latent wrapper). Base emitted all its records only at completion. This was observable progress, not a live timer stall; no mode was switched. After interruption the head was already dead, so a goroutine dump could not be obtained. The logs contain no crash diagnosis.', '',
        'The first head attempt remains preserved: it had 3,865 completed timers before interruption at 20:45 UTC. The explicitly requested foreground rerun is a separate attempt with 3,153 completed timers. Neither emitted a file finding record, so their timer streams cannot be combined into a paired finding ledger. No additional census was launched after the foreground attempt died. No zero result is inferred from either lost buffer.', '',
        'Both headers agree and report first-error mode with zero checker diagnostics (`checker_rejected=false`). The tool retains its fixed measurement label verbatim in the streams. Ordinary Load/Lower output guards were enabled. Compiler and dependency trees are clean, and all 746 adapted source hashes still match the pre-run manifest.', '',
        '## Evidence and validation', '',
        'The compressed streams preserve every saved record, with the shared absolute adapted prefix normalized to relative paths as in the reference method. Full and partial timer streams, logs, source hashes, stock unit catalog, compiler/dependency pins, binary hashes/build metadata and process snapshots are under `evidence/`. `compare.py` uses main stage3/census/pair and refuses to manufacture a comparison when either stream is incomplete. `validate.py` checks the saved evidence and explicit unknown answers; see `validation.txt`.', '']
    (out / 'REPORT.md').write_text('\n'.join(lines))
    print(json.dumps({'status': 'incomplete', 'any_member_findings': None,
                      'regressions': None, 'coverage': result['coverage'],
                      'pairing_error': pairing_error, 'unit_time_sums': sums}, indent=2))
