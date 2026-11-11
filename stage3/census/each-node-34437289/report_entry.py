from summary import *
from collections import defaultdict
pair=json.loads((root/'pair-delta.json').read_text())
static=json.loads((root/'entry-static-units.json').read_text())
reconcile={}
for n in ['parent','head']:
 rs=rows('entry-'+n); units=[u for r in rs for u in r.get('units',[])]
 assert rs[0]['status']=='measurement' and rs[0]['latent_mode']=='first-error' and not rs[0]['diagnostics']
 assert len(units)==static['total']
 assert len([r for r in rs if r.get('file')])==len(static['files'])
 reconcile[n]={'units':len(units),'source_records':len(static['files']),'status_counts':dict(Counter(u['status'] for u in units)),'checker_diagnostics':len(rs[0]['diagnostics'])}
(root/'entry-reconciliation.json').write_text(json.dumps(reconcile,indent=2)+'\n')
text=(root/'debug-report-draft.md').read_text().replace('The tsc entry comparison is still running.','The completed tsc entry comparison is reported below.').replace('The final report commit will have','The final report commit has')
lines=['Both entry censuses exited 0, are checker-clean, and contain all 81 source records and 4,562 top-level units. entry-reconciliation.json records the eligibility/status census.','', 'The requested head includes the adaptation branch\'s main ancestry. Global entry deltas measure that whole scratch merge. The restoring-mutant debug run isolates the old-body substitution.','', 'The pair tool is unmodified. Because the requested inputs differ in debug.ts, entry-parent-mapped.jsonl maps the parent root and only locations on exactly equal complete debug source lines into head coordinates. It rejects an edited-line location or any other src file difference. pair-location-map.json retains every mapping; source-manifest.json retains both original hashes. The original streams are retained separately. The pair tool then validates matching headers, unit identities and eligibility and normalizes property/element access chains using pinned stock TypeScript.','', '| Measure | Raw | Distinct |','| --- | ---: | ---: |']
for label,field in [('Cleared, no remaining finding','lowered'),('Changed reason at an already finding-bearing site','moved'),('New findings, including exposed reasons','new_findings'),('New sites','new_sites'),('Units whose first stop changed','unit_first_stop_delta'),('Regressions','regressions')]:
 lines.append('| '+label+' | '+str(len(pair['raw_findings'][field]))+' | '+str(len(pair['distinct_reads'][field]))+' |')
lines += ['| All disappeared finding keys, cleared plus changed | '+str(len(pair['raw_findings']['lowered'])+len(pair['raw_findings']['moved']))+' | '+str(len(pair['distinct_reads']['lowered'])+len(pair['distinct_reads']['moved']))+' |','', '| Kind | Parent raw | Head raw | Parent distinct | Head distinct |','| --- | ---: | ---: | ---: | ---: |']
kinds=sorted(set(pair['raw_findings']['totals']['before'])|set(pair['raw_findings']['totals']['after'])|set(pair['distinct_reads']['totals']['before'])|set(pair['distinct_reads']['totals']['after']))
for kind in kinds:
 vals=[pair[v]['totals'][s].get(kind,0) for v,s in [('raw_findings','before'),('raw_findings','after'),('distinct_reads','before'),('distinct_reads','after')]]
 lines.append('| '+kind+' | '+' | '.join(map(str,vals))+' |')
lines += ['', '### Every moved unit, with old and new stops','', 'A changed first stop in an already-blocked unit is listed here separately from regressions. A unit that previously had no lowering stop and now stops is a regression.','', '| Views | Unit | Parent stop | Head stop |','| --- | --- | --- | --- |']
def stop(f):
 return 'No lowering stop' if not f else f['kind']+' at '+short(f['where'])+': '+f['reason']
moves={}
for view in ['raw_findings','distinct_reads']:
 for r in pair[view]['unit_first_stop_delta']:
  u=r['unit']; label=short(u['where'])+' '+u.get('name',u.get('kind',''))
  key=(label,stop(r['base']),stop(r['head']))
  moves.setdefault(key,[]).append(view)
for (u,b,h),views in sorted(moves.items()): lines.append('| '+', '.join(views)+' | '+md(u)+' | '+md(b)+' | '+md(h)+' |')
if not moves: lines += ['| Both | None | None | None |']
lines += ['', '### Every regression','']
reg={}
for view in ['raw_findings','distinct_reads']:
 for r in pair[view]['regressions']:
  key=json.dumps(r,sort_keys=True);reg.setdefault(key,[]).append(view)
if not reg: lines += ['None. No previously lowered unit now stops, no previously finding-free refusal-scan site now finds, and no new panic is reported.']
for encoded,views in reg.items():
 r=json.loads(encoded); f=r.get('finding') or r.get('head') or r.get('base')
 lines += ['- '+r['type']+' ('+', '.join(views)+'): `'+short(f['where'])+'`: '+f['kind']+': '+f['reason']]
lines += ['', '### Entry run resources','', '| Run | Wall seconds | Peak RSS KiB | Exit |','| --- | ---: | ---: | ---: |']
for n in ['parent','head']:
 t=json.loads((root/('entry-'+n+'.time')).read_text());lines.append(f'| {n} | {t["wall_seconds"]:.3f} | {t["peak_rss_kib"]} | {t["child_exit"]} |')
lines += ['', 'pair-report.md contains all per-reason counts and the raw/distinct summaries. pair-delta.json retains every cleared finding, reason transition, new finding, first-stop change and regression. Large artifacts have a .gz suffix.','']
text=text.replace('Pending: both censuses are active. No regression conclusion is claimed yet.','\n'.join(lines))
assert '\u2014' not in text and '\u2013' not in text
(dest/'REPORT.md').write_text(text)
(root/'final-report.md').write_text(text)
print('report completed; moved-unit rows:',len(moves),'regression records:',len(reg))
