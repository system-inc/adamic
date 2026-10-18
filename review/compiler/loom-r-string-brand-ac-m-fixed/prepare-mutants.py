import json,subprocess
from pathlib import Path
out=Path(__file__).resolve().parent;folder=out/'mutants';folder.mkdir(exist_ok=True)
old=Path('review/compiler/loom-string-brand-2/run-mutants.py');ctx={'__file__':str(old.resolve())};exec(old.read_text().split('results=[]')[0],ctx);cases=[]
for topic,file,before,after,pattern,catcher in ctx['mutants']:
 text=Path(file).read_text();assert text.count(before)==1,(topic,text.count(before));replacement=folder/(topic+'.go.txt');replacement.write_text(text.replace(before,after,1));overlay=folder/(topic+'.json');overlay.write_text(json.dumps({'Replace':{str(Path(file).resolve()):str(replacement)}}));cases.append({'topic':topic,'overlay':str(overlay),'packages':['./internal/lower' if topic in ('literal-refinement','unchecked-missing-cast') else './internal/oracle'],'pattern':pattern,'catcher':catcher})
file=Path('internal/lower/expression.go');text=file.read_text();before='if stringShape, _ := l.stringBrandRepresentation(proven); stringShape {';assert text.count(before)==1;replacement=folder/'paths-representation-regression.go.txt';replacement.write_text(text.replace(before,'if stringShape, _ := l.stringBrandRepresentation(proven); false && stringShape {',1));overlay=folder/'paths-representation-regression.json';overlay.write_text(json.dumps({'Replace':{str(file.resolve()):str(replacement)}}));cases.append({'topic':'paths-representation-regression','overlay':str(overlay),'packages':['./internal/oracle'],'pattern':'^TestElementAccessCompilerAreaBoundaries$/paths$','catcher':'a function returning Path | undefined'});(folder/'cases.json').write_text(json.dumps(cases,indent=2))
s=subprocess.check_output(['git','show','05ab3f8a:review/compiler/loom-r-string-brand-ac-m/mutant-check.py'],text=True);(out/'mutant-check.py').write_text(s)
prefix=s.split("standard=['-count=1'")[0]
rest='''
standard=['-count=1','-timeout','90s','-parallel','4','-json']
audit=json.loads((out/'stage3-audit.json').read_text())
targets=[audit['diagnostic_updates'][0],audit['pending_native_proof'][0]]
records=[]
for item in targets:
 p=Path(item['manifest']);original=p.read_text();decoder=json.JSONDecoder();cursor=0;mutated=None
 while cursor<len(original):
  if original[cursor]!='{':cursor+=1;continue
  entry,end=decoder.raw_decode(original,cursor)
  if entry.get('file')==item['file']:
   raw=original[cursor:end];match=re.search(r'"stage0"\\s*:\\s*',raw);_,value_end=decoder.raw_decode(raw,match.end());mutated=original[:cursor+match.end()]+json.dumps(item['before'],separators=(',',':'))+original[cursor+value_end:];break
  cursor=end
 assert mutated is not None and mutated!=original
 label='mutant-record-'+Path(item['manifest']).parent.name
 try:
  p.write_text(mutated)
  bucket=p.parent.name;test='TestFixtures'+bucket.title()
  run(label,['go','test','./stage3/fixtures','-run','^'+test+'$/'+bucket+'$/'+re.escape(item['file'])+'$/stage0$',*standard],1)
  assert 'gap changed' in (out/(label+'.log')).read_text()
  records.append({'manifest':str(p),'file':item['file'],'mutation':'restore the stale main stage0 record','caught_by':'exact stage0 record comparison','exit':1})
 finally:p.write_text(original)
run('stage3-records-restored',['go','test','./stage3/fixtures','-run','^TestFixtures(Assertions|Namespaces)$',*standard])
(out/'record-mutants.json').write_text(json.dumps(records,indent=2))
'''
(out/'record-mutants.py').write_text(prefix+rest)
