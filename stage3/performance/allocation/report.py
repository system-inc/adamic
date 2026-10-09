"""Render measurement tables without filling unobservable totals with estimates."""
import json
from pathlib import Path


def read(path):
    return json.loads(path.read_text())


def write_report(output):
    output = Path(output)
    lines = ['# TypeScript 6.0.3 allocation observations', '',
             'Exact constructor calls and snapshot live objects are separate from sampled allocation estimates.', '',
             '| Input | Observed peak heap MiB | Peak phase | Replay pre-snapshot heap MiB | Replay post-GC heap MiB | After check post-GC heap MiB | Released post-GC heap MiB |',
             '|---|---:|---|---:|---:|---:|---:|']
    for name in ('compiler', 'mitt'):
        folder = output / name
        observation = read(folder / 'observe/observations.json')
        replay = read(folder / 'peak/observations.json')['boundaries'][0]
        lifetime = read(folder / 'lifetime/observations.json')
        boundaries = {row['name']: row for row in lifetime['boundaries']}
        mib = 1024 * 1024
        lines.append(f"| {name} | {observation['peak']['heapUsed']/mib:.2f} | {observation['peak']['phase']} | {replay['before']['heapUsed']/mib:.2f} | {replay['after']['heapUsed']/mib:.2f} | {boundaries['check']['after']['heapUsed']/mib:.2f} | {boundaries['released']['after']['heapUsed']/mib:.2f} |")
    for name in ('compiler', 'mitt'):
        folder = output / name
        summary = read(folder / 'lifetime/summary.json')
        lines += ['', '## ' + name, '',
                  '| Phase | Constructor | Exact calls | New live at boundary | New live self KiB | Survive check | Dead before boundary | Gone by check | Survive release |',
                  '|---|---|---:|---:|---:|---:|---:|---:|---:|']
        for row in summary['phase_cohorts']:
            if row['constructor'] == 'Other': continue
            allocated = 'N/A' if row['exact_constructor_calls'] is None else str(row['exact_constructor_calls'])
            dead = 'N/A' if row['gone_before_phase_snapshot_count'] is None else str(row['gone_before_phase_snapshot_count'])
            lines.append(f"| {row['phase']} | {row['constructor']} | {allocated} | {row['observed_new_live_count']} | {row['new_live_self_bytes']/1024:.1f} | {row['survives_check_count']} | {dead} | {row['gone_by_check_count']} | {row['survives_release_count']} |")
        lines += ['', 'Whole live graph by requested constructor after checking (shallow self bytes):', '',
                  '| Constructor | Live count | Self KiB | Live after Program release | Self KiB after release |',
                  '|---|---:|---:|---:|---:|']
        for category, row in summary['live_at_boundaries']['check'].items():
            released = summary['released'][category]
            lines.append(f"| {category} | {row['count']} | {row['self_bytes']/1024:.1f} | {released['count']} | {released['self_bytes']/1024:.1f} |")
        sampled = read(folder / 'observe-profile-summary.json')
        stock = read(folder / 'stock-profile-summary.json')
        lines += ['', f"Snapshot-free inspector sampling including collected objects: {sampled['total_estimated_bytes']/1024/1024:.2f} MiB estimated, {sampled['sample_records']} sample records.",
                  f"Unmodified CLI --heap-prof: {stock['total_estimated_bytes']/1024/1024:.2f} MiB estimated in the exit profile, {stock['sample_records']} sample records.", '',
                  '| Allocation stack attribution | Estimated MiB, including collected | Sample records |',
                  '|---|---:|---:|']
        for category, value in sampled['estimated_bytes_by_stack'].items():
            lines.append(f"| {category} | {value/1024/1024:.2f} | {sampled['sample_records_by_stack'].get(category,0)} |")
    lines += ['', 'N/A is not zero. Sampling carries allocation call stacks, not allocated-object constructors or exact total counts. Stack attribution is heuristic and is not constructor-byte accounting.', '',
              'Cohorts are objects newly observed at a post-GC phase boundary. The end is checking complete with the Program still rooted; released is after its compilation frame has unwound. Birth self bytes need not equal end sizes. Backing arrays are V8 storage nodes, separate from JavaScript Array objects; Map storage is not included in Map self bytes.', '',
              'Peak is the highest heapUsed observation at 256-constructor intervals and phase boundaries in a run without snapshots. The peak snapshot is replayed at the same deterministic constructor event and boundary location; replay GC and heap usage can differ. It is not proof of the continuous global maximum.', '']
    (output / 'REPORT.md').write_text('\n'.join(lines))
