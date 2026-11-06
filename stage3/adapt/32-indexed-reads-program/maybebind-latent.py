#!/usr/bin/env python3
"""Measure CW051 on all four complete source configurations, retaining first stops."""
import json
import os
from pathlib import Path
import subprocess
import sys
repo,before,after,extra,output=map(lambda s:Path(s).resolve(),sys.argv[1:])
output.mkdir(parents=True,exist_ok=False)
node_env=dict(os.environ,NODE_PATH=str(repo/'stage3/api/node_modules'))
selection_script="""const fs=require('fs'),ts=require('typescript');const f=process.argv[1],s=ts.createSourceFile(f,fs.readFileSync(f,'utf8'),ts.ScriptTarget.Latest,true);const names=['performBuild','createWatchOfConfigFile','createWatchOfFilesAndCompilerOptions'];console.log(JSON.stringify(names.map(name=>{const a=s.statements.filter(n=>ts.isFunctionDeclaration(n)&&n.name?.text===name);if(a.length!==1)throw Error(name);const p=s.getLineAndCharacterOfPosition(a[0].getStart(s));return {name,where:'src/compiler/executeCommandLine.ts:'+(p.line+1)+':'+(p.character+1)};})));"""
rows=[]
for label,tree in [('main',before),('main-plus-32',after),('main-plus-99x',extra/'before'),('main-plus-32-plus-99x',extra/'after')]:
    if not tree.is_dir():
        rows.append(dict(configuration=label,status='blocked',reason='unchanged 99x refuses handed-off maybeBind site inventory drift after the twelve rewrites'))
        continue
    destination=output/label;destination.mkdir()
    units=json.loads(subprocess.check_output(['node','-e',selection_script,str(tree/'src/compiler/executeCommandLine.ts')],env=node_env,text=True))
    selected=dict(schema=1,files=[dict(file='src/compiler/executeCommandLine.ts',all_units=False,units=[u['where'] for u in units])])
    selection=destination/'selection.json';selection.write_text(json.dumps(selected,indent=2)+'\n')
    env=dict(os.environ,LATENT_FULL='0',LATENT_ASSERT_NO_OUTPUT='1',ANY_MEMBERS_CENSUS_PIECE=str(selection),ANY_MEMBERS_CENSUS_ROOT=str(tree),LATENT_TIMES=str(destination/'times.jsonl'))
    command=['/workspace/scratch/32-latent-build/latent',str(tree/'src/compiler'),str(destination/'records.jsonl')]
    with (destination/'process.log').open('w') as log:r=subprocess.run(command,cwd=repo,env=env,stdout=log,stderr=subprocess.STDOUT)
    assert r.returncode==0,label
    records=[json.loads(s) for s in (destination/'records.jsonl').read_text().splitlines()]
    rows.append(dict(configuration=label,tree=str(tree),units=units,command=command,exit=r.returncode,records=records))
(output/'comparison.json').write_text(json.dumps(rows,indent=2)+'\n')
print(json.dumps(rows,indent=2))
