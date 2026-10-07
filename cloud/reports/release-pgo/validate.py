#!/usr/bin/env python3
"""Hold the parse driver and full parser AST witness to the external Go oracle."""
import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys

s = Path(sys.argv[1]).resolve()
lto = Path(sys.argv[2]).resolve()
old = Path(sys.argv[3]).resolve()
builds = json.loads((s/'builds.json').read_text())
compiler = '/workspace/adamic-tools/llvm/bin/clang'
log = (s/'validation-commands.jsonl').open('w')
def run(args, stem, env=None):
    log.write(json.dumps(list(map(str,args)))+'\n');log.flush()
    with (s/(stem+'.stdout')).open('wb') as out,(s/(stem+'.stderr')).open('wb') as err:
        subprocess.run(list(map(str,args)),stdout=out,stderr=err,env=env,check=True)

def ast_build(mode, source, binary):
    flags=[('-fprofile-use='+str(s/'ast-training.profdata')) if f.startswith('-fprofile-use=') else f for f in builds[mode]['link_flags']]
    archive=s/mode/'runtime.a'
    if mode in ['pgo','split']:
        directory=s/('ast-runtime-'+mode)
        archive=directory/'runtime.a'
        if not archive.exists():
            directory.mkdir(exist_ok=True)
            objects=[]
            for unit in sorted((s/'runtime').glob('*.c')):
                obj=directory/(unit.stem+'.o')
                run([compiler,*[f for f in flags if f!='-fuse-ld=lld'],'-c',unit,'-o',obj],mode+'-ast-runtime-'+unit.stem)
                objects.append(obj)
            run(['/workspace/adamic-tools/llvm/bin/llvm-ar','rcs',archive,*objects],mode+'-ast-runtime-archive')
    run([compiler,*flags,'-I',s/'runtime','-o',binary,source,'-Xlinker','--whole-archive',archive,'-Xlinker','--no-whole-archive','-lm'],mode+'-'+binary.name+'-build')
# The AST witness is a different generated entrypoint with different FE profile keys.
# Train its own main C on the same 38 files; reuse the parse experiment's runtime archives.
generation=builds['generate']['link_flags']
run([compiler,*generation,'-I',s/'runtime','-o',s/'generate/ast',s/'ast.c','-Xlinker','--whole-archive',s/'generate/runtime.a','-Xlinker','--no-whole-archive','-lm'],'ast-generate-build')
run(['taskset','-c','3',s/'generate/ast','--manifest',s/'train.txt','--whole'],'ast-training-run',dict(os.environ,LLVM_PROFILE_FILE=str(s/'ast-training.profraw')))
run(['/workspace/adamic-tools/llvm/bin/llvm-profdata','merge',s/'ast-training.profraw','-o',s/'ast-training.profdata'],'ast-profile-merge')
run(['taskset','-c','3',old/'go-parse','--manifest',s/'held-out.txt','--count'],'go-held-count')
if (s/'go-held-count.stdout').read_bytes()!=b'0\n' or (s/'go-held-count.stderr').read_bytes():
    raise RuntimeError('Go count differs from all four variants')
run(['taskset','-c','3',lto/'ast-go','--manifest',s/'held-out.txt','--whole'],'go-held-ast')
want=(s/'go-held-ast.stdout').read_bytes()
truth=dict(bytes=len(want),sha256=hashlib.sha256(want).hexdigest())
for mode in ['o2','thin','pgo','split']:
    binary=s/mode/'ast'
    ast_build(mode,s/'ast.c',binary)
    run(['taskset','-c','3',binary,'--manifest',s/'held-out.txt','--whole'],mode+'-held-ast')
    got=(s/(mode+'-held-ast.stdout')).read_bytes()
    if got!=want or (s/(mode+'-held-ast.stderr')).read_bytes():
        raise RuntimeError('MISCOMPILE: full held-out AST differs from Go: '+mode)
    print(mode,'full AST Go byte parity PASS',json.dumps(truth),flush=True)
# A real same-length output mutant must be caught by this exact comparison.
source=(s/'ast.c').read_text()
if 'SourceFile' not in source:raise RuntimeError('AST mutant literal missing')
mutant_directory=s/'mutant-source';mutant_directory.mkdir(exist_ok=True)
mutant=mutant_directory/'ast.c';mutant.write_text(source.replace('SourceFile','XourceFile',1))
mutant_binary=s/'ast-mutant'
ast_build('pgo',mutant,mutant_binary)
run(['taskset','-c','3',mutant_binary,'--manifest',s/'held-out.txt','--whole'],'ast-mutant')
got=(s/'ast-mutant.stdout').read_bytes()
if got==want:raise RuntimeError('byte-comparison mutant survived')
truth['mutant_first_difference']=next(i for i,(a,b) in enumerate(zip(got,want)) if a!=b)
truth['mutant_bytes']=len(got)
(s/'validation.json').write_text(json.dumps(truth,indent=2)+'\n')
print('compiled AST output mutant caught at byte',truth['mutant_first_difference'],flush=True)
log.close()
