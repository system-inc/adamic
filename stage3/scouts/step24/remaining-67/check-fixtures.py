#!/usr/bin/env python3
"""Measure headers and source Node witnesses; mutate a real initializer/read."""
import json, os, re, subprocess, sys, tempfile
from pathlib import Path
unit=Path(__file__).resolve().parent
repo=unit.parents[3]
compiler=sys.argv[1]
results=[]
for file in sorted((unit/'fixtures').glob('*.a')):
    source=file.read_text()
    source=re.sub(r'^// a-check:.*\n','',source)
    file.write_text(source)
    checked=subprocess.run([compiler,'types',str(file)],capture_output=True,text=True)
    code=re.search(r'\bTS(\d+)\b',checked.stdout+checked.stderr)
    if not code: raise RuntimeError((file.name,checked.returncode,checked.stdout,checked.stderr))
    file.write_text('// a-check: type error TS'+code[1]+'\n'+source)
    node=subprocess.run(['node','--disable-warning=ExperimentalWarning',str(repo/'oracle/node.mjs'),str(file)],capture_output=True,text=True)
    if node.returncode: raise RuntimeError(node.stderr)
    with tempfile.TemporaryDirectory(prefix='remaining67-fixture-') as scratch:
        mutant=Path(scratch)/file.name
        if 'value?()' in source and 'value: undefined' in source: changed=source.replace('value: undefined','value: () => 1',1)
        elif 'values[0]' in source: changed=source.replace('values[0]','1',1)
        elif 'value?()' in source: changed=source.replace('host.value = undefined','host.value = () => 1',1)
        else: changed=source.replace('undefined;', '1;',1) if 'return undefined;' in source else source.replace('= undefined;', '= 1;',1) if '= undefined;' in source else source.replace('value: undefined','value: 1',1)
        assert changed!=source,file.name
        mutant.write_text(changed)
        observation=subprocess.run(['node','--disable-warning=ExperimentalWarning',str(repo/'oracle/node.mjs'),str(mutant)],capture_output=True,text=True)
        if (node.returncode,node.stdout,node.stderr)==(observation.returncode,observation.stdout,observation.stderr):raise RuntimeError('surviving Node mutant: '+file.name)
    results.append(dict(file=file.name,code=int(code[1]),compilerExit=checked.returncode,diagnostic=checked.stdout+checked.stderr,nodeExit=node.returncode,stdout=node.stdout,stderr=node.stderr,mutantSource=changed,mutantStdout=observation.stdout,mutantExit=observation.returncode,caught='Node byte comparison'))
(unit/'fixture-results.json').write_text(json.dumps(results,indent=2)+'\n')
(unit/'counts.md').write_text('# Witness counts\n\n'+str(len(results))+' source witnesses; '+str(len(results))+' real-input mutants caught by Node output comparison. These are checker-error scout witnesses, not internal/oracle runtime fixtures.\n')
print(json.dumps({'witnesses':len(results),'mutantsCaught':len(results)}))
