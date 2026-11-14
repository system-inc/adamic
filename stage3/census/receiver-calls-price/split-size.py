import pathlib,json,difflib,hashlib
scratch=pathlib.Path('/workspace/scratch');rewrite=json.loads((scratch/'receiver-split-rewrite.json').read_text());rows=[];patch=[]
for site in rewrite['files']:
 file=site['file'];a=(scratch/'receiver-baseline'/file).read_text().splitlines();b=(scratch/'receiver-split'/file).read_text().splitlines();add=remove=0
 for tag,i,j,k,l in difflib.SequenceMatcher(a=a,b=b,autojunk=False).get_opcodes():
  if tag!='equal':remove+=j-i;add+=l-k
 rows.append({'file':file,'removedLines':remove,'addedLines':add,'changedLinesTotal':remove+add,'parserCallsRewritten':sum(m.get('rewrittenUses',0) for m in site['members']),'transformerArrows':len(site['members']) if site['kind']=='TransformationContext' else 0,'directUsesRetained':sum(m.get('directUses',0) for m in site['members']),'valueUsesRetained':sum(m.get('valueUses',0) for m in site['members'])})
 patch.extend(difflib.unified_diff(a,b,fromfile='baseline/'+file,tofile='split/'+file,lineterm=''))
text='\n'.join(patch)+'\n';(scratch/'receiver-split-source.patch').write_text(text)
(scratch/'receiver-split-results/rewrite-size.json').write_text(json.dumps({'files':rows,'privatePatch':{'bytes':len(text.encode()),'sha256':hashlib.sha256(text.encode()).hexdigest()}},indent=2)+'\n')
