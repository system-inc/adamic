import os,sys,pathlib,difflib,subprocess,json
from importlib.machinery import SourceFileLoader
r=SourceFileLoader('audit','/workspace/u107-run.py').load_module();p=r.p;root=r.root
# The first entry probe left unreachable statements, losing TypeScript's narrowing. Replace the whole body instead.
f=root/'stage1/cohere/lint/main.ts';before=f.read_text();a=before.index('function run(');b=before.index('\n\nconst args = programArguments();');after=before[:a]+'function run(row: string, countOnly: boolean): number {\n    return 0;\n}'+before[b:]
(p/'P1.diff').rename(p/'P1-invalid-probe.diff.txt');(p/'P1B.diff').write_text(''.join(difflib.unified_diff(before.splitlines(True),after.splitlines(True),fromfile='a/stage1/cohere/lint/main.ts',tofile='b/stage1/cohere/lint/main.ts')))
try:
 f.write_text(after);env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/workspace/u107-cache/P1B';r.run('P1B',json.loads((p/'M1-time.json').read_text())['command'].split(' -run ',1)[1],env)
finally:f.write_text(before)
f=root/'internal/testguard/guard.go';before=f.read_text();a=before.index('func Run(');after=before[:a]+'func Run(command *exec.Cmd, budget, ceiling time.Duration) error { return nil }\n';after=after.replace('\n\t"strings"','')
(p/'P2.diff').write_text(''.join(difflib.unified_diff(before.splitlines(True),after.splitlines(True),fromfile='a/internal/testguard/guard.go',tofile='b/internal/testguard/guard.go')))
try:
 f.write_text(after)
 with (p/'P2-vet.log').open('w') as log:v=subprocess.run(['go','vet','./internal/testguard/'],cwd=root,stdout=log,stderr=subprocess.STDOUT,timeout=90)
 assert v.returncode==0
 r.run('P2','^TestChildCPUWaitGuard$')
finally:f.write_text(before)
# The emitted mismatch row gets one individual clean run. Do not audit a witness that cannot finish its own preconditions.
d=r.run('emitted-individual-baseline','^TestEmittedJavaScriptMismatch_000$')
if d['exit']==0:
 for i in range(2):r.run('emitted-timing-'+str(i),'^TestEmittedJavaScriptMismatch_000$')
 f=root/'stage1/cohere/lint/lint_test.go';before=f.read_text();old='if diff := difference(side.run.output, want.output); diff != "" {';assert before.count(old)==1;after=before.replace(old,'if diff := difference(side.run.output, want.output); diff != "" && side.name != "emitted JavaScript" {')
 (p/'W5.diff').write_text(''.join(difflib.unified_diff(before.splitlines(True),after.splitlines(True),fromfile='a/stage1/cohere/lint/lint_test.go',tofile='b/stage1/cohere/lint/lint_test.go')))
 try:
  f.write_text(after)
  with (p/'W5-vet.log').open('w') as log:v=subprocess.run(['go','vet','./stage1/cohere/lint/'],cwd=root,stdout=log,stderr=subprocess.STDOUT,timeout=90)
  assert v.returncode==0
  r.run('W5','^TestEmittedJavaScriptMismatch_000$')
 finally:f.write_text(before)
