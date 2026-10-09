import pathlib,time,json,subprocess,difflib,os
out=pathlib.Path('review/test-audit/stage1-cohere-lint-yepesta_independent_shards');deadline=time.monotonic()+120
while not any(r['name']=='S2' for r in json.loads((out/'runs.json').read_text())):
 if time.monotonic()>deadline:raise SystemExit('waiting for permitted construction checks cooked')
 time.sleep(1)
time.sleep(1)
# Import the runner definitions without reexecuting its audit loop.
ns={};src=(out/'run_audit.py').read_text();exec(src[:src.index('for group,pat in groups.items():')],ns)
run=ns['run'];groups=ns['groups'];plan=json.loads((out/'frozen-plan.json').read_text());m=plan['mutants'][3];f=pathlib.Path(m['file']);original=subprocess.check_output(['git','show','HEAD:'+str(f)],text=True)
try:
 f.write_text(original.replace(m['from'],m['to']))
 for group,pat in groups.items():run('M4-isolated-'+group,pat)
finally:f.write_text(original)
f=pathlib.Path('stage1/cohere/lint/main.ts');original=subprocess.check_output(['git','show','HEAD:'+str(f)],text=True);start=original.index('function run(');end=original.index('\nconst args = programArguments();');modified=original[:start]+'function run(row: string, countOnly: boolean): number {\n    return 0;\n}\n'+original[end:]
(out/'P1-invalid.rejected.txt').write_text((out/'P1.patch').read_text());(out/'P1.patch').write_text(''.join(difflib.unified_diff(original.splitlines(True),modified.splitlines(True),fromfile='a/'+str(f),tofile='b/'+str(f))))
try:
 f.write_text(modified);r=run('P1-valid')
 if r['cooked']:r=run('P1-valid-warm-lowered')
 if r['cooked']:
  for group,pat in groups.items():run('P1-isolated-'+group,pat)
finally:f.write_text(original)
plan['probe']['change']='replace entire run body with return 0; initial early-return insertion failed TypeScript unreachable narrowing diagnostics and is excluded';(out/'frozen-plan.json').write_text(json.dumps(plan,indent=2)+'\n')
# Check each replay patch against an isolated index at the starting commit.
index='/tmp/u113-patch-index';env=dict(os.environ,GIT_INDEX_FILE=index)
with (out/'apply-check.log').open('w') as log:
 for patch in sorted(out.glob('*.patch')):
  subprocess.run(['git','read-tree','HEAD'],env=env,stdout=log,stderr=log,check=True)
  log.write(str(patch)+'\n');log.flush();subprocess.run(['git','apply','--cached','--check',str(patch)],env=env,stdout=log,stderr=log,check=True)
