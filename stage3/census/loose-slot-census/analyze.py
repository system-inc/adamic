"""Render completed receipts; never turn missing measurements into zero."""
from pathlib import Path
import collections,json,re
out=Path(__file__).resolve().parent
s=out.parents[3].parent # executable scripts use the explicit scratch root below
s=Path('/workspace/loose-slot-scratch')
def normal(record,which):
 v=record.copy();v.pop('seconds',None);v.pop('roots',None)
 prefix=str(s/('adapted-'+which)/'src/compiler')+'/'
 for key in ['where','reason','fix']:
  if key in v:v[key]=v[key].replace(prefix,'')
 return v
def label(r):return '; '.join([r.get('where','—'),r['kind'],r.get('reason','—')])
def esc(s):return s.replace('|','\\|').replace('\n',' ')
roots=sorted((s/'adapted-head/src/compiler').rglob('*.ts'));assert len(roots)==79
pairs=[]
for i,root in enumerate(roots,1):
 rows=[]
 for which in ['main','head']:
  p=out/'evidence'/which/f'root-{i:02}.json'
  resource=json.loads(p.with_suffix('.time.json').read_text());assert resource['exit_code']==0
  rows.append(normal(json.loads(p.read_text()),which))
 pairs.append((str(root.relative_to(s/'adapted-head/src/compiler')),*rows))
differences=[x for x in pairs if x[1]!=x[2]]
(out/'NEXT-STOPS.md').write_text('# All 79 production first stops\n\n| Root | Main | Head |\n| --- | --- | --- |\n'+'\n'.join('| '+esc(n)+' | '+esc(label(a))+' | '+esc(label(b))+' |' for n,a,b in pairs)+'\n')
summary={'roots':79,'unchanged':79-len(differences),'differences':[{'root':n,'main':a,'head':b}for n,a,b in differences]}
(out/'evidence/comparison.json').write_text(json.dumps(summary,indent=2)+'\n')
p=out/'evidence/type-whole.json';assert p.exists(),'whole-program type count not complete'
type_data=json.loads(p.read_text());assert type_data['Files']==79
resource=json.loads(p.with_suffix('.time.json').read_text());assert resource['exit_code']==0
events=type_data['Events'];byform=collections.defaultdict(list);byfile=collections.defaultdict(list);sites=collections.defaultdict(list)
for e in events:
 match=re.fullmatch(r'(.+):(\d+):(\d+)',e['Where']);assert match,e
 file,line,col=match.groups();file=str(Path(file).relative_to(s/'adapted-head/src/compiler'));e=dict(e,File=file,Line=int(line),Column=int(col));sites[(file,int(line),int(col))].append(e);byfile[file].append(e);byform[e['Form']].append(e)
def distinct(es):return len({(x['File'],x['Line'],x['Column'])for x in es})
listing=['# Every observed type-level site','', 'Locations refer to the adapted head source. Each entry quotes its full source line. Multiple events at one exact position are grouped.','']
for (file,line,col),es in sorted(sites.items()):
 text=(s/'adapted-head/src/compiler'/file).read_text().splitlines()[line-1]
 listing += [f'## {file}:{line}:{col}', '',f'{len(es)} event(s); forms: '+', '.join(sorted({x['Form']for x in es}))+'.','','```typescript',text,'```','']
(out/'SITES.md').write_text('\n'.join(listing))
report=(out/'REPORT.md').read_text();controls=report[report.index('## Control first'):report.index('## Production next stops')]
for name,label,obs in [('arrays-production','production head','NotYet at arrays.a:48:1'),('arrays-latent','latent, LATENT_FULL=1','0 findings'),('arrays-type','type-level instrument','9 events; includes arrays.a:48:1'),('arrays-mutant','stop-disabled mutant','0 events')]:
 v=json.loads((out/'evidence/controls'/(name+'.time.json')).read_text());controls=re.sub(r'^\| '+re.escape(label)+r' \|.*$',f'| {label} | {obs} | {v["seconds"]:.6f} | {v["peak_rss_KiB"]} |',controls,flags=re.M)
lines=['# Loose-slot census','',controls,'## Production next stops','',f'**{len(differences)} roots differ; {79-len(differences)}/79 have the same exact first-stop tuple.** CommonJSBundle; production `LowerWithOptions` through the existing census inspector, one root per process. Both specified SHAs were freshly built and measured. The full table is [NEXT-STOPS.md](NEXT-STOPS.md); raw observations, times and peak RSS are under [evidence/main](evidence/main) and [evidence/head](evidence/head).','']
for n,a,b in differences:lines += [f'- `{n}`: {label(a)} → {label(b)}.']
if not differences:lines+=['There are no changed roots to list.']
groups=collections.Counter((a['kind'],a.get('where'),a.get('reason'))for _,a,_ in pairs)
for (kind,where,reason),count in groups.items():lines += [f'- {count} root(s): `{where or "—"}`; {kind}; {reason or "—"}.']
lines+=['','For the two main roots that lower whole:']
for n,a,b in pairs:
 if n in ['corePublic.ts','hostErrors.ts']:lines += [f'- `{n}` at head: **{label(b)}**.']
lines+=['','`findCycles` is late, not literally last: `lower.go:215` follows all module scans, declarations, bodies and preceding tail transformations; bind/view checks and finalization follow it. Earlier walls cannot reach this stop first. `graph_types.go:132` invokes the loose-slot stop before the no-seeds return. The whole commit message and all of `loose_slots.go` were read.','', '## Sites past the walls: partial type-level source count','',f'**{len(events)} observed events; {len(sites)} distinct exact source sites.** The exhaustive production site count remains **UNKNOWN**. This is a **partial type-level count, not a lowering**, and is conditional on the reconstructed source metadata described below. It is not a count of newly failing production roots. H was gathered over the complete 79-file input before judgment: {type_data["Nodes"]} non-type AST nodes; {type_data["Candidates"]} candidate entries; {type_data["HTypes"]} distinct explicit H types; H.every={type_data["HEvery"]}; {type_data["Shapes"]} shapes.','',f'Whole-program type run: **{resource["seconds"]:.6f} seconds**, **{resource["peak_rss_KiB"]} KiB** peak RSS. Each site, with its quoted source line, is in [SITES.md](SITES.md). The raw event list retains each event’s source, slot and holder types.','', '| Form | Events | Distinct sites |','| --- | ---: | ---: |']
for form,es in sorted(byform.items()):lines += [f'| {form} | {len(es)} | {distinct(es)} |']
for form in ['write','literal init','captured']:
 if form not in byform:lines += [f'| {form} | 0 | 0 |']
lines+=['','| File | Events | Distinct sites | write events/sites | literal init events/sites | captured events/sites |','| --- | ---: | ---: | ---: | ---: | ---: |']
for file,es in sorted(byfile.items()):
 cells=[]
 for form in ['write','literal init','captured']:
  selected=[x for x in es if x['Form']==form];cells.append(f'{len(selected)}/{distinct(selected)}')
 lines += [f'| {file} | {len(es)} | {distinct(es)} | '+' | '.join(cells)+' |']
lines += ['','Distinct sites are `(adapted file, start line, start column)`, not source lines, edits or repeated root visits. Per-form site counts can overlap; events are never silently deduplicated. Files without observed events are absent from the file table, not from the input.','', '### Method and limits','', 'The declaration/overload boundary documented in `whole-program-scout/REPORT.md` at `0d078a8d` occurs before replaceable bodies. Placeholder bodies would also lose captures, concrete instantiations and stores. This work uses the explicitly permitted type-level fallback rather than claiming coherent continued IR. The private overlay adds a measurement-only API and no emission route.','', 'The instrument calls head’s own `looseEntry` on the source walk, its shape/use/made methods, `noteLooseInit`, `holdsLoose`, `viewPairScalar` and `looseReaches`. It completes H before judging candidates. It reconstructs direct member/index assignments and recognized collection calls from source, gathers literal fields/elements/Map/Set entries with the head helper, and models captured cells by checker symbol identity and function nesting. The stop-disabled control retains gathering but bypasses the judgment. k17 and k18 also match production at the exact captured-cell and literal-init sites, respectively; their receipts are in the controls directory.','', '**What this instrument cannot see or certify:** actual lowered IR write sites, synthetic library stores, concrete generic instantiations made only by lowering, explicit carried lowering slots, closed-frame-input proofs, transformed async/generator frames, accessor capture environments, static registration and bound-function metadata. Destructuring/loop/parameter-property stores are not fully reconstructed; computed collection method dispatch and record helper stores are not exhaustive. Source capture retention can overapproximate production. It therefore establishes observed source/type judgments, not an exhaustive number of production stops or a proven upper/lower bound on one. No ordinary next-stop observation is inferred from this count.','', '## Setup, reproduction and resources','', 'Main: `fe5a1a54f045e9bce4e5aa32ca74885a317fbebd`; head: `7e8cfaafb74a61f82f1610a4b738157b9918343d`, exactly one commit on main. Each used a fresh detached worktree under `/workspace/loose-slot-scratch`, initialized recursive shallow submodules, locked `npm ci` in `stage3/api`, and its own `stage3/apply.sh`. The result branch was created only in the head scratch worktree. The shared checkout was not used for edits or measurements.','', 'TypeScript v6.0.3 source was freshly cloned at `050880ce59e30b356b686bd3144efe24f875ebc8`, supplied as ADAMIC_TYPESCRIPT_SOURCE. Completed adapted compiler manifests are identical at both SHAs: 79 TypeScript roots. Their hashes and submodule pins are in [evidence/provenance.json](evidence/provenance.json). GOTMPDIR, TMPDIR, GOCACHE and Adamic build caches were under scratch.','', 'Scripts archive the exact driver/instrument as `.go.txt`. `build_drivers.py` overlays only the census command’s main; `build_type.py` overlays a private lower file plus its command. `groups.py` runs indices in groups of ten and pushes each completed group at both SHAs. `measure.py` records wall time and Linux wait4 peak RSS in KiB per child; these include load/check/lower or type scan, and exclude dependency installation, adaptation and Go build. All measured commands and per-run resources are retained in `.time.json`.','']
for which in ['main','head']:
 resources=[json.loads((out/'evidence'/which/f'root-{i:02}.time.json').read_text())for i in range(1,80)]
 lines += [f'- {which}: summed wall {sum(x["seconds"]for x in resources):.6f} s; largest per-root peak RSS {max(x["peak_rss_KiB"]for x in resources)} KiB.']
lines+=['','Nothing was submitted or merged. Only the census directory is committed on the authorized new branch. No production compiler edit or native-output success is claimed.']
(out/'REPORT.md').write_text('\n'.join(lines)+'\n')
print(summary, 'events',len(events),'sites',len(sites))
