import collections as C, gzip, hashlib, json, pathlib, shutil
root=pathlib.Path('/tmp/void-meter'); out=pathlib.Path('/tmp/void-report/stage3/meter/runs/void-rows-519819228d84');out.mkdir(parents=True,exist_ok=True)
records=[json.loads(x) for x in (root/'latent.jsonl').read_text().splitlines()]
assert records[0]['latent_mode']=='first-error'
expected={str(p) for p in (root/'adapted/src/compiler').rglob('*') if p.suffix in ('.ts','.a')}
assert {r['file'] for r in records[1:]}==expected
assert len(records)-1==len(expected)
fs=[f for r in records[1:] for f in r.get('findings') or []]
prefix="a read of a void container's elements"
def family(f):return 'void-container read' if f['reason'].startswith(prefix) else 'instantiated at void'
def selected(f):return f['reason'].startswith(prefix) or 'instantiated at void' in f['reason']
def norm(s):return s[s.rindex('/src/')+1:] if '/src/' in s else s
rows=[f for f in fs if selected(f)]
shapes=json.loads((root/'shapes.json').read_text())
groups=C.defaultdict(list); families=C.defaultdict(list); units=C.defaultdict(set); exactunits=C.defaultdict(set)
for f in rows:
 groups[f['reason']].append(f);families[family(f)].append(f);units[norm(f['where'])].add(norm(f['unit']));exactunits[f['reason'],norm(f['where'])].add(norm(f['unit']))
def sites(rs):return {norm(f['where']) for f in rs}
def hist(rs):
 u=C.defaultdict(set)
 for f in rs:u[norm(f['where'])].add(norm(f['unit']))
 return ', '.join(f'{n} unit(s): {c} sites' for n,c in sorted(C.Counter(map(len,u.values())).items()))
with gzip.open(out/'void-rows.jsonl.gz','wt',encoding='utf8') as stream:
 for f in rows:
  row=dict(f);row['where']=norm(f['where']);row['unit']=norm(f['unit']);row['family']=family(f);row['site_unit_count']=len(units[row['where']]);row['exact_reason_site_unit_count']=len(exactunits[f['reason'],row['where']]);row.update(shapes[f['where']]);stream.write(json.dumps(row,ensure_ascii=False)+'\n')
raw=C.Counter(f['kind'] for f in fs);unique=C.Counter(k[0] for k in {tuple(f[k] for k in ('kind','where','reason','text')) for f in fs})
seconds=int((root/'latent-seconds.txt').read_text())
lines=['# Stage 3 tsc void rows', '', 'Main: `519819228d84d2a8eefce34a6bf76b79c7eeff63`.', '', 'Measured a detached main worktree with only `55c1c38fbb0840fbd4c2b93579c8722d4c1df35b` cherry-picked (local measurement HEAD `2f2c6c146`). The report branch is cut directly from main and contains no compiler change. Upstream TypeScript: v6.0.3, `050880ce59e30b356b686bd3144efe24f875ebc8`.', '', 'The build copied `build_census` from `stage3/meter/twice-daily.sh`, including both overlays and all three binaries. It ran the same lockfile-pinned npm ci flags and apply command. The foreground latent command used `LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1` over adapted `src/compiler`. Setup, all builds, apply, and latent exited 0. No code was edited.', '', f'Latent elapsed: **{seconds} seconds ({seconds//60}m {seconds%60}s)**. Coverage: **79/79 source files**. Header: checker-rejected program, first-error mode. These are measurement findings, not proof of native output.', '', '## Headline totals', '', 'The meter counts unique `(kind, where, reason, text)` findings across attempts. Raw rows are shown separately; they must not be confused with distinct source sites.', '', '| Kind | Meter total | Raw rows |', '| --- | ---: | ---: |']
for k in ('NotYet','Refused','panic','error','SkippedDependency'):lines.append(f'| {k} | {unique[k]:,} | {raw[k]:,} |')
lines+=['',f'NotYet + Refused: **{unique["NotYet"]+unique["Refused"]:,}** meter findings; **{raw["NotYet"]+raw["Refused"]:,}** raw rows. The latter is 13,401 in this exact run; the motivating 13,402 is not substituted for measured data.', '', '## Selected families', '', '| Reason predicate | Raw findings | Distinct where sites | Distinct units per site |','| --- | ---: | ---: | --- |']
for fam in ('void-container read','instantiated at void'):
 rs=families[fam];lines.append(f'| {fam} | {len(rs):,} | {len(sites(rs)):,} | {hist(rs)} |')
lines+=['',f'Combined: **{len(rows):,} rows, {len(sites(rows)):,} sites**. Six sites occur in both families. Site identity is the finding’s `where`, normalized to `src/…:line:column`; it is never `site_where` or the read location embedded in a generic reason. Units are distinct normalized `unit` values, not row counts. All 9,504 direct void-container sites occur under exactly one unit; the 106 excess rows in that family repeat sites within the same unit. This rise is predominantly many sites, not many units multiplying a few sites.', '', 'Every extracted row retains its original finding fields (except normalized `where` and `unit`), exact reason, source line, syntax kind, expression, and unit counts. The exact reason retains its original temporary path when one is embedded in the message. `site_unit_count` covers both families; `exact_reason_site_unit_count` covers that exact reason. Thus the gzip records the count for every site, including all one-unit sites.', '', '### Sites under multiple units', '', '| Site | Units | Source line |','| --- | ---: | --- |']
for w,us in sorted(units.items()):
 if len(us)>1:
  orig=next(f['where'] for f in rows if norm(f['where'])==w);src=shapes[orig]['line'].replace('|','\\|');lines.append(f'| `{w}` | {len(us)} | `{src}` |')
lines+=['', 'In particular, `utilitiesPublic.ts:1280:12` is `find` instantiated with `T = JSDocTag`: 21 rows, one call site, 21 units. The source is `return find(getJSDocTagsWorker(node, noCache), predicate);`; the reason points to the generic body’s read at `core.ts:170:23`. It is a generic call, not a getter on JSDocTag.', '', '## Top three shapes', '', 'Grouping is **exact reason text first, then the TypeScript expression syntax at the reported site**, ranked by distinct normalized `where`. Argument expressions are resolved by AST parent membership in call/new arguments; this distinguishes a named value, a property read, and a nested call result even when they start at the same column. These are syntactic shapes, not a claim about checker types or a proved root cause.', '', 'The three largest exact-reason groups before splitting by construct are argument (5,620 sites), value (897), and returned value (880). After splitting by construct, all top three shapes belong to the argument reason. Ordinary node-shaped arguments dominate; these are not limited to explicit `void[]` or named callbacks.']
rank=C.defaultdict(set)
for f in rows:rank[f['reason'],shapes[f['where']]['syntax']].add(f['where'])
top=sorted(rank,key=lambda k:(-len(rank[k]),k))[:3]
examples={
'Identifier':['src/compiler/binder.ts:333:30','src/compiler/checker.ts:54204:24','src/compiler/commandLineParser.ts:2008:23','src/compiler/core.ts:780:38','src/compiler/debug.ts:561:39'],
'PropertyAccessExpression':['src/compiler/binder.ts:1900:18','src/compiler/commandLineParser.ts:2459:38','src/compiler/debug.ts:1115:50','src/compiler/emitter.ts:2224:28','src/compiler/factory/nodeConverters.ts:101:34'],
'CallExpression':['src/compiler/executeCommandLine.ts:556:28','src/compiler/factory/utilities.ts:255:32','src/compiler/factory/nodeFactory.ts:1557:55','src/compiler/factory/emitHelpers.ts:212:29','src/compiler/binder.ts:1732:44']}
names={'Identifier':'Named value passed as an argument','PropertyAccessExpression':'Property access passed as an argument','CallExpression':'Call result passed as an argument'}
for i,key in enumerate(top,1):
 reason,syntax=key; ws=rank[key]; rs=[f for f in rows if f['reason']==reason and shapes[f['where']]['syntax']==syntax]
 lines+=['',f'### {i}. {names[syntax]}', '',f'**{len(ws):,} sites; {len(rs):,} rows; one unit per site.** AST: `{syntax}`. Exact reason: `{reason}`.', '', '| Site | Expression | Source line |', '| --- | --- | --- |']
 for w in examples[syntax]:
  orig=next(x for x in ws if norm(x)==w);s=shapes[orig];esc=lambda x:x.replace('|','\\|').replace('\n',' ')
  lines.append(f'| `{w}` | `{esc(s["expression"])}` | `{esc(s["line"])}` |')
lines+=['', '## Every exact reason', '', 'Raw findings count every matching row. Sites deduplicate by `where` within this exact reason. The unit histogram gives the number of distinct units for each site. Reasons below are exact emitted strings; paths embedded in them are deliberately retained.', '', '| Exact reason | Findings | Sites | Units per site |', '| --- | ---: | ---: | --- |']
for reason,rs in sorted(groups.items(),key=lambda p:(-len(sites(p[1])),p[0])):
 lines.append(f'| {reason.replace(chr(124),chr(92)+chr(124))} | {len(rs):,} | {len(sites(rs)):,} | {hist(rs)} |')
lines+=['', '## Artifacts and reproduction', '', '- `run.sh`: exact measurement script used; requires the requested setup and fresh throwaway worktree paths. Reuses the setup-verified identical cohere gitlink via a local symlink.', '- `shapes.cjs`: TypeScript 6.0.3 parser classification of selected sites.', '- `report.py`: extraction, coverage checks, unit accounting, and this report.', '- `void-rows.jsonl.gz`: only matching finding rows, with source and per-site metadata; no full latent census is committed.', '', 'Local logs: `/tmp/setup.log`, `/tmp/void-fetch.log`, `/tmp/void-worktree.log`, `/tmp/void-run.log`, and `/tmp/void-meter/logs/{build,latent-overlay,entry-overlay,entry-build,latent-build,upstream,packages,apply,latent}.log`. Full latent remains local at `/tmp/void-meter/latent.jsonl`.', '', f'Full latent SHA-256 (local, uncommitted): `{hashlib.sha256((root/"latent.jsonl").read_bytes()).hexdigest()}`.', '']
(out/'VOID.md').write_text('\n'.join(lines))
for src,name in [('/tmp/void-run.sh','run.sh'),('/tmp/void-shapes.cjs','shapes.cjs'),('/tmp/void-report.py','report.py')]:shutil.copyfile(src,out/name)
print(out);print('reasons',len(groups),'rows',len(rows),'sites',len(sites(rows)));print('top',[(k,len(rank[k])) for k in top])
