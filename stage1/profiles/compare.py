#!/usr/bin/env python3
"""Build the five named parse candidates, including Go's automatic default.pgo."""
import argparse
import gzip
import hashlib
import json
import os
from pathlib import Path
import subprocess
import shutil
import sys
sys.dont_write_bytecode=True
import regenerate

REPO=regenerate.REPO

def require_output(stdout, stderr, expected, name):
    if stdout.read_bytes()!=expected or stderr.read_bytes():
        raise RuntimeError('MISCOMPILE: output differs: '+name)


def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--typescript',type=Path,required=True)
    parser.add_argument('--work',type=Path,required=True)
    args=parser.parse_args();work=args.work.resolve();work.mkdir(parents=True,exist_ok=True)
    os.environ['CGO_ENABLED']='0'
    os.environ['GOMAXPROCS']='1'
    train=regenerate.corpus(args.typescript.resolve())
    benchmark=json.loads((REPO/'stage1/profiles/benchmarks.json').read_text())
    (work/'compiler.txt').write_text(''.join(str(args.typescript.resolve()/r['path'].removeprefix('typescript/'))+'\n' for r in benchmark['files']))
    (work/'training.txt').write_text(''.join(str(p)+'\n' for p in train))
    commands=[]
    def run(command,name,cwd=REPO):
        commands.append(list(map(str,command)))
        (work/'commands.json').write_text(json.dumps(commands,indent=2)+'\n')
        with (work/(name+'.stdout')).open('wb') as out,(work/(name+'.stderr')).open('wb') as err:
            subprocess.run(list(map(str,command)),cwd=cwd,stdout=out,stderr=err,check=True)
    build=work/'adamic-stage1';run(['go','build','-o',build,'./cmd/adamic-stage1'],'build-tool')
    for policy in ['o2','thin','profile']:
        run([build,'-driver','parse','-policy',policy,'-emit',work/(policy+'.c'),'-o',work/policy],policy+'-build')
        if (work/(policy+'-build.stderr')).read_bytes():raise RuntimeError('candidate build fell back or warned: '+policy)
    assert (work/'o2.c').read_bytes()==(work/'thin.c').read_bytes()==(work/'profile.c').read_bytes()
    root=REPO/'cohere';virtual=root/'adamic_parse.go'
    overlay=work/'go-overlay.json';overlay.write_text(json.dumps({'Replace':{str(virtual):str(REPO/'stage1/cohere/parse/testdata/go_parse.go.txt')}})+'\n')
    run(['go','build','-trimpath','-ldflags=-s -w','-pgo=off','-overlay='+str(overlay),'-o',work/'go-plain',virtual],'go-plain-build',root)
    run(['taskset','-c','3',work/'go-plain','--manifest',work/'training.txt','--count','--cpu-profile',work/'default.pgo','--repeat','200'],'go-training',root)
    if (work/'go-training.stdout').read_bytes()!=b'0\n' or (work/'go-training.stderr').read_bytes():raise RuntimeError('Go training output differs')
    # Go's automatic profile lookup uses a real file, not an overlay-only path.
    default = root/'default.pgo'
    if default.exists():raise RuntimeError('refusing to replace an existing cohere default.pgo')
    try:
        shutil.copyfile(work/'default.pgo',default)
        run(['go','build','-trimpath','-ldflags=-s -w','-pgo=auto','-overlay='+str(overlay),'-o',work/'go-profile',virtual],'go-profile-build',root)
        run(['go','version','-m',work/'go-profile'],'go-profile-settings',root)
        settings=(work/'go-profile-settings.stdout').read_text()
        if '-pgo=' not in settings:raise RuntimeError('default.pgo was not applied by Go')
    finally:
        default.unlink(missing_ok=True)
    run(['go','tool','pprof','-top',work/'go-plain',work/'default.pgo'],'go-profile-top',root)
    size_tool=Path(shutil.which('clang')).resolve().parent/'llvm-size'
    records={}
    for name in ['o2','thin','profile','go-plain','go-profile']:
        run(['taskset','-c','3',work/name,'--manifest',work/'compiler.txt','--count'],name+'-preflight')
        require_output(work/(name+'-preflight.stdout'),work/(name+'-preflight.stderr'),b'0\n',name+' count')
        run([size_tool,'--format=sysv',work/name],name+'-sections')
        sections=(work/(name+'-sections.stdout')).read_text().splitlines()
        text=sum(int(row.split()[1]) for row in sections if row.split() and row.split()[0]=='.text')
        if text<=0:raise RuntimeError('missing executable .text section: '+name)
        records[name]={'binary_sha256':hashlib.sha256((work/name).read_bytes()).hexdigest(),'bytes':(work/name).stat().st_size,'text_bytes':text}
    # The exact three native timing binaries also emit the AST oracle bytes.
    tsc=root/'TypeScript/tsc';ast_virtual=tsc/'adamic_parser_oracle.go';ast_overlay=work/'ast-overlay.json'
    ast_overlay.write_text(json.dumps({'Replace':{str(ast_virtual):str(REPO/'stage1/typescript/parser/testdata/oracle.go')}})+'\n')
    run(['go','build','-pgo=off','-overlay='+str(ast_overlay),'-o',work/'ast-go',ast_virtual],'ast-go-build',tsc)
    run([work/'ast-go','--manifest',work/'compiler.txt','--whole'],'ast-go')
    expected=(work/'ast-go.stdout').read_bytes()
    for name in ['o2','thin','profile','go-plain','go-profile']:
        run([work/name,'--manifest',work/'compiler.txt','--ast'],name+'-ast')
        require_output(work/(name+'-ast.stdout'),work/(name+'-ast.stderr'),expected,name+' benchmark AST')
    records['ast']={'bytes':len(expected),'sha256':hashlib.sha256(expected).hexdigest()}
    # Correctness also covers the actual training inputs; they remain untimed.
    run([work/'ast-go','--manifest',work/'training.txt','--whole'],'training-ast-go')
    training_expected=(work/'training-ast-go.stdout').read_bytes()
    for name in ['o2','thin','profile','go-plain','go-profile']:
        run([work/name,'--manifest',work/'training.txt','--ast'],name+'-training-ast')
        require_output(work/(name+'-training-ast.stdout'),work/(name+'-training-ast.stderr'),training_expected,name+' training AST')
    records['training_ast']={'bytes':len(training_expected),'sha256':hashlib.sha256(training_expected).hexdigest()}
    records['source_sha256']=hashlib.sha256((work/'profile.c').read_bytes()).hexdigest()
    records['typescript_root']=str(args.typescript.resolve())
    records['benchmark_manifest_sha256']=hashlib.sha256((work/'compiler.txt').read_bytes()).hexdigest()
    records['training_manifest_sha256']=hashlib.sha256((work/'training.txt').read_bytes()).hexdigest()
    (work/'builds.json').write_text(json.dumps(records,indent=2)+'\n')
    print('all five exact timed binaries match count and AST outputs against the independent Go oracle',flush=True)
    print(json.dumps(records,indent=2),flush=True)

if __name__=='__main__':main()
