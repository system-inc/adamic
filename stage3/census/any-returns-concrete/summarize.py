"""Recount a complete census with the frozen hidden-byte attribution code."""
import gzip
import hashlib
import json
from pathlib import Path
import subprocess
import sys
import types

repository, corpus, ledger, output = map(Path, sys.argv[1:5])

def frozen(ref, path):
 return subprocess.check_output(['git','show',ref+':'+path],cwd=repository)

# Reference Adamic's measurement implementation; do not import its topic branch.
for name, path in [('hidden','stage3/census/hidden/hidden.py'),('ranking','stage3/census/hidden-ranking/ranking.py')]:
 module = types.ModuleType(name)
 module.__file__ = str(repository/path)
 sys.modules[name] = module
 exec(compile(frozen('6c4fc1af',path),module.__file__,'exec'),module.__dict__)
hidden, ranking = sys.modules['hidden'], sys.modules['ranking']
stock = json.loads(gzip.decompress(frozen('6c4fc1af','stage3/census/hidden/evidence/stock.json.gz')))
for name, metadata in stock.items():
 data = (corpus/name).read_bytes()
 assert len(data)==metadata['bytes'] and hashlib.sha256(data).hexdigest()==metadata['sha256'],name
rows = hidden.read_rows(ledger)
measured = hidden.calculate(rows,stock,corpus)
entries = ranking.boundaries(rows,stock,corpus)
result = ranking.partition(entries,measured['files'],{})
original = json.loads(frozen('fb2b782a','stage3/census/any-returns/RESULT.json'))
reason = 'a function returning any'
row = next((row for row in result['ranked_reasons'] if row['kind']=='NotYet' and row['reason']==reason),None)
remaining = [entry for entry in entries if entry['kind']=='NotYet' and entry['reason']==reason]
callees = {}
for entry in remaining:
 diagnostic = entry['diagnostic']
 location = diagnostic.split('/src/compiler/',1)[1].split(': stage 0',1)[0]
 name, line, column = location.rsplit(':',2)
 matched = next((c for c in original['ranked_callees'] if c['declaration'].rsplit(':',2)[:2]==['src/compiler/'+name,line]),None)
 key = matched['callee'] if matched else name+':'+line
 stats = callees.setdefault(key,{'declaration':name+':'+line,'stock_return':matched['stock_return'] if matched else None,'boundaries':set(),'bytes':0})
 stats['boundaries'].add((entry['file'],entry['start'],entry['end']))
 stats['bytes'] += entry['attributed_hidden_bytes']
for stats in callees.values():
 stats['boundaries']=len(stats['boundaries'])
 stats['reason']='source any contract' if stats['stock_return']=='any' else 'checker or measurement state unresolved; non-any stock declaration'
summary = {'before':{'boundaries':original['boundaries'],'hidden_bytes':original['hidden_bytes']},'after':{'boundaries':row['boundary_count'] if row else 0,'hidden_bytes':row['bytes_revealed_if_fixed_alone'] if row else 0},'remaining_callees':callees,'provenance':{'compiler_base':'ed6e29751ee47d86fad450cd1674139883bc0f70','correction':'checker mapper identity preserved in measurement snapshots','production_branch_is_not_measured':True,'ledger_sha256':hashlib.sha256(ledger.read_bytes()).hexdigest(),'verified_source_files':len(stock)}}
output.write_text(json.dumps(summary,indent=2)+'\n')
print(json.dumps(summary['after']))
