"""Hold the saved real-project replay to the completed census and top-ranked row."""
import gzip
import json
from pathlib import Path

out=Path(__file__).resolve().parent
summary=json.loads((out/'summary.json').read_text())
command=json.loads((out/'evidence/replay-command.json').read_text())
replay=json.loads((out/'evidence/replay.json').read_text())
with gzip.open(out/'full.jsonl.gz','rt') as stream:records=[json.loads(line) for line in stream]
reason=command[command.index('-reason')+1]
where=command[command.index('-where')+1]
kind=command[command.index('-kind')+1]
top=next(r for r in summary['rows'] if not r['context_sensitive'])
assert reason==top['reason'] and kind=='NotYet', 'top non-context-sensitive reason'
assert where.removeprefix('/tmp/stage3-notyet-adapted/') in top['examples'], 'table example'
assert len(replay['units'])==1 and replay['units'][0]['status']=='attempted', 'one eligible replay unit'
unit=replay['units'][0]['where']
full=next(r for r in records[1:] if r['file']==replay['file'])
expected=[f for f in full['findings'] if f['phase']=='lowering' and f['unit']==unit]
actual=replay['findings']
assert any((f['phase'],f['kind'],f['where'],f['reason'])==('lowering',kind,where,reason) for f in expected), 'full signature'
assert any((f['phase'],f['kind'],f['where'],f['reason'])==('lowering',kind,where,reason) for f in actual), 'replay signature'
# Entry census and replay intentionally use different scope labels.
strip_label=lambda fs:[{k:v for k,v in f.items() if k!='measurement'} for f in fs]
assert strip_label(actual)==strip_label(expected), 'ordered selected-unit findings'
positive=(out/'evidence/replay.log.txt').read_text()
negative=(out/'evidence/replay-real-mutant.log.txt').read_text()
assert 'reproduced NotYet: '+reason+' at '+where in positive, 'successful command output'
assert 'replay signature did not reproduce: NotYet '+reason+' at '+where in negative, 'real selection mutant'
print('PASS: top non-context-sensitive reason, table example, exact phase/kind/where/reason signature, one eligible unit, all '+str(len(actual))+' ordered findings (apart from scope labels), and real sibling-selection mutant')
print('selected unit: '+unit)
print(positive.strip())
