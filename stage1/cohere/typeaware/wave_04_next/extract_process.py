#!/usr/bin/env python3
"""Extract production process-rule fixtures; retain original TypeScript module specifiers."""
from pathlib import Path
import argparse,json,re,subprocess,os
parser=argparse.ArgumentParser();parser.add_argument('directory',type=Path);args=parser.parse_args()
r=Path(__file__).resolve().parents[4];o=args.directory.resolve();o.mkdir(parents=True,exist_ok=True)
s=(r/'cohere/internal/lint/rules/nexus/correctness_no_process_exit_after_output_test.go').read_text()
token=r'"(?:[^"\\]|\\.)*"|`[^`]*`'
read=lambda body:[t[1:-1] if t.startswith('`') else json.loads(t) for t in re.findall(token,body)]
def variable(name):
 m=re.search(r'var '+name+r' = correctnessNoProcessExitAfterOutputLines\((.*?)\n\)',s,re.S)
 if m is None:raise RuntimeError('missing fixture '+name)
 return '\n'.join(read(m[1]))+'\n'
for variable_name,filename in [('NodeTypes','node.d.ts'),('WebConsole','web-console.d.ts'),('HelpModule','Help.ts'),('Script','ScriptHelp.ts')]:
 (o/filename).write_text(variable('correctnessNoProcessExitAfterOutput'+variable_name))
prefix=variable('correctnessNoProcessExitAfterOutputPrelude')
blocks=re.findall(r'\[\]string\{\s*((?:(?:'+token+r')\s*,?\s*)+)\}',s,re.S)
sources=[]
for body in blocks:
 lines=read(body)
 if any(';' in line or '{' in line or '}' in line for line in lines):sources.append('\n'.join(lines)+'\n')
# Run the authoritative dynamically assembled real-site controls, too.
export=o/'fixtures_test.go'
export.write_text('package nexus\nimport("testing";"encoding/json";"os")\nfunc TestWave04ExportProcess(t *testing.T){sources:=[]string{'+','.join('correctnessNoProcessExitAfterOutput'+name+'('+fixed+')' for name in ['Statement','Sync','CardBalances','Godword'] for fixed in ['false','true'])+',};data,err:=json.Marshal(sources);if err!=nil{t.Fatal(err)};if err=os.WriteFile(os.Getenv("WAVE04_PROCESS_EXPORT"),data,0600);err!=nil{t.Fatal(err)}}\n')
virtual=r/'cohere/internal/lint/rules/nexus/wave04_process_export_test.go';overlay=o/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(virtual):str(export)}}))
with (o/'export.log').open('wb') as log:
 subprocess.run(['go','test','-overlay',str(overlay),'./internal/lint/rules/nexus','-run','^TestWave04ExportProcess$','-count=1'],cwd=r/'cohere',env=dict(os.environ,WAVE04_PROCESS_EXPORT=str(o/'real-sites.json')),stdout=log,stderr=log,check=True)
real_sources=json.loads((o/'real-sites.json').read_text())
sources += [source.removeprefix(prefix) for source in real_sources]
sources += ["declare const optionalReceiver: { fn(value: never): void } | undefined; console.log(output); optionalReceiver?.fn(process.exit(1)); process.exit(2);\n"]
paths=[]
for at,source in enumerate(sources):
 p=o/f'process-{at:03d}.a';p.write_text(prefix+source);paths.append(p)
(o/'tsconfig.json').write_text(json.dumps({'compilerOptions':{'strict':True,'target':'ES2022','lib':['ES2022'],'moduleDetection':'auto','types':[]},'files':['node.d.ts','web-console.d.ts','ScriptHelp.ts']}))
(o/'manifest').write_text('\n'.join(str(p) for p in paths+[o/'Help.ts',o/'ScriptHelp.ts'])+'\n')
(o/'provenance.json').write_text(json.dumps(dict(authority='cohere/internal/lint/rules/nexus/correctness_no_process_exit_after_output_test.go',roots=len(paths),adaptation="original ./Help specifiers and .ts helper fixtures; source prefix and declarations copied verbatim"),indent=2)+'\n')
print(len(paths),'process source controls')
