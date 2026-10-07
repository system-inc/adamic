#!/usr/bin/env python3
"""Prepare a scratch compatibility proposal; never change shared repository code."""
import argparse,json,os,subprocess
from pathlib import Path
p=Path(__file__).resolve().parent
repo=p.parents[4]
a=argparse.ArgumentParser()
a.add_argument('--scratch',type=Path,required=True)
a.add_argument('--typescript',type=Path)
a.add_argument('--run',default='^TestContinuationCorpus$')
args=a.parse_args()
if args.typescript:
 observed = subprocess.check_output(['git', '-C', str(args.typescript), 'rev-parse', 'HEAD'], text=True).strip()
 if observed != '050880ce59e30b356b686bd3144efe24f875ebc8':
  raise SystemExit('TypeScript compiler corpus pin differs')
s=args.scratch.resolve();s.mkdir(parents=True,exist_ok=True)
paths=['stage1/cohere/lint/registry/registry.go','stage1/cohere/lint/lint_test.go']
for path in paths:
 t=s/path;t.parent.mkdir(parents=True,exist_ok=True);t.write_text((repo/path).read_text())
with (s/'patch.log').open('w') as log:
 subprocess.run(['patch','--batch','-p1','-d',str(s),'-i',str(p/'compatibility.patch')],stdout=log,stderr=subprocess.STDOUT,check=True)
# Preserve upstream suffixes. Tailwind's file gate cannot be compared by renaming
# every captured JavaScript fixture to TypeScript.
t=s/paths[1];text=t.read_text()
for before, after in [
 ('Rule, Source, Outcome, FixedSource string', 'Rule, File, Source, Outcome, FixedSource string'),
 ('row.Rule, row.Options, row.Source)', 'row.Rule+"\\t"+row.File, row.Options, row.Source)'),
 ('fmt.Sprintf("case-%03d.ts", i)', 'fmt.Sprintf("case-%03d%s", i, filepath.Ext(row.File))'),
]:
 if text.count(before) != 1:
  raise SystemExit('capture compatibility anchor changed: '+before)
 text = text.replace(before, after, 1)
t.write_text(text)
replacements={str(repo/path):str(s/path) for path in paths}
replacements[str(repo/'stage1/cohere/lint/continuation_overlay_test.go')]=str(p/'validation_test.go.txt')
overlay=s/'overlay.json';overlay.write_text(json.dumps({'Replace':replacements}))
env=os.environ.copy();env['ADAMIC_GATE_UNCACHED']='1'
if args.typescript:env['ADAMIC_TYPESCRIPT_SOURCE']=str(args.typescript.resolve())
cmd=['go','test','-overlay='+str(overlay),'./stage1/cohere/lint','-count=1','-v','-timeout=20m','-run',args.run]
print(' '.join(cmd),flush=True)
with (s/'test.log').open('w') as log:
 result=subprocess.run(cmd,cwd=repo,env=env,stdout=log,stderr=subprocess.STDOUT)
print('exit='+str(result.returncode)+' log='+str(s/'test.log'),flush=True)
raise SystemExit(result.returncode)
