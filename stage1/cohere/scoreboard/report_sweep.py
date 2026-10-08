#!/usr/bin/env python3
"""Close the full-tree receipt against the reduction set and render the numbers."""
import argparse,collections,csv,gzip,hashlib,json,pathlib
p=argparse.ArgumentParser();p.add_argument('results');a=p.parse_args();out=pathlib.Path(a.results)
s=json.loads((out/'summary.json').read_text());w=json.loads((out/'witnesses.json').read_text())
# Registered Next descriptors and the unported @next/next namespace are one family.
if '@next' in s['per_family']:
 old=s['per_family'].pop('@next');target=s['per_family']['next']
 for key,value in old.items():
  if isinstance(value,dict):
   for name,n in value.items():target[key][name]=target[key].get(name,0)+n
  else:target[key]+=value
 (out/'summary.json').write_text(json.dumps(s,indent=2)+'\n')
if not s['complete'] or s['files']!=235377 or s['lint_files']!=201196:raise SystemExit('full tree denominator differs')
if s['total']['checks']!=99626201 or len(s['port_rule_order'])!=93 or len(s['unported_ranked'])!=399:raise SystemExit('rule/cell denominator differs')
if any(t['checks']!=201196 for name,t in s['per_rule'].items() if not name.startswith('format/')):raise SystemExit('per-rule input coverage differs')
if s['oracle_byte_verified_files']!=s['files']:raise SystemExit('oracle transport audit incomplete')
expected=set();inputs={};formats=collections.Counter()
for f in sorted(out.glob('files-*.gz')):
 for line in gzip.open(f,'rt'):
  r=json.loads(line);inputs[r['index']]=r['source_sha256'];formats[r['format_state']]+=1
  expected.update((r['index'],rule) for rule in r['divergences'])
actual=set();unisolated=collections.Counter()
for line in gzip.open(out/'divergences-minimized.jsonl.gz','rt'):
 r=json.loads(line);key=(r['index'],r['rule'])
 if r['source_sha256']!=inputs.get(r['index']):raise SystemExit('reduction source differs '+str(key))
 if key in actual:raise SystemExit('duplicate reduction '+str(key))
 actual.add(key)
 if r['cause'].startswith('rule emission'):unisolated[r['rule']]+=1
if expected!=actual:
 (out/'reduction-gaps.json').write_text(json.dumps(dict(missing=sorted(expected-actual),unexpected=sorted(actual-expected)),indent=2)+'\n')
 raise SystemExit('reduction closure failed: missing '+str(len(expected-actual))+'; unexpected '+str(len(actual-expected)))
if len(expected)!=s['divergences'] or len(expected)!=w['cells']:raise SystemExit('divergence count differs')
for kind in ['rule','family']:
 with (out/('per-'+kind+'.csv')).open('w') as f:
  c=csv.writer(f,lineterminator="\n");c.writerow([kind,'checks','agree','diverge','blocked','agreement_among_successful_cells','go_findings','node_findings','blocked_causes'])
  for name,t in sorted(s['per_'+kind].items()):
   n=t['agree']+t['diverge'];c.writerow([name,*[t[k] for k in ['checks','agree','diverge','blocked']],f"{100*t['agree']/n:.6f}%" if n else '',t['go_findings'],t['node_findings'],json.dumps(t['blocked_causes'],sort_keys=True)])
with (out/'unported-ranked.csv').open('w') as f:
 c=csv.writer(f,lineterminator="\n");c.writerow(['rank','rule','findings','finding_files','census_status'])
 for i,r in enumerate(s['unported_ranked'],1):c.writerow([i,r['rule'],r['findings'],r['files'],r['status']])
with (out/'blocked-causes.csv').open('w') as f:
 c=csv.writer(f,lineterminator="\n");c.writerow(['cause','cells'])
 c.writerows(sorted(s['blocked_cells_by_cause'].items()))
t=s['total'];lines=['# Step 43: all 235,377 tracked files at the 23 public pins','',
 'Both snapshots are **dependency-uninstalled**. This is the identified public subset of Kirk’s quiet hundred; it does not include the other Mac-only snapshots. The port is the source Node execution, not a native Adamic lint build.','',
 f"**{s['files']:,} files; {s['lint_files']:,} script files; {s['bytes']:,} input bytes.** All eight requested extensions are included. 51 tracked symlinks use their Git blob bytes, preserving the original paths.",'',
 f"**{t['checks']:,} cells: {t['agree']:,} agree, {t['diverge']:,} diverge, {t['blocked']:,} blocked.** Agreement among successfully compared cells: {100*t['agree']/(t['agree']+t['diverge']):.6f}%. Blocked cells are excluded from that percentage and never counted as matching.",'',
 'Scope: 93 registered port rules, 399 unported registry rules, full all-fixes, a separately labeled 90-listener syntax-fixes comparison, and one formatter per input. Full all-fixes remains blocked when typed rules require checker facts. Syntax-fixes certifies only the measured syntax listener scope.','',
 '## Agreement per family','', '| Family | Checks | Agree | Diverge | Blocked |','| --- | ---: | ---: | ---: | ---: |']
for name,r in sorted(s['per_family'].items()):lines.append('| '+name+' | '+' | '.join(f"{r[k]:,}" for k in ['checks','agree','diverge','blocked'])+' |')
lines+=['','Every rule’s numbers, findings and successful-cell percentage are in `per-rule.csv`; every file’s states, hashes, finding counts, stdout digests, exit codes and timings are in `files-*.jsonl.gz`. Rule order and the state alphabet are in `summary.json` and SCOUT.md.','', '## Blocked cells','', '| Cause | Cells |','| --- | ---: |']
for name,n in sorted(s['blocked_cells_by_cause'].items()):lines.append(f'| {name} | {n:,} |')
lines+=['', 'Execution faults are kept separately where the receipt cannot establish one of the requested semantic causes. Unported rules have an additional Go census status histogram; those statuses overlap the primary unported block and must not be added to the primary total.','', '## Divergences and responsibility','',f"All **{len(actual):,} divergence cells** have same-path replayed, source-subsequence reductions, complete reduced Go/Node answers and checked implementation references in `divergences-minimized.jsonl.gz`. There are **{w['distinct_witnesses']:,} distinct rule/signature/input witnesses** in `witnesses.json`. Reduction is a single-rune deletion fixed point with Go parse validity held for scripts; global shortestness is not asserted.",'', 'The inspected causal sites distinguish comment/BOM loss, blank lines, hashbang comment guards, import phases, constructor/static-block parsing and missing JSDoc attachment. Every cell retains actual rule emission/fix or formatter entry sites; the witness catalog adds the inspected causal sites. The final paired tree checks are retained in `additional-trees.json`.','', 'Cases with emission-site attribution only: '+json.dumps(dict(unisolated),sort_keys=True)+'.','', '## Unported rules ranked by measured findings','', '| Rank | Rule | Go findings | Finding files | Census |','| ---: | --- | ---: | ---: | --- |']
for i,r in enumerate(s['unported_ranked'][:20],1):lines.append(f"| {i} | {r['rule']} | {r['findings']:,} | {r['files']:,} | {r['status']} |")
lines+=['','The complete 399-rule ranking is `unported-ranked.csv`. An incomplete census is a lower bound: missing checker programs, required options, option decoder failures and parse diagnostics remain counted and named. This measures registry defaults, not installed project policies.','', '## Receipt integrity','',f"All {s['oracle_byte_verified_files']:,} oracle records passed the byte-transport audit; {s['oracle_corrected_files']:,} files needed corrected answers. JSON’s only lossy string substitution is U+FFFD: explicit raw-byte fields and old strings without that sentinel are lossless; every ambiguous old string is independently replayed.",'', 'Source Node and the real Go oracle ran every applicable input. Persistent host workers reuse transport, not language implementations. CLI-versus-worker fixtures, an independent isolated-rule census, the live message mutant, invalid-byte mutant and stderr collector regression pass in the package validation receipt.','']
(out/'REPORT.md').write_text('\n'.join(lines))
checks=dict(files=s['files'],lint_files=s['lint_files'],cells=t['checks'],divergence_closure=len(actual),distinct_witnesses=w['distinct_witnesses'],oracle_byte_verified_files=s['oracle_byte_verified_files'],emission_site_only_cells=dict(unisolated),formatter_states=dict(formats))
(out/'closure.json').write_text(json.dumps(checks,indent=2)+'\n')
with (out/'artifacts.sha256').open('w') as target:
 for f in sorted(out.iterdir()):
  if f.is_file() and f.name not in {'artifacts.sha256','reduction-gaps.json'}:
   target.write(hashlib.file_digest(f.open('rb'),'sha256').hexdigest()+'  '+f.name+'\n')
print(json.dumps(checks,sort_keys=True))
