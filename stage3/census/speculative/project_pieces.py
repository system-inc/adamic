"""Publish transparent byte-cost extrapolations and bounded-run budget ceilings.
Usage: project_pieces.py PLANS SAMPLE_RUN TS_RUN WATCH_UNION TS_UNION OUTPUT
A projection is an inference, not a completion claim or statistical interval.
"""
import heapq
import hashlib
import json
from pathlib import Path
import statistics
import sys
plans, sample, ts_run, watch_union, ts_union, output = (Path(p).resolve() for p in sys.argv[1:7])
for union in (watch_union,ts_union):
    assert json.loads((union/'UNION.json').read_text())['exact_match'], 'union parity failed; do not project'
workload = json.loads((plans/'WORKLOAD.json').read_text())
checker = json.loads((plans/'checker.ts.plan.json').read_text())
expected = sorted(checker['pieces'],key=lambda p:(-p['bytes'],p['id']))[:6]
measurements = []
for piece in expected:
    metrics = json.loads((sample/f"piece-{piece['id']:03d}.metrics.json").read_text())
    if metrics['exit'] == 0:
        record = sample/'records'/f"piece-{piece['id']:03d}.jsonl"
        assert record.exists(), 'successful sample record dropped'
        assert hashlib.sha256(record.read_bytes()).hexdigest() == record.with_suffix('.sha256').read_text().strip()
    measurements.append(dict(piece=piece['id'], bytes=piece['bytes'], atoms=piece['atoms'], **metrics))
all_pieces = [dict(file=f['file'], **p) for f in workload for p in json.loads((plans/f['plan']).read_text())['pieces']]
transformer_floor = min(json.loads(p.read_text())['wall_seconds'] for p in ts_run.glob('*.metrics.json'))
prepasses = [m['phase_first_wall_seconds']['walk'] for m in measurements if 'walk' in m.get('phase_first_wall_seconds',{})]
startup = statistics.median(prepasses) if prepasses else None
completed = [m for m in measurements if m['exit'] == 0]
rates = [(m['wall_seconds']-m['phase_first_wall_seconds']['walk'])/m['bytes'] for m in completed]
assert not rates or min(rates) > 0

def schedule(durations, cores):
    queue = [(0.0,c) for c in range(cores)]
    heapq.heapify(queue)
    for duration in sorted(durations,reverse=True):
        finish,core = heapq.heappop(queue)
        heapq.heappush(queue,(finish+duration,core))
    return max(finish for finish,core in queue)

projections = {}
if rates:
    for label,rate,overhead in [('fast_sample',min(rates),min(prepasses)),('median_sample',statistics.median(rates),startup),('slow_sample',max(rates),max(prepasses))]:
        # Completed samples retain their times; failed samples retain only attempt-time floors.
        observed = {m['piece']:m['wall_seconds'] for m in measurements}
        times = [observed[p['id']] if p['file'] == 'checker.ts' and p['id'] in observed else overhead+rate*p['bytes'] for p in all_pieces]
        projections[label] = dict(seconds_per_byte=rate, fixed_overhead_seconds=overhead, core_hours=sum(times)/3600,
                                 largest_piece_seconds=max(times),
                                 wall_seconds={str(c):schedule(times,c) for c in (14,32,64)})
limit = measurements[0]['time_limit_seconds']
assert all(m['time_limit_seconds'] == limit for m in measurements)
result = dict(definition='Single-CPU wall time is the core-cost proxy; fixed project overhead plus byte cost extrapolated from the six largest checker pieces. Sample extrema are scenarios, not confidence bounds.',
              assumptions=['Same per-core throughput and sufficient RAM on the larger box.',
                           'Repeated loader, refusal scan and registration are included in every piece.',
                           'Measured checker load/scan/registration overhead is charged to every file. This is deliberately cautious about repeated prepasses but is not a proven upper bound.',
                           'Unmeasured bodies may be slower, fail, or overflow the stack; byte size alone is not a cost proof.',
                           'A censored sample contributes its observed attempt time only; completion requires unknown extra time. Scenarios then have an unresolved completion tail, not a finite finish forecast.',
                           'Top-level containers outside createTypeChecker are indivisible; more cores cannot shorten that critical path.',
                           'Coverage remains the previously published 65 files; these six samples do not finish checker.'],
              fixed_overhead_seconds=startup, transformer_empirical_floor_seconds=transformer_floor, measured=measurements,
              prepass_censored=not prepasses,
              measured_sample_cpu_core_hours=sum(m.get('user_cpu_seconds',0)+m.get('system_cpu_seconds',0) for m in measurements)/3600,
              completion_projection_censored=len(completed) != len(measurements),
              completed_samples=len(completed), censored_or_failed_samples=[m['piece'] for m in measurements if m['exit'] != 0],
              remaining_files=len(workload), remaining_source_bytes=sum(f['source_bytes'] for f in workload),
              total_pieces=len(all_pieces), total_atom_bytes=sum(p['bytes'] for p in all_pieces),
              workload=workload, projections=projections,
              measured_sample_wall_core_hours=sum(m['wall_seconds'] for m in measurements)/3600,
              bounded_attempt_budget=dict(per_piece_seconds=limit, core_hours=len(all_pieces)*limit/3600,
                                         wall_seconds={str(c):schedule([limit]*len(all_pieces),c) for c in (14,32,64)},
                                         note='Nominal time-limit budget for attempts, not completed coverage; KILL grace, supervisor and publication overhead excluded.'))
output.mkdir(parents=True,exist_ok=True)
(output/'PROJECTION.json').write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps({k:v for k,v in result.items() if k not in ('measured','workload')},indent=2))
