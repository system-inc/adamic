import json,os,sys,subprocess,hashlib
from pathlib import Path
from runner import run,out,raw,repo
adapted=Path('/workspace/require-row-adapted')
roots=sorted((adapted/'src/compiler').rglob('*.ts'));assert len(roots)==79
(out/'roots.json').write_text(json.dumps(list(map(str,roots)),indent=2)+'\n')
for side,sha in [('base','71b126b0'),('head','738b3f31')]:
 source=repo if side=='base' else Path('/workspace/require-row-head')
 if side=='head':
  run('head-worktree',['git','worktree','add','--detach',str(source),sha])
  # Both commits have identical submodule pins. Share the already initialized dependency tree.
  if not (source/'cohere').is_symlink():
   (source/'cohere').rmdir();(source/'cohere').symlink_to(repo/'cohere',target_is_directory=True)
  if not (source/'stage3/api/node_modules').exists(): (source/'stage3/api/node_modules').symlink_to(repo/'stage3/api/node_modules',target_is_directory=True)
 run(side+'-compiler-source-clean',['git','diff','--exit-code',sha,'--','internal','stage3/census/tool','stage3/census/latent'],cwd=source)
 binary=raw/(side+'-census')
 run(side+'-build',['go','build','-o',str(binary),'./stage3/census/tool'],cwd=source)
 (out/(side+'-binary.sha256')).write_text(hashlib.sha256(binary.read_bytes()).hexdigest()+'\n')
 run(side+'-buildinfo',['go','version','-m',str(binary)],cwd=source)
 # Generate full latent overlay in scratch. Rename generated Go files to .go.txt before building.
 overlay=raw/(side+'-overlay')
 run(side+'-overlay',['python3','stage3/census/latent/make_overlay.py',str(source),str(overlay)],cwd=source)
 mapping=json.loads((overlay/'overlay.json').read_text())
 for key,val in mapping['Replace'].items():
  path=Path(val)
  if path.suffix=='.go':
   renamed=path.with_suffix('.go.txt');path.rename(renamed);mapping['Replace'][key]=str(renamed)
 (overlay/'overlay.json').write_text(json.dumps(mapping,indent=2)+'\n')
 run(side+'-overlay-format',['gofmt','-w',*sorted(set(mapping['Replace'].values()))],cwd=source)
 latent=raw/(side+'-latent')
 run(side+'-latent-build',['go','build','-buildvcs=false','-overlay='+str(overlay/'overlay.json'),'-o',str(latent),'./stage3/census/latent/tool'],cwd=source)
 env={k:v for k,v in os.environ.items() if not k.startswith('LATENT_')}
 times=raw/(side+'-unit-times.jsonl');stream=raw/(side+'-core.jsonl')
 env.update(LATENT_FULL='1',LATENT_ONLY_FILE=str(adapted/'src/compiler/core.ts'),LATENT_ASSERT_NO_OUTPUT='1',LATENT_TIMES=str(times))
 run(side+'-core',[str(latent),str(adapted/'src/compiler'),str(stream)],cwd=source,env=env,watch=[times,stream])
 (out/(side+'-latent-binary.sha256')).write_text(hashlib.sha256(latent.read_bytes()).hexdigest()+'\n')
 run(side+'-checkpoint-core',[sys.executable,str(out/'checkpoint.py'),side+' full core sites'])

for start in range(0,79,20):
 for side in ["base","head"]:
  source=repo if side=="base" else Path("/workspace/require-row-head")
  binary=raw/(side+"-census")
  group=start//20+1;rows=[]
  for index,root in enumerate(roots[start:start+20],start):
   stream=raw/f'{side}-root-{index+1:02}.jsonl'
   run(f'{side}-root-{index+1:02}',[str(binary),str(root),str(stream)],cwd=source,watch=[stream])
   observations=[json.loads(l) for l in stream.read_text().splitlines()];assert len(observations)==2
   assert {k:v for k,v in observations[0].items() if k!='seconds'}=={k:v for k,v in observations[1].items() if k!='seconds'}
   rows.append(observations[0])
  (out/f'{side}-group-{group}.json').write_text(json.dumps(rows,indent=2)+'\n')
  run(f'{side}-checkpoint-group-{group}',[sys.executable,str(out/'checkpoint.py'),f'{side} group {group}'])
