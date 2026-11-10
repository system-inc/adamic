import pathlib,subprocess,sys,json,gzip
out=pathlib.Path('/workspace/never-call-report/stage3/census/never-call-392cb60b');repo=out.parents[2]
for case in ['debug-full','semver-full','tsc-first']:
 def stream(side):
  p=out/(case+'-'+side+'.jsonl');return p if p.exists() else p.with_suffix('.jsonl.gz')
 cmd=['python3',str(repo/'stage3/census/pair/pair.py'),str(stream('parent')),str(stream('head')),'/workspace/never-call-adapted','--typescript',str(repo/'stage3/api/node_modules/typescript'),'--json',str(out/(case+'-paired.json')),'--report',str(out/(case+'-paired.md'))]
 with (out/(case+'-pair.log')).open('wb') as f:r=subprocess.run(cmd,cwd=repo,stdout=f,stderr=subprocess.STDOUT)
 print(case,'pair child exit code:',r.returncode,flush=True)
 if r.returncode:sys.exit(r.returncode)
 d=json.loads((out/(case+'-paired.json')).read_text())
 print(case,{name:{k:len(v[k]) for k in ['lowered','moved','new_findings','new_sites','unit_first_stop_delta','regressions']} for name,v in d.items() if name in ['raw_findings','distinct_reads']},flush=True)
