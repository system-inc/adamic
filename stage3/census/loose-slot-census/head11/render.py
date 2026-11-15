"""Render the -11 follow-up above the byte-preserved published -7 report."""
from pathlib import Path
import json,collections,hashlib,subprocess,gzip
out=Path(__file__).resolve().parent;repo=out.parents[3];s=Path('/workspace/loose-slot-scratch/eleven');oldsha='8bf06ebedd36f13c251aa6c3050aa7a5382d7cf1'
def read(p):return json.loads(p.read_text())
def rel(w):return w.split('/src/compiler/')[-1]
def source(w):
 path,line,col=w.rsplit(':',2);return Path(path).read_text().splitlines()[int(line)-1]
def esc(t):return str(t).replace('|','\\|').replace('\n',' ')
def quote(lines,w):lines += [f'`{rel(w)}`','','```typescript',source(w),'```','']
original=subprocess.check_output(['git','show',oldsha+':stage3/census/loose-slot-census/REPORT.md'],cwd=repo,text=True)
typed=read(out/'evidence/type-whole.json');audit=read(out/'evidence/independent-audit.json');old=read(out.parent/'evidence/type-whole.json')
oldsites={rel(e['Where']):e for e in old['Events']};newsites={rel(e['Where']):e for e in typed['Events']}
refs={rel(r['Where']):rel(c['BindingWhere'])for c in audit['Catches']for r in c['References']}
still=sorted(oldsites.keys()&newsites.keys());gone=sorted(oldsites.keys()-newsites.keys());new=sorted(newsites.keys()-oldsites.keys())
captures=[c for c in typed['Catches']if c['Captured']];stops=[c for c in typed['Catches']if c['Stopped']];bound=[c for c in typed['Catches']if c['Bound']];calls=typed['UnknownCatchCalls']or[];callcount=len({c['Where']for c in calls})
assert len(typed['Catches'])==audit['CatchClauses']and len(bound)==audit['BoundCatchClauses']and len(captures)==audit['CapturedCatchBindings']
assert {rel(c['Where'])for c in calls}=={rel(c['Where'])for c in audit['UnknownCalls']}
catchdoc=['# Catch inventory at -11','',f'{len(typed["Catches"])} AST catch clauses; {len(bound)} have bindings; {len(captures)} have bindings captured by closures; {len(stops)} captured catch bindings stop. There are {callcount} distinct calls ({len(calls)} argument entries) passing a catch binding directly to an unknown-typed parameter. Parentheses are unwrapped; derived values such as String(error) are not counted as the caught binding.','', 'Emitter helper templates and comments are source text, not catch AST nodes. An independent TypeScript 6.0.3 AST/checker walk agrees with the Go/checker inventory.','', '## Stopped captured catch bindings','']
if not stops:catchdoc += ['None. There is no stopped catch-cell clause/closure/H-witness triple to quote in this program.','']
for c in stops:
 quote(catchdoc,c['ClauseWhere'])
 for closure in c['Closures']:catchdoc += [f'Capturing closure `{rel(closure["Where"])}`:','','```typescript',closure['Text'],'```','']
 for call in c['UnknownCalls']:quote(catchdoc,call['Where'])
catchdoc += ['## Every bound catch clause','']
for c in bound:
 catchdoc += [f'### {rel(c["Where"])} — {c["Name"]}','',f'Captured by a closure: **{str(c["Captured"]).lower()}**. Stops as a captured catch binding: **{str(c["Stopped"]).lower()}**.',''];quote(catchdoc,c['ClauseWhere'])
 for closure in c['Closures']:catchdoc += [f'Closure `{rel(closure["Where"])}`:','','```typescript',closure['Text'],'```','']
 if not c['Closures']:catchdoc += ['Capturing closures: none.','']
catchdoc += ['## Every catch-binding call to an unknown parameter','']
for call in calls:
 catchdoc += [f'### {rel(call["Where"])}','',f'Binding `{rel(call["BindingWhere"])}`; argument type `{call["ArgumentType"]}`; parameter `{call["Parameter"]}: {call["ParameterType"]}` at `{rel(call["ParameterWhere"])}`. Independently makes H every shape: **{str(call["MakesHEvery"]).lower()}**.',''];quote(catchdoc,call['Where']);catchdoc += ['Parameter declaration:',''];quote(catchdoc,call['ParameterWhere'])
(out/'CATCHES.md').write_text('\n'.join(catchdoc))
siteDoc=['# Paired type-level sites: -7 to -11','', 'Pair key: compiler-relative file, line, column. Both adapted compiler inputs have identical file hashes. This compares partial type-level observations, not successful lowerings.','',f'{len(still)} stand; {len(gone)} gone; {len(new)} new. Captured catch-cell stops and direct caught-value reads are marked separately.','', '| Site | -7 form | -11 form | Status | Captured catch-cell stop | Caught-value binding read at site |','| --- | --- | --- | --- | --- | --- |']
for key in sorted(oldsites.keys()|newsites.keys()):
 a=oldsites.get(key);b=newsites.get(key);status='stands'if a and b else'gone'if a else'new';siteDoc += [f'| {key} | {a["Form"]if a else"—"} | {b["Form"]if b else"—"} | {status} | {"yes"if b and b["Catch"]else"no"} | {refs.get(key,"—")} |']
siteDoc += ['', 'The `sys.ts:1643:72` entry is the literal initialization of an `error` field from the catch binding at `sys.ts:1642:24`. It stands; it is not a captured catch-cell stop.','', '## Every observed site, with source line','']
for key in sorted(newsites):siteDoc += [f'### {key}','','```typescript',source(newsites[key]['Where']),'```','']
(out/'SITE-PAIRS.md').write_text('\n'.join(siteDoc))
rows={};resource={}
for which in ['base','head']:
 rows[which]={};resource[which]=[]
 for p in (out/'evidence'/which).glob('root-*.time.json'):
  r=read(p)
  if r['exit_code']!=0:continue
  i=int(p.name.split('-')[1].split('.')[0]);obs=read(p.with_name(p.name.replace('.time.json','.json')));obs.pop('seconds',None);root=obs.pop('roots')[0];obs['root']=root.split('/src/compiler/')[-1]
  for k in ['where','reason','fix']:
   if k in obs:obs[k]=obs[k].replace(str(s/('adapted-'+which)/'src/compiler')+'/','')
  rows[which][i]=obs;resource[which].append(r)
paired=sorted(rows['base'].keys()&rows['head'].keys());differences=[i for i in paired if rows['base'][i]!=rows['head'][i]]
comparison={'pairs_complete':len(paired),'unchanged':len(paired)-len(differences),'different_indices':differences,'still_sites':still,'gone_sites':gone,'new_sites':new,'catch_clauses':len(typed['Catches']),'bound_catch_clauses':len(bound),'captured_catch_bindings':len(captures),'stopped_catch_bindings':len(stops),'unknown_catch_calls':callcount}
(out/'evidence/comparison.json').write_text(json.dumps(comparison,indent=2)+'\n')
def stop(r):return '; '.join([r.get('where','—'),r['kind'],r.get('reason','—')])
nextdoc=['# Production next stops at -11 and its main base','', '| Root | Base 89614cc0f | Head 8843e0ab |','| --- | --- | --- |']
for i in paired:nextdoc += ['| '+esc(rows['head'][i]['root'])+' | '+esc(stop(rows['base'][i]))+' | '+esc(stop(rows['head'][i]))+' |']
(out/'NEXT-STOPS.md').write_text('\n'.join(nextdoc)+'\n')
section=['<!-- HEAD11 START -->','# -11 catch follow-up','', 'Head: `8843e0ab7b4cfaa547c08e438ca39442a4774fc6`. Parent and main base: **`89614cc0f9d6aa7b7fe96b56c5c6c1c3eace9825`**. It is one commit on that main. Its parent diff was read first: it contains the full loose-slot feature; the catch amendment relative to -7 adds `looseCaughtRoad`, caught-value collection, the catch-cell links, and the caught-source judgment. The seven older blocking roads remain premises of the narrowing.','', 'This follow-up uses the same source/capture reconstruction, rebuilt against -11 with its own catch judgment: **a partial type-level count, not a lowering**. The published -7 report is preserved below.','', '## Controls at -11','', 'The logging-helper shape described in the request stops; the message-extraction twin lowers. Exact Language u00/u00b files were not present in this repository, so these fixtures are explicitly labeled reconstructions of the specified shape, not byte-identical Language originals. The committed c04 fixture is also tested unchanged. See [control sources and provenance](head11/controls).','', '| Control | Production | Probe events / sites | Captured catch binding | H.every / caughtEvery | Wall seconds: production / probe | Peak RSS KiB: production / probe |','| --- | --- | ---: | --- | --- | ---: | ---: |']
for name in ['c04','u00-reconstructed','u00b-reconstructed']:
 d=out/'evidence/controls';p=read(d/(name+'-production.json'));t=read(d/(name+'-type.json'));pr=read(d/(name+'-production.time.json'));tr=read(d/(name+'-type.time.json'));section += [f'| {name} | {p["kind"]} | {len(t["Events"])}/{len({e["Where"]for e in t["Events"]})} | {sum(c["Captured"]for c in t["Catches"])} | {str(t["HEvery"]).lower()} / {str(t["CaughtEvery"]).lower()} | {pr["seconds"]:.6f} / {tr["seconds"]:.6f} | {pr["peak_rss_KiB"]} / {tr["peak_rss_KiB"]} |']
section += ['', 'u00’s probe reports two Error-source events at **one** catch binding (explicit throw and library Error); production stops at that binding. u00b keeps the unknown-parameter call and H.every=true but captures the extracted string. c04 has a captured catch binding, H.every=false, and lowers. Thus the positive/negative controls exercise the narrowing and the requested H interaction.','', '## 1. Catch bindings and denominators','',f'**{len(stops)} stopped captured catch bindings.** Of {len(typed["Catches"])} real catch clauses, {len(bound)} have a binding and {len(typed["Catches"])-len(bound)} omit it; **{len(captures)} bindings are captured by a closure**. Separately, **{callcount} calls** pass a catch binding directly to an unknown-typed parameter, all with an unknown-typed source and each independently making H every shape. The two necessary conditions of the requested u00 shape do not meet here: unknown calls exist, captured catch bindings do not.','', 'An independent TypeScript 6.0.3 AST/checker audit agrees exactly, including the five call locations. Template strings in `factory/emitHelpers.ts` contain textual catches and closures but are not compiler catch AST nodes. The seven actual bound clauses and all five calls/unknown parameter declarations are quoted in [CATCHES.md](head11/CATCHES.md). There is no stopped catch binding for which a capturing-closure/H-witness triple can be supplied.','',f'The whole-program walk has H.every={str(typed["HEvery"]).lower()}, caughtEvery={str(typed["CaughtEvery"]).lower()}, {typed["Candidates"]} candidates, {typed["HTypes"]} explicit H types, and {typed["Shapes"]} shapes over {typed["Nodes"]:,} non-type AST nodes. The head’s caught-road collection is global; it is not a per-try throw analysis.','', '## 2. The original 22, paired by site','',f'**{len(still)} stand; {len(gone)} gone; {len(new)} new.** The rebuilt probe reads {len(typed["Events"])} events at {len(newsites)} distinct sites: '+', '.join(f'{count} {form}'for form,count in sorted(collections.Counter(e['Form']for e in typed['Events']).items()))+'. None is a captured catch-cell stop. The existing caught-value literal initialization at `sys.ts:1643:72` is marked separately. Every old/new site pair, its classification, and quoted source line are in [SITE-PAIRS.md](head11/SITE-PAIRS.md).','', '## 3. Production next-stop table','']
if len(paired)==79:
 section += [f'**{len(differences)} changed; {79-len(differences)}/79 unchanged first-stop tuples** between the exact base and head, in CommonJSBundle format. [Full 79-root table](head11/NEXT-STOPS.md). Both sides were freshly built and measured, one root per child, in groups of ten with normal checkpoint pushes.']
 wall=[i for i in paired if rows['base'][i].get('where')=='sys.ts:1387:5'];section += ['',f'The {len(wall)} roots at main’s `sys.ts:1387:5` wall have {sum(rows["base"][i]==rows["head"][i]for i in wall)} unchanged tuples.']
 for i in paired:
  if rows['head'][i]['root']in['corePublic.ts','hostErrors.ts']:section += [f'- `{rows["head"][i]["root"]}` at head: **{rows["head"][i]["kind"]}**.']
else:section += [f'Checkpoint: {len(paired)}/79 pairs completed; {len(paired)-len(differences)} unchanged and {len(differences)} changed among completed pairs. Unmeasured roots are not assumed unchanged. [Completed rows](head11/NEXT-STOPS.md).']
section += ['', '## Method, limits, resources and reproduction','', 'The original probe’s partial source model and limitations remain: synthetic lowering-created stores, concrete generic instantiations, carried slots, retained/bound/accessor captures, closed-frame proofs, IR scalar facts, and transformed async frames are not reconstructed. Catch binding registration and its source capture inventory are explicit; the new head’s `looseEntry` gathers `looseCaughtRoad`, and captured catches use `looseCaughtCell`/`looseCaught`/`looseReaches` after gathering. No production lowering source is overlaid in the next-stop binaries. Inventory calls are direct identifier arguments (parentheses allowed); aliases and derived values are outside that denominator. No census result here is an exhaustive production lowering count.','', 'Fresh detached worktrees were made for both SHAs, with recursive depth-1 submodules, locked API installs, pinned TypeScript `050880ce59e30b356b686bd3144efe24f875ebc8`, and separate stage3/apply outputs. GOTMPDIR, TMPDIR, GOCACHE and Adamic caches remain in scratch. All 79 compiler input files are byte-identical across the two new worktrees and the -7 run; pairing needs no line relocation. Source/binary hashes and the parent diff are archived under [head11/evidence](head11/evidence).','']
for name,label in [('type-whole','Head partial type-level count'),('independent-audit','Independent TypeScript inventory')]:
 r=read(out/'evidence'/(name+'.time.json'));section += [f'- {label}: **{r["seconds"]:.6f} s**, **{r["peak_rss_KiB"]} KiB** peak RSS.']
for which in ['base','head']:
 rs=resource[which]
 if rs:section += [f'- {which}: {len(rs)} completed children; summed wall **{sum(r["seconds"]for r in rs):.6f} s**; maximum child peak **{max(r["peak_rss_KiB"]for r in rs)} KiB**.']
section += ['', 'Every reported process has a wall/peak-RSS wait4 receipt. Child wall times can overlap; these are resource records, not isolated speed comparisons. Setup/build times are excluded. Control receipts retain the final exercised sources. `head11/build.py`, `build_type.py`, `runs.py`, `groups.py` and `independent_audit.cjs` archive the exact procedure; scripts use the recorded scratch prefix. Only `.go.txt` probe sources are committed. Nothing submitted or merged.','<!-- HEAD11 END -->','','---','']
(out.parent/'REPORT.md').write_text('\n'.join(section)+'\n'+original)
print(comparison)
