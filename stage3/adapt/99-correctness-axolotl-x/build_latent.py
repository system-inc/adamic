#!/usr/bin/env python3
"""Build main's measurement-only overlay with the provided scheduling selector."""
import difflib
import hashlib
import json
from pathlib import Path
import subprocess
import sys
repo, output = map(lambda s: Path(s).resolve(), sys.argv[1:])
subprocess.run(['python3', str(repo/'stage3/census/latent/make_overlay.py'), str(repo), str(output)], cwd=repo, check=True)
hook=output/'internal_lower_latent_hook.go'
original=hook.read_text(); offset=original.index('func latentRun('); prefix,body=original[:offset],original[offset:]
for needle,addition in [('\tfor _, file := range program.Files() {\n','\t\tif !anyMembersCensusSelected(program.FileName(file), "") { continue }\n'),('\t\tfor _, node := range file.Statements.Nodes {\n','\t\t\tif !anyMembersCensusSelected(program.FileName(file), program.Where(node)) { continue }\n')]:
    assert body.count(needle)==1
    body=body.replace(needle,needle+addition,1)
hook.write_text(prefix+body)
(output/'selector.patch').write_text(''.join(difflib.unified_diff(original.splitlines(True),(prefix+body).splitlines(True),fromfile='original-hook',tofile='selected-hook')))
mapping=json.loads((output/'overlay.json').read_text())
selector=Path(__file__).resolve().parent/'latent-selector.go.txt'
mapping['Replace'][str(repo/'internal/lower/any_members_census_piece.go')]=str(selector)
for key,value in list(mapping['Replace'].items()):
    file=Path(value)
    if file.suffix=='.go':
        renamed=file.with_suffix('.go.txt'); file.rename(renamed); mapping['Replace'][key]=str(renamed)
(output/'overlay.json').write_text(json.dumps(mapping,indent=2)+'\n')
command=['go','build','-buildvcs=false','-overlay='+str(output/'overlay.json'),'-o',str(output/'latent'),'./stage3/census/latent/tool']
subprocess.run(command,cwd=repo,check=True)
(output/'provenance.json').write_text(json.dumps(dict(command=command,compiler_commit=subprocess.check_output(['git','rev-parse','HEAD'],cwd=repo,text=True).strip(),binary_sha256=hashlib.sha256((output/'latent').read_bytes()).hexdigest(),selector_sha256=hashlib.sha256(selector.read_bytes()).hexdigest()),indent=2)+'\n')
