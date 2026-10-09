"""Project the committed full latent census, preserving pins and exclusions."""
import collections,csv,gzip,hashlib,json,sys
from pathlib import Path
repo,tree,out=map(Path,sys.argv[1:]);run=repo/'stage3/meter/runs/20261008T035244Z.latent-full';owners=json.loads((repo/'stage3/meter/owners.json').read_text());manifest=json.loads((run/'compiler/source-manifest.json').read_text());stock=json.loads((run/'compiler/stock-units.json').read_text())
metadata=json.loads((run/'metadata.json').read_text());selected_files={'src/compiler/parser.ts','src/compiler/scanner.ts','src/compiler/factory/nodeFactory.ts','src/compiler/factory/nodeTests.ts','src/compiler/utilities.ts','src/compiler/utilitiesPublic.ts'}
def norm(s):
 i=s.find('src/compiler/');return s[i:] if i>=0 else s

def owner(reason):
 if 'seen as' in reason:return owners.get('seen as','OWNER BLANK')
 if reason in owners:return owners[reason]
 matches=[k for k in owners if reason.startswith(k)];return owners[max(matches,key=len)] if matches else 'OWNER BLANK'
def wall(reason):
 if 'as a condition' in reason or 'PrefixUnaryExpression' in reason:return 'truthiness/unary frontend (codex/taste-not-soundness); owner '+owner(reason)
 if 'function inside a function' in reason:return 'nested function/closure frontend (codex/nested-functions); numbered step not recorded'
 if reason.startswith('reading ') and owner(reason)=='OWNER BLANK':return 'isolated census context boundary; replay before assigning a feature wall'
 if 'cast' in reason or 'seen as' in reason or 'refinement' in reason:return 'checked-cast/type-shape wall (#b5w3ycg); step09 scout'
 if 'type predicate' in reason or 'asserts' in reason:return 'predicate/assertion body proof; owner '+owner(reason)
 if 'non-null' in reason or 'NonNull' in reason:return 'non-null/default proof; owner '+owner(reason)
 if any(x in reason for x in ['JSDocState','SyntaxKind','CharacterCodes','NodeFlags','enum','EnumDeclaration']):return 'enum and namespace frontend; owner '+owner(reason)
 if 'namespace' in reason:return 'namespace frontend; owner '+owner(reason)
 if 'any' in reason or 'unknown' in reason:return 'any/unknown source adaptation; step09 scout'
 if 'Uint16Array' in reason or 'typed array' in reason:return 'typed-array wall; step12 scout'
 return 'wall step not recorded; owner '+owner(reason)
rows=[];unit_rows=[];boundaries=[];files=[];unique=set()
with gzip.open(run/'compiler/full.jsonl.gz','rt') as f:
 for line in f:
  record=json.loads(line);file=norm(record.get('file',''))
  if file not in selected_files:continue
  original=next(x for x in manifest if x['file']==file);current=tree/file;matches=hashlib.sha256(current.read_bytes()).hexdigest()==original['sha256'];files.append(dict(original,current_source_equal=matches))
  units=record.get('units',[]);anchors=[u for u in units if 'JSDoc' in u.get('name','')]
  chosen=[u for u in units if u in anchors or any(u.get('body_start',-1)>=a.get('body_start',1<<60) and u.get('body_end',1<<60)<=a.get('body_end',-1) for a in anchors)]
  selected={norm(u['where']) for u in chosen};unit_rows.extend(dict(u,file=file) for u in chosen)
  # The original source matches parser/factory exactly, allowing independent AST owner mapping
  # of scan findings attempted from an enclosing namespace, rather than just unit matching.
  text=current.read_bytes().decode('utf-8') if matches else None
  def nearest(where):
   if text is None:return None
   location=norm(where);parts=location.rsplit(':',2)
   if len(parts)!=3 or parts[0]!=file:return None
   ln,col=map(int,parts[1:]);lines=text.splitlines(keepends=True);prefix=lines[ln-1].encode('utf-16-le')[:2*(col-1)].decode('utf-16-le');offset=len((''.join(lines[:ln-1])+prefix).encode())
   candidates=[u for u in stock.get(file,[]) if u['start']<=offset<u['end']]
   return min(candidates,key=lambda u:u['end']-u['start']) if candidates else None
  for finding in record.get('findings',[]):
   unit=norm(finding.get('unit',''));n=nearest(finding.get('where',''));in_scope=unit in selected or n is not None and n['where'] in selected
   if not in_scope:continue
   if finding['kind'] not in ['Refused','NotYet']:boundaries.append(finding);continue
   key=tuple(finding[k] for k in ['kind','where','reason','text'])
   if key in unique:continue
   unique.add(key);rows.append(dict(finding,file=norm(finding['where']).rsplit(':',2)[0],attempting_file=file,owner_function=n['name'] if n else next((u.get('name','') for u in chosen if norm(u['where'])==unit),''),owner=owner(finding['reason']),wall_step=wall(finding['reason'])))
  for u in chosen:
   if u.get('status')!='attempted':boundaries.append(u)
summary=dict(collections.Counter(x['kind'] for x in rows));reason_counts=collections.Counter((x['kind'],x['reason'],x['wall_step']) for x in rows)
report=dict(provenance=dict(run=str(run.relative_to(repo)),compiler_source_commit=metadata['compiler_source_commit'],instrumentation_core_commit=metadata['instrumentation_core_commit'],measurement='historical full latent census on a checker-rejected program; not a current-main recount'),selection='named JSDoc functions and nested declarations within them, six parser/scanner/factory/helper files; matching-source scan rows joined to smallest stock AST declaration',limits=['Shared generic parser/scanner helpers outside selected JSDoc declarations are not an exhaustive transitive closure.', 'Scanner/utilities source hashes differ; their rows retain historical coordinates and are selected by attempted unit only.', 'A failed construct hides child expressions; excluded bodies and census boundaries do not prove no blockers.', 'Wall labels are reason-based routing proposals. No numbered wall assignment is invented where the repository records none.'],source_files=files,counts=summary,units=unit_rows,excluded_or_other=boundaries,findings=rows,reasons=[dict(kind=k[0],reason=k[1],wall_step=k[2],count=v) for k,v in reason_counts.most_common()])
(out/'latent-ledger.json').write_text(json.dumps(report,indent=2)+'\n')
with (out/'latent-sites.csv').open('w',newline='') as f:
 writer=csv.DictWriter(f,lineterminator='\n',fieldnames=['file','where','kind','reason','owner_function','owner','wall_step','text']);writer.writeheader();writer.writerows({k:x.get(k,'') for k in writer.fieldnames} for x in rows)
print('PASS pinned census projection:',summary,'units',len(unit_rows),'excluded/other',len(boundaries));print('source equal:',[(f['file'],f['current_source_equal']) for f in files])
