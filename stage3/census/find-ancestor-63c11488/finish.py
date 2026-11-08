"""Recount, render and seal a report-only commit after all paired runs complete."""
import hashlib,json,subprocess,sys,os
from collections import Counter
from pathlib import Path
out=Path(__file__).resolve().parent;repo=out.parents[2];tree=Path('/workspace/scratch/fa-adapted')
def load(name):return [json.loads(s) for s in (out/(name+'.jsonl')).read_text().splitlines()]
def esc(s):return str(s).replace('|','&#124;').replace('\n',' ')
def loc(s):return s.split('/src/',1)[-1]
def why(f):return f['kind']+': '+f['reason']+' at '+loc(f['where']) if f else 'finding-free'
def git(*args):return subprocess.check_output(['git',*args],cwd=repo,text=True).strip()
base=git('rev-parse','1060360b');head=git('rev-parse','63c11488')
assert git('rev-parse','HEAD')==base
assert git('rev-parse',head+'^')==base
expected=json.loads((out/'source-hashes.json').read_text())
for file,sha in expected.items():assert hashlib.sha256((tree/file).read_bytes()).hexdigest()==sha,file
provenance=json.loads((out/'provenance.json').read_text())
for side,pin in [('parent',base),('head',head)]:
 root=Path('/workspace/scratch/fa-'+side)
 assert provenance[side]['sha']==pin==subprocess.check_output(['git','rev-parse','HEAD'],cwd=root,text=True).strip()
 assert subprocess.run(['git','diff','--quiet',pin,'--','internal','stage3/census/latent','stage3/meter'],cwd=root).returncode==0
 for kind,sha in provenance[side]['binaries'].items():
  path=Path('/workspace/scratch/fa-'+side+'-'+kind+('-bin' if kind=='entry' else ''))
  assert hashlib.sha256(path.read_bytes()).hexdigest()==sha
assert not list(out.rglob('*.go'))
assert json.loads((out/'pair-tests.status.json').read_text())['exit']==0
for label,prefix in [('entry','entry-run'),('utilitiesPublic','full-utilitiesPublic'),('core','full-core')]:
 a,b=load('parent-'+prefix),load('head-'+prefix)
 assert a[0]==b[0] and a[0]['status']=='measurement' and not a[0]['diagnostics']
 delta=json.loads((out/(label+'-paired.json')).read_text())
 def raw(rows):
  fs={};us={};first={}
  for row in rows[1:]:
   for u in row.get('units') or []:us[u['where']]=u
   for f in row.get('findings') or []:
    fs.setdefault((f['where'],f['kind'],f['reason'],f['text']),f)
    if f['phase']=='lowering':first.setdefault(f['unit'],f)
  return fs,us,first
 bf,bu,bs=raw(a);hf,hu,hs=raw(b);assert bu.keys()==hu.keys()
 assert dict(Counter(f['kind'] for f in bf.values()))==delta['raw_findings']['totals']['before']
 assert dict(Counter(f['kind'] for f in hf.values()))==delta['raw_findings']['totals']['after']
 stop=lambda f:(f['where'],f['kind'],f['reason']) if f else None
 changed={u for u in bu if stop(bs.get(u))!=stop(hs.get(u))}
 assert len(changed)==len(delta['raw_findings']['unit_first_stop_delta'])
 assert len(hf.keys()-bf.keys())==len(delta['raw_findings']['new_findings'])
 if label=='entry':
  assert a[0]['latent_mode']=='first-error' and len(a)==1+len(a[0]['sources']) and len(b)==len(a)
  assert set(a[0]['sources'])=={r['file'] for r in a[1:]}=={r['file'] for r in b[1:]}
  assert {str(Path(p).relative_to(tree)) for p in a[0]['sources']}==set(expected)
  entry=a;entry_delta=delta
for name in ['named.py','full-sites.py']:
 rc=subprocess.run([sys.executable,str(out/name)],cwd=repo).returncode
 print(name,'child_exit='+str(rc),flush=True);assert rc==0
lines=['# Stacked findAncestor: paired latent census','','Task 448a51k. Report only. No compiler or adaptation change is committed.','',
 '| Side | Compiler pin |','| --- | --- |',f'| Parent and apply | `{base}` |',f'| Head | `{head}` |','',
 'The remote optimizer/find-ancestor-result-stacked tip was checked before setup and remained 63c11488. Its parent is road (B), 1060360b, on b5c0aa64 and main 3fbe3a2a. Both internal/lower and the rest of each compiler are built from their own pin. The pinned checker is shared unchanged. [Input pins and carried-directory hashes](input-pins.json) and [binary provenance](provenance.json) retain the checks.','',
 'One scratch checkout of the parent received both complete directories stage3/adapt/99-correctness-core-or-predicate and stage3/adapt/99-correctness-core-multimap from eeb87728. Their bytes were checked against that commit before apply. Apply succeeded, exit 0, on TypeScript 6.0.3 at 050880ce59e30b356b686bd3144efe24f875ebc8. Both sides use this single adapted tree. [Patch set](patch-set.md), [apply log](apply.log), and [source hashes](source-hashes.json).','',
 f"The entry first-error comparison reaches {len(entry[0]['sources'])} files from src/tsc/tsc.ts, with zero checker diagnostics on each side. All census runs set LATENT_ASSERT_NO_OUTPUT=1 and exercise both no-output guards. Full-file runs use LATENT_FULL=1 and LATENT_ONLY_FILE set to the selected adapted file, with src/compiler as the loader root. Their historical raw measurement label says checker-rejected, but their checker_rejected field is false and diagnostics are empty. No native program is built or run.",'',
 'The method follows efb78625 stage3/census/find-ancestor-5a86e3b3/REPORT.md: own-pin guarded entry overlays, access-chain pairing, and separate focused full-unit inventories for the six nested checker callers. The specifically named core-sites-af42d3dc report was absent from the fetched history and the available typescript/core-sites tip; the available core-sites-4186581e runner was read, and the exact full-mode flags requested here were used.','',
 '## Entry counts','','| View | Cleared | New findings | Changed reasons | New sites | Unit first-stop changes | Regressions |','| --- | ---: | ---: | ---: | ---: | ---: | ---: |']
for view,title in [('raw_findings','Raw'),('distinct_reads','Distinct outermost access chains')]:
 d=entry_delta[view];lines.append('| '+' | '.join(map(esc,[title,*[len(d[k]) for k in ['lowered','new_findings','moved','new_sites','unit_first_stop_delta','regressions']]]))+' |')
lines+=['','Cleared means the old finding disappeared with no remaining finding at that source identity. Changed reasons retain another finding there. New findings include replacement reasons, so new and changed are not disjoint. Raw counts use exact raw-position keys, not event multiplicity. Distinct counts use the stock TypeScript 6.0.3 outermost access chain. A different first stop in an already-blocked unit is reported separately from regressions. The pair tool flags newly stopped units, newly finding refusal-scan sites and new panics. These observations do not prove native correctness.','',
 '[Raw parent stream](parent-entry-run.jsonl), [raw head stream](head-entry-run.jsonl), [paired JSON](entry-paired.json), and [paired ledger](entry-paired.md).','',
 '## Answers to the four questions','',
 '1. Yes. In both the full-file and focused probes, the attempted findAncestor overload/implementation units at 787 and 788 move from the prologue Refused at utilitiesPublic.ts:786:1, citing return undefined at 792:13, to no recorded finding. Exact old reason: `overload 1 of findAncestor result T | undefined cannot be served by implementation result Node | undefined (return at /workspace/scratch/fa-adapted/src/compiler/utilitiesPublic.ts:792:13 isn\'t the value its predicate checked)`. New first stop: none. The generic declaration at 786 is deferred on both sides.','',
 '2. No, the old helper-arrow stops are not reproduced. Neither full core stream records a finding in orPredicate2 or orPredicate3; their generic declarations remain deferred. In the focused head callers, the former checked-predicate-result ArrowFunction reason is absent at core.ts:2463:17 and 2467:17, and the four formerly blocked call sites have no local failed boundary or callee-helper finding. Parent probes at 4770, 4781 and 50077 stop earlier at findAncestor overload validation, so absence of a parent helper finding does not prove that helper body lowered. The complete core streams are byte-for-byte identical.','',
 '3. The exact caller coordinates and old/new stops are in the six-row table below. At 4751 the binder.ts:2258:13 source-slot certificate stop remains. At 48437 the checker.ts:48398:31 AliasDeclarationNode prologue stop remains. Calls at 4770, 4781 and 50077 lose the findAncestor return-at-795 refusal; 50086 loses its enclosing reading-result stop at checker.ts:50085:22. Those four head sites have no failed boundary containing the call. Earlier whole-unit stops still exist, as the focused table states; no whole-unit success is claimed.','',
 '4. Entry regression entries: '+str(len(entry_delta['distinct_reads']['regressions']))+' distinct, '+str(len(entry_delta['raw_findings']['regressions']))+' raw. The full-file regression and panic table below gives both additional comparisons. Changed first stops in already-blocked units are listed separately.','',
 '## Every unit whose first stop changes','']
for label in ['entry','utilitiesPublic','core']:
 d=json.loads((out/(label+'-paired.json')).read_text());changes=d['distinct_reads']['unit_first_stop_delta']
 lines += ['### '+label,'','| Unit | Location | Parent first stop | Head first stop | Interpretation |','| --- | --- | --- | --- | --- |']
 for r in changes:
  u=r['unit'];interpret='changed stop in blocked unit' if r['base'] and r['head'] else ('cleared first stop' if r['base'] else 'newly stopped unit')
  lines.append('| '+' | '.join(map(esc,[u.get('name') or u['kind'],loc(u['where']),why(r['base']),why(r['head']),interpret]))+' |')
 if not changes:lines+=['| None | | | | |']
 lines+=['']
lines+=['## Entry findAncestor status','','The generic overload declaration at 786 is deferred; its absence of a direct lowering stop is not a successful body probe. The non-generic overload and implementation units at 787 and 788 exercise the prologue.','','| Side | Unit | Status | Instantiated | First lowering stop |','| --- | --- | --- | --- | --- |']
for side in ['parent','head']:
 rows=load(side+'-entry-run');fs,us,stops=raw(rows)
 for u in sorted((u for u in us.values() if u.get('name')=='findAncestor' and '/utilitiesPublic.ts:' in u['where']),key=lambda u:u['where']):
  lines.append('| '+' | '.join(map(esc,[side,loc(u['where']),u['status'],u.get('instantiated',''),why(stops.get(u['where']))]))+' |')
lines+=['', (out/'named-sites.md').read_text(), (out/'full-sites.md').read_text(), '## Regressions and panics','', '| Scope | Regressions, raw | Regressions, distinct | Parent panics | Head panics |','| --- | ---: | ---: | ---: | ---: |']
for label in ['entry','utilitiesPublic','core']:
 d=json.loads((out/(label+'-paired.json')).read_text());lines.append('| '+' | '.join(map(esc,[label,len(d['raw_findings']['regressions']),len(d['distinct_reads']['regressions']),len(d['raw_findings']['panics']['before']),len(d['raw_findings']['panics']['after'])]))+' |')
 for r in d['distinct_reads']['regressions']:lines+=['','- '+label+' '+r['type']+': '+why(r.get('finding') or r.get('head'))]
lines+=['','## Every census run: time and peak memory','','| Run | Own child exit | Wall seconds | Peak RSS MiB |','| --- | ---: | ---: | ---: |']
run_names=['parent-entry-run','head-entry-run','parent-full-utilitiesPublic','head-full-utilitiesPublic','parent-full-core','head-full-core']+[side+'-focus-'+str(i) for side in ['parent','head'] for i in sorted(json.loads((out/'focus-index.json').read_text()).values())]
for name in run_names:
 s=json.loads((out/(name+'.status.json')).read_text());m=json.loads((out/(name+'.time.json')).read_text());assert s['exit']==0 and not s['stopped']
 lines.append(f"| {name} | {s['exit']} | {s['wall_seconds']:.3f} | {m['peak_rss_kib']/1024:.1f} |")
lines+=['','RSS is Linux child getrusage peak RSS, converted from KiB to MiB. Wall times include up to one second of wrapper polling. Entry runs overlap and use identical GOMEMLIMIT=5GiB and GOGC=70 settings; these are shared-host observations, not benchmarks. Each run retains its exact command, own exit, unit timings and log. The wrapper terminates a process group after 30 minutes without output or unit progress, and cleans up remaining group members after exit.','',
 '## Validation and checkpoints','','Setup, npm ci, apply and all six guarded binary builds completed successfully. All 15 existing paired-census tests passed. Source hashes, equal measurement headers, equal unit identities, raw finding totals, new-finding counts, first-stop changes, compiler pins and binary hashes were independently checked by finish.py. Go evidence uses .go.txt. No measured source was removed or changed.','',
 'The initial controller stopped before any census when the entry binary output path collided with its overlay directory. The already built binaries were moved from the directory to explicit -entry-bin paths, and the runner was corrected. This is retained in controller-initial.log; no partial census from that attempt contributes to counts.','',
 'A checkpoint was pushed after successful apply and after every census run, to separate typescript/find-ancestor-stacked-census-checkpoint-* refs. Each checkpoint has a JSON receipt and push log beside this report. This allows the final branch to contain one report-only commit directly on 1060360b without force-pushing. No main ref is pushed.','']
report='\n'.join(lines).replace('\u2013','-').replace('\u2014','-')
(out/'REPORT.md').write_text(report)
(out/'audit.json').write_text(json.dumps(dict(source_hashes_unchanged=True,source_files=len(expected),compiler_pins_and_binaries_verified=True,headers_equal=True,unit_identities_equal=True,raw_totals_and_first_stops_recounted=True,entry_files=len(entry[0]['sources']),checker_diagnostics=0,census_runs=len(run_names),pair_tests=15),indent=2)+'\n')
print('report and audit complete child_exit=0',flush=True)
