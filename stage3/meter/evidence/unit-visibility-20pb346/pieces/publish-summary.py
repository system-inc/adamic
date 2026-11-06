"""Wait for the bounded queue, independently verify it, then publish its summary.

This process starts no census. It exits without a summary commit if the queue
fails, any proof check fails, or a census process remains active.
"""
from pathlib import Path
import json, os, subprocess, time
from publication_bridge import publish_commit
repo=Path('/workspace/adamic')
root=Path('/workspace/scratch/20pb346-2')
folder=Path(__file__).resolve().parent
while not (root/'unit-supervisor.done').exists():
    if 'Traceback' in (root/'unit-supervisor.log').read_text():
        raise RuntimeError('bounded queue failed; completed checkpoints remain published')
    pid=int((root/'unit-supervisor.pid').read_text())
    status=Path('/proc')/str(pid)/'stat'
    assert status.exists() and status.read_text().split(') ',1)[1][0]!='Z','bounded queue exited without completion'
    time.sleep(55)
subprocess.run(['python3',str(folder/'finish-pieces.py')],cwd=repo,check=True)
summary=json.loads((folder/'summary.json').read_text())
execution=json.loads((folder/'execution-plan.json').read_text())
expected_pairs=2+len(execution['completed_original_pieces'])+len(execution['pieces'])
assert summary['identical_pairs']==expected_pairs and summary['files']==79 and summary['units']==10714
names=('before-entry','after-entry','before-compiler','after-compiler','census','pieces-before','pieces-after','bounded-before','bounded-after','integration-entry','integration-census')
executables={str(root/name).encode() for name in names}
active=[]
for path in Path('/proc').glob('[0-9]*/cmdline'):
    try:
        args=path.read_bytes().split(b'\0')
        if args and args[0] in executables:active.append(int(path.parent.name))
    except (FileNotFoundError,ProcessLookupError):pass
assert not active,('task census processes remain',active)
report=folder.parent/'REPORT.md'
text=report.read_text()
text=text.replace('Measurements remain pinned to the preparation base above.', 'Entry growth, entry profiles, and the initial capped measurements remain pinned\nto the preparation base above. The bounded compiler comparisons use the\nintegration base `d4c6b07aca1a7c55785226563c2ad79d552c1adf` with the same frozen\nadapted input; each source hash is recorded in the selections and plan.')
addition='''## Completed bounded replacement proof

At the user's request, no further whole pass was run. All %d matching
before/after selections completed and have byte-identical latent.jsonl outputs:
two file-group pairs cover 20 files and 809 candidates; %d unit-piece pairs
cover the remaining 59 files and 9,905 candidates. The independent final verifier
checks exact disjoint coverage of all 79 compiler files and 10,714 candidates,
zero exit statuses, matching raw bytes and hashes, and the recorded time bounds.
The largest individual wall time is %.3f seconds, below 30 minutes.

See `pieces/summary.json` and the selected inputs, compressed outputs, unit times,
and per-side metrics beside each result. Oversized files use a symmetric
scratch-only full-mode unit selector, validated against the original binaries;
it keeps the private-field cache path active. The shipped census code has no
new selector. Two independent pairs may overlap; before then after remains
sequential within each pair. These wall times are bounds, not a speed benchmark.
Separate pieces are not concatenated to claim whole-pass identity, because
generic call witnesses have pass scope.

The meter Python suite also passes all 66 tests using freshly rebuilt ordinary,
entry, and latent binaries from the integration base. See
`integration-python-tests.txt`. The earlier complete Go, Python, pair, and race
results remain saved. The implementation did not change during piece collection.
No task census process remains active.

'''%(summary['identical_pairs'],summary['unit_piece_pairs'],summary['maximum_wall_seconds'])
assert '## Completed bounded replacement proof' not in text
text=text.replace('Task: #20pb346',addition+'Task: #20pb346')
report.write_text(text)
subprocess.run(['git','diff','--check'],cwd=repo,check=True)
subprocess.run(['git','add','stage3/meter/evidence/unit-visibility-20pb346'],cwd=repo,check=True)
subprocess.run(['git','commit','-m','meter: verify complete bounded compiler comparison'],cwd=repo,check=True)
publish_commit(repo,root/'final-connector-publication.log')
sha=subprocess.check_output(['git','rev-parse','HEAD'],cwd=repo,text=True).strip()
remote=subprocess.check_output(['git','-c','credential.helper=','ls-remote','--heads','origin','typescript/meter-stream-records-2'],cwd=repo,text=True).split()[0]
assert sha==remote
(root/'final-publication.json').write_text(json.dumps(dict(sha=sha,branch='typescript/meter-stream-records-2',pairs=summary['identical_pairs'],files=79,units=10714,active_census_processes=active),indent=2)+'\n')
