import hashlib
import json
import re
import subprocess
from pathlib import Path

root = Path(__file__).resolve().parents[4]
def git(*args):
    return subprocess.check_output(['git', *args], cwd=root, text=True).strip()

symbols = ['ForFunctionWithoutManualMemoization', 'AsCompilationUnit',
           'UnconditionalBlocks', 'ControlDominators',
           'InlineImmediatelyInvokedFunctionExpressionsIncludingMemoCallbacks']
refs = [r for r in git('for-each-ref', '--format=%(refname)', 'refs/remotes/origin').splitlines()
        if not r.endswith('/HEAD')]
result = {'head': git('rev-parse', 'HEAD'), 'native_declarations': {},
          'production_sources': {}, 'nproc': subprocess.check_output(['nproc'], text=True).strip()}
sources = {}
for ref in refs:
    paths = git('ls-tree', '-r', ref, 'stage1/cohere').splitlines()
    matches = []
    for entry in paths:
        metadata, path = entry.split('\t', 1)
        blob = metadata.split()[2]
        if Path(path).suffix not in {'.a', '.ts'} or '/validation' in path or '/testdata/' in path:
            continue
        if blob not in sources:
            sources[blob] = git('show', blob)
        source = sources[blob]
        for symbol in symbols:
            if re.search(r'\b(?:function|class)\s+' + re.escape(symbol) + r'\b', source):
                matches.append({'path': path, 'symbol': symbol})
    result['native_declarations'][ref] = matches
for name in ['set_state_in_effect', 'set_state_in_render', 'static_components']:
    path = root / 'cohere/internal/lint/rules/react' / (name + '.go')
    data = path.read_bytes()
    result['production_sources'][name] = {
        'sha256': hashlib.sha256(data).hexdigest(), 'lines': len(data.splitlines()),
        'hir_calls': sorted(set(re.findall(rb'high_level_intermediate_representation\.(\w+)\(', data)[i].decode()
                                for i in range(len(re.findall(rb'high_level_intermediate_representation\.(\w+)\(', data)))))
    }
print(json.dumps(result, indent=2, sort_keys=True))
