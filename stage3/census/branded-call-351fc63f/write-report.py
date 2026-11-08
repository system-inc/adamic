"""Render measured evidence, leaving incomplete entry comparisons explicitly pending."""
import json,re,subprocess,sys
from pathlib import Path
out=Path(__file__).resolve().parent;repo=out.parents[2]
sys.path.insert(0,str(repo/'stage3/census/pair'))
import pair

def esc(s):return str(s).replace('|','&#124;').replace('\n',' ')
def loc(s):return s.split('/src/',1)[-1] if s else ''
def ready(name):
 p=out/(name+'.status.json')
 return p.exists() and json.loads(p.read_text())['exit']==0

def delta(label,scope,base,head):
 if not ready(base) or not ready(head):return None
 p=out/f'{label}-{scope}-pair.json'
 if not p.exists():
  d=pair.compare(pair.load(out/(base+'.jsonl')),pair.load(out/(head+'.jsonl')),Path('/tmp/branded-'+label),repo/'stage3/api/node_modules/typescript')
  p.write_text(json.dumps(d,indent=2)+'\n');(out/f'{label}-{scope}-pair.md').write_text(pair.report(d))
 return json.loads(p.read_text())

subprocess.run([sys.executable,str(out/'named-ledger.py')],check=True)
subprocess.run([sys.executable,str(out/'core-ledger.py')],check=True)
entries={a:delta(a,'entry',f'{a}-parent-entry',f'{a}-head-entry') for a in ['A','B']}
cores={a:delta(a,'core',f'{a}-parent-core',f'{a}-head-core') for a in ['A','B']}
for a in ['A','B']:
 for i in range(1,5):delta(a,f'focus-{i}',f'{a}-parent-focus-{i}',f'{a}-head-focus-{i}')
delta('A','serializer','A-parent-serializer','A-head-serializer')
complete=all(entries.values())
lines=['# Paired latent census: branded call 351fc63f against b5c0aa64','', 'Task fx9waqy. Compiler request, November 8, 2026. Report only; no compiler or adaptation change lands.','']
if (out/'STOP_REASON.txt').exists():lines+=['Stopped: '+(out/'STOP_REASON.txt').read_text().strip(),'']
if complete:
 regs=sum(len(d['distinct_reads']['regressions']) for d in entries.values())
 lines+=[f'Both entry pairings completed. Normalized entry regression entries: **{regs}**. The named expected improvements are not demonstrated by the core and focused scopes; the exact earlier stops are below.','']
else:lines+=['**Entry comparisons are still pending. No complete entry regression verdict is claimed.** Full-core and focused observations have completed.','']
lines+=['## Pins and applied inputs','', '| Role | Pin |','| --- | --- |', '| Parent compiler and A apply | `b5c0aa648a7a35b2f17899895d009eabd4147000` |', '| Head compiler | `351fc63fa2573036e194656fb92c35fe03858afe` |', '| Parent main | `3fbe3a2ab40f22f8837a66daaad3b5fb845376e7` |', '| B adapter source | `eeb87728b743235cf1df68dfed17241f63416734` |', '| TypeScript 6.0.3 | `050880ce59e30b356b686bd3144efe24f875ebc8` |','',
 'A uses the unmodified parent apply. B uses a separate parent checkout carrying exactly `99-correctness-core-or-predicate` and `99-correctness-core-multimap` from the adapter pin. Both applies succeeded; B needed no pin repair. The two 81-source measured graphs differ only in `src/compiler/core.ts` and `src/compiler/checker.ts`. Every completed run has zero checker diagnostics.', '',
 'Compiler and measurement source bytes match their own pins. Both compilers use the same pinned checker dependency `06af2e399c6c1fb93fd570b774027942747400b4`. [Provenance](provenance.json), [input comparison](input-comparison.json), [A hashes](A-source-hashes.json), [B hashes](B-source-hashes.json), and [patch sets](A-patch-set.md) ([B](B-patch-set.md)) retain the evidence.','',
 '## Entry first-error pairing','',
 'These are the guarded meter entry driver over `src/tsc/tsc.ts`, with all 81 reached source files. `LATENT_ASSERT_NO_OUTPUT=1` is required. No speculative, mutant or forced-generic mode is used. Raw below means pair.py raw-position findings, deduplicated by its exact key; distinct means its outermost-access-chain normalization. New findings include replacement reasons and therefore are not disjoint from changed findings. New sites and first-stop changes are shown separately.', '',
 '| Pairing | View | Cleared | New findings | Changed findings | New sites | First-stop unit changes | Regression entries |', '| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: |']
for a,d in entries.items():
 for v,title in [('raw_findings','Raw'),('distinct_reads','Distinct')]:
  if d:
   r=d[v];vals=[len(r[k]) for k in ['lowered','new_findings','moved','new_sites','unit_first_stop_delta','regressions']]
  else:vals=['pending']*6
  lines.append('| '+' | '.join(map(esc,[a,title,*vals]))+' |')
lines+=['','A previously finding-free unit that now stops, a previously clear scan site that now finds, and a new panic are regressions. A different first stop in an already-blocked unit is reported separately. Normalized identities control the site judgment.','', '| Pairing | Unit | Parent first stop | Head first stop | Classification |','| --- | --- | --- | --- | --- |']
for a,d in entries.items():
 if not d:lines.append(f'| {a} | pending | | | |');continue
 us=d['raw_findings']['unit_first_stop_delta']
 if not us:lines.append(f'| {a} | No unit first stop moved | | | |')
 for u in us:
  def why(f):return (f['kind']+': '+f['reason']+' at '+loc(f['where'])) if f else 'finding-free'
  kind='already-blocked first-stop change' if u['base'] and u['head'] else 'cleared unit' if u['base'] else 'regression'
  lines.append('| '+' | '.join(map(esc,[a,u['unit'].get('name','')+' '+loc(u['unit']['where']),why(u['base']),why(u['head']),kind]))+' |')
lines+=['','Raw streams and complete deltas: [A parent](A-parent-entry.jsonl), [A head](A-head-entry.jsonl), [A paired JSON](A-entry-pair.json), [B parent](B-parent-entry.jsonl), [B head](B-head-entry.jsonl), [B paired JSON](B-entry-pair.json). A link to an unfinished stream is partial evidence, not a complete census.','', '## Full core.ts pairing','',
 'The reference procedure is preserved in [reference-run-census.py](reference-run-census.py) and its [report](reference-core-census-REPORT.md). Each compiler loads the full compiler-directory program and selects only the exact adapted `core.ts` using `LATENT_FULL=1`, `LATENT_ONLY_FILE` and `LATENT_ASSERT_NO_OUTPUT=1`. No speculative or replay flags apply to these four runs. The legacy checker-rejected label remains verbatim despite zero actual checker diagnostics.', '',
 'Within each pairing the complete parent and head streams are **byte-identical** ([hashes](core-equality.json)). Each has 270 units, 14 NotYet events, 4 Refused events and 12 recorder Boundary events. Thus cleared/new/changed are **0/0/0 in both raw and distinct views**, and no other core.ts site moves.','',
 'A does not contain `orPredicate2` or `orPredicate3`. In B, both helper declarations and all three remaining `or` declarations are deferred and uninstantiated. There is no recorded site to quote in any of those bodies. This is not evidence that a deferred body lowers.', '',
 (out/'core-sites.md').read_text().replace('# Core declaration site ledger','### Every requested declaration and recorded site').replace('## Pairing A','#### Pairing A').replace('## Pairing B','#### Pairing B'),
 'Full streams: [A parent](A-parent-core.jsonl), [A head](A-head-core.jsonl), [B parent](B-parent-core.jsonl), [B head](B-head-core.jsonl). Paired ledgers: [A](A-core-pair.json), [B](B-core-pair.json).','',
 '## Named checker callers and expected moves','',
 'The completed parent A entry census stops `createTypeChecker` before these nested sites. It stops at `checker.ts:2969:55`: `a value of type ParameterPropertyDeclaration`. Focused full-mode replay observations retain the complete entry graph and use the existing replay selector. The serializer is additionally observed through the real nongeneric `symbolTableToDeclarationStatements` initializer context, without forcing a generic declaration.', '',
 'The focused observations are unchanged on parent and head in both pairings. The hoped-for serializer `?.` wall is not reached: the caller prologue refuses `serializePropertySymbolsForClassOrInterface` before that body can be observed. Of the six requested findAncestor source sites, only 4767, 4778 and 50073 have the #448a51k refusal as their enclosing statement stop. Site 4748 stops earlier on a source-slot certificate, 48433 on `AliasDeclarationNode` in its prologue, and 50082 on `reading result`. These earlier observations must not be relabeled as six independent findAncestor failures.', '',
 (out/'named-sites.md').read_text().replace('# Named caller observations','### Exact named-site observations').replace('## Pairing A','#### Pairing A').replace('## Pairing B','#### Pairing B'),
 'The raw JSON [named-site attribution](named-sites.json) retains the matching Boundary range and exact underlying finding for every quoted statement/prologue stop. [A observing selectors](A-observing-sites.json) and [B selectors](B-sites.json) preserve requested coordinates, actual coordinates, lexical owners and call text. The parent apply shifts the supplied checker lines by 1 or 2; the B adaptation shifts them by 3 or 4.','',
 '## Time and peak memory','',
 'Wall seconds are the wrapper observation, with up to one second of polling resolution. Peak RSS uses Linux `getrusage(RUSAGE_CHILDREN)` in KiB; each census is one Go child. CPU user/system time is retained separately. Runs overlap as described in [the concurrency note](concurrency-note.txt); these single observations are not performance benchmarks. Completed short runs are reused by matching command and compiler-binary hash, not rerun.','',
 '| Pairing | Scope | Compiler | Child exit | Wall seconds | Peak RSS MiB |','| --- | --- | --- | ---: | ---: | ---: |']
resources=[]
for a in ['A','B']:
 for scope in ['entry','core']:
  for side in ['parent','head']:
   name=f'{a}-{side}-{scope}';s=out/(name+'.status.json');t=out/(name+'.time.json')
   if s.exists() and t.exists():
    r=json.loads(s.read_text());mem=json.loads(t.read_text());vals=[a,scope,side,r['exit'],f"{r['wall_seconds']:.3f}",f"{mem['peak_rss_kib']/1024:.1f}"]
   else:vals=[a,scope,side,'running','pending','pending']
   lines.append('| '+' | '.join(map(esc,vals))+' |')
for s in sorted(out.glob('*.status.json')):
 name=s.name.removesuffix('.status.json');t=out/(name+'.time.json')
 if not t.exists():continue
 r=json.loads(s.read_text());m=json.loads(t.read_text());resources.append(dict(run=name,**r,**m))
(out/'resources.json').write_text(json.dumps(resources,indent=2)+'\n')
r=['# All child run resources','','| Run | Exit | Wall seconds | Peak RSS MiB |','| --- | ---: | ---: | ---: |']
for x in resources:r.append('| '+' | '.join(map(esc,[x['run'],x['exit'],f"{x['wall_seconds']:.3f}",f"{x['peak_rss_kib']/1024:.1f}"]))+' |')
(out/'resources.md').write_text('\n'.join(r)+'\n')
lines+=['','Every focused run, apply, build and preliminary failure is included in [all run resources](resources.md) ([JSON](resources.json)). Each `*.status.json` records the exact child command and exit; each `*.time.json` records peak RSS and CPU time. Logs are files and were not piped through head or tail.','',
 '## Setup, verification and limits','',
 'Setup was run with `/workspace/adamic-tools/env.sh`; successful [setup retry](setup-retry.log) used Go 1.27.1, Node 24.19.0, clang 20.1.8 and the pinned TypeScript 6.0.3 API ([npm ci](npm-ci.log)). [Build environment](build-environment.json), [parent workspace](parent-workspace.txt), [head workspace](head-workspace.txt), and the generated guarded [overlays](overlays/) are retained. All Go evidence filenames end in `.go.txt`.', '',
 'The existing pairing suite passed all 15 tests, including normalization and regression mutants ([log](pair-tests.log)). Source hashes are checked before and after each completed pairing. The runtime no-output loader and lowerer assertions execute in every census. No native acceptance, backend agreement or full gate is claimed.', '',
 'Preliminary infrastructure failures are preserved: the initial A apply lost its optional-declarations Node child to SIGKILL with a cgroup OOM kill recorded; both concurrent census builds were terminated and A was reapplied successfully from a clean tree. An initial scratch-workspace replacement error and a missing bridge directory were corrected before successful builds. Setup was retried after restoring an omitted stage1 package. These are not compiler regressions. No later OOM kill has been observed.', '',
 'To reserve RAM, unused scratch checkout directories, the upstream source clone after both applies, and the upstream `tests` directories in the adapted trees were removed. No measured source file was dropped or changed; all 81 source hashes are retained. Completed guarded streams remain the evidence.', '',
 'Checkpoints use separate suffixed branch refs so the final requested branch can contain exactly one commit directly on b5c0aa64 without force-pushing. [Checkpoint records](checkpoints.json) retain the SHAs. No main branch is pushed.', '']
text='\n'.join(lines)
assert '\u2013' not in text and '\u2014' not in text
(out/'REPORT.md').write_text(text)
print('report written; entry pairings complete='+str(complete),flush=True)
