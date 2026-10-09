from pathlib import Path
import json,difflib
root=Path('/workspace/adamic');out=root/'review/test-audit/stage1-cohere-yaml-scalars_match_go_shards'; rep={};spec=[]
f='stage1/cohere/yaml/scalars_match_go_shards_test.go';orig=(root/f).read_text();s=orig
for id,fn in [('P1','scalarsMatchGoLoweredProduct'),('P2','scalarsMatchGoNativeProduct'),('P3','scalarsMatchGoGoProduct')]:
 old='func '+fn+'(t *testing.T) string {';new=old+'\n\tif os.Getenv("ADAMIC_AUDIT_SELECTOR") == "'+id+'" { return "" }';s=s.replace(old,new,1);spec.append(dict(id=id,file=f,old=old,new=new,kind='empty product entry probe'))
old='func scalarsMatchGoDisagrees(actual, expected []byte) bool { return !bytes.Equal(actual, expected) }';new='func scalarsMatchGoDisagrees(actual, expected []byte) bool { if os.Getenv("ADAMIC_AUDIT_SELECTOR") == "W1" { return false }; return !bytes.Equal(actual, expected) }';s=s.replace(old,new);spec.append(dict(id='W1',file=f,old=old,new=new,kind='weaken witnessed disagreement'))
# Scratch construction switch only changes options, never expected answers.
helper='\nfunc auditOption(normal, broken, id string) string { if os.Getenv("ADAMIC_AUDIT_SELECTOR") == id { return broken }; return normal }\n'
s+=helper
for id,old,new in [('S1','filepath.Abs("scalar_main.ts")','filepath.Abs(auditOption("scalar_main.ts", "missing.ts", "S1"))'),('S2','os.ReadFile(filepath.Join(lowered, "scalars.c"))','os.ReadFile(filepath.Join(lowered, auditOption("scalars.c", "missing.c", "S2")))'),('S3','"-o", filepath.Join(directory, "go-scalars")','auditOption("-o", "-invalid", "S3"), filepath.Join(directory, "go-scalars")')]:
 assert s.count(old)>=1;s=s.replace(old,new,1);spec.append(dict(id=id,file=f,old=old,new=new,kind='construction option change'))
p=Path('/tmp/u153/scratch/scalars_match_go_shards_test.go');p.parent.mkdir(exist_ok=True);p.write_text(s);rep[str(root/f)]=str(p)
for id,file in [('W2','schema_test.go'),('W3','unist_test.go')]:
 f='stage1/cohere/yaml/'+file;orig=(root/f).read_text();old='if bytes.Equal(side.out, expected) {';assert orig.count(old)==1
 new='if bytes.Equal(side.out, expected) || os.Getenv("ADAMIC_AUDIT_SELECTOR") == "'+id+'" {';p=Path('/tmp/u153/scratch')/file;p.write_text(orig.replace(old,new));rep[str(root/f)]=str(p);spec.append(dict(id=id,file=f,old=old,new=new,kind='weaken witnessed equality'))
Path('/tmp/u153/harness-overlay.json').write_text(json.dumps({'Replace':rep}));(out/'harness-menu.json').write_text(json.dumps(spec,indent=2))
for r in spec:
 orig=(root/r['file']).read_text();r['line']=orig[:orig.index(r['old'])].count('\n')+1
 edited=orig.replace(r['old'],r['new'],1)
 if r['id'].startswith('S'):edited+=helper
 (out/(r['id']+'.diff')).write_text(''.join(difflib.unified_diff(orig.splitlines(True),edited.splitlines(True),fromfile='a/'+r['file'],tofile='b/'+r['file'])))
(out/'harness-menu.json').write_text(json.dumps(spec,indent=2))
