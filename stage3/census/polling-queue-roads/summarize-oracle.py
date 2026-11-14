from pathlib import Path
import json,shutil,collections
root=Path(__file__).parent;p=root/'evidence'
source=Path('/tmp/polling-oracle-results-complete')
report=json.loads((source/'report.json').read_text())
expected=json.loads((root.parents[1]/'lane/expected.json').read_text())
assert report['status']==expected['oracle_status']
assert report['counts']==expected['platforms']['linux/x64/v24.19.0']
assert {k:v['exit']for k,v in report['phases'].items()}==expected['phase_exits']
assert [f['title']for f in report['failures']]==expected['failed_tests']
assert report['baseline_diffs']==expected['baseline_diffs']
shutil.copytree(source,p/'oracle',dirs_exist_ok=True)
shutil.copytree('/tmp/polling-oracle-counts-complete',p/'oracle-counts',dirs_exist_ok=True)
rows=[json.loads(f.read_text())for f in (p/'oracle-counts').glob('*.json')]
testrows=[r for r in rows if '/tmp/polling-oracle-tree-complete/built/local/run.js'in r['loaded']]
assert len(testrows)==9,('counter did not retain parent and all eight test workers',len(testrows))
props=collections.Counter()
for r in rows:props.update(r['byProperty'])
summary={'oracle':{k:report.get(k)for k in ['status','counts','workers','runners','tests','node','phases','baseline_diffs']},'matches_expected_lane_oracle_outcome':True,'recorded_processes':len(rows),'test_bundle_processes':len(testrows),'factory_calls':sum(r['factories']for r in rows),'queue_operations':sum(r['operations']for r in rows),'byProperty':dict(props),'test_factory_calls':sum(r['factories']for r in testrows),'test_queue_operations':sum(r['operations']for r in testrows),'definition':'Every queue string-key get/set, including initialization, push lookup and its internal index/length operations; excludes identity tests and passing/returning/binding a queue.'}
assert summary['queue_operations']==sum(props.values())
assert props.get('set:pollingInterval',0)==summary['factory_calls']
(p/'oracle-summary.json').write_text(json.dumps(summary,indent=2)+'\n')
text=(root/'REPORT.md').read_text();note=f"## Original-tree lane oracle coverage\n\ncreatePollingIntervalQueue ran **{summary['factory_calls']:,} times**, with **{summary['queue_operations']:,} queue property operations** ({summary['test_factory_calls']:,} / {summary['test_queue_operations']:,} in the test bundle processes). The counter retained the test parent and all eight workers. Counts include the three field initializations per queue; see the [per-property totals and oracle result](evidence/oracle-summary.json), [raw per-process counters](evidence/oracle-counts/) and [full oracle report](evidence/oracle/report.json).\n\n"
if not summary['factory_calls']:note+='The oracle never exercised dynamic-priority queue creation in this run; zero operations are coverage absence, not behavioral evidence. The Node witness supplies explicit queue-transition coverage.\n\n'
note+=f"Oracle counts: `{json.dumps(report.get('counts',{}),sort_keys=True)}`; runners `{report.get('runners')}`, workers `{report.get('workers')}`, test filter `{report.get('tests')}`. Oracle status is `{report.get('status')}`; the sole failure is the expected `api/typescript.d.ts` public-API acknowledgment. Counts, phase exits, failure title and baseline diff match stage3/lane/expected.json for Linux x64 / Node v24.19.0; this checks the oracle outcome, not the outer lane declaration guard.\n\n"
text=text.replace('## What this measurement cannot see',note+'## What this measurement cannot see');(root/'REPORT.md').write_text(text)
print(json.dumps(summary,indent=2))
