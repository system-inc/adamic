"""Compare complete any censuses with the original hidden-byte attribution."""
import hashlib
import json
from pathlib import Path
import re
import subprocess
import sys
import types

repository, corpus, stock_path, before_path, after_path, output = map(Path, sys.argv[1:7])
for name, path in [('hidden','stage3/census/hidden/hidden.py'), ('ranking','stage3/census/hidden-ranking/ranking.py')]:
 module = types.ModuleType(name)
 module.__file__ = str(repository/path)
 sys.modules[name] = module
 source = subprocess.check_output(['git','show','6c4fc1af:'+path], cwd=repository)
 exec(compile(source,module.__file__,'exec'),module.__dict__)
hidden, ranking = sys.modules['hidden'], sys.modules['ranking']
stock = json.loads(stock_path.read_text())
actual = {path.relative_to(corpus).as_posix() for path in corpus.rglob('*') if path.is_file()}
assert actual == set(stock), 'manifest must cover every compiler file'
for name, metadata in stock.items():
 data = (corpus/name).read_bytes()
 assert len(data) == metadata['bytes'] and hashlib.sha256(data).hexdigest() == metadata['sha256'], name

def is_any(kind, reason):
 return kind in {'NotYet', 'Refused'} and re.search(r'\bany\b', reason) is not None

def snapshot(ledger, sha):
 rows = hidden.read_rows(ledger)
 measured = hidden.calculate(rows,stock,corpus)
 entries = ranking.boundaries(rows,stock,corpus)
 partition = ranking.partition(entries,measured['files'],{})
 selected = [e for e in entries if is_any(e['kind'],e['reason'])]
 reasons = [r for r in partition['ranked_reasons'] if is_any(r['kind'],r['reason'])]
 sites = {(f['kind'],f['where'],f['reason'],f['text']) for row in rows[1:] for f in row['findings'] if is_any(f['kind'],f['reason'])}
 timer = [e for e in selected if re.search(r'/sys\.ts:52:\d+:', e['diagnostic'])]
 returns = [e for e in selected if e['reason'] in {'a function returning any', 'a generic function whose resolved return type is any'}]
 return dict(compiler_commit=sha,
  boundaries=len({(e['file'],e['start'],e['end']) for e in selected}),
  credited_hidden_bytes=sum(r['bytes_revealed_if_fixed_alone'] for r in reasons),
  diagnostic_sites=len(sites),
  clearTimeout_annotation=dict(boundaries=len({(e['file'],e['start'],e['end']) for e in timer}),
   attributed_hidden_bytes=sum(e['attributed_hidden_bytes'] for e in timer)),
  return_any=dict(boundaries=len({(e['file'],e['start'],e['end']) for e in returns}),
   credited_hidden_bytes=sum(r['bytes_revealed_if_fixed_alone'] for r in reasons if r['reason'] in {'a function returning any','a generic function whose resolved return type is any'})),
  all_hidden_bytes=measured['hidden_bytes'],
  reasons=[{k:r[k] for k in ('kind','reason','boundary_count','bytes_revealed_if_fixed_alone')} for r in reasons],
  diagnostics=[dict(kind=k,where=hidden.local_where(w,corpus),reason=r,text=t) for k,w,r,t in sorted(sites)],
  ledger_sha256=hashlib.sha256(ledger.read_bytes()).hexdigest(),
  checker_rejected=rows[0]['checker_rejected'], measured_files=len(rows)-1)

before = snapshot(before_path,'3ce9a33b7ece5d51ec32645b8d66af1763db7802')
after = snapshot(after_path,sys.argv[7])
result = dict(before=before,after=after,
 reduction={key:before[key]-after[key] for key in ('boundaries','credited_hidden_bytes','diagnostic_sites','all_hidden_bytes')},
 provenance=dict(scout_commit='4e6121d3',adaptation_commit='1e920a31b601422c75a26ead083e9c1423c34b66',
  verified_source_files=len(stock),stock_sha256=hashlib.sha256(stock_path.read_bytes()).hexdigest(),
  measurement='measured on a checker-rejected program',
  definition='Any bucket: NotYet/Refused reason text containing the word any; hidden-byte credit uses the unchanged outermost-cause partition. Conflicts remain separate. These are refused measurement boundaries, not compiled programs.'))
output.write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps({'before':{k:before[k] for k in ('boundaries','credited_hidden_bytes')},'after':{k:after[k] for k in ('boundaries','credited_hidden_bytes')},'reduction':result['reduction']}))
