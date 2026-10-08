#!/usr/bin/env python3
"""Regenerate the parse and lint profiles from the committed, separate corpus."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import platform
import subprocess
import sys

REPO = Path(__file__).resolve().parents[2]

def disjoint(training, benchmarks):
    train_names = [r['path'] for r in training]
    train_hashes = {r['sha256'] for r in training}
    benchmark_names = {r['path'] for r in benchmarks}
    benchmark_hashes = {r['sha256'] for r in benchmarks}
    if len(train_names) != len(set(train_names)):
        raise RuntimeError('duplicate training path')
    if set(train_names) & benchmark_names or train_hashes & benchmark_hashes:
        raise RuntimeError('training overlaps a benchmark path or file hash')

def corpus(typescript):
    training = json.loads((REPO/'stage1/profiles/training.json').read_text())
    benchmarks = json.loads((REPO/'stage1/profiles/benchmarks.json').read_text())
    disjoint(training['files'], benchmarks['files'])
    pin = subprocess.check_output(['git','-C',str(typescript),'rev-parse','HEAD'],text=True).strip()
    if pin != training['typescript_commit'] or pin != benchmarks['typescript_commit']:
        raise RuntimeError('TypeScript checkout differs from corpus pin')
    for r in training['files'] + benchmarks['files']:
        path = typescript / r['path'].removeprefix('typescript/')
        if hashlib.sha256(path.read_bytes()).hexdigest() != r['sha256']:
            raise RuntimeError('corpus bytes changed: '+r['path'])
    actual = {'typescript/'+p.relative_to(typescript).as_posix() for p in (typescript/'src/compiler').rglob('*.ts')}
    if actual != {r['path'] for r in benchmarks['files']}:
        raise RuntimeError('batch 8 benchmark manifest no longer covers all compiler files')
    return [typescript/r['path'].removeprefix('typescript/') for r in training['files']]

def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--typescript',type=Path,required=True)
    parser.add_argument('--work',type=Path,required=True)
    parser.add_argument('--llvm-profdata',required=True)
    parser.add_argument('--driver',choices=['parse','lint','both'],default='both')
    args=parser.parse_args()
    work=args.work.resolve();work.mkdir(parents=True,exist_ok=True)
    files=corpus(args.typescript.resolve())
    manifest=work/'training.txt';manifest.write_text(''.join(str(p)+'\n' for p in files))
    training_hash=hashlib.sha256((REPO/'stage1/profiles/training.json').read_bytes()).hexdigest()
    def run(command,name,env=None):
        with (work/(name+'.log')).open('wb') as log:
            subprocess.run(list(map(str,command)),cwd=REPO,env=env,stdout=log,stderr=subprocess.STDOUT,check=True)
    build=work/'adamic-stage1'
    run(['go','build','-o',build,'./cmd/adamic-stage1'],'build-tool')
    target=platform.system().lower()+'-'+{'x86_64':'amd64','aarch64':'arm64','arm64':'arm64'}[platform.machine()]
    for driver in ['parse','lint'] if args.driver=='both' else [args.driver]:
        dest=REPO/'stage1/cohere'/driver/'profiles'/target;dest.mkdir(parents=True,exist_ok=True)
        binary=work/(driver+'-generate')
        run([build,'-driver',driver,'-policy','generate','-emit',work/(driver+'.c'),'-o',binary],driver+'-generate-build')
        raw=work/(driver+'.profraw')
        env=dict(os.environ,LLVM_PROFILE_FILE=str(raw))
        # No benchmark file may appear in this argv's manifest.
        run([binary,'--manifest',manifest,'--count'],driver+'-train',env)
        text=dest/'profile.txt'
        run([args.llvm_profdata,'merge','--text',raw,'-o',text],driver+'-merge-text')
        run([build,'-driver',driver,'-write-manifest',text,'-training-hash',training_hash],driver+'-manifest')
        run([build,'-driver',driver,'-policy','profile','-o',work/(driver+'-profile')],driver+'-profile-build')
        if (work/(driver+'-profile-build.log')).read_bytes():
            raise RuntimeError('profile-built artifact must not fall back: '+driver)
        print(driver, 'profile regenerated from',len(files),'separate files at',dest.relative_to(REPO),flush=True)

if __name__=='__main__':
    main()
