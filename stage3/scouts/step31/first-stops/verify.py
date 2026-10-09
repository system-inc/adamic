#!/usr/bin/env python3
"""Independent Node witnesses; no native-pass assertion. Set STEP31_TYPESCRIPT."""
import json
import os
from pathlib import Path
import subprocess
import tempfile

HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[3]
TS = os.environ['STEP31_TYPESCRIPT']
NODE = os.environ.get('STEP31_NODE', 'node')

def run(source):
    with tempfile.TemporaryDirectory(prefix='step31-node-') as folder:
        file = Path(folder) / 'witness.a'
        file.write_text(source)
        result = subprocess.run([NODE, '--disable-warning=ExperimentalWarning', str(ROOT / 'oracle/node.mjs'), str(file)], capture_output=True, text=True)
        return {'stdout': result.stdout, 'stderr': result.stderr, 'exit': result.returncode}

def diagnostics(source, unchecked=True, optional=True):
    with tempfile.TemporaryDirectory(prefix='step31-check-') as folder:
        file = Path(folder) / 'witness.ts'
        file.write_text(source)
        script = '''const ts=require(process.argv[1]); const file=process.argv[2];
const p=ts.createProgram([file],{strict:true,noUncheckedIndexedAccess:process.argv[3]==="true",exactOptionalPropertyTypes:process.argv[4]==="true",noEmit:true,target:ts.ScriptTarget.ESNext});
console.log(JSON.stringify({version:ts.version,diagnostics:ts.getPreEmitDiagnostics(p).map(d=>({code:d.code,message:ts.flattenDiagnosticMessageText(d.messageText,"\\n")}))}));'''
        result = subprocess.run([NODE, '-e', script, TS, str(file), str(unchecked).lower(), str(optional).lower()], capture_output=True, text=True, check=True)
        return json.loads(result.stdout)

observations = []
for name, code, expected in [('checker-index.a', 18048, '42\n'), ('emitter-optional.a', 2322, 'index.ts\npath,extension,packageId,originalPath,resolvedUsingTsExtension\ntrue\n')]:
    source = (HERE / name).read_text()
    node = run(source)
    assert node == {'stdout': expected, 'stderr': '', 'exit': 0}, node
    checker = diagnostics(source)
    assert checker['version'] == '6.0.3', checker
    assert [d['code'] for d in checker['diagnostics']] == [code], checker
    ablated = diagnostics(source, unchecked=code != 18048, optional=code != 2322)
    assert ablated['diagnostics'] == [], ablated
    if code == 18048:
        mutant = source.replace('declarations: [{ endFlowNode: 42 }]', 'declarations: []')
        changed = run(mutant)
        assert changed['exit'] != 0 and 'TypeError' in changed['stderr'], changed
        catcher = 'Node throws TypeError on the empty indexed input'
    else:
        mutant = source.replace('            originalPath: result.resolvedModule.originalPath,\n', '')
        changed = run(mutant)
        assert changed['exit'] == 0 and changed['stdout'] != expected, changed
        catcher = 'Node own-key output differs when the undefined-valued field is omitted'
    observations.append({'file': name, 'node': node, 'stock_strict_checker': checker, 'responsible_option_ablation': ablated, 'mutant': {'catcher': catcher, 'exit': changed['exit'], 'stdout': changed['stdout']}, 'native': 'pending: pinned TypeScript submodule unavailable'})
(HERE / 'evidence/witness-results.json').write_text(json.dumps(observations, indent=2) + '\n')
print('2 Node witnesses match; 2 exact stock strict diagnostics match; 2 mutants caught; native pending')
