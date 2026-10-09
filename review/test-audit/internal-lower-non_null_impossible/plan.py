import pathlib,subprocess,json,difflib,re
p=pathlib.Path('review/test-audit/internal-lower-non_null_impossible');files=['non_null.go','refusals.go','optional_indexing_map.go','optional_indexing.go','optional_widening.go','generic.go','class_inheritance.go','lower.go'];original={f:subprocess.check_output(['git','show','HEAD:internal/lower/'+f],text=True) for f in files};items=[]
reached=[l for l in (p/'functions.log').read_text().splitlines() if not l.startswith('total:') and float(l.split()[-1].strip('%'))>0];(p/'reached-functions.txt').write_text('\n'.join(reached)+'\n')
def add(id,f,old,new,menu='flip condition',mode='bool'):
 assert old in original[f],id;items.append(dict(id=id,file='internal/lower/'+f,line=original[f][:original[f].index(old)].count('\n')+1,old=old,new=new,menu=menu,mode=mode))
add('M01','non_null.go','".ts"','".a"','change constant','string')
add('M02','non_null.go','value.Type() == ir.Number || value.Type() == ir.Boolean','value.Type() == ir.MaybeNumber || value.Type() == ir.Boolean','change constant')
add('M03','non_null.go','"non-null assertion failed at "','""','change constant','string')
add('M04','non_null.go','return value, nil','return ir.NumberConstant{Value: 1}, nil','return early with constant','return')
add('M05','refusals.go','"use a Map, which keeps keys in the order they were added"','"use a Map"','change constant','string')
add('M06','optional_indexing_map.go','return hazard','return false','return early','return')
add('M07','optional_indexing.go','index.Type() != ir.Number','index.Type() != ir.String','change constant')
add('M08','optional_widening.go','!isClassInstance(source)','isClassInstance(source)')
add('M09','optional_widening.go','skip[property.Name]','false','change option')
add('M10','optional_widening.go','target = l.impliedTarget(node)','target = nil','change option','statement')
add('M11','optional_widening.go','" on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"','" on the source type, or build a fresh object with known fields (adamic/no-optional-view)"','change constant','string')
add('M12','generic.go','len(present) == 1 && given.Flags()&checker.TypeFlagsUndefined == 0','len(present) == 0 && given.Flags()&checker.TypeFlagsUndefined == 0','off-by-one bound')
add('M13','generic.go','missing != nil','missing == nil')
add('M14','generic.go','if _, isSet := into[declared]; !isSet {','if _, isSet := into[declared]; isSet {','flip condition','statement')
add('M15','class_inheritance.go','checkABI && (!knownA || !knownB || a != b)','checkABI && (!knownA || !knownB || a == b)')
add('M16','class_inheritance.go','!l.classAssignable(accepts, override)','!l.classAssignable(override, accepts)','swap two arguments')
entries=[('P01','lower.go','func Lower('),('P02','refusals.go','func (l *lowering) refuse('),('P03','optional_widening.go','func (l *lowering) optionalAtSite('),('P04','generic.go','func (l *lowering) inferTypes(')]
for id,f,fn in entries:
 s=original[f];start=s.index(fn);end=s.index('\n}',start)+2;old=s[start:end];ret={'P01':'return nil, nil','P02':'return nil','P03':'return nil','P04':'return'}[id];new=old[:old.index('{')+1]+'\n\t'+ret+'\n}';add(id,f,old,new,'empty-answer probe','probe')
(p/'manifest.json').write_text(json.dumps(items,indent=2));(p/'diffs').mkdir(exist_ok=True);modified=original.copy()
for x in items:
 f=x['file'].split('/')[-1];new=original[f].replace(x['old'],x['new'],1)
 if x['mode']=='probe':
  # Remove imports made unused by dropping the whole entry body.
  for line in new.splitlines(True):
   m=re.match(r'\s*"([^\"]+)"\s*$',line)
   if m:
    alias=m[1].split('/')[-1]
    if alias+'.' not in new:new=new.replace(line,'',1)
 (p/'diffs'/(x['id']+'.diff')).write_text(''.join(difflib.unified_diff(original[f].splitlines(True),new.splitlines(True),fromfile='a/'+x['file'],tofile='b/'+x['file'])))
 old=x['old'];id=x['id'];mode=x['mode']
 if mode=='bool':repl='auditU040Bool("'+id+'", func() bool { return '+old+' }, func() bool { return '+x['new']+' })'
 elif mode=='string':repl='auditU040String("'+id+'", '+old+', '+x['new']+')'
 elif mode=='return':repl='if auditU040Selected("'+id+'") { '+x['new']+' }; '+old
 elif mode=='statement':
  if id=='M14':repl='if _, isSet := into[declared]; auditU040Bool("M14", func() bool { return !isSet }, func() bool { return isSet }) {'
  else:repl='if auditU040Selected("'+id+'") { '+x['new']+' } else { '+old+' }'
 else:
  header=old[:old.index('{')+1];old=header;repl=header+'\n if auditU040Selected("'+id+'") { '+x['new'].split('\n')[1].strip()+' }'
 modified[f]=modified[f].replace(old,repl,1)
(p/'plan.md').write_text('Fixed plan before mutant outcomes. CODE UNDER TEST is Adamic Go lowering and relation analysis; ORACLES are handwritten assertions. All reached functions listed in reached-functions.txt. Sixteen mutations spread across seven files, plus four separate entry probes. Budget limits this below the twenty-mutant aim. Census input is the existing initializer fixture copied to a strict TypeScript project.\n'+ '\n'.join(f"{x['id']} {x['file']}:{x['line']} {x['menu']}: {x['old']} -> {x['new']}" for x in items if not x['id'].startswith('P'))+'\n')
patch=''
for f,s in modified.items():patch+=''.join(difflib.unified_diff(original[f].splitlines(True),s.splitlines(True),fromfile='a/internal/lower/'+f,tofile='b/internal/lower/'+f))
(p/'switched-source.diff').write_text(patch)
(p/'original-files.json').write_text(json.dumps(original))
print('Reached functions',len(reached),'mutants',16)
