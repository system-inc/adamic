#!/usr/bin/env python3
"""Inspect reduced tree disagreements with the existing pinned parser drivers."""
import argparse,concurrent.futures,difflib,json,pathlib,subprocess
p=argparse.ArgumentParser();p.add_argument('witnesses');p.add_argument('oracle');p.add_argument('scratch');p.add_argument('out');a=p.parse_args()
root=pathlib.Path(__file__).resolve().parents[3];scratch=pathlib.Path(a.scratch);scratch.mkdir(parents=True,exist_ok=True)
w=json.loads(pathlib.Path(a.witnesses).read_text());rows=[r for r in w['witnesses'] if r['cause'].startswith('rule emission')]
paths=[]
for i,r in enumerate(rows):
 f=scratch/('witness-'+str(i)+pathlib.Path(r['example_file']).suffix);f.write_text(r['input']);paths.append(str(f))
manifest=scratch/'manifest.txt';manifest.write_text('\n'.join(paths)+'\n')
commands=[[a.oracle,'--manifest',str(manifest),'--whole','--recovery'],['node','--disable-warning=ExperimentalWarning',str(root/'oracle/node.mjs'),str(root/'stage1/typescript/parser/main.ts'),'--manifest',str(manifest),'--whole','--recovery']]
def run(command):
 r=subprocess.run(command,cwd=root,capture_output=True,text=True,timeout=60)
 if r.returncode:raise SystemExit('tree evidence execution failed: '+r.stderr)
 result={};current=None
 for line in r.stdout.splitlines():
  if line.startswith('case '):current=int(line.split()[1]);result[current]=[]
  elif current is not None:result[current].append(line)
 return result
with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:go,node=list(pool.map(run,commands))
records=[]
for i,r in enumerate(rows):
 x,y=go[i],node[i];records.append(dict(rule=r['rule'],input=r['input'],original_example=r['example_file'],go_tree=x,node_tree=y,tree_diff=list(difflib.unified_diff(x,y,fromfile='Go',tofile='Node',lineterm=''))))
pathlib.Path(a.out).write_text(json.dumps(records,indent=2)+'\n');print('inspected',len(records),'reduced trees')
