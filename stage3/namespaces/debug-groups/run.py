"""Compare a group's real shape and semantic mutant on independent source Node."""
import json
import os
from pathlib import Path
import subprocess
import sys

root = Path(__file__).resolve().parents[3]
group = sys.argv[1]
folder = Path(__file__).resolve().parent
spec = json.loads((folder / (group + '.json')).read_text())
scratch = Path('/tmp/debug-groups') / group
scratch.mkdir(parents=True, exist_ok=True)
source = folder / (group + '.a')
text = source.read_text()
changed = text
for old, new in spec['mutations']:
    assert old in changed, old
    changed = changed.replace(old, new)
assert changed != text
mutant = scratch / 'mutant.a'
mutant.write_text(changed)
def run(command):
    result = subprocess.run(command, cwd=root, capture_output=True, timeout=120)
    return dict(command=command, stdout=result.stdout.decode(), stderr=result.stderr.decode(), exit=result.returncode)
node = run(['node', '--disable-warning=ExperimentalWarning', 'oracle/node.mjs', str(source)])
bad = run(['node', '--disable-warning=ExperimentalWarning', 'oracle/node.mjs', str(mutant)])
assert node['exit'] == 0 and not node['stderr'], node
assert bad['stdout'] != node['stdout'] or bad['exit'] != node['exit'], (node, bad)
compiler = os.environ.get('DEBUG_GROUP_COMPILER', '/tmp/namespaces-graph-matrix-delivery/adamic')
build = run([compiler, 'build', str(source), '-o', str(scratch / 'native'), '--sanitize'])
result = dict(group=group, sites=spec['sites'], node=node, mutant=bad, build=build, judgment=spec['judgment'])
if build['exit'] == 0:
    native = run([str(scratch / 'native')]); assert {k:native[k] for k in ['stdout','stderr','exit']} == {k:node[k] for k in ['stdout','stderr','exit']}, native
    result['native'] = native
else:
    assert spec['boundary'] in build['stderr'], build
(folder / (group + '.results.json')).write_text(json.dumps(result,indent=2)+'\n')
print(group, 'Node:', repr(node['stdout']), 'mutant caught; build exit', build['exit'])
