import pathlib,json,gzip,collections,re
out=pathlib.Path(__file__).resolve().parents[1];follow=out/'continuation'
def read_stream(path):
 if not path.exists():path=pathlib.Path(str(path)+'.gz')
 if not path.exists():return []
 rows=[]
 with (gzip.open(path,'rt') if path.suffix=='.gz' else path.open()) as stream:
  for line in stream:
   try:rows.append(json.loads(line))
   except ValueError:pass
 return rows
def location(where):
 if '/src/compiler/' in where:return where.split('/src/compiler/',1)[1]
 return where

def within(where,site):
 where=location(where);m=re.fullmatch(r'(.+):(\d+):(\d+)',where)
 if not m or m[1]!=site['file']:return False
 if 'call_span' not in site:return where==site['location']
 span=site.get('_comparison_span',site['call_span']);pos=(int(m[2])-1,int(m[3])-1)
 return (span['start']['line'],span['start']['character'])<=pos<(span['end']['line'],span['end']['character'])
sites=json.loads((out/'sites.json').read_text());grouped={}
for site in sites:
 grouped.setdefault(site['location'],dict(site,baseline_findings=[]))['baseline_findings'].append(site)
scopes=collections.Counter((s['file'],json.dumps(s.get('call_span'),sort_keys=True)) for s in grouped.values() if s.get('call_span'))
for s in grouped.values():
 if s.get('call_span') and scopes[(s['file'],json.dumps(s['call_span'],sort_keys=True))]>1:s['_comparison_span']=s['argument_span']
files=json.loads((follow/'variant-files.json').read_text());result=[];summaries=[]
for label in ['a','b']:
 raw=[];units=[];status={};go_sites={}
 for name in files:
  stem=name.replace('/','__').removesuffix('.ts');p=follow/label/'raw'/(stem+'.status.json');assert p.exists(),str(p)
  s=json.loads(p.read_text());rows=read_stream(p.with_name(stem+'.jsonl'));s['complete_stream']=s['child_exit']==0 and any(r.get('status')=='relation_report' for r in rows);status[name]=s
  for row in rows:
   raw.extend(row.get('findings') or []);units.extend(row.get('units') or [])
   for d in row.get('diagnostic_sites') or []:
    key=(d.get('file'),d.get('start'),d.get('end'),d.get('text'))
    if key not in go_sites:
     p=pathlib.Path(d.get('file',''));start=d.get('start',-1)
     if p.is_file() and start>=0:
      prefix=p.read_bytes()[:start];go_sites[key]=dict(d,mapped_location=location(str(p))+':'+str(prefix.count(b'\n')+1)+':'+str(len(prefix)-prefix.rfind(b'\n')))

 diagnostics=json.loads((follow/label/'whole-src-diagnostics.json').read_text());seen=set();findings=[]
 for f in raw:
  key=tuple(f.get(k,'') for k in ['kind','where','reason','text','phase','unit'])
  if key not in seen:seen.add(key);findings.append(f)
 variant=[]
 for loc,site in grouped.items():
  matched=[f for f in findings if within(f.get('where',''),site) or within(f.get('site_where',''),site)]
  diag=[d for d in diagnostics if d.get('file','').startswith('src/compiler/') and within(d['file'].removeprefix('src/compiler/')+':'+str(d['line'])+':'+str(d['column']),site)]
  go_diag=[d for d in go_sites.values() if within(d['mapped_location'],site)]
  unit=[u for u in units if location(u.get('where',''))==location(site['unit'])]
  observed_roots=site.get('observed_in') or [site['file']];complete=all(status.get(n,{}).get('complete_stream',False) for n in observed_roots)
  eligible=any(not u.get('checker_diagnostics') and u.get('status') in ['attempted','instantiated'] for u in unit)
  verdict='incomplete stream' if not complete else 'checker diagnostic at original call' if diag or go_diag else 'finding at original call' if matched else 'enclosing unit excluded or not attempted' if not eligible else 'no finding at original call; unit attempted'
  row=dict(variant=label,location=loc,callee=site['callee'],baseline_findings=site['baseline_findings'],findings=matched,checker_diagnostics=diag,go_checker_diagnostics=go_diag,unit_statuses=unit,complete_stream=complete,verdict=verdict)
  result.append(row);variant.append(row)
 summary=json.loads((follow/label/'summary.json').read_text());summary.update({'census_files':len(files),'complete_streams':sum(s['complete_stream'] for s in status.values()),'baseline_locations':len(grouped),'site_verdicts':dict(collections.Counter(r['verdict'] for r in variant)),'matched_finding_kinds':dict(collections.Counter(f['kind'] for r in variant for f in r['findings'])),'first_stop':read_stream(follow/label/'core-first-stop.jsonl')[0]});summaries.append(summary)
(follow/'site-comparison.json').write_text(json.dumps(result,indent=2)+'\n');(follow/'variant-summary.json').write_text(json.dumps(summaries,indent=2)+'\n')
print(json.dumps(summaries,indent=2))
