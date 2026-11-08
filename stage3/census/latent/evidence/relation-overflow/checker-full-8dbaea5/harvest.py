"""Retain one bounded checker-only observation without inventing missing answers."""
from pathlib import Path
import gzip, hashlib, json, shutil
base = Path('/workspace/scratch/8dbaea5')
run = base / 'checker-full'
out = Path('/workspace/adamic/stage3/census/latent/evidence/relation-overflow/checker-full-8dbaea5')
out.mkdir(parents=True, exist_ok=True)
def rows(path):
    result = []
    if path.exists():
        for line in path.read_text().splitlines():
            try: result.append(json.loads(line))
            except json.JSONDecodeError: break
    return result
observations, queries, units = [rows(run / name) for name in ('latent.jsonl', 'relation-times.jsonl', 'unit-times.jsonl')]
trailers = [r for r in observations if r.get('status') == 'relation_report']
assert not trailers or (len(trailers) == 1 and observations[-1] is trailers[0])
trailer = trailers[0] if trailers else None
metrics = json.loads((run / 'metrics.json').read_text())
provenance = json.loads((base / 'measurement-provenance.json').read_text())
files = [r for r in observations if 'file' in r]
assert all(r['file'] == provenance['environment']['LATENT_ONLY_FILE'] for r in files)
assert observations and observations[0]['status'] == 'measurement'
assert observations[0]['roots'] == [provenance['entry']]
assert observations[0]['latent_mode'] == 'full'
assert all(r['file'] == provenance['environment']['LATENT_ONLY_FILE'] for r in units)
assert metrics['command'][1] == provenance['entry']
assert metrics['time_limit_seconds'] == 6000

program_queries = [r for r in queries if r['where'] == 'program']
after_load = trailer['relation_overflow']['after_load'] if trailer else program_queries[0]['answer'] if program_queries else None
end = trailer['relation_overflow']['end'] if trailer else None
if trailer:
    assert len(files) == 1
    assert observations[0]['roots'] == [provenance['entry']]
    assert metrics['exit'] == 0
    assert len(program_queries) == 2
    assert isinstance(after_load, str) and isinstance(end, str)
    assert [r['answer'] for r in program_queries] == [after_load, end]

reason = 'normal completion' if trailer else '100-minute wall bound' if metrics['exit'] == 124 else 'resident-memory bound' if metrics['rss_limit_killed'] else 'process stopped without a completion trailer'
if (run / 'stop-reason.txt').exists(): reason = (run / 'stop-reason.txt').read_text().strip()
result = dict(task='8dbaea5', complete=trailer is not None, after_load=after_load, end=end,
              answer_source='relation_report trailer' if trailer else 'after-load query timing row; end unavailable',
              stop_reason=reason, child_exit=metrics['exit'], wrapper_exit=int((run / 'wrapper.exit').read_text()),
              serialized_file_records=len(files), serialized_units=sum(len(r.get('units', [])) for r in files),
              completed_timed_units=len(units), last_completed_timed_unit=units[-1] if units else None,
              program_relation_queries=len(program_queries), metrics=metrics, provenance=provenance)
if trailer: result['relation_report'] = trailer
(out / 'result.json.txt').write_text(json.dumps(result, indent=2) + '\n')
for path in [run / 'latent.jsonl', run / 'unit-times.jsonl', run / 'relation-times.jsonl', run / 'run.log', base / 'apply.log', base / 'main-npm.log']:
    if path.exists():
        with path.open('rb') as inp, (out / (path.name + '.gz')).open('wb') as raw, gzip.GzipFile(filename='', mode='wb', fileobj=raw, mtime=0) as compressed:
            shutil.copyfileobj(inp, compressed)
for path in [run / 'metrics.json', base / 'measurement-provenance.json', base / 'measure.sh', run / 'wrapper.log', base / 'measure-wrapper.log', run / 'wrapper.exit', base / 'main-npm.exit', base / 'apply.exit', run / 'stop-request.json', run / 'stop-reason.txt', run / 'process-status-at-stop.txt']:
    if path.exists(): shutil.copyfile(path, out / (path.name + '.txt' if path.suffix == '.log' or path.name == 'measurement-provenance.json' else path.name))
shutil.copyfile(__file__, out / 'harvest.py')
if (base / 'adapted/patch-set.md').exists(): shutil.copyfile(base / 'adapted/patch-set.md', out / 'adapted-patch-set.md')
source_paths = observations[0].get('sources', []) if observations else []
source_paths = sorted(set(source_paths + [provenance['entry'], provenance['environment']['LATENT_ONLY_FILE']]))
sources = {str(Path(name).relative_to(base / 'adapted')): dict(bytes=Path(name).stat().st_size, sha256=hashlib.sha256(Path(name).read_bytes()).hexdigest()) for name in source_paths if Path(name).is_file()}
(out / 'adapted-source-sha256.json').write_text(json.dumps(sources, indent=2) + '\n')
last = units[-1]['where'] if units else 'none'
text = f'''# checker.ts full census: task 8dbaea5

- Code pushed before the run: {provenance['code_commit']}.
- Main adaptations: {provenance['main_adaptations_commit']}.
- TypeScript source: {provenance['typescript_commit']} (6.0.3).
- Entry: {provenance['entry']}.
- Selection: LATENT_FULL=1, LATENT_ONLY_FILE={provenance['environment']['LATENT_ONLY_FILE']}.
- Per-unit relation queries: disabled. Only load and end endpoints are queried.
- Run complete: {result['complete']}; supervised child exit: {metrics['exit']}; wrapper exit: {result['wrapper_exit']}.
- Stop reason: {reason}.
- RelationOverflow after load: {json.dumps(after_load)}.
- RelationOverflow at run end: {json.dumps(end)}.
- Answer source: {result['answer_source']}.
- Completed timed units: {len(units)}; last completed: {last}.
- Serialized file records: {len(files)}; serialized units: {result['serialized_units']}.
- Wall time: {metrics['wall_seconds']:.3f} seconds; peak RSS: {metrics['peak_rss_kib'] / 1024:.1f} MiB.

An empty string means that query detected no relation overflow. A null end
answer means the completion trailer is absent and the end answer is unavailable.
File observations are buffered until lowering returns. A partial header-only\nJSONL is not a zero-findings result.\nTimed units count completed lowering attempts, not skipped/deferred catalogue
entries. A last timing location is the last completed unit, not a claim about
the exact node executing when a cutoff occurred.

The runtime is the meter entry overlay on this branch's latent overlay. Its
resolved source reach keeps the tsc entry first, while the file selection lowers
only checker.ts. The adapted tree was generated with archived main adaptation
scripts. measure.sh is the actual wrapper. The run has a 6000-second GNU timeout
and a 12288-MiB resident bound. No native output is claimed.
Job 1 validation and both red reader mutants are in ../meter-readers/.
'''
(out / 'report.md').write_text(text)
print(json.dumps({k: result[k] for k in ('complete', 'after_load', 'end', 'stop_reason', 'child_exit', 'completed_timed_units', 'last_completed_timed_unit')}, indent=2))
