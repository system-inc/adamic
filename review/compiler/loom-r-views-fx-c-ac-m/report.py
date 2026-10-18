from pathlib import Path
import json,subprocess,collections
out=Path(__file__).resolve().parent;meta=json.loads((out/'base.json').read_text());checks=json.loads((out/'checks.json').read_text());counts=json.loads((out/'counts.json').read_text());mutants=json.loads((out/'mutant-results.json').read_text());scan=json.loads((out/'final.json').read_text());headers=json.loads((out/'header-mutants.json').read_text());events=[];skips=[]
for p in sorted(out.glob('*.log')):
 for line in p.read_text().splitlines():
  try:e=json.loads(line)
  except ValueError:continue
  if not isinstance(e,dict) or 'Test' not in e or e.get('Action') not in ('pass','fail','skip'):continue
  row=dict(log=p.name,package=e.get('Package'),test=e['Test'],seconds=e.get('Elapsed',0),result=e['Action']);(skips if e['Action']=='skip' else events).append(row)
groups=set()
for e in events:
 parts=e['test'].split('/')
 for n in range(1,len(parts)):groups.add((e['log'],e['package'],'/'.join(parts[:n])))
for e in events:e['leaf']=(e['log'],e['package'],e['test']) not in groups
leaves=[e for e in events if e['leaf']];assert leaves and all(e['seconds']<60 for e in leaves)
assert all(c['exit']==(1 if c['label']=='ir-all' else 0) for c in checks),checks
assert len(mutants)==12 and all(m['verified'] for m in mutants),mutants
assert len(scan)==63 and len(headers)==2 and all(h['exit']==1 for h in headers)
assert len(counts['rows'])==18 and all(r['identical_to_source'] for r in counts['rows']) and not counts['removed']
paths=json.loads((out/'paths.json').read_text())
for path in paths:assert Path(path).read_bytes()==subprocess.check_output(['git','show',meta['source']+':'+path]),path
(out/'test-times.json').write_text(json.dumps(events,indent=2));(out/'skips.json').write_text(json.dumps(skips,indent=2));(out/'test-times.tsv').write_text('log\tpackage\ttest\tseconds\tresult\tleaf\n'+''.join(f"{e['log']}\t{e['package']}\t{e['test']}\t{e['seconds']}\t{e['result']}\t{e['leaf']}\n" for e in events))
summary=dict(collections.Counter(r['status'] for r in scan));(out/'a-check-summary.json').write_text(json.dumps(summary,indent=2))
report=f'''Built: Views fix-forward C recut as one commit on current main, retaining both sides of the counts conflict.
Commits: source {meta['source']}; base {meta['base']}; delivery SHA is the pushed compiler/loom-r-views-fx-c-ac-m tip.
Commands and outputs: full requested compiler package trees, fixture consumers, guards, vet, gofmt and Linux counts passed; commands and exits below.
Mutants: eleven semantic faults caught, one green isolation control passed, and both header predicate mutants caught.
Not covered: full repository gate, all other compiler-reader packages, opt-in lanes in skips.json, ahra/PR/Loom submission or queue changes.

'''+(out/'resolution.md').read_text()+'\n\n'
report+='All source non-count additions/deletions are identical to the source commit (source-delta.json). All 63 .a files, including review evidence, remain byte-identical to that source. No new header or positional expectation correction is required. Native runtime source and recut lines match exactly (runtime-lines.json); additional runtime C/H lines for clearance: NONE.\n\n'
report+='Counts were regenerated once from main, never hand-merged. Exactly 15 additions and 3 existing-row updates, all with source-identical final values; no removed row and no unrelated change. The existing changes are node-indicator-good.a retains 15 to 14; callable-producer-wrong.a changes from 2/0/5/4/2/0 to 3/3/6/11/3/0; undefined-read-write.a retains 5 to 4. Exact complete table delta follows below. Main\'s callback widening and weak-reference rows survive unchanged.\n\n'
report+='Validation uses GOMAXPROCS=4, ADAMIC_GATE_UNCACHED=1, -count=1, -parallel 4 and -p 1, with JSON output written directly to logs. Whole package commands have -timeout 12m and an 840-second outer limit; explicit reader/maps and parallel guards have -timeout 90s. No test was removed, weakened or rerun after exceeding its leaf budget. test-times.json/tsv distinguishes grouping parents from leaves and retains all measured leaf times, including deliberate mutation failures. Full oracle holds source Node, native release, ASan/UBSan/LeakSanitizer, backend JavaScript and recorded counts, including the source\'s producers, union results and snapshot fixtures.\n\n'
report+='Full internal/ir includes SortedMap and call-target readers; internal/lower includes native-agreement registration; internal/oracle includes exact runtime-set witnesses. TestCallTargetReaders and TestEveryMapAProgramHoldsIsSorted are also recorded with the requested 90-second guard limit. The cmd/adamic-gate parallel guard passed. Full internal/javascript, flow/fresh fixture consumers and stage3/fixtures passed. No new compiler-consuming Go package, IR Program map, runtime feature set or product recipe was introduced.\n\n'
report+='A-check: the unchanged README-pinned Gate.aCheck invokes this branch\'s actual compiler on every one of the 63 changed .a paths, including review files. scan.py retains actual exit statuses, full diagnostics, source hashes and the predicate\'s judgments. Results: '+json.dumps(summary)+'. The four existing refusal headers remain correct. Removed-header and wrong-reason mutants use the README\'s explicitly documented replay of the current compiler\'s captured real diagnostic; both fail the unchanged predicate and restore the real review witness byte for byte. This is a predicate proof, not a claim to rerun compiler diagnostics on the temporary header variants.\n\n'
report+='Mutants: source\'s eleven corrected faults and one isolation control run as precise changes over the current files, preserving main, with every original byte restored in finally blocks. Only test assertion/behavior failures count; Go/C build errors and timeouts do not. Separate read-through Go cache writes for the mutant phase use root-volume scratch while reading immutable successful actions from the positive run\'s cache (mutant-cache.json), avoiding a second gigabyte export product and the earlier unit\'s disk-full failure. The packed snapshot fault is held directly by payload bytes; the combined retain fault is held by the Node/native snapshot fixture; the green control proves the retain guard masks a stale payload in isolation. Exact patches, commands, exits and catchers follow.\n\n'
report+='Setup: GOPROXY=https://proxy.golang.org|direct; timeout 600 bash cloud/setup.sh --wasi-sdk; source /workspace/adamic-tools/env.sh. Linux clang 20.1.8 (SDK clang 20.1.8-wasi-sdk), Go 1.27.1, Node 24.19.0; nproc=5, cpu.max=400000 100000 (four CPUs). Exact timing lines:\n\n```text\n'+'\n'.join(line for line in (out/'setup.log').read_text().splitlines() if line.startswith('setup:'))+'\n```\n\n'
report+='Commands and results:\n\n'+'\n'.join(f"- {c['label']}: exit {c['exit']}, {c['seconds']}s; `{' '.join(c['command'])}`." for c in checks)+'\n\n'
report+='Every changed count row (A/F/retain/release/peak/region):\n\n```text\n'+'\n'.join(('before: '+r['before']+'\n' if r['before'] else 'before: absent\n')+'after:  '+r['after'] for r in counts['rows'])+'\n```\n\n'
report+='Mutant results:\n\n'
for m in mutants:report+=f"- {m['name']}: exit {m['exit']}, expected green={m['expected_green']}, verified={m['verified']}, {m['seconds']}s; "+', '.join(e['Test'] for e in m['catchers'])+'.\n'
report+=f"\nMeasured leaves: {len(leaves)}; slowest {max(e['seconds'] for e in leaves)} seconds.\n"
(out/'REPORT.md').write_text(report);print('Report ready',summary,'max leaf',max(e['seconds'] for e in leaves))
