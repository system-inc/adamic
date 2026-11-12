import collections,csv,gzip,hashlib,json,pathlib,shutil,subprocess,sys
R=pathlib.Path(__file__).resolve().parent;E=R/'evidence';S=pathlib.Path('/workspace/lwp2');E.mkdir(exist_ok=True)
def rows(p):return [json.loads(x)for x in p.read_text().splitlines()if x.strip()]
def norm(w):return 'src/'+w.split('/src/')[1] if '/src/'in w else w
def archive(name):
 d=S/name;o=E/name;o.mkdir(exist_ok=True);status=json.loads((d/'status.json').read_text());assert status['child_exit_code']==0 and not status['deadline_expired'] and status['wall_seconds']<1800
 shutil.copy2(d/'status.json',o/'status.json')
 for p in d.iterdir():
  if p.name=='status.json':continue
  if p.is_file():(o/(p.name+'.gz')).write_bytes(gzip.compress(p.read_bytes(),mtime=0))
def git(*args):
 p=subprocess.run(['git',*args],capture_output=True,text=True)
 if p.returncode:print(p.stderr);sys.exit(p.returncode)
 return p.stdout.strip()
sha=git('rev-parse','origin/main');step=git('rev-parse','b4884718');fix=git('rev-parse','9ae9f9888')
for commit in [step,fix]:git('merge-base','--is-ancestor',commit,sha)
provenance={'main':sha,'step1':step,'view_entry_fix':fix,'loom_change':'chg_p17rtqz1fksx39vh60errayw4r','lower_tree':git('rev-parse',sha+':internal/lower'),'worktree_lower_tree':git('rev-parse','HEAD:internal/lower'),'original_report_also_contains_both':True,'variants':{'traced':'shallow preflight probe; observations only','isolated3':'same source, preflight lower-state snapshot only'},'binaries':{}}
assert provenance['lower_tree']==provenance['worktree_lower_tree']
for b in ['trace2-census','trace3-census','whole','pin3','pin4']:
 p=S/b
 if p.exists():provenance['binaries'][b]=hashlib.sha256(p.read_bytes()).hexdigest()
(E/'provenance.json').write_text(json.dumps(provenance,indent=2)+'\n')
for n in ['traced','isolated3','whole-fixture','scoped-fixture','whole-predicate','scoped-predicate','isolated3-predicate','whole-literal','scoped-literal','whole-tsc']:archive(n)
for n in ['trace-all','isolated3-all','whole-fixture-trace','whole-predicate-trace','scoped-predicate-trace','isolated3-predicate-trace','whole-literal-trace','scoped-literal-trace']:
 p=S/(n+'.jsonl');(E/(n+'.jsonl.gz')).write_bytes(gzip.compress(p.read_bytes(),mtime=0))
for n in ['widened','predicate','literal-overload']:
 shutil.copy2(S/'fixture'/(n+'.ts'),E/(n+'.ts.txt'))
 stdout=S/'fixture'/('node.stdout'if n=='widened'else n+'.stdout');shutil.copy2(stdout,E/(n+'.node.stdout'))
trace=rows(S/'trace-all.jsonl');fail={norm(x['unit']):x for x in trace if 'read_where'in x};probes={norm(x['unit']):x for x in trace if x.get('event')=='discarded_preflight_probe'}
control=rows(S/'isolated3-all.jsonl');cf={norm(x['unit']):x for x in control if 'read_where'in x}
old=json.loads((R.parent/'evidence/new-stops.json').read_text());old={norm(x['unit']):x for x in old if x['pass_name']=='checkViewWidenedReads'}
assert set(old)==set(fail) and len(fail)==265
assert sum('read_where'in x for x in trace)==265
for run,count in [('traced',265),('isolated3',22)]:
 data=rows(S/run/'census.jsonl');assert data[0]['status']=='measurement' and not data[0]['checker_rejected'] and data[0]['latent_mode']=='first-error'
 rel=data[-1]['relation_overflow'];assert data[-1]['status']=='relation_report' and not rel['after_load'] and not rel['end'] and rel['cache_sizes']['flaggedPairs']==0
 units=[u for row in data for u in row.get('units',[])];assert len(units)==265 and {norm(u['where'])for u in units}==set(old) and all(u['status']=='attempted'for u in units)
 findings=[f for row in data for f in row.get('findings',[])];assert len(findings)==count and all(f['phase']=='whole_pass:checkViewWidenedReads' and f['reason']==old[norm(f['unit'])]['reason']for f in findings)

names=json.loads((S/'names.json').read_text());ledger=[];split=collections.Counter();contributions=collections.Counter()
for unit,x in fail.items():
 real=x['entries'];kind='orphan preflight predicate registration'if not real else ', '.join(sorted({e['kind']for e in real}))
 assert kind in ['orphan preflight predicate registration','CheckedCast','predicate narrowing']
 split[kind]+=1
 sources=real or probes[unit]['entries']
 for e in sources:e['source_function']=names.get(e['where'],{}).get('function');e['source_ast_kinds']=names.get(e['where'],{}).get('node_kinds_at_start',[])
 for fn in {e.get('source_function')for e in sources if e.get('source_function')}:contributions[(kind,fn)]+=1
 if not real:assert x['field']=='kind'and x['receiver']=='Node'and x['unknown']and probes[unit]['shares_result'] and all('TypePredicate'in e['source_ast_kinds']for e in sources)
 assert not x['overload_helpers']
 ledger.append({'unit':unit,'name':old[unit]['name'],'old_reason':old[unit]['reason'],'read_where':norm(x['read_where']),'classification':kind,'recorded_entries':real,'discarded_probe_entries':probes.get(unit,{}).get('entries',[]),'flow_unknown':x['unknown'],'flow_reasons':x['reasons'],'flow_sites':x['sites'],'inside':x['inside'],'view_origins':x['view_origins'],'isolated_probe_still_stops':unit in cf,'overloaded_function':None})
assert split=={'orphan preflight predicate registration':243,'CheckedCast':8,'predicate narrowing':14}
assert len(cf)==22 and set(cf)=={u for u,x in fail.items()if x['entries']}
assert all(not x['shares_result']for x in control if x.get('event')=='discarded_preflight_probe')
for u,x in cf.items():assert norm(x['read_where'])==norm(fail[u]['read_where'])
(E/'widened-ledger.json').write_text(json.dumps(sorted(ledger,key=lambda x:x['unit']),indent=2)+'\n')
with (E/'entry-functions.csv').open('w')as f:
 w=csv.writer(f,lineterminator='\n');w.writerow(['classification','source_function','units_nonexclusive']);w.writerows((k,fn,n)for(k,fn),n in sorted(contributions.items()))
assert json.loads((S/'whole-predicate/census.jsonl').read_text())['kind']=='accepted'
literal=json.loads((S/'whole-literal/census.jsonl').read_text());assert literal['reason']=="a read of field kind through type Node, which may hold a checked view's value"
assert {x['event']for x in rows(S/'whole-literal-trace.jsonl')if 'event'in x}>={'production_tail_enter','production_widened_reached'}
assert (S/'fixture/predicate.stdout').read_text()=='true\n'and (S/'fixture/literal-overload.stdout').read_text()=='1\n'
summary={'main':sha,'units':265,'origin_split':dict(split),'overload_functions':{},'original_kind_Node':243,'isolated_kind_Node':0,'isolated_remaining':len(cf),'all_shared_probe_events':len(probes),'whole_predicate':'accepted','whole_literal_overload':literal['reason']}
(E/'summary.json').write_text(json.dumps(summary,indent=2)+'\n')
text=f'''# Widened-read attribution on current main

Main: **`{sha}`**. It contains #z5d0jye step 1 **`{step}`** and the view-entry fix **`{fix}`**, `chg_p17rtqz1fksx39vh60errayw4r`. The original REPORT base a93ac0b94 already contained both. The same worktree now merges current main without rewriting the prior report branch. Compiler sources match this main's lower tree; only scratch overlays add observations.

**The 243 are census state-leak artifacts, not overload-check entries.** All 265 historical widened-read units reproduce on this main's own adapted checker-clean tsc entry. None has an overload-check helper. The 243 `.kind`/`Node` failures have **no recorded viewSources entry at all**.

## Split of all 265

| Requested entry category | Units | What the trace establishes |
|---|---:|---|
| (a) overload check call | 0 | No overloadCheckCalls helper in any of the 265 products. Per overloaded function: none; skipPartiallyEmittedExpressions, skipParentheses and getThisContainer each contribute zero to this set. |
| (b) CheckedCast | 8 | Recorded cast entries; source expressions and targets listed per unit. |
| (c) recorded predicate narrowing | 14 | Recorded predicate entries; all source functions and locations listed per unit. |
| (d) orphan checked-field registration from a discarded predicate preflight | 243 | All are the reported `.kind` through Node words. The registry survives, but its view records do not. There is no IR view entry to name. Each discarded registration's actual TypePredicate source location, function and target is named in the ledger. |
| Total | 265 | |

By *intended registration origin* the predicate family is 257 (14 recorded + 243 orphaned); it is wrong to present the latter 243 as actual flow entries. `followup/evidence/entry-functions.csv` gives counts per source function (nonexclusive when a unit has several entries); `followup/evidence/widened-ledger.json` preserves every one of the 265 units, exact stop/read location, all recorded and discarded entry sources, types, source-function names, unknown-flow reasons and control outcome. There are no overloaded-function counts to apportion: category (a) is empty.

The pass intentionally merges real view entries into sentinel sites and weighs all `viewLie`s jointly. Multiple source entries in a unit are reported as a set, not as a fabricated unique producer. Unknown flow is distinguished from a reached entry. All 8 cast units and 14 recorded-predicate units remain after isolation; this does not certify them as whole-tsc production regressions.

## Cause and discriminating control

`stage3/census/latent/lower.go.txt`'s `latentDeferredStop` says it probes without changing unit state, but uses `probe := *l`. Only `probe.refusal` is reset. `probe.result == l.result`, so `refuseNode` → predicate registration → `viewSchema` writes CheckedFields into the real unit IR. Appending to `probe.viewSources` changes the copy's slice header; the original unit's header stays empty. The trace records the shared result pointer and the discarded registrations, rather than inferring an overload from a checked field name.

At `checkViewWidenedReads`, the keyed field-name registry is nonempty, receiver flow is Unknown, and `viewVouched` on an empty lies list returns false. Therefore `inside && !viewVouched(...)` refuses a raw `.kind` read with no view entry. This is the direct path for all 243. Source names are TypePredicate AST nodes, including registrations whose narrows flag was never finalized.

A scratch-only control replaces that shallow preflight copy with `*l.latentSnapshot()`. It changes no production check and keeps the same 265 unit selection, body-lowering path and tail order. **All 243 disappear; exactly the other 22 remain at the same read sites.** All 251 logged preflight-probe events switch from shared to separate result pointers. The first attempted control patch failed to match formatted whitespace and was rejected as invalid; only the asserted, corrected snapshot control (`isolated3`) supplies these results. No census/production source fix is committed here.

## Production whole-program answer, held to Node

The whole probe calls real `load.Load` and **production `LowerWithOptions(..., Options{{}})`**. Its overlay only logs tail entry and widened-pass entry/failure; it does not disable, skip or replace a pass, and it emits no native program. This uses the fixture alternative in the task, not a claim that every compiler root was swept.

| Whole root | Node output | Production whole | Scoped prototype | Isolated-preflight control |
|---|---|---|---|---|
| predicate.ts: `isIdentifier(node: Node): node is Identifier`, body reads node.kind | true | accepted; tail and widened pass reached | same `.kind`/`Node` NotYet at 3:64 | accepted |
| widened.ts: Node-preserving narrow branded overload, result read through Node | 1 | accepted; tail and widened pass reached | accepted | not needed |
| literal-overload.ts: Broad.kind:number → narrow overload returning NarrowNode → Node.kind:1&#124;2 | 1 | NotYet at 11:21 | identical reason and location at 11:21 | not needed |

The last fixture's exact reason is:

> a read of field kind through type Node, which may hold a checked view's value

Its production trace identifies **tag_overload_check**, source call 9:17, source Broad, target NarrowNode, known entry flow (site 1, Unknown=false), and an incompatible held field type. Thus a genuine overload-entry widened read **does** stop in whole-program lowering with the same words. The preceding predicate fixture demonstrates that the 243-shaped scoped failure can instead be an artifact: production accepts it, and isolating the census preflight removes it. Do not undo the overload view-entry fix to silence this census artifact.

The actual tsc entry was also tried in production. It stops earlier at `sys.ts:1654:13`, `NotYet: node:inspector.Session.post`, before any tail marker. I do not infer that no other tsc root can reach the tail; the decisive comparisons above are complete, small whole programs. The failed broader-object overload pilot stopped before the widened pass and was not used as the qualifying fixture.

## Evidence and checks

All final children exit 0 with their actual status retained; each has a 1,700-second bound. Exit 0 means the measurement completed, not that lowering accepted. Independent whole-file refusal scans are omitted on both scoped variants, as in REPORT.md; production whole-program runs retain their real preflights. The historical 265 source-unit locations exactly match the current 265 measured locations. The two scoped streams have complete relation trailers without overflow. `followup/report.py` asserts exhaustive unit identity, the split, all 243 removals, the unchanged 22, whole-tail markers and Node outputs before rendering this report.

Raw streams, status records, provenance, source-function metadata and fixture source are under `followup/evidence/`. Go observation code is `.go.txt`. The companion [PIN.md](PIN.md) records full-mode generator results separately, including any body wall; it is not folded into these counts.
'''
(R.parent/'WIDENED.md').write_text(text)
print(json.dumps(summary,indent=2))
