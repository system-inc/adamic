"""Run every requested gate with fresh profiling artifacts and pinned inputs."""
import json, os, pathlib, subprocess, time
root=pathlib.Path(__file__).resolve().parent
repo=root.parents[4]
profile='/tmp/adamic-fixture-extensions-profile'
env=dict(os.environ, ADAMIC_TYPESCRIPT_SOURCE='/workspace/scratch/typescript-6.0.3', ADAMIC_PARSER_BENCH='1', ADAMIC_LINT_BENCH='1', ADAMIC_LINT_PROFILE_DIR=profile, ADAMIC_LINT_PROFILE_SNAPSHOTS=profile, ADAMIC_RECOVERY_ARTIFACTS='/tmp/adamic-fixture-extensions-recovery')
results={}
for label, package, selector in [('parser','./stage1/typescript/parser',None),('compiler','./stage1/cohere/lint','^TestCompilerAndStage1Agree$'),('rules','./stage1/cohere/lint','^TestRulesAgree$'),('profiles','./stage1/cohere/lint','^TestProfile(Artifacts|SnapshotsAgree)$')]:
 command=['go','test',package,'-count=1','-json','-v','-failfast=false','-timeout=90m']
 if selector: command += ['-run',selector]
 before=os.getloadavg();started=time.monotonic()
 with (root/f'{label}.jsonl').open('w') as log:
  result=subprocess.run(command,cwd=repo,env=env,stdout=log,stderr=subprocess.STDOUT)
 results[label]={'command':command,'exit':result.returncode,'wall_seconds':time.monotonic()-started,'load_before':before,'load_after':os.getloadavg()}
 (root/'results.json').write_text(json.dumps(results,indent=2)+'\n')
 print(label,results[label],flush=True)
(root/'finished.json').write_text(json.dumps({'successful':all(r['exit']==0 for r in results.values())})+'\n')
