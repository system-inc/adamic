#!/usr/bin/env python3
"""Build all parse units with a consistent profile from a different generated program."""
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys

s=Path(sys.argv[1]).resolve()
builds=json.loads((s/'builds.json').read_text())
compiler='/workspace/adamic-tools/llvm/bin/clang'
archiver='/workspace/adamic-tools/llvm/bin/llvm-ar'
profdata='/workspace/adamic-tools/llvm/bin/llvm-profdata'
source_directory=s/'mismatched-source';source_directory.mkdir(exist_ok=True)
# Keep the FE file key parse.c, but use the full-AST program's different main/IDs/CFGs.
foreign=source_directory/'parse.c';shutil.copyfile(s/'ast.c',foreign)
commands=(s/'stale-commands.jsonl').open('w')
def run(args,stem,env=None,check=True):
    commands.write(json.dumps(list(map(str,args)))+'\n');commands.flush()
    with (s/(stem+'.stdout')).open('wb') as out,(s/(stem+'.stderr')).open('wb') as err:
        r=subprocess.run(list(map(str,args)),stdout=out,stderr=err,env=env)
    if check and r.returncode:raise RuntimeError('command failed: '+stem)
    return r.returncode
run([compiler,*builds['generate']['link_flags'],'-I',s/'runtime','-o',s/'generate/mismatched',foreign,'-Xlinker','--whole-archive',s/'generate/runtime.a','-Xlinker','--no-whole-archive','-lm'],'stale-generate-build')
run(['taskset','-c','3',s/'generate/mismatched','--manifest',s/'train.txt','--whole'],'stale-training-run',dict(os.environ,LLVM_PROFILE_FILE=str(s/'mismatched.profraw')))
run([profdata,'merge',s/'mismatched.profraw','-o',s/'mismatched.profdata'],'stale-merge')
run([profdata,'show','--all-functions','--counts',s/'mismatched.profdata'],'stale-profile-functions')
flags=[('-fprofile-use='+str(s/'mismatched.profdata')) if f.startswith('-fprofile-use=') else f for f in builds['split']['compile_flags']]
directory=s/'stale';directory.mkdir(exist_ok=True)
objects=[]
for unit in sorted((s/'runtime').glob('*.c')):
    obj=directory/(unit.stem+'.o')
    run([compiler,*flags,'-c',unit,'-o',obj],'stale-runtime-'+unit.stem)
    objects.append(obj)
archive=directory/'runtime.a'
run([archiver,'rcs',archive,*objects],'stale-archive')
run([compiler,*flags,'-fuse-ld=lld','-I',s/'runtime','-o',directory/'parse',s/'parse.c','-Xlinker','--whole-archive',archive,'-Xlinker','--no-whole-archive','-lm'],'stale-parse-build')
run(['taskset','-c','3',directory/'parse','--manifest',s/'held-out.txt','--count'],'stale-held-count')
if (s/'stale-held-count.stdout').read_bytes()!=(s/'go-held-count.stdout').read_bytes() or (s/'stale-held-count.stderr').read_bytes():
    raise RuntimeError('MISCOMPILE: different-source profile changed held-out parse output')
# The same foreign profile also validates full AST output from the original ast.c.
# Only this diagnostic-only file-name mismatch needs the additional unprofiled warning suppression.
run([compiler,*flags,'-Wno-profile-instr-unprofiled','-fuse-ld=lld','-I',s/'runtime','-o',directory/'ast',s/'ast.c','-Xlinker','--whole-archive',archive,'-Xlinker','--no-whole-archive','-lm'],'stale-ast-build')
run(['taskset','-c','3',directory/'ast','--manifest',s/'held-out.txt','--whole'],'stale-held-ast')
want=(s/'go-held-ast.stdout').read_bytes();got=(s/'stale-held-ast.stdout').read_bytes()
if got!=want or (s/'stale-held-ast.stderr').read_bytes():
    raise RuntimeError('MISCOMPILE: different-source profile changed full AST bytes')
# Expose incompatibility diagnostics without changing the accepted binary's policy.
probe=[f for f in flags if f not in ['-Werror','-Wno-profile-instr-out-of-date']]
code=run([compiler,*probe,'-I',s/'runtime','-c',s/'parse.c','-o',directory/'diagnostic.o'],'stale-diagnostic',check=False)
if code!=0 or 'out of date' not in (s/'stale-diagnostic.stderr').read_text():
    raise RuntimeError('did not expose the stale profile diagnostic')
result=dict(training_files=38,held_out_files=39,parse_source_sha256=hashlib.sha256((s/'parse.c').read_bytes()).hexdigest(),foreign_source_sha256=hashlib.sha256(foreign.read_bytes()).hexdigest(),good_profile_sha256=hashlib.sha256((s/'training.profdata').read_bytes()).hexdigest(),foreign_profile_sha256=hashlib.sha256((s/'mismatched.profdata').read_bytes()).hexdigest(),held_ast_bytes=len(got),held_ast_sha256=hashlib.sha256(got).hexdigest(),parse_output='0\n',flags=flags)
(s/'stale.json').write_text(json.dumps(result,indent=2)+'\n')
print('different-source profile: held parse and full AST Go byte parity PASS',json.dumps(result),flush=True)
commands.close()
