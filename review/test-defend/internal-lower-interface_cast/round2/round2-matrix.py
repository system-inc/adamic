import pathlib,subprocess,json,time,os,difflib
R=pathlib.Path('/workspace/adamic');P=R/'review/test-defend/internal-lower-interface_cast/round2';PKG='./internal/lower/'
plans=[
 dict(id='G03',test='TestIteratorGapsAreExplicit',file='internal/lower/object.go',old='\t\t\treturn nil, l.notYet(node, "a spread whose element representation differs from its destination")',new='',kind='drop statement',difference='last input: sole custom spread into number|undefined array, absent from receiver-view subsumer'),
 dict(id='F01',test='TestDefaultTaggedInterfaceNeedsNoFlag',file='internal/lower/interface_cast.go',old='(field.Flags&ast.SymbolFlagsOptional != 0 || access.QuestionDotToken != nil || accessorSymbol(field))',new='(field.Flags&ast.SymbolFlagsOptional == 0 || access.QuestionDotToken != nil || accessorSymbol(field))',kind='flip condition',difference='direct inline checked-interface cast followed by payload read; no exclusive line or independent behavior found versus admission table'),
 dict(id='F02',test='TestDefaultTaggedInterfaceNeedsNoFlag',file='internal/lower/interface_cast.go',old='Allowed: []ir.Expression{allowed}',new='Allowed: []ir.Expression{allowed}[:0]',kind='off-by-one implicit bound',difference='direct inline cast tag-membership boundary; remove its sole allowed tag without changing source Node oracle'),
 dict(id='F03',test='TestDefaultTaggedInterfaceNeedsNoFlag',file='internal/lower/interface_cast.go',old='Field: field, FieldType:',new='Field: "pos", FieldType:',kind='change constant',difference='direct inline cast must check the kind tag rather than the numeric position payload'),
 dict(id='G01',test='TestIteratorGapsAreExplicit',file='internal/lower/iteration.go',old='\t\t\t\t\t\thazard = l.notYet(where, "an iterator built by object spread (own method presence is not proved)")',new='',kind='drop statement',difference='literal protocol-copy/object-spread path absent from receiver-view subsumer'),
 dict(id='G02',test='TestIteratorGapsAreExplicit',file='internal/lower/iteration.go',old='done.Flags&ast.SymbolFlagsOptional != 0',new='done.Flags&ast.SymbolFlagsOptional == 0',kind='flip condition',difference='optional done input versus required boolean done in subsumer')]
orig={x['file']:(R/x['file']).read_text() for x in plans}
for x in plans:
 s=orig[x['file']];assert s.count(x['old'])==1,(x['id'],s.count(x['old']));x['line']=s[:s.index(x['old'])].count('\n')+1
 (P/(x['id']+'.diff')).write_text(''.join(difflib.unified_diff(s.splitlines(True),s.replace(x['old'],x['new'],1).splitlines(True),fromfile='a/'+x['file'],tofile='b/'+x['file'])))
(P/'plan.json').write_text(json.dumps(plans,indent=2));runs=[]
def run(id,cmd,env=None):
 t=time.monotonic()
 with (P/(id+'.log')).open('w') as f:r=subprocess.run(cmd,cwd=R,stdout=f,stderr=subprocess.STDOUT,env=env)
 item=dict(id=id,exit=r.returncode,seconds=time.monotonic()-t,command=cmd,selector=(env or {}).get('ADAMIC_MUTANT'),build_cache=(env or {}).get('ADAMIC_BUILD_CACHE_DIR'));runs.append(item);(P/'runs.json').write_text(json.dumps(runs,indent=2));print(id,r.returncode,round(item['seconds'],2),flush=True)
for x in plans:
 f=R/x['file'];f.write_text(orig[x['file']].replace(x['old'],x['new'],1))
 try:
  run(x['id']+'-vet',['timeout','90','go','vet',PKG]);assert runs[-1]['exit']==0
 finally:f.write_text(orig[x['file']])
for f,s in orig.items():
 for x in (x for x in plans if x['file']==f):
  id=x['id']
  if id in ['G03','G01']:new='if !round2Mutant("'+id+'") { '+x['old'].strip()+' }'
  elif id=='F01':new=x['old'].replace('field.Flags&ast.SymbolFlagsOptional != 0','round2Condition("F01", field.Flags&ast.SymbolFlagsOptional != 0)')
  elif id=='G02':new='round2Condition("'+id+'", '+x['old']+')'
  elif id=='F02':new='Allowed: round2Allowed([]ir.Expression{allowed})'
  else:new='Field: round2Field(field), FieldType:'
  s=s.replace(x['old'],new,1)
 (R/f).write_text(s)
helper=R/'internal/lower/defense_round2_selector.go';helper.write_text('''package lower
import("os"; "github.com/system-inc/adamic/internal/ir")
func round2Mutant(id string)bool{return os.Getenv("ADAMIC_MUTANT")==id}
func round2Condition(id string,v bool)bool{if round2Mutant(id){return !v};return v}
func round2Allowed(v []ir.Expression)[]ir.Expression{if round2Mutant("F02"){return v[:0]};return v}
func round2Field(v string)string{if round2Mutant("F03"){return "pos"};return v}
''')
def events(id):
 out=[]
 for line in (P/(id+'.log')).read_text().splitlines():
  try:out.append(json.loads(line))
  except:pass
 return out
try:
 run('switch-vet',['timeout','90','go','vet',PKG]);assert runs[-1]['exit']==0;gap_defended=False
 for x in plans:
  if x['id'] in ['G01','G02'] and gap_defended:continue
  env=os.environ.copy();env['ADAMIC_MUTANT']=x['id'];env['ADAMIC_BUILD_CACHE_DIR']='/workspace/defend-lower-round2-cache/'+x['id']
  run(x['id'],['timeout','120','go','test','-json','-count=1','-timeout','90s',PKG,'-run','.'],env)
  es=events(x['id']);fails=[e['Test'] for e in es if e.get('Action')=='fail' and e.get('Test') and '/' not in e['Test']];print('FAILS',x['id'],fails,flush=True)
  if x['id']=='G03' and fails==['TestIteratorGapsAreExplicit']:gap_defended=True
finally:
 for f,s in orig.items():(R/f).write_text(s)
 helper.unlink()
