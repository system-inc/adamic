"""Run completed-tree verification, retaining each command's complete log."""
from pathlib import Path
import subprocess, os, json, hashlib
repo=Path('/workspace/adamic');out=repo/'stage3/adapt/99-correctness-axolotl-y';env=dict(os.environ,NODE_PATH=str(repo/'stage3/api/node_modules'))
results={}
def run(name,args,cwd=repo):
 with (out/(name+'.log')).open('wb') as log:
  r=subprocess.run(args,cwd=cwd,env=env,stdout=log,stderr=log)
 results[name]={'command':args,'cwd':str(cwd),'exit':r.returncode}
 (out/'checks.json').write_text(json.dumps(results,indent=2)+'\n')
 assert r.returncode==0,(name,r.returncode)
 print(name,'exit 0',flush=True)
tree=Path('/workspace/y-after');modules=tree/'node_modules'
if not modules.exists():modules.symlink_to('/workspace/typescript-source/node_modules')
roots=sorted((tree/'src/compiler').rglob('*.ts'));assert len(roots)==79
run('types-after',['/workspace/scratch/y-bin/adamic','types',*[str(p) for p in roots]])
run('build-after',['npm','run','build:compiler'],tree)
run('comparison',['node',str(out/'compare.cjs'),'/workspace/y-before',str(tree),'/workspace/y-owner-check'])
comparison=json.loads((out/'comparison.log').read_text());assert comparison['emittedJavascriptFiles']==9
(out/'comparison.json').write_text(json.dumps(comparison,indent=2)+'\n')
run('adapter-replay',['node',str(out/'adapt.cjs'),str(tree)])
assert json.loads((out/'adapter-replay.log').read_text())=={'files':0,'edits':0,'javascriptChanging':False}
run('measure-after',['node',str(out/'measure.cjs'),str(tree)])
# The restored original must reproduce the exact reported first stop and source bytes.
mutant=Path('/workspace/y-mutant');before=Path('/workspace/y-before');names=sorted(p.relative_to(before/'src') for p in (before/'src').rglob('*') if p.is_file())
assert all((before/'src'/n).read_bytes()==(mutant/'src'/n).read_bytes() for n in names)
rows=[json.loads(line) for line in (out/'mutant-binder.jsonl').read_text().splitlines()];assert len(rows)==2
assert all(r['kind']=='Refused' and r['where'].endswith('/core.ts:966:72') and r['reason']=='a value of type T[] seen as {}[], which can write {} where T is read' for r in rows)
(out/'mutant-comparison.json').write_text(json.dumps({'sourceFiles':len(names),'mutantEqualsBaseline':True,'kind':rows[0]['kind'],'where':'core.ts:966:72','reason':rows[0]['reason'],'matchesReportedFirstStop':True},indent=2)+'\n')
print('all completed-tree checks passed',flush=True)
