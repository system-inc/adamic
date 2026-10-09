import pathlib,json,difflib,re
p=pathlib.Path('review/test-audit/internal-oracle-census_overload_result');(p/'diffs').mkdir(exist_ok=True)
root='internal/lower/';original={n:pathlib.Path(root+n).read_text() for n in ['census_overload_proof.go','census_small.go','lower.go']};(p/'original-files.json').write_text(json.dumps(original))
items=[
('M01','census_overload_proof.go','!l.censusHasUndefined(produced)','l.censusHasUndefined(produced)','flip condition','bool'),
('M02','census_overload_proof.go','!reachable(flow)','reachable(flow)','flip condition','bool'),
('M03','census_overload_proof.go','len(flows) > 256','len(flows) > 0','change constant','bool'),
('M04','census_overload_proof.go','!trusted','trusted','flip condition','bool'),
('M05','census_overload_proof.go','\t\tl.typeMapper = newTypeMapper(sources, targets)\n','','drop statement','drop'),
('M06','census_overload_proof.go','!unchanged','unchanged','flip condition','bool'),
('M07','census_small.go','\t\t\tordinal := 0','\t\t\tordinal := 1','change constant','ordinal'),
('M08','census_small.go','"overload %d of %s result: expected %s, got undefined"','"overload %d of %s result: expected type %s, got undefined"','change constant','string'),
('M09','census_overload_proof.go','admitted[index].Flags()&checker.TypeFlagsUndefined == 0','admitted[index].Flags()&checker.TypeFlagsUndefined != 0','flip condition','bool')]
s=original['lower.go'];start=s.index('func Lower(');brace=s.index('{',start);depth=1;end=brace+1
while depth:
 if s[end]=='{':depth+=1
 elif s[end]=='}':depth-=1
 end+=1
body=s[start:end];header=s[start:brace+1]
items += [('P01','lower.go',body,header+'\n\treturn nil, nil\n}','empty-answer probe','probe'),('P02','lower.go',body,header+'\n\treturn &ir.Program{}, nil\n}','empty IR companion probe','probe')]
manifest=[];switched=original.copy()
for id,file,old,new,menu,mode in items:
 s=original[file];assert s.count(old)==1,(id,s.count(old));line=s[:s.index(old)].count('\n')+1;x=dict(id=id,file=root+file,line=line,old=old,new=new,menu=menu,mode=mode);manifest.append(x);one=s.replace(old,new,1)
 if mode=='probe':
  for l in one.splitlines(True):
   m=re.match(r'\s*"([^\"]+)"\s*$',l)
   if m and m[1].split('/')[-1]+'.' not in one:one=one.replace(l,'',1)
 diff=''.join(difflib.unified_diff(s.splitlines(True),one.splitlines(True),fromfile='a/'+root+file,tofile='b/'+root+file));(p/'diffs'/(id+'.diff')).write_text(diff)
 if mode=='bool':replacement='auditU056Bool("'+id+'", func() bool { return '+old+' }, func() bool { return '+new+' })'
 elif mode=='drop':replacement='\t\tif !auditU056Selected("'+id+'") { l.typeMapper = newTypeMapper(sources, targets) }\n'
 elif mode=='ordinal':replacement='\t\t\tordinal := auditU056Int("'+id+'", 0, 1)'
 elif mode=='string':replacement='auditU056String("'+id+'", '+old+', '+new+')'
 else:
  switched[file]=switched[file].replace(header,header+'\n if auditU056Selected("'+id+'") { '+('return nil, nil' if id=='P01' else 'return &ir.Program{}, nil')+' }',1);continue
 switched[file]=switched[file].replace(old,replacement,1)
(p/'manifest.json').write_text(json.dumps(manifest,indent=2));(p/'plan.md').write_text('Fixed before mutant outcomes. CODE UNDER TEST: Adamic Go lowering, entry Lower, overload relation and return-flow proof. ORACLES: handwritten IR/stop assertions plus untouched-source Node execution. The stopped-source-only row reaches no Adamic code and cannot receive a production mutant or entry probe. Every reached lowering function is listed in reached-functions.txt. Nine fixed-menu mutants, P01 nil entry probe, P02 companion empty-IR probe. No oracle, harness or fixture edits.\n\n'+'\n'.join(x['id']+' '+x['file']+':'+str(x['line'])+' '+x['menu']+' '+repr(x['old'])+' -> '+repr(x['new']) for x in manifest)+'\n')
patch=''.join(''.join(difflib.unified_diff(original[f].splitlines(True),switched[f].splitlines(True),fromfile='a/'+root+f,tofile='b/'+root+f)) for f in original);(p/'switched-source.diff').write_text(patch)
helper='''package lower
import "os"
func auditU056Selected(id string) bool { return os.Getenv("ADAMIC_MUTANT") == id }
func auditU056Bool(id string, old, changed func() bool) bool { if auditU056Selected(id) { return changed() }; return old() }
func auditU056Int(id string, old, changed int) int { if auditU056Selected(id) { return changed }; return old }
func auditU056String(id, old, changed string) string { if auditU056Selected(id) { return changed }; return old }
'''
pathlib.Path(root+'audit_u056.go').write_text(helper);(p/'switched-helper.go.txt').write_text(helper)
