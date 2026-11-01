#!/usr/bin/env python3
"""Run the repository's pinned native compiler on both standalone witnesses."""
import json, pathlib, subprocess, sys, tempfile
binary=str(pathlib.Path(sys.argv[1]).resolve())
results=[]
for text in ['export class C extends f() extends {}','class C extends 0 extends {}']:
    with tempfile.TemporaryDirectory(prefix='99e-native-') as directory:
        root=pathlib.Path(directory)
        (root/'input.ts').write_text(text+'\n')
        run=subprocess.run([binary,'--declaration','input.ts'],cwd=root,capture_output=True,text=True,timeout=60)
        assert run.returncode==2,run
        assert 'TS1172' in run.stdout,run.stdout
        output=(root/'input.d.ts').read_text()
        results.append(dict(input=text,exit=run.returncode,stdout=run.stdout,stderr=run.stderr,declaration=output))
assert results[0]['declaration']=='declare const C_base: any;\nexport declare class C extends C_base {\n}\nexport {};\n'
print(json.dumps(results,indent=2))
