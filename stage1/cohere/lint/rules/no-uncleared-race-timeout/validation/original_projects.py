import argparse, gzip, json, pathlib, re, shutil, subprocess, time
parser=argparse.ArgumentParser()
parser.add_argument('artifacts',type=pathlib.Path)
parser.add_argument('cases',type=pathlib.Path)
parser.add_argument('evidence',type=pathlib.Path)
parser.add_argument('--drop-companions',action='store_true')
args=parser.parse_args()
root=pathlib.Path(__file__).resolve().parents[6]
source=(root/'cohere/internal/lint/rules/nexus/correctness_no_uncleared_race_timeout_test.go').read_text()
block=source.split('var correctnessNoUnclearedRaceTimeoutNodeTimers = strings.Join([]string{',1)[1].split('}, "\\n")',1)[0]
timers='\n'.join(json.loads(s) for s in re.findall(r'^\s*("(?:[^"\\]|\\.)*"),$',block,re.M))
helper=(root/'cohere/internal/lint/testing/program.go').read_text()
config=json.loads(re.search(r'const defaultTsConfig = `(.*?)`',helper,re.S).group(1))
with gzip.open(args.cases,'rt') as f:cases=json.load(f)
assert len(cases) == 21, ("missing upstream captures", len(cases))
args.evidence.mkdir(parents=True,exist_ok=True)
results=[];started=time.monotonic()
for index,case in enumerate(cases):
 directory=args.artifacts/f'case-{index:03}'
 if directory.exists():shutil.rmtree(directory)
 path=directory/case['file'].lstrip('/')
 path.parent.mkdir(parents=True,exist_ok=True);path.write_text(case['source'])
 if case.get('otherFiles',0):
  auxiliary=directory/'repository/types/node/web-globals/timers.d.ts'
  auxiliary.parent.mkdir(parents=True,exist_ok=True)
  if not args.drop_companions:auxiliary.write_text(timers)
  elif auxiliary.exists():auxiliary.unlink()
 project=directory/'tsconfig.json';project.write_text(json.dumps(config))
 manifest=directory/'manifest.txt';manifest.write_text(f"program {project}\n{path}\t{case['rule']}\n")
 transcript=directory/'transcript'
 commands={'Go':[str(args.artifacts/'oracle'),'--manifest',str(manifest)],'sanitized native':[str(args.artifacts/'scanner-sanitized'),'--manifest',str(manifest),'--record',str(transcript)],'Node':['node','--disable-warning=ExperimentalWarning',str(root/'oracle/node.mjs'),str(root/'stage1/cohere/lint/main.ts'),'--manifest',str(manifest),'--replay',str(transcript)],'emitted JavaScript':['node','--disable-warning=ExperimentalWarning',str(root/'oracle/node.mjs'),str(args.artifacts/'lint.mjs'),'--manifest',str(manifest),'--replay',str(transcript)]}
 outputs={}
 for name,command in commands.items():
  slug=name.replace(' ','-');out=args.evidence/f'{index:03}-{slug}.stdout';err=args.evidence/f'{index:03}-{slug}.stderr'
  with out.open('wb') as stdout,err.open('wb') as stderr:run=subprocess.run(command,cwd=root,stdout=stdout,stderr=stderr)
  assert run.returncode==0,(index,name,run.returncode,str(err))
  outputs[name]=out.read_bytes()
 expected=len(case.get('findings',[]));actual=sum(line.startswith(b'range ') for line in outputs['Go'].split(b'\n'))
 assert actual==expected,('original fixture finding count',index,expected,actual,'drop_companions',args.drop_companions)
 assert all(value==outputs['Go'] for value in outputs.values()),('runtime mismatch',index)
 results.append({'case':index,'companion_files':case.get('otherFiles',0),'findings':actual,'bytes_identical':len(outputs['Go'])})
summary={'cases':len(results),'findings':sum(x['findings'] for x in results),'wall_seconds':time.monotonic()-started,'config':config,'results':results,'note':'Reconstructs each original upstream Node timer declaration project, not the shared single-subject capture. Original captured finding counts independently prevent matching zero outputs from certifying a dropped companion.'}
(args.evidence/'original-project-results.json').write_text(json.dumps(summary,indent=2)+'\n')
print(json.dumps({k:v for k,v in summary.items() if k!='results'}))
