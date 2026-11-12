"""Read-only audit of committed evidence; no external scratch tree is required."""
import collections,gzip,json,pathlib
R=pathlib.Path(__file__).resolve().parent;E=R/'evidence'
def rows(p):
 with (gzip.open(p,'rt')if p.suffix=='.gz'else p.open())as f:return [json.loads(x)for x in f if x.strip()]
def norm(s):return 'src/'+s.split('/src/')[1]if '/src/'in s else s
def status(d):
 s=json.loads((d/'status.json').read_text());assert s['child_exit_code']==0 and not s['deadline_expired']and s['wall_seconds']<1800
 return s
old={norm(x['unit']):x for x in json.loads((R.parent/'evidence/new-stops.json').read_text())if x['pass_name']=='checkViewWidenedReads'}
ledger=json.loads((E/'widened-ledger.json').read_text());assert len(ledger)==265 and {x['unit']for x in ledger}==set(old)
counts=collections.Counter(x['classification']for x in ledger);assert counts=={'orphan preflight predicate registration':243,'CheckedCast':8,'predicate narrowing':14}
for variant,n in [('traced',265),('isolated3',22)]:
 d=E/variant;status(d);r=rows(d/'census.jsonl.gz');assert r[0]['latent_mode']=='first-error'and not r[0]['checker_rejected'];assert r[-1]['status']=='relation_report'
 rel=r[-1]['relation_overflow'];assert not rel['after_load']and not rel['end']and rel['cache_sizes']['flaggedPairs']==0
 u=[u for row in r for u in row.get('units',[])];assert len(u)==265 and {norm(u['where'])for u in u}==set(old)
 f=[f for row in r for f in row.get('findings',[])];assert len(f)==n and all(f['phase']=='whole_pass:checkViewWidenedReads'and f['reason']==old[norm(f['unit'])]['reason']for f in f)
raw=rows(E/'trace-all.jsonl.gz');fail=[x for x in raw if 'read_where'in x];assert len(fail)==265 and all(not x['overload_helpers']for x in fail)
orphans=[x for x in ledger if x['classification'].startswith('orphan')]
assert all(not x['recorded_entries']and not x['isolated_probe_still_stops']and all('TypePredicate'in e['source_ast_kinds']for e in x['discarded_probe_entries'])for x in orphans)
assert all(x['isolated_probe_still_stops']for x in ledger if x not in orphans)
assert len([x for x in raw if x.get('shares_result')is True])==251
control=rows(E/'isolated3-all.jsonl.gz');assert len([x for x in control if x.get('shares_result')is False])==251
for variant in ['whole-fixture','whole-predicate','whole-literal','whole-tsc','scoped-fixture','scoped-predicate','scoped-literal','isolated3-predicate']:status(E/variant)
assert rows(E/'whole-predicate/census.jsonl.gz')[0]['kind']=='accepted'
assert rows(E/'whole-fixture/census.jsonl.gz')[0]['kind']=='accepted'
whole=rows(E/'whole-literal/census.jsonl.gz')[0];scoped=[x for r in rows(E/'scoped-literal/census.jsonl.gz')for x in r.get('findings',[])if x.get('phase')=='whole_pass:checkViewWidenedReads'];assert len(scoped)==1 and scoped[0]['reason']==whole['reason']and scoped[0]['where']==whole['where']
assert whole['reason']=="a read of field kind through type Node, which may hold a checked view's value"
events=rows(E/'whole-literal-trace.jsonl.gz');assert {x.get('event')for x in events}>={'production_tail_enter','production_widened_reached'}
origin=next(x for x in events if 'entries'in x);assert origin['entries'][0]['function']=='tag'and origin['entries'][0]['kind']=='overload check call'and not origin['unknown']
assert (E/'predicate.node.stdout').read_text()=='true\n'and (E/'literal-overload.node.stdout').read_text()=='1\n'
for label in ['parent','87028b08','isolated-parent','isolated-child','old-stop-mutant']:
 d=E/'pin'/label;status(d);r=rows(d/'census.jsonl.gz');assert r[0]['latent_mode']=='full'and r[-1]['status']=='relation_report'
 events=rows(d/'events.jsonl.gz');sites=[x for x in events if x['event']=='statement_end'];assert len(sites)==5 and sum(x.get('tail_result')=='admitted'for x in sites)==4
 blocked=[x for x in sites if 'body_error'in x];assert len(blocked)==1 and blocked[0]['call'].endswith('21783:34')and 'Map or Set interface'in blocked[0]['body_error']
 joined=next(x for x in events if x['event']=='combined')
 if label=='old-stop-mutant':assert 'a step names the one function'in joined['error']
 else:assert joined['result']=='admitted'and joined['saved_sites']==4
 origins=next(x for x in events if x['event']=='generatorFunctions'and x['active']=='combined-five-source-statements');assert len(origins['functions'])==3 and {x['yields']for x in origins['functions']}=={'an object'}
assert rows(E/'pin/parent/events.jsonl.gz')==rows(E/'pin/87028b08/events.jsonl.gz')
assert rows(E/'pin/isolated-parent/events.jsonl.gz')==rows(E/'pin/isolated-child/events.jsonl.gz')
print(json.dumps({'units':265,'split':dict(counts),'243_removed_by_isolation':True,'whole_program_fixtures_verified':True,'pin_four_admitted_one_body_blocked':True,'old_stop_mutant_red':True},indent=2,sort_keys=True))
