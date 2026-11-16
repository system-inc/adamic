#!/usr/bin/env python3
"""Measurement only. Run after the README's scratch overlay build; checkpoint in batches."""
import argparse, concurrent.futures, datetime, gzip, hashlib, json, os, pathlib, subprocess, time
from collections import Counter
HERE=pathlib.Path(__file__).resolve().parent
ROOT=HERE.parents[2]
FAMILY=', whose readonly field pos becomes writable:'
def utc(): return datetime.datetime.now(datetime.timezone.utc).isoformat()
def dump(path,data):
    temp=path.with_suffix(path.suffix+'.tmp'); temp.write_text(json.dumps(data,indent=2,sort_keys=True)+'\n'); temp.replace(path)
def render(scratch):
    states=[json.loads(p.read_text()) for p in sorted((scratch/'pieces').glob('*.json'))]
    counts=Counter(); values={}; meterkeys=set()
    for s in states:
        for f in s.pop('rows',[]):
            key=(f['file'],f['line'],f['col'],f['reason'],f['unit'])
            counts[key]+=1; values[key]=f
            meterkeys.add((f['kind'],f['where'],f['reason'],f['text']))
    rows=[]
    for key in sorted(counts):
        r=values[key].copy(); r['multiplicity']=counts[key]; rows.append(r)
    with (HERE/'rows.jsonl.gz').open('wb') as f:
        with gzip.GzipFile(filename='',fileobj=f,mode='wb',mtime=0) as z:
            for r in rows: z.write((json.dumps(r,sort_keys=True,ensure_ascii=False)+'\n').encode())
    dump(HERE/'pieces.json',states)
    manifest=json.loads((scratch/'manifest.json').read_text())
    meterpath=ROOT/'stage3/meter/runs/20261115T131735Z.PG4H0R/main/report.json'
    meter=json.loads(meterpath.read_text()); reasons={k:v for k,v in meter['latent_lowering']['per_reason'].items() if k.startswith('Refused: ') and FAMILY in k}
    targets=Counter({k:0 for k in ('TextRange','SourceMapRange | undefined','TextRange | undefined','Mutable<Node>','others')}); callers=Counter(); kinds=Counter()
    for r in rows:
        target=r['reason'].split(FAMILY,1)[0].rsplit(' seen as ',1)[1]
        targets[target if target in ('TextRange','SourceMapRange | undefined','TextRange | undefined','Mutable<Node>') else 'others']+=1
        callers[r['unit_name'] or r['unit']]+=1; kinds[r['unit_kind']]+=1
    ok=[s for s in states if s['complete']]; failed=[s for s in states if not s['complete']]
    def table(c): return '\n'.join('| '+str(k).replace('&','&amp;').replace('<','&lt;').replace('>','&gt;').replace('|','\\|')+' | '+str(v)+' |' for k,v in c.most_common())
    report=f'''# Pos-writable rows — #aryr7ea

Status: {'complete' if len(states)==len(manifest['files']) and not failed else 'incomplete checkpoint'} at {utc()}.

Measured main tip `{manifest['sha']}` in a fresh detached worktree; TypeScript v6.0.3
`050880ce59e30b356b686bd3144efe24f875ebc8`. Recursive shallow submodules initialized;
`npm ci` in stage3/api; adapted using this tip's stage3/apply.sh. No compiler or adaptation changes.
Scratch: `{scratch}`. Build follows stage3/census/latent/README.md: make_overlay.py,
gofmt, then go build -buildvcs=false -overlay=... ./stage3/census/latent/tool.
TMPDIR, GOTMPDIR, GOCACHE, GOMODCACHE and stage3/npm caches are in scratch.

Each process loads the entire adapted src/compiler directory, with one exact absolute
LATENT_ONLY_FILE and LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 CENSUS_MODULE_FORMAT=CommonJSBundle.
Speculative mode is unset. Three files at most run concurrently (GOMEMLIMIT=3GiB,
GOMAXPROCS=2 per process). A piece is complete only
on exit 0 with a full-mode header, matching module format, exactly the selected file record,
and a relation_report trailer. The per-process cutoff is 6600 seconds, below two hours.

## Validation basis — #3dgsny7

The full audit's baseline was stale: full_audit.py expected string and number conditions
plus var; this main lowers the two conditions, so only var remains. full_audit.py is unchanged.
Production was built at c41391ea without an overlay. The exact full fixture stops at its
intentional checker errors. Its exact extracted three function reaches lowering and refuses
only var; replacing var with let succeeds. Separate observable string and number branches
also succeed and emit if (Boolean(...)). production-check.json records commands, timings,
outputs and binary hash; the small fixture inputs and emitted branch excerpt are retained.
The full-mode continuation proof rests on the passing legacy audit plus that production
check, as authorized for #3dgsny7; the stale full audit and its unverified mutants are not
claimed to pass. The initial stop is retained in preflight.json.

**A full-mode per-file census is a floor where whole-program facts differ.** Isolated
state, cross-file fact discovery and census-wide generic instantiation witnesses can differ.
Failures can block child traversal. This ledger is not proof of exhaustive lowering coverage
or successful compilation. Census runs invoke no backend and produce no usable IR.

## Counts

Completed {len(ok)}/{len(manifest['files'])} compiler .ts files; failed pieces: {len(failed)}.
Distinct requested keys (file, line, col, reason, unit): **{len(rows)}**.
Matching raw records (sum of multiplicity): **{sum(counts.values())}**.
Meter-style distinct (kind, where, reason, text), omitting unit: **{len(meterkeys)}**.

The saved daily meter 20261115T131735Z.PG4H0R, main `{meter['adamic_commit']}`,
counts **{sum(reasons.values())}** pos-writable sites across **{len(reasons)}** exact reasons.
It uses the same reason text generated by internal/lower/invariance.go (readonly field pos
becomes writable), confirmed against its per_reason ledger. Family selection is Refused
with the literal `{FAMILY}` in reason, across every source and target type.
The exact Node-to-TextRange reason alone has {reasons.get('Refused: a value of type Node seen as TextRange'+FAMILY+' a readonly field may hold something narrower than number, which a write of number would replace',0)} meter sites.
The row and record counts retain attempting-unit contexts and repeated observations;
see the reconciliation below. The baseline and measured compiler pins differ, so this
is count comparison rather than a paired whole-program identity check.

rows.jsonl.gz is sorted by (file, line, col, reason, unit), with one JSON object per
distinct key. multiplicity is the number of matching raw records for that key.
unit_kind and unit_name are copied from the attempting unit ledger, not inferred
from the nearest declaration at the diagnostic site.

## Target type (distinct requested keys)

| Target | Sites |
| --- | ---: |
{table(targets)}

## Attempting unit kind

| Kind | Sites |
| --- | ---: |
{table(kinds)}

## Callee / receiving slot

The findings schema records the attempting unit, location, reason and text, but does not
name the callee or receiving slot for this refusal. No callee counts are inferred from
reason types. The following first cut groups the attempting function/declaration name;
these are callers/owners, not identified callees. All rows retain unit location and kind. Locations are adapted-root relative, with one-based
lines and UTF-16 columns as emitted by the census; exact reason strings are unchanged.

| Attempting name (or location when unnamed) | Sites |
| --- | ---: |
{table(callers)}

## Run times and failures

Initial setup manifest {manifest['started_utc']}; this checkpoint {utc()}.
Actual per-file census start/end times are recorded in pieces.json.
Each piece's UTC start/end, wall time, exit status, complete flag, raw SHA256, and log path
are in pieces.json. Source file inventory and adapted source SHA256s are in manifest.json.
Rows from partial pieces, if any, are retained and explicitly marked piece_complete=false.

| Failed file | Exit | Reason |
| --- | ---: | --- |
'''
    for s in failed: report+=f"| {s['file']} | {s['exit_code']} | {s['error']} |\n"
    if not failed: report+='| none | | |\n'
    if states:
        first=min(state['started_utc'] for state in states)
        last=max(state['ended_utc'] for state in states)
        duration=(datetime.datetime.fromisoformat(last)-datetime.datetime.fromisoformat(first)).total_seconds()
        report+='\n## Count reconciliation and measured time window\n\n'
        report+=f"Matching records {sum(counts.values())} = {len(rows)} distinct requested keys + {sum(counts.values())-len(rows)} repeat records.\n\n"
        report+=f"Distinct requested keys {len(rows)} = {len(meterkeys)} meter-style sites + {len(rows)-len(meterkeys)} additional attempting-unit contexts.\n\n"
        if len(meterkeys)==sum(reasons.values()):
            report+=f"The observed meter-style total equals the saved meter's {sum(reasons.values())}. Exact-reason count agreement is recorded below. The pins differ; count agreement does not establish position or whole-program fact identity.\n\n"
        else:
            report+=f"The observed meter-style count differs from the saved {sum(reasons.values())} by {len(meterkeys)-sum(reasons.values()):+d}. This different-pin, separate-process comparison is not a paired whole-program rerun; the delta cannot be uniquely attributed to pin changes versus file-scoped facts alone.\n\n"
        report+=f"Per-file runs started {first}; latest completed piece ended {last}; observed span {duration:.3f} seconds. Sum of piece wall times: {sum(state['wall_seconds'] for state in states):.3f} seconds (pieces overlap).\n\n"
        report+='| Slowest piece | Wall seconds | Complete |\n| --- | ---: | --- |\n'
        for state in sorted(states,key=lambda state:state['wall_seconds'],reverse=True)[:5]:
            report+=f"| {state['file']} | {state['wall_seconds']:.3f} | {state['complete']} |\n"
    old_reasons={reason.removeprefix('Refused: '):count for reason,count in reasons.items()}
    actual_reasons=Counter(key[2] for key in meterkeys)
    row_reasons=Counter(row['reason'] for row in rows)
    record_reasons=Counter()
    for key,count in counts.items(): record_reasons[key[3]]+=count
    reason_comparison=[dict(reason=reason,meter_sites=old_reasons.get(reason,0),observed_sites=actual_reasons.get(reason,0),distinct_rows=row_reasons.get(reason,0),records=record_reasons.get(reason,0),delta=actual_reasons.get(reason,0)-old_reasons.get(reason,0)) for reason in sorted(set(old_reasons)|set(actual_reasons))]
    dump(HERE/'meter-comparison.json',dict(baseline_sha=meter['adamic_commit'],measured_sha=manifest['sha'],baseline_sites=sum(old_reasons.values()),observed_sites=len(meterkeys),distinct_rows=len(rows),records=sum(counts.values()),reasons=reason_comparison))
    changes=sorted((row for row in reason_comparison if row['delta']),key=lambda row:(-abs(row['delta']),row['reason']))
    report+='\n## Exact-reason comparison with the saved meter\n\n'
    if not changes:
        report+='All exact-reason site counts agree with the saved meter. Attempting-unit contexts and repeated records still use the separate counts above.\n'
    else:
        report+=f"{len(changes)} exact reasons have different site counts. meter-comparison.json contains every reason, count and delta; the largest differences are below. This identifies the observations that differ, not a unique causal attribution across the different pins and file scopes.\n\n"
        report+='| Exact reason | Meter sites | Observed sites | Delta |\n| --- | ---: | ---: | ---: |\n'
        for row in changes[:10]:
            label=row['reason'].replace('&','&amp;').replace('<','&lt;').replace('>','&gt;').replace('|','\\|')
            report+=f"| {label} | {row['meter_sites']} | {row['observed_sites']} | {row['delta']:+d} |\n"
    verification=HERE/'verification.json'
    if verification.exists():
        checked=json.loads(verification.read_text())
        if checked['distinct_rows']==len(rows) and checked['complete_pieces']==len(ok) and len(checked['failed_pieces'])==len(failed):
            report+='\nIndependent verification: **pass** (verification.json). Every row key, multiplicity, exact reason, attempting-unit metadata, raw stream SHA256 and adapted compiler source hash was reconciled by verify.py. The compiler/adaptation diff is empty; full_audit.py is unchanged.\n'
    (HERE/'REPORT.md').write_text(report)
    dump(HERE/'manifest.json',manifest)
    print(f"checkpoint: {len(ok)}/{len(manifest['files'])} complete, {len(failed)} failed, {len(rows)} rows",flush=True)
def one(file,scratch):
    adapted=scratch/'adapted'; path=adapted/file; slug=file.replace('/','__'); raw=scratch/'raw'/f'{slug}.jsonl'; log=scratch/'raw'/f'{slug}.log'
    dest=scratch/'pieces'/f'{slug}.json'
    if dest.exists(): return
    env=os.environ.copy(); env.pop('LATENT_SPECULATIVE',None)
    env.update(LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1',CENSUS_MODULE_FORMAT='CommonJSBundle',LATENT_ONLY_FILE=str(path),GOMEMLIMIT='3GiB',GOMAXPROCS='2',LATENT_TIMES=str(scratch/'raw'/f'{slug}.times.jsonl'))
    state=dict(file=file,started_utc=utc(),log=str(log)); start=time.monotonic()
    try:
        with log.open('w') as out: code=subprocess.run([str(scratch/'latent-census'),str(adapted/'src/compiler'),str(raw)],env=env,stdout=out,stderr=subprocess.STDOUT,timeout=6600).returncode
    except subprocess.TimeoutExpired: code=124
    state.update(exit_code=code,wall_seconds=time.monotonic()-start,ended_utc=utc(),complete=False,error='',rows=[])
    records=[]
    if raw.exists():
        state['raw_sha256']=hashlib.sha256(raw.read_bytes()).hexdigest()
        for line in raw.open():
            try: records.append(json.loads(line))
            except json.JSONDecodeError: state['error']='truncated JSONL'
    header=records[0] if records else {}; files=[r for r in records if r.get('file')]
    state['complete']=(code==0 and header.get('latent_mode')=='full' and header.get('module_format')=='CommonJSBundle' and len(files)==1 and files[0]['file']==str(path) and records[-1].get('status')=='relation_report')
    if not state['complete']: state['error']=state['error'] or 'nonzero exit or incomplete/mismatched stream'
    units={u['where']:u for r in files for u in r.get('units',[])}
    def relative(s): return s.replace(str(adapted)+'/', '')
    for record in records:
        for f in record.get('findings',[]):
            if f.get('kind')!='Refused' or FAMILY not in f.get('reason',''): continue
            where=relative(f['where']); fn,line,col=where.rsplit(':',2); unit=units.get(f['unit'],{})
            state['rows'].append(dict(file=fn,line=int(line),col=int(col),where=where,reason=f['reason'],text=relative(f['text']),kind=f['kind'],phase=f['phase'],unit=relative(f['unit']),unit_kind=unit.get('kind','unknown'),unit_name=unit.get('name',''),piece_file=file,piece_complete=state['complete']))
    dump(dest,state); print(file,code,round(state['wall_seconds'],2),len(state['rows']),flush=True)
def main():
    ap=argparse.ArgumentParser(); ap.add_argument('scratch',type=pathlib.Path); ap.add_argument('--render',action='store_true'); ap.add_argument('--limit',type=int); args=ap.parse_args(); s=args.scratch.resolve()
    for folder in ['raw','pieces']: (s/folder).mkdir(exist_ok=True)
    manifest=s/'manifest.json'
    if not manifest.exists():
        files=sorted(str(p.relative_to(s/'adapted')) for p in (s/'adapted/src/compiler').rglob('*.ts'))
        dump(manifest,dict(sha='c41391ea3346727231dfa3472d061826e26f6aad',started_utc=utc(),files=files,source_sha256={f:hashlib.sha256((s/'adapted'/f).read_bytes()).hexdigest() for f in files}))
    preflight=HERE/'preflight.json'
    if preflight.exists() and json.loads(preflight.read_text())['full']['exit_code'] != 0 and not json.loads(preflight.read_text()).get('stale_baseline_confirmed'):
        raise SystemExit('stopped: saved full-mode audit failed; see REPORT.md')
    if args.render: render(s); return
    files=[f for f in json.loads(manifest.read_text())['files'] if not (s/'pieces'/(f.replace('/','__')+'.json')).exists()]
    if args.limit: files=files[:args.limit]
    with concurrent.futures.ThreadPoolExecutor(max_workers=3) as pool:
        pending=[pool.submit(one,f,s) for f in files]
        for future in concurrent.futures.as_completed(pending):
            future.result(); render(s)
            subprocess.run(['git','add','stage3/census/pos-writable-rows'],cwd=ROOT,check=True)
            subprocess.run(['git','commit','-m','census: checkpoint pos-writable rows'],cwd=ROOT,check=True)
            if os.environ.get('POS_ROWS_API_PUBLISH') != '1':
                subprocess.run(['git','push','origin','HEAD:refs/heads/typescript/pos-writable-rows'],cwd=ROOT,check=False,timeout=90)
if __name__=='__main__': main()
