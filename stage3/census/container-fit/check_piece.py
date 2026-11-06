"""Pair one completed schedule, verify public bytes, and publish its reviewable checkpoint."""
import collections,gzip,hashlib,json,pathlib,re,subprocess,sys
O=pathlib.Path(__file__).resolve().parent;R=O.parents[2];E=O/'evidence';pid=sys.argv[1];assert re.fullmatch('[pq][0-9]{2}',pid)
P=E/'pair-pieces';P.mkdir(exist_ok=True)
for label in ['base','log']:
 D=E/'latent'/label/pid;m=json.loads((D/'metrics.json').read_text());assert m['returncode']==0 and not m.get('validation_error')
 for k,v in m.items():
  if k.endswith('_sha256'):
   with gzip.open(D/(k[:-7]+'.gz'),'rb') as f:assert hashlib.sha256(f.read()).hexdigest()==v
subprocess.run([sys.executable,str(R/'stage3/census/pair/pair.py'),str(E/'latent/base'/pid/'records.jsonl.gz'),str(E/'latent/log'/pid/'records.jsonl.gz'),'/workspace/container-fit-adapted','--typescript','/workspace/container-fit-base/stage3/api/node_modules/typescript','--json',str(P/(pid+'.json')),'--report',str(P/(pid+'.md'))],check=True,timeout=120)
md=P/(pid+'.md');md.write_text(md.read_text().replace(chr(92)*2+'|',chr(92)+'|'))
d=json.loads((P/(pid+'.json')).read_text())['distinct_reads'];byte=None
plan=json.loads((E/'plan/INDEX.json').read_text())
if pid in plan.get('public_head_check_pieces',[]):
 a=gzip.open(E/'latent/head'/pid/'records.jsonl.gz','rb').read();b=gzip.open(E/'latent/log'/pid/'records.jsonl.gz','rb').read();assert a==b
 B=E/'byte-checks';B.mkdir(exist_ok=True);byte=dict(equal=True,bytes=len(a),raw_sha256=hashlib.sha256(a).hexdigest());(B/(pid+'.json')).write_text(json.dumps(byte,indent=2)+'\n')
with gzip.open(E/'latent/log'/pid/'records.jsonl.gz','rt') as f:rows=list(map(json.loads,f))
caps=[x for row in rows[1:] for x in row['findings'] if re.search('a fit walk past .* distinct pairs',x['reason'])]
summary=dict(piece=pid,first_stop_changes=len(d['unit_first_stop_delta']),regressions=len(d['regressions']),cap_rows=caps,head_log_byte_identity=byte)
(P/(pid+'-summary.json')).write_text(json.dumps(summary,indent=2)+'\n')
summaries=[json.loads(p.read_text()) for p in sorted(P.glob('*-summary.json'))]
(E/'latent/INDEX.json').write_text(json.dumps(dict(completed_pairs=summaries,planned_pieces=plan['pieces'],failed_sizing_piece='base/p00; excluded'),indent=2)+'\n')
report=O/'REPORT.md';text=report.read_text();marker='\n## Live completed-pair index\n';text=text.split(marker)[0]
text+=marker+'\nThese are partial paired results. Final cohort tables, full regression lists, walk distributions and common-unit sums wait for the entire partition. Each public chunk also has an exact head/log byte check.\n\n| Piece | Changed first stops | Regressions | Cap rows | Public byte check |\n| --- | ---: | ---: | ---: | --- |\n'
for r in summaries:text+=f'| {r["piece"]} | {r["first_stop_changes"]} | {r["regressions"]} | {len(r["cap_rows"])} | {"equal" if r["head_log_byte_identity"] else "not requested"} |\n'
if caps:text+='\n**N=4096 IS TOO LOW HERE.** Every cap row for this piece is in evidence/pair-pieces/'+pid+'-summary.json.\n'
report.write_text(text)
subprocess.run(['git','add','-f',str(P/(pid+'.json')),str(P/(pid+'.md')),str(P/(pid+'-summary.json')),str(E/'latent/INDEX.json'),str(report),*([str(E/'byte-checks'/(pid+'.json'))] if byte else [])],cwd=R,check=True)
paths=subprocess.check_output(['git','diff','--cached','--name-only'],cwd=R,text=True).splitlines();assert all(p.startswith('stage3/census/container-fit/') for p in paths)
subprocess.run(['git','diff','--cached','--check'],cwd=R,check=True)
if paths:subprocess.run(['git','commit','-m','stage3/census: verify paired container fit '+pid],cwd=R,check=True)
with (E/'publication.log').open('a') as f:subprocess.run(['timeout','--kill-after=10s','120','git','push','origin','HEAD:refs/heads/typescript/container-fit-census'],cwd=R,stdout=f,stderr=subprocess.STDOUT,check=True)
print(json.dumps(dict(piece=pid,first_stop_changes=summary['first_stop_changes'],regressions=summary['regressions'],cap_rows=len(caps),byte_identity=bool(byte))),flush=True)
if caps:print('N=4096 IS TOO LOW: see saved cap rows for '+pid,flush=True)
