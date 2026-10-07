#!/usr/bin/env python3
"""Add the ruling's disjoint partition without losing the original witnesses."""
import json
from pathlib import Path
import sys
report=json.loads(Path(sys.argv[1]).read_text())
out=Path(sys.argv[2])
rows=report['rows']
assert len(rows)==100
assert len({(r['File'],r['Line'],r['Column'],r['Kind']) for r in rows})==100
assert report['counts']=={b:sum(r['bucket']==b for r in rows) for b in ('a','b','c','rest')}
assert sum(report['counts'].values())==100
assert all(r['writes'] for r in rows)
assert all(r['Source']=='never' for r in rows if r['bucket']=='b')
assert all(r['proof'].get('guards') for r in rows if r['bucket']=='c')
assert all(r['proof'].get('immediate_write') for r in rows if r['bucket']=='a')
# Keep the original complete list and its stable witness anchors.
saved=out.read_text()
start=saved.index('## Complete write-site list for @system_adamic')
old=saved[start:].replace('## Complete write-site list for @system_adamic','## Original 100-site write inventory and witness anchors',1)
old=old.replace('**100 sites; 58 distinct assignment witnesses.** Each W identifier','**100 original sites; 58 distinct assignment witnesses.** The new rest bucket is above. Each W identifier',1)
operations={}
for r in rows:
 for w in r['writes']:operations.setdefault((w['where'],w['access'],w['property'],w['written']),w)
ids={k:f'W{i:03}' for i,k in enumerate(sorted(operations),1)}
def cell(v):return str(v).replace('|','&#124;').replace('\n','<br>').replace('`','&#96;')
def site(r):return f"{r['File']}:{r['Line']}:{r['Column']}"
def witnesses(r):
 return '<br>'.join(f"[{ids[w['where'],w['access'],w['property'],w['written']]}](#{ids[w['where'],w['access'],w['property'],w['written']].lower()}): {cell(w['property'])} {cell(w['operation'])} {cell(w['written'])}" for w in r['writes'])
a,b,c,rest=(report['counts'][k] for k in ('a','b','c','rest'))
lines=[f'(a) Fresh literal initialization: **{a} sites**.',f'(b) Adamic source `never`: **{b} sites**; stock-checker discrepancy retained below.',f'(c) Tag-checked class downcast: **{c} sites**.',f'Rest: **{rest} sites** stay refused; every site has source, target and write witnesses below.','All **100** write-reaching sites occur once; the **291** read-only sites and existing adaptation edits are unchanged.','', '# October 7 05:30 ruling', '',
'This replaces the 04:35 write handoff with exactly one bucket per original write-reaching location.',
'The full original 100-site inventory and all 58 witnesses remain below for traceability.', '',
'## Proof boundaries', '',
'Case (a) uses the expressly sanctioned `sys.ts:155` initial-construction pattern. The literal is the allocation arm of `(v || (v = {}))[k] = x`: no second alias to that allocation exists before the immediate initializing write. The primary wider-view binding is part of that expression. This does not preserve freshness after storing, passing, capturing or returning the value. Later writes on the old-value arm have a source binding whose static type already declares the indexed optional members, and are not additional `{}` widening sites.', '',
'**Case (b) is a classification of the Adamic census relation, not an independent dead-code proof for upstream tsc.** All 47 records say `Source: never`, including the ruling examples at utilities.ts:4416 (two relations), factory/utilities.ts:363 and factory/nodeFactory.ts:7105. The stock checker does not set `TypeFlags.Never` at any of these 47 located expressions. At checker.ts:52739 its printed type is `never` but its flags are the error/any representation; 46 print inhabited types. The original census and bound stock evidence are both preserved. Under the ruling these 47 static Adamic relations go in (b); the compiler must reconcile the discrepancy before using them as a native unreachable-code proof. No declaration or reachability adaptation is proposed here.', '',
'Case (c) requires a positive, dominating tag condition on the same checker-bound value as the downcast. Parentheses and assertions are unwrapped; identifier identity is a checker symbol, not spelling. The analyzer recognizes true branches of `if`, conditional expressions and the right-hand evaluation of `&&`. `TypeFlags.TypeParameter` proves TypeParameter; `TypeFlags.Union` proves UnionType; `TypeFlags.Object` together with `ObjectFlags.Anonymous` proves AnonymousType. Enum members resolve to their declarations. Composite TypeVariable tags, calls returning allegedly appropriate types, and unproved mutable flag aliases do not qualify.', '',
'The proof is deliberately local and conservative. The original write graph joins callers and paths. A rest witness is a possible alias-reaching write, not a claim that every path executes it. Rest therefore includes relations for which constructor/result provenance or a flag alias needs more proof, even when a future compiler analysis may establish that they are allowed. Guards license the class view, not unrelated writes to a different shape.', '',
'## (a) Fresh literal initialization', '', '| Site in src/compiler | Source | Target | Allocation and immediate write | Escape proof |','|---|---|---|---|---|']
for r in rows:
 if r['bucket']=='a':
  p=r['proof'];lines.append(f"| {site(r)} | {cell(r['Source'])} | {cell(r['Target'])} | {p['allocation']}; {p['immediate_write']}: {cell(p['expression'])} | {cell(p['escape_analysis'])} Primary binding: {cell(', '.join(p['owning_binding']))}. |")
lines+=['','## (b) Adamic static source never','','| Site in src/compiler | Adamic source | Target | Stock bound expression type | Never evidence |','|---|---|---|---|---|']
for r in rows:
 if r['bucket']=='b':lines.append(f"| {site(r)} | never | {cell(r['Target'])} | {cell(r['proof']['stock_source'])} | Original refusal census Source=never; stock Never flag={r['proof']['stock_never']}. Compiler reconciliation required. |")
lines+=['','## (c) Guarded class view','','| Site in src/compiler | Source | Target | Dominating tag check on the same symbol | Declared member evidence |','|---|---|---|---|---|']
for r in rows:
 if r['bucket']=='c':
  guards='<br>'.join(f"{g['where']}: {cell(g['condition'])}; bound receiver={g['receiver_symbol']}" for g in r['proof']['guards'])
  lines.append(f"| {site(r)} | {cell(r['Source'])} | {cell(r['Target'])} | {guards} | {cell(', '.join(r['proof']['target_declarations']))}; writes: {witnesses(r)} |")
lines+=['','## Rest: full refused write list','','No further source adaptation is made. These require a source declaration or further compiler proof within the three permitted cases. A fresh result from a helper is not a fresh object literal; a tag supplied to a constructor is not itself a dominating tag check; a composite TypeVariable tag does not prove TypeParameter. The generic and sentinel relations retain the original conservative aliases and their exact write witnesses.','','| Site in src/compiler | Source type | Target type | First refused p | What is written |','|---|---|---|---|---|']
for r in rows:
 if r['bucket']=='rest':lines.append(f"| {site(r)} | {cell(r['Source'])} | {cell(r['Target'])} | {cell(r['Property'])} | {witnesses(r)} |")
lines+=['','## Reproduction and failing proofs','','The saved wave-2 tree supplies the exact line coordinates of the 100-site handoff. Source pin and original alias analysis are in [provenance](evidence/read-write/provenance.json) and [analysis](evidence/read-write/analysis.json.gz). The rebucketing uses the same strict checker options and adds no source edits. The exact per-site proof is in [ruling evidence](evidence/ruling/analysis.json.gz).','','```sh','NODE_PATH=STOCK_6_0_3_NODE_MODULES node stage3/adapt/75-optional-widening/rebucket.cjs TREE stage3/adapt/75-optional-widening/evidence/read-write/analysis.json.gz ruling.json > ruling.log 2>&1','NODE_PATH=STOCK_6_0_3_NODE_MODULES node stage3/adapt/75-optional-widening/probe-ruling.cjs MUTANT_TREE mutants.json > mutants.log 2>&1','python3 stage3/adapt/75-optional-widening/render-ruling.py ruling.json stage3/adapt/75-optional-widening/READ_WRITE.md > render.log 2>&1','```','','Mutant results and commands are recorded in [mutants](evidence/ruling/mutants.json). Full-apply revalidation is independent of this saved 391-site classification; its result is recorded in README.md when complete.','',old]
text='\n'.join(lines).rstrip()+'\n';assert '\u2014' not in text
out.write_text(text)
print(json.dumps(report['counts']))
