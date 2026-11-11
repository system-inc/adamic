from pathlib import Path
import subprocess,json,hashlib,shutil
S=Path(__file__).resolve().parent;project=S/'small-project';project.mkdir(exist_ok=True)
(project/'main.ts').write_text('export type Choice = { kind: "left"; value: string } | { kind: "right"; value: number };\nexport type Selected<T> = T extends { value: infer V } ? V : never;\nexport const pair: [Choice, boolean] = [{ kind: "left", value: "trace" }, true];\nexport function map<T, U>(xs: readonly T[], f: (x: T) => U): U[] { return xs.map(f); }\nexport const result = map([1, 2], x => ({ value: x, label: String(x) }));\n')
rows=[];written=[];cli=S/'trace-cli'
if cli.exists():shutil.rmtree(cli)
shutil.copytree('/workspace/scratch/dp24r5m-base/built/local',cli)
for side in ['base','after']:
 tree=Path('/workspace/scratch/dp24r5m-'+side);trace=project/'trace'
 if side=='after':
  for f in (tree/'built/local').rglob('*.js'):shutil.copyfile(f,cli/f.relative_to(tree/'built/local'))
 cmd=['node',cli/'tsc.js']
 cmd=[str(x) for x in cmd]+['--ignoreConfig','--strict','--noEmit','--ignoreDeprecations','6.0','--generateTrace',str(trace),str(project/'main.ts')]
 with (S/(side+'-trace-witness.log')).open('w') as log:p=subprocess.run(cmd,cwd=project,stdout=log,stderr=subprocess.STDOUT)
 row=dict(side=side,sourceTree=str(tree),cwd=str(project),command=cmd,exit=p.returncode);rows.append(row);print(json.dumps(row),flush=True);assert p.returncode==0
 data=(trace/'types.json').read_bytes();written.append(data);(S/(side+'-written-types.json')).write_bytes(data);row.update(bytes=len(data),sha256=hashlib.sha256(data).hexdigest())
assert written[0]==written[1]
(S/'trace-witness-results.json').write_text(json.dumps(dict(equal=True,runs=rows),indent=2)+'\n')
