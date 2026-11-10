from pathlib import Path
import json,difflib,hashlib,gzip
scratch=Path('/workspace/scratch');unit=Path('/workspace/adamic/stage3/adapt/99-correctness-axolotl-t');reports=[];full_patch=[];counts=dict.fromkeys(['a','b','c'],0)
hashtext=lambda s:hashlib.sha256(s.encode()).hexdigest()
for stem in ['_tsc','typescript','run']:
 prefix=scratch/('debug-display-2-class-'+stem);meta=json.loads(Path(str(prefix)+'.json').read_text());file=meta['file']
 original=(scratch/'debug-display-2-before/built/local'/file).read_text();after=(scratch/'debug-display-2-candidate/built/local'/file).read_text();expected=Path(str(prefix)+'.expected.js.txt').read_text()
 a=expected.splitlines(True);b=after.splitlines(True);na=Path(str(prefix)+'.expected.normalized.txt').read_text().splitlines(True);nb=Path(str(prefix)+'.after.normalized.txt').read_text().splitlines(True);assert len(a)==len(na) and len(b)==len(nb)
 removed=meta['removed']
 # Least fixed point: an accepted deletion can only refer to reviewed owners
 # or previously accepted deleted declarations. Unanchored cycles fail closed.
 by_name={d['name']:d for d in removed};accepted={};edges={}
 for d in removed:
  refs=[]
  for ref in d['references']:
   parent=None
   if not ref['inside_five_owner']:
    name=(ref['enclosing_function'] or '').rstrip('0123456789')
    candidate=by_name.get(name)
    if candidate and candidate['original_declaration_line']<=ref['line']<=candidate['original_declaration_line']+candidate['end_line']-candidate['start_line']:parent=name
   refs.append({'line':ref['line'],'column':ref['column'],'owner':ref['inside_five_owner'],'removed_declaration':parent})
  edges[d['name']]=refs
 changed=True
 while changed:
  changed=False
  for name,refs in edges.items():
   if name not in accepted and all(ref['owner'] or ref['removed_declaration'] in accepted for ref in refs):
    chains=[]
    for ref in refs:
     if ref['owner']:chains.append([name,ref['owner']])
     else:chains.extend([[name]+chain for chain in accepted[ref['removed_declaration']]])
    accepted[name]=chains;changed=True
 for d in removed:
  d['class_b_transitively_closed']=d['name'] in accepted;d['reference_edges']=edges[d['name']];d['reference_chains']=accepted.get(d['name'],[])
 drop={i for d in removed for i in range(d['start_line'],d['end_line']+1)};canonical_after_deletions=''.join(line for i,line in enumerate(na) if i not in drop)
 global_equal=canonical_after_deletions==''.join(nb)
 hunks=[]
 for number,group in enumerate(difflib.SequenceMatcher(None,a,b).get_grouped_opcodes(3),1):
  i1,i2=group[0][1],group[-1][2];j1,j2=group[0][3],group[-1][4];old=''.join(na[i1:i2]);new=''.join(nb[j1:j2]);declarations=[d for d in removed if d['start_line']>=i1 and d['end_line']<i2];deleted_lines={i for d in declarations for i in range(d['start_line'],d['end_line']+1)};pruned=''.join(na[i] for i in range(i1,i2) if i not in deleted_lines)
  if old==new:classification='a';reason='Byte-identical after terminal decimal suffix normalization on AST Identifier nodes.'
  elif declarations and pruned==new and all(d['class_b_transitively_closed'] for d in declarations):classification='b';reason='Only listed declarations disappear after identifier normalization; every original symbol reference is inside a reviewed owner or a transitively accepted removed declaration.'
  else:
   classification='c'
   if declarations and pruned==new:reason='Declaration removal has references outside the three owners: '+', '.join(d['name'] for d in declarations if not d['class_b_transitively_closed'])
   else:reason='Not identical after identifier normalization and deletion of inventoried declarations.'
  counts[classification]+=1;hunks.append({'hunk':number,'class':classification,'expected_lines':[i1+1,i2],'after_lines':[j1+1,j2],'old_raw_sha256':hashtext(''.join(a[i1:i2])),'new_raw_sha256':hashtext(''.join(b[j1:j2])),'old_normalized_sha256':hashtext(old),'new_normalized_sha256':hashtext(new),'normalized_after_deletions_sha256':hashtext(pruned),'declarations':[d['name'] for d in declarations],'removal_chains':{d['name']:d['reference_chains'] for d in declarations},'reason':reason})
 full_patch.extend(difflib.unified_diff(original.splitlines(True),after.splitlines(True),fromfile='before/'+file,tofile='after/'+file))
 local={c:sum(h['class']==c for h in hunks) for c in counts};reports.append({**meta,'counts':local,'hunks':hunks,'all_residual_bytes_equal_after_normalization_and_declaration_deletion':global_equal,'normalized_expected_sha256':hashtext(''.join(na)),'normalized_after_sha256':hashtext(''.join(nb)),'normalized_expected_after_deletions_sha256':hashtext(canonical_after_deletions)})
report={'status':'stop' if counts['c'] else 'pass','counts':counts,'hunk_method':'Replace exactly the three before-owner emissions (including exactly two namespace export assignments) with their after-owner texts, then group every remaining line-diff opcode with three context lines. A mixed rename/removal hunk is class b only if normalization plus declaration deletion exactly matches and every removed symbol reference is inside one of the three owners or another removed declaration accepted by the least transitive closure. Unanchored cycles fail closed.','normalization':'Only AST Identifier nodes have terminal decimal suffixes removed. No strings, comments, numeric literals, whitespace or punctuation are rewritten.','removed_declaration_rule':'A removed declaration is class b when every reference is inside a reviewed owner or inside another removed declaration that is itself class b, closed transitively.','bundles':reports,'ready_to_land':False}
(unit/'debug-display-2-bundle-classification.json').write_text(json.dumps(report,indent=2)+'\n')
with gzip.open(unit/'debug-display-2-all-emitted-javascript-hunks.patch.gz','wb') as f:f.write(''.join(full_patch).encode())
print(json.dumps({'status':report['status'],'counts':counts,'bundles':[{'file':r['file'],'counts':r['counts'],'global_canonical_equality':r['all_residual_bytes_equal_after_normalization_and_declaration_deletion']} for r in reports]}))
raise SystemExit(1 if counts['c'] else 0)
