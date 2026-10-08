#!/usr/bin/env python3
"""Compare Go-captured helper observations and compiling output mutants on all backends."""
import hashlib
import json
from pathlib import Path
import re
import subprocess
import tempfile

owned = Path(__file__).resolve().parent
root = owned.parents[4]
helpers = owned.parent
metadata = json.loads((owned / 'testdata/provenance.json').read_text())
pin = subprocess.check_output(['git', '-C', str(root/'cohere'), 'rev-parse', 'HEAD'], text=True).strip()
assert pin == metadata['cohere'], 'cohere pin drift'
for file, digest in metadata['sha256'].items():
    assert hashlib.sha256((root/'cohere'/file).read_bytes()).hexdigest() == digest, file + ' drift'
rows = json.loads((owned/'testdata/cases.json').read_text())
assert len(rows) == 1540
coverage = json.loads((owned/'testdata/coverage.json').read_text())
assert sum(r['invocations'] for r in coverage.values()) == len(rows)

def written(text):
    units = text.encode('utf-16-le', errors='surrogatepass')
    result = ''
    for offset in range(0, len(units), 2):
        code = int.from_bytes(units[offset:offset+2], 'little')
        result += chr(code) if 32 <= code <= 126 and code != 92 else '\\u' + format(code, '04x')
    return result

want = ''.join((str(r['wantStart'])+' '+str(r['wantEnd']) if r['kind']=='range' else
                written(r['want']) if r['kind']=='name' else str(r['want']).lower())+'\n' for r in rows).encode()

with tempfile.TemporaryDirectory(prefix='typescript-helper-verify-') as temporary:
    scratch = Path(temporary)
    def run(command):
        result = subprocess.run(command, cwd=root, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        assert result.returncode == 0 and not result.stderr, repr(command)+'\n'+result.stderr.decode(errors='replace')
        return result.stdout
    compiler = scratch/'adamic'
    run(['go','build','-o',str(compiler),'./cmd/adamic'])
    corpus = str(owned/'testdata/cases.json')
    runner = str(root/'oracle/node.mjs')
    def build(entry, directory):
        native = directory/'native'
        run([str(compiler),'build',str(entry),'-o',str(native),'--sanitize'])
        emitted = directory/'emitted.mjs'
        emitted.write_bytes(run([str(compiler),'js',str(entry)]))
        return {'Node':['node','--disable-warning=ExperimentalWarning',runner,str(entry),corpus],
                'emitted JavaScript':['node','--disable-warning=ExperimentalWarning',runner,str(emitted),corpus],
                'sanitized native':[str(native),corpus]}
    baseline = scratch/'baseline';baseline.mkdir()
    for side, command in build(owned/'main.a', baseline).items():
        got = run(command)
        assert got == want, 'Go mismatch on '+side
        print('Go agreement on '+side+': '+str(len(rows))+' helper observations, '+str(len(got))+' bytes', flush=True)
    changes = [('sourcename/treated_as.a', "stem !== '' && !stem.endsWith('/')", "stem !== ''"),
               ('typescript/is_typescript_source_file.a', "fileName.endsWith('.cts')", 'false'),
               ('typescript/non_null_assertion_operator_range.a', 'end - 1, end', 'end - 2, end')]
    files = [owned/'main.a',owned/'is_typescript_source_file.a',owned/'non_null_assertion_operator_range.a',helpers/'sourcename/treated_as.a']
    for number,(file,old,new) in enumerate(changes):
        directory = scratch/('mutant-'+str(number));directory.mkdir()
        copies = {f:directory/f.relative_to(helpers) for f in files}
        for original, target in copies.items():
            text = original.read_text()
            if original == helpers/file:
                assert text.count(old)==1, 'mutant anchor'
                text = text.replace(old,new,1)
            def rewrite(match):
                path=match.group(1)
                if path=='adamic':return match.group(0)
                resolved=(original.parent/path).resolve()
                return "from '"+str(copies.get(resolved,resolved))+"'"
            text=re.sub(r"from '([^']+)'",rewrite,text)
            target.parent.mkdir(parents=True,exist_ok=True);target.write_text(text)
        for side,command in build(copies[owned/'main.a'],directory).items():
            got=run(command)
            assert got != want, file+' mutant survived on '+side
            first=next(i for i,(a,b) in enumerate(zip(got.splitlines(),want.splitlines()),1) if a!=b)
            print(file+' semantic mutant caught on '+side+' at output line '+str(first),flush=True)
