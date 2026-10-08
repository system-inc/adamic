"""Focused source Node, backend Node, counted/native/sanitized and mutant checks."""
from pathlib import Path
import json
import os
import subprocess
import tempfile

here = Path(__file__).resolve().parent
repo = here.parents[2]
compiler = os.environ.get('STEP24_COMPILER', '/tmp/step24-adamic')
results=[]
def run(command, stem):
    p=subprocess.run(command, cwd=repo, capture_output=True)
    (here/'evidence'/f'{stem}.stdout').write_bytes(p.stdout)
    (here/'evidence'/f'{stem}.stderr').write_bytes(p.stderr)
    return p

def header(source, diagnostic):
    first=source.splitlines()[0]
    return first.startswith('// a-check: refused ') and first[20:] in diagnostic

with tempfile.TemporaryDirectory(prefix='step24-fixtures-') as tmp:
    scratch=Path(tmp)
    for name in ['context','speculation','factory-cast','recursion']:
        source=here/'fixtures'/f'{name}.a'
        args=['8'] if name=='recursion' else []
        truth=run(['node','--disable-warning=ExperimentalWarning',str(repo/'oracle/node.mjs'),str(source),*args],name+'-node')
        assert truth.returncode==0 and not truth.stderr,truth.stderr
        binary=scratch/name
        built=run([compiler,'build',str(source),'-o',str(binary),'--sanitize'],name+'-build')
        record={'name':name,'node_exit':truth.returncode,'node_stdout':truth.stdout.decode(),'build_exit':built.returncode,'diagnostic':built.stderr.decode()}
        if name=='factory-cast':
            assert built.returncode==1 and header(source.read_text(),built.stderr.decode()),built.stderr
            assert not header(source.read_text().replace('a cast','a made-up cast',1),built.stderr.decode())
            assert not header(source.read_text().split('\n',1)[1],built.stderr.decode())
            record['header_mutants']=['wrong-reason rejected','removed-header rejected']
        elif name=='speculation':
            assert built.returncode==1 and 'assigning a field of a value' in built.stderr.decode(),built.stderr
            record['status']='NotYet, diagnostic-array length write'
        else:
            assert built.returncode==0,built.stderr
            native=run([str(binary),*args],name+'-native')
            assert (native.returncode,native.stdout,native.stderr)==(truth.returncode,truth.stdout,truth.stderr)
            js=run([compiler,'js',str(source)],name+'-js-build')
            assert js.returncode==0,js.stderr
            emitted=scratch/(name+'.mjs');emitted.write_bytes(js.stdout)
            backend=run(['node','--disable-warning=ExperimentalWarning',str(repo/'oracle/node.mjs'),str(emitted),*args],name+'-backend')
            assert (backend.returncode,backend.stdout,backend.stderr)==(truth.returncode,truth.stdout,truth.stderr),backend.stderr
            counted=run([compiler,'build',str(source),'-o',str(binary),'--count'],name+'-count-build')
            assert counted.returncode==0,counted.stderr
            counts=run([str(binary),*args],name+'-counts')
            assert counts.returncode==0 and counts.stdout==truth.stdout,counts.stderr
            record['counts']=counts.stderr.decode()
        results.append(record)
    for name,fixture,before,after in [
        ('context-restore','context','setContextFlag(true, contextFlagsToClear);','setContextFlag(false, contextFlagsToClear);'),
        ('context-cache-bits','context','setContextFlag(false, contextFlagsToSet);','contextFlags = 5;'),
        ('token-rewind','speculation','currentToken = saveToken;','currentToken = 99;'),
        ('error-rewind','speculation','parseErrorBeforeNextFinishedNode = saveParseErrorBeforeNextFinishedNode;','parseErrorBeforeNextFinishedNode = true;'),
        ('diagnostic-rewind','speculation','parseDiagnostics.length = saveParseDiagnosticsLength;','parseDiagnostics.length = parseDiagnostics.length;'),
        ('reparse-diagnostics','speculation','if (speculationKind !== 2)','if (true)'),
    ]:
        text=(here/'fixtures'/f'{fixture}.a').read_text(); assert text.count(before)==1
        source=scratch/(name+'.a');source.write_text(text.replace(before,after))
        observed=run(['node','--disable-warning=ExperimentalWarning',str(repo/'oracle/node.mjs'),str(source)],name+'-mutant-node')
        truth=next(r for r in results if r['name']==fixture)['node_stdout'].encode()
        assert observed.returncode==0 and not observed.stderr and observed.stdout!=truth
        record={'mutant':name,'exit':0,'caught_by':'stdout against unchanged source Node','stdout':observed.stdout.decode()}
        if fixture=='context':
            binary=scratch/name
            built=run([compiler,'build',str(source),'-o',str(binary),'--sanitize'],name+'-mutant-build')
            assert built.returncode==0,built.stderr
            native=run([str(binary)],name+'-mutant-native')
            assert native.returncode==0 and not native.stderr and native.stdout!=truth
            record['native']='built, completed, changed stdout'
        results.append(record)
(here/'evidence/fixtures.json').write_text(json.dumps(results,indent=2)+'\n')
lines=['# Scout fixture counts','','Local fixtures are outside internal/oracle fixture discovery. Counts from current main --count.','', '| Fixture | Runtime counts |','| --- | --- |']
for r in results:
    if 'counts' in r: lines.append('| '+r['name']+'.a | '+r['counts'].strip().replace('\n','; ')+' |')
(here/'counts.md').write_text('\n'.join(lines)+'\n')
print(json.dumps(results,indent=2))
