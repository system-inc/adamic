"""Recount completed evidence and pack raw results after all workers have exited."""
import collections,gzip,hashlib,json,os,re,subprocess,tarfile
from pathlib import Path
repo=Path('/workspace/adamic');out=repo/'stage3/adapt/99-correctness-axolotl-r';s=Path('/workspace/scratch/tosorted/step2');raw=s/'raw'
assert (raw/'census-driver.log').read_text().endswith('group 4: 19 roots per side, committed and pushed\n')
assert (raw/'replay-driver.log').read_text().endswith('All non-99r outcomes identical\n')
assert (raw/'verify-driver.log').read_text().endswith('All verification checks passed\n')
checks=json.loads((raw/'verification.json').read_text());assert all(r['exit']==0 for r in checks.values())
contract=json.loads((raw/'contract.log').read_text());assert len(contract['refusals'])==86;assert contract['replay']=={'files':0,'edits':0,'writes':0,'bytesAndMtimePreserved':True}
replay=json.loads((raw/'replay.json').read_text());assert [r['adapter'] for r in replay]==sorted(p.name for p in (repo/'stage3/adapt').iterdir() if p.is_dir())
for row in replay:
 if row['adapter']=='99-correctness-axolotl-r':assert row['before']['exit']==1 and row['after']['exit']==0
 else:assert row['before']==row['after'],row['adapter']
mutants=json.loads((raw/'relation-mutants.log').read_text());assert mutants['allRelationMutantsRed'] and mutants['ownerHashChecksNotInvoked'];assert len(mutants['rows'])==3
comparison=json.loads((out/'tosorted-comparison.json').read_text());assert comparison['changedSources']==['compiler/core.ts'] and comparison['emittedJavascriptFiles']==9 and len(comparison['declarations'])==2
manifest=json.loads((s/'ready-for-census.json').read_text());assert manifest=={str(p.relative_to(s/'after/src')):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((s/'after/src').rglob('*')) if p.is_file()}
assert all((s/'before/src'/p).read_bytes()==(s/'mutable-arm-mutant/src'/p).read_bytes() for p in manifest)
(raw/'source-manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
rows=json.loads((out/'tosorted-census.json').read_text());assert len(rows)==79 and len({r['root'] for r in rows})==79
reason="a value of type never seen as T, a type parameter whose constraint any can be written, so it can write what never can't hold"
def normalized(row):
 return (row['kind'],row.get('where','').split('/src/compiler/')[-1],row.get('reason',''))
def wall(row):return normalized(row)==('Refused','core.ts:1049:34',reason)
cohort=[r for r in rows if wall(r['before'])];assert len(cohort)==73
assert all(normalized(r['after'])==('Refused','core.ts:1076:100',"a cast the runtime can't check") for r in cohort)
binder=next(r['after'] for r in rows if r['root']=='binder.ts')
ordinary=[json.loads(l) for l in (raw/'ordinary-after-binder.jsonl').read_text().splitlines()];assert len(ordinary)==2
assert all({k:v for k,v in binder.items() if k!='seconds'}=={k:v for k,v in r.items() if k!='seconds'} for r in ordinary)
mutable=[json.loads(l) for l in (raw/'mutable-arm-mutant.jsonl').read_text().splitlines()];assert len(mutable)==2 and all(wall(r) for r in mutable)
for i in range(1,5):
 archive=out/f'tosorted-group-{i}.tar.gz';assert hashlib.sha256(archive.read_bytes()).hexdigest()==(out/f'tosorted-group-{i}.sha256').read_text().split()[0]
 with tarfile.open(archive) as t:
  entries=t.extractfile(f'group-{i}/SHA256SUMS').read().decode().splitlines()
  for line in entries:
   sha,name=line.split('  ',1);assert hashlib.sha256(t.extractfile(f'group-{i}/'+name).read()).hexdigest()==sha
  n=19 if i==4 else 20
  for side in ['37f8314b','after']:assert len(t.extractfile(f'group-{i}/'+side+'.jsonl').read().splitlines())==n
source=json.loads((raw/'new-stop.json').read_text());assert source['where']=='core.ts:1076:100' and source['lines']==11 and source['expression']=='Array.prototype as any'
proof=Path('internal/lower/cast_proof.go').read_text().splitlines();assert 'return castProof{}, refused' in proof[196] and 'TypeFlagsAny' in proof[195]
# The only census overlay replaces the driver. The compiler remains ordinary.
prefix=(repo/'stage3/census/tool/main.go').read_text().split('func main() {')[0];assert (out/'census-batch.go.txt').read_text().split('func main() {')[0]==prefix
for label,root in [('base',Path('/workspace/scratch/tosorted-base')),('current',repo)]:
 mapping=json.loads((raw/f'batch-{label}-overlay.json').read_text())['Replace'];assert mapping=={str(root/'stage3/census/tool/main.go'):str(out/'census-batch.go.txt')}
 assert (root/'stage3/census/tool/main.go').read_text().split('func main() {')[0]==prefix
history=[]
for line in subprocess.check_output(['git','log','--reverse','--format=%H %s','--grep=99r: checkpoint toSorted compiler census group'],cwd=repo,text=True).splitlines():
 sha,subject=line.split(' ',1);history.append({'sha':sha,'subject':subject})
assert len(history)==4
(raw/'checkpoint-history.json').write_text(json.dumps(history,indent=2)+'\n')
tables={}
for side in ['before','after']:
 groups=collections.defaultdict(list)
 for row in rows:groups[normalized(row[side])].append(row['root'])
 tables[side]=[{'count':len(v),'kind':k[0],'where':k[1] or '-','reason':k[2] or '-','roots':v} for k,v in sorted(groups.items(),key=lambda x:(-len(x[1]),x[0]))]
summary={'roots':79,'groups':[20,20,20,19],'movedFrom1049To1076':len(cohort),'beforeCompiler':'37f8314b98577c8856a382f8641e7d8a2ca091e7','afterCompiler':'73cdea837469a7a9e27ff692e747421646a068dd','tables':tables,'otherChanges':[r['root'] for r in rows if not wall(r['before']) and normalized(r['before'])!=normalized(r['after'])],'standardToolBinderCrosscheck':True,'checkpoints':history}
(out/'tosorted-census-summary.json').write_text(json.dumps(summary,indent=2)+'\n')
lines=['# toSorted amendment census\n','Independent ordinary Load/Lower observations on all 79 roots. Groups of 20, 20, 20 and 19 were each committed and plainly pushed upon completion. Refused and NotYet are recorded outcomes, not native compilation successes.\n']
for side,label in [('before','Main 37f8314b, original complete adapted tree'),('after','Main 73cdea837 compiler, rebased complete tree with the 99r amendment')]:
 lines+=['## '+label+'\n','| Roots | Kind | First stop | Exact reason |\n| ---: | --- | --- | --- |\n']
 for g in tables[side]:lines.append(f"| {g['count']} | {g['kind']} | `{g['where']}` | {g['reason'].replace('|',chr(92)+'|')} |\n")
 lines.append('\n')
lines+=['## New 73-root group: exact source\n',"There is no enclosing function: the refused cast is in the top-level elementAt initializer, outside either arrow's body. Its complete 11-line declaration is:\n",'```ts\n'+source['declaration'].replace('\r\n','\n')+'\n```\n',"`castRefusal` calls `castProof`. In internal/lower/cast_proof.go:196-197, the source-or-target TypeFlagsAny branch returns the Refused value constructed at line 188. The target of `Array.prototype as any` is any. Neither this initializer nor its casts were edited.\n",'## Census method and provenance\n',"census-batch.go.txt preserves the standard stage3/census/tool imports, run record and inspect function byte-for-byte at both compiler revisions. Only main is replaced through a one-file Go overlay to take 1 to 20 independent root paths and omit the standard tool's redundant aggregate row. No load, checker, lowering, preflight or backend code is overlaid. Each root still calls ordinary Load on that one file and ordinary Lower on its complete dependency graph. The after binder observation equals both standard-tool rows, excluding elapsed time. Group 3 exposed a driver ordering assertion: Python Path component order differs from Go string order for declarations.ts and declarations/diagnostics.ts. All 20 observations per side had completed. The driver was corrected to require the exact unique root set and associate each before/after row by its absolute root path. Those completed compiler measurements were retained without rerunning them; the original error log and recovery ledger are archived.\n",'tosorted-census.json retains every paired root and exact outcome. tosorted-census-summary.json retains full groups and checkpoint commits. Four tar.gz archives retain raw rows, root lists, logs and internal SHA256SUMS; adjacent sha256 files verify each archive. The full main before apply is 73cdea837, while the comparison census explicitly retains the requested 37f8314b baseline. The restored-arm mutant is byte-identical to the fresh 73cdea837 before source tree and reproduces 1049:34 on that compiler.\n']
(out/'TOSORTED-CENSUS.md').write_text(re.sub(r'(?m)(\|[^\n]*\|)\n\n(?=\|)',r'\1\n','\n'.join(lines)))
prooftext=f'''# 99r toSorted readonly-arm proof\n\nAdamic authorized the pure upcast on task b5w3ycg, Nov 6 at 20:05Z, applying docs/rulings.md:14. Only toSorted's after text and after hash change in owners.json; its before state and all eight other owners are unchanged. The source edit is `emptyArray as readonly T[]`.\n\nMain moved under 95-checker-axolotl-h, so the branch was rebased onto 73cdea837469a7a9e27ff692e747421646a068dd before editing. The published measurement commit is retained as an ancestry parent to permit plain pushes without force. All authored changes stay inside 99r. Step 1's original archive and closeout remain historical evidence.\n\n| Check | Result |\n| --- | --- |\n| Full before/after apply, --build-product | Both exit 0; before is exact main 73cdea837 |\n| Owner contract | 86 no-write refusals; pristine before and new after accepted; bytes/mtime and LF/CRLF preserved |\n| Compiled sound-relation contract | Actual extracted cast and satisfies witness pass adamic types and adamic c for unconstrained T instantiated as number and string |\n| Required readonly [] mutant | Relation contract red; no owner-hash gate used |\n| Other downcast, never[] plus required branded field | Checker accepts actual cast; castProof refuses it; relation contract red |\n| Mutable T[] control | Checker accepts both probes; Adamic refuses both with invariant-mutable |\n| Complete paired adapter replay | All {len(replay)} adapters attempted in order on both complete trees; every non-99r outcome identical |\n| adamic types after | All 79 roots together; exit 0, zero diagnostics |\n| Complete core JavaScript | Byte-identical with comments kept and removed |\n| All nine emitted JavaScript files | Byte-identical |\n| Both public declarations | Byte-identical |\n| Complete source comparison | {comparison['sourceFiles']} files; only the exact toSorted owner edit differs |\n| FlatMap and other emptyArray sites | Unchanged |\n| Mutable-arm mutant | Exact core.ts:1049:34 stop in both standard ordinary rows; all source bytes equal the fresh main-before source |\n| Per-file census | All 79 roots; groups 20, 20, 20, 19 committed and pushed individually |\n| 73-root wall | Moves to core.ts:1076:100, Refused: a cast the runtime can't check |\n| Batched meter control | Standard tool's two binder rows equal the batched observation, excluding elapsed time |\n\nThe contract extracts the exact adapted empty-array cast and declaration from core.ts. For each generated probe, adamic types demonstrates checker admission and adamic c invokes ordinary Load and LowerWithOptions. The actual cast follows castProof's upcast candidate path at cast_proof.go:219-229 and provenRelation at proven_relations.go:11-30, including mutable invariance, optional-field compatibility and nominal identity. The relation-only satisfies witness forbids acceptance merely as a checked downcast. The probes carry the shared declaration and unconstrained T; they isolate this arm from unrelated compiler dependencies. The required mutants call this same relation contract directly, bypassing owner hashes, and the mutable-view control proves the test is stronger than checker assignability.\n\nTOSORTED-CENSUS.md contains both complete grouped tables, the exact unchanged elementAt declaration for the new stop, and the refusal line/function. No native compilation of the blocked roots is claimed. The runtime oracle is not rerun because the full emitted JavaScript and public declarations are byte-identical.\n\nThe before and after apply processes froze their ordered adapter lists and executed unchanged transformation inputs. Later evidence files change apply's product key, not these source transformations. Compiler provenance, exact commands, input roots, complete logs, comparison hashes, mutant probes and outputs are packed in tosSorted-step2.tar.gz with internal SHA256SUMS and an external archive hash. Each census group has its own verified archive. All Go evidence files use .go.txt.\n'''
(out/'TOSORTED-PROOF.md').write_text(prooftext.replace('tosSorted-step2','tosorted-step2'))
(raw/'full-apply-results.json').write_text(json.dumps({'before':{'head':'73cdea837469a7a9e27ff692e747421646a068dd','command':'ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source /workspace/scratch/tosorted-main/stage3/apply.sh --build-product /workspace/scratch/tosorted/step2/before','exit':0},'after':{'command':'ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source /workspace/adamic/stage3/apply.sh --build-product /workspace/scratch/tosorted/step2/after','exit':0}},indent=2)+'\n')
# All task workers must already have completed. Ignore this verifier itself.
live=[]
for p in Path('/proc').iterdir():
 if not p.name.isdigit() or int(p.name)==os.getpid():continue
 try:
  comm=(p/'comm').read_text().strip();state=(p/'stat').read_text().split()[2]
  if comm in ['node','go','compile','tool','adamic','census-batch','census-batch-ba','python3'] and state!='Z':live.append({'pid':int(p.name),'command':comm})
 except OSError:pass
assert not live,live
(raw/'process-audit.json').write_text(json.dumps({'liveTaskWorkers':live,'allStartedWorkersExited':True},indent=2)+'\n')
rawfiles=[p for p in sorted(raw.rglob('*')) if p.is_file() and not any(part.startswith('census-group-') for part in p.relative_to(raw).parts) and p.name!='STEP2-SHA256SUMS']
(raw/'STEP2-SHA256SUMS').write_text(''.join(hashlib.sha256(p.read_bytes()).hexdigest()+'  '+str(p.relative_to(raw))+'\n' for p in rawfiles))
archive=out/'tosorted-step2.tar.gz'
with tarfile.open(archive,'w:gz') as t:
 for p in rawfiles+[raw/'STEP2-SHA256SUMS']:t.add(p,arcname='step2/'+str(p.relative_to(raw)))
(out/'tosorted-step2.sha256').write_text(hashlib.sha256(archive.read_bytes()).hexdigest()+'  '+archive.name+'\n')
changed=subprocess.check_output(['git','diff','--name-only','73cdea837','HEAD'],cwd=repo,text=True).splitlines()
untracked=subprocess.check_output(['git','ls-files','--others','--exclude-standard'],cwd=repo,text=True).splitlines()
assert len(set(changed+untracked))<150
for name in changed+untracked:
 assert name.startswith('stage3/adapt/99-correctness-axolotl-r/'),name
 p=repo/name;assert p.suffix!='.go',name
 if p.suffix not in ['.gz']:
  text=p.read_text();assert '\u2013' not in text and '\u2014' not in text,name
print(json.dumps({'roots':79,'movedRoots':73,'otherChanges':summary['otherChanges'],'adapters':len(replay),'refusalFixtures':86,'step2ArchiveSha256':hashlib.sha256(archive.read_bytes()).hexdigest(),'liveTaskWorkers':live},indent=2))
