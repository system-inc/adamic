import difflib, gzip, hashlib, json, pathlib, shutil, subprocess
out = pathlib.Path(__file__).resolve().parent
repo = pathlib.Path('/workspace/adamic')
required = ['before-apply','final-apply','guards','before-proofs','after-proofs','types','comparison','before-first-stop','after-first-stop','mutant-first-stop','stage71-idempotence']
statuses = {name: json.loads((out/(name+'.status.json')).read_text()) for name in required}
assert all(s['child_exit'] == 0 for s in statuses.values()), statuses
original_hashes = json.loads((out/'stage71-before-hashes.json').read_text())
replayed_tree = pathlib.Path('/workspace/scratch/line-map-stage71')
replayed_hashes = {str(p.relative_to(replayed_tree)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((replayed_tree/'src').rglob('*')) if p.is_file()}
assert original_hashes == replayed_hashes
(out/'stage71-idempotence-result.json').write_text(json.dumps(dict(whole_adaptation_replayed=True,files=len(original_hashes),byte_identical=True),indent=2)+'\n')
comparison = json.loads((out/'comparison.json').read_text())
assert not comparison['lane_api']['errors']
assert comparison['public_api']['normalized_changes'] == 0
assert comparison['lane_api']['sanctioned_api_unchanged']
roots = json.loads((out/'after-roots.json').read_text()); assert len(roots) == 79
stops = {label: [json.loads(line) for line in (out/(label+'-first-stop.jsonl')).read_text().splitlines()] for label in ['before','after','mutant']}
assert all(r['where'].endswith('/tracing.ts:206:67') for label in ['before','mutant'] for r in stops[label])
assert all(r['where'].endswith('/tracing.ts:232:34') for r in stops['after'])
for label in ['before','after']:
    services=json.loads((out/label/'services.json').read_text()); assert services['TS2386'] == 0
    assert not services['diagnostics']
    emit=json.loads((out/label/'declaration-emit.json').read_text()); assert not emit['diagnostics'];assert len(emit['files']) == 278
for name in ['scanner.ts','types.ts']:
    a=(pathlib.Path('/workspace/scratch/line-map-before/src/compiler')/name).read_bytes().decode()
    b=(pathlib.Path('/workspace/scratch/line-map-final/src/compiler')/name).read_bytes().decode()
    (out/(name+'.patch')).write_text(''.join(difflib.unified_diff(a.splitlines(True),b.splitlines(True),fromfile='before/'+name,tofile='after/'+name)))
report = '''# LineMapSource, baobab variant (e)

Base: `efe887763`; pinned TypeScript: `050880ce59e30b356b686bd3144efe24f875ebc8`.
Ruling: Adamic, baobab, Nov 11 11:28Z, accepted by Compiler.
Measurement precedent: the full fourth section of `ddbfff50`, and
getPositionOfLineAndCharacter's existing internal overload.

Owner is `71-writable-views`, which already guards SourceFileLike.lineMap.
The full-source comparison finds exactly the measured two-file edit. The whole
documented SourceFileLike and both helper bodies remain unchanged. No compiler
backend or rule changes are included.

| Proof | Result |
| --- | --- |
| Final keyed apply | Child exit 0 |
| Production adamic types | All 79 compiler roots, child exit 0 |
| Services full program | 0 diagnostics, 0 TS2386 |
| Ten emitted JavaScript files | Byte-identical |
| Built runtime export keys | 2,256, identical |
| Parser and SourceMapSource-shaped witnesses | Byte-identical at six positions, including 0, a line start, and end |
| stripInternal declaration emit | 278 files, 0 emit diagnostics |
| Upstream declaration bundler | Child exit 0 |
| Normalized public API delta | 0 |
| Actual stage3/lane api_check | 0 errors and 0 unsanctioned changes |
| sanctioned-api.json | Unchanged |
| Guard and whole adaptation idempotence | Second application is a no-op, all source bytes unchanged |
| Drift and half-state controls | All 11 rejected before writes |
| Baseline core.ts first stop | tracing.ts:206:67 |
| Edited core.ts first stop | tracing.ts:232:34, a cast the runtime cannot check |
| Internal-overload restoring mutant | tracing.ts:206:67 |

The original production census driver is used unchanged and reports both its
per-file record and its one-root whole-program repetition. Their results agree.
The mutant retains LineMapSource, the new implementation parameter and the new
getLineStarts parameter; only the first internal overload is removed, leaving
the public SourceFileLike overload first. Its source mutation and hashes are retained.

The bundler implementation is unchanged. Its TypeScript package import selects
the pinned stage3 API parser by absolute path; its other imports use upstream's
lockfile dependencies. The original script hash is retained in bundler-status.json.
The runtime witness bundles each actual completed source tree with esbuild.

All subprocess logs are files and statuses report the child's own exit. The
abandoned after-apply build was terminated and reaped after documentation was
finalized; only final-apply is used as proof. The initial setup control began
before main was fetched and failed on its old cohere pin; setup was repeated
successfully at main. The initial bundler dependency-resolution failure is under
controls and is excluded from final counts. No API sanction changed.

Large raw artifacts are gzip-compressed after comparison. artifact-index.json
records their original hashes and sizes; SHA256SUMS covers the retained files.
No raw proof artifact is added under stage3/adapt in the implementation commit.
'''
(out/'REPORT.md').write_text(report)
(out/'proof-summary.json').write_text(json.dumps(dict(status='pass',required_statuses=statuses,roots=79,stops=stops,comparison=comparison),indent=2)+'\n')
for p in out.rglob('node_modules'):
    if p.is_symlink(): p.unlink()
artifacts=[]
for p in sorted(out.rglob('*')):
    if not p.is_file() or p.name in ['artifact-index.json','SHA256SUMS']:continue
    data=p.read_bytes();record=dict(path=str(p.relative_to(out)),size=len(data),sha256=hashlib.sha256(data).hexdigest())
    if len(data)>1_000_000 and p.suffix != '.gz':
        target=p.with_name(p.name+'.gz')
        with target.open('wb') as file:
            with gzip.GzipFile(filename='',mode='wb',fileobj=file,mtime=0) as stream:stream.write(data)
        p.unlink();record['stored_path']=str(target.relative_to(out))
    artifacts.append(record)
(out/'artifact-index.json').write_text(json.dumps(artifacts,indent=2)+'\n')
(out/'SHA256SUMS').write_text(''.join(hashlib.sha256(p.read_bytes()).hexdigest()+'  '+str(p.relative_to(out))+'\n' for p in sorted(out.rglob('*')) if p.is_file() and p.name != 'SHA256SUMS'))
print(json.dumps(dict(status='pass',artifacts=len(artifacts),roots=79,api_errors=0)))
