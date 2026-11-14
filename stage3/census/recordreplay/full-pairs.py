from pathlib import Path
import os,subprocess as s,json
D=Path('/workspace/scratch/recordreplay');R=D/'base';O=R/'stage3/census/recordreplay/full-sys';O.mkdir(parents=True,exist_ok=True)
for label,tree,head in [('A',D/'control',D/'control-sys-full.jsonl'),('B',D/'candidate2',D/'candidate2-sys-full.jsonl')]:
 out=D/(label+'-main-sys-full.jsonl');env=dict(os.environ,LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1',CENSUS_MODULE_FORMAT='CommonJSBundle',LATENT_ONLY_FILE=str(tree/'src/compiler/sys.ts'))
 s.run(['python3',str(D/'run.py'),label+'-main-sys-full',str(D/'latent-base'),str(tree/'src/compiler/sys.ts'),str(out)],cwd=R,env=env,check=True)
 s.run(['python3',str(R/'stage3/census/pair/pair.py'),str(out),str(head),str(tree),'--typescript',str(D/'head/stage3/api/node_modules/typescript'),'--json',str(O/(label+'.json')),'--report',str(O/(label+'.md'))],cwd=R,check=True)
 delta=json.loads((O/(label+'.json')).read_text())
 # Preserve every changed raw finding and continuation boundary in full.
 lines=['# '+label+' full sys.ts changes','']
 for field in ['lowered','moved','new_findings','unit_first_stop_delta']:
  lines += ['## '+field,'','```json',json.dumps(delta['raw_findings'][field],indent=2),'```','']
 (O/(label+'-every-change.md')).write_text('\n'.join(lines))
 print(label,'full paired complete',flush=True)
