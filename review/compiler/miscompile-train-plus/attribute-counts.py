from pathlib import Path
import subprocess,difflib
out=Path('review/compiler/miscompile-train-plus/count-attribution');out.mkdir(exist_ok=True)
paths=['internal/oracle/testdata/review/agree/fxspptb_oct9_native_p04_undefined_slot_write.a']+['internal/lower/testdata/scalar_union_views/p54_'+n+'.a' for n in ['n','b','s']]+['stage3/interface-downcasts/v2/snapshot-maybe-'+n+'.a' for n in ['boolean','number']]+['stage3/interface-downcasts/v2/undefined-read-write.a']
with (out/'merged-build.log').open('w') as f:subprocess.run(['go','build','-o','/tmp/train-plus-merged-adamic','./cmd/adamic'],stdout=f,stderr=subprocess.STDOUT,timeout=120,check=True)
for p in paths:
 name=Path(p).stem
 with (out/(name+'-merged.c.txt')).open('w') as f:subprocess.run(['/tmp/train-plus-merged-adamic','c',p],stdout=f,stderr=subprocess.STDOUT,timeout=30,check=True)
 a=(out/(name+'-train.c.txt')).read_text();b=(out/(name+'-merged.c.txt')).read_text()
 (out/(name+'.diff')).write_text(''.join(difflib.unified_diff(a.splitlines(True),b.splitlines(True),fromfile='train',tofile='merged')))
# Isolate reference-tag guard: restoring it must recover p04's two and p54's one null retain.
p=Path('internal/native/view_unions_read.go');original=p.read_text()
a='\te.line("else if (%s.kind >= adamic_view_union_string && %s.kind <= adamic_view_union_function) %s = adamic_retain(%s.payload.reference);", snapshot, snapshot, boxed, snapshot)\n\te.line("else %s = NULL;", boxed)'
b='\te.line("else %s = adamic_retain(%s.payload.reference);", boxed, snapshot)'
try:
 assert original.count(a)==1
 p.write_text(original.replace(a,b))
 with (out/'unguarded-build.log').open('w') as f:subprocess.run(['go','build','-o','/tmp/train-plus-unguarded-adamic','./cmd/adamic'],stdout=f,stderr=subprocess.STDOUT,timeout=120,check=True)
 for fixture in paths[:4]+paths[-1:]:
  name=Path(fixture).stem
  with (out/(name+'-unguarded-build.log')).open('w') as f:subprocess.run(['/tmp/train-plus-unguarded-adamic','build',fixture,'-o','/tmp/train-plus-unguarded-'+name,'--count'],stdout=f,stderr=subprocess.STDOUT,timeout=90,check=True)
  with (out/(name+'-unguarded-count.log')).open('w') as f:subprocess.run(['/tmp/train-plus-unguarded-'+name],stdout=f,stderr=subprocess.STDOUT,timeout=30,check=True)
finally:p.write_text(original)
# Isolate spread checking only for snapshot fixture count attribution, not a proposed code change.
p=Path('internal/lower/object.go');original=p.read_text()
a='\t\t\tspread, err = l.checkedViewMembers(property, property.AsSpreadAssignment().Expression, spread, nil)\n\t\t\tif err != nil {\n\t\t\t\treturn nil, err\n\t\t\t}\n'
try:
 assert original.count(a)==1
 p.write_text(original.replace(a,''))
 with (out/'spread-bypass-build.log').open('w') as f:subprocess.run(['go','build','-o','/tmp/train-plus-spread-bypass-adamic','./cmd/adamic'],stdout=f,stderr=subprocess.STDOUT,timeout=120,check=True)
 for fixture in paths[4:6]:
  name=Path(fixture).stem
  with (out/(name+'-spread-bypass-build.log')).open('w') as f:subprocess.run(['/tmp/train-plus-spread-bypass-adamic','build',fixture,'-o','/tmp/train-plus-spread-'+name,'--count'],stdout=f,stderr=subprocess.STDOUT,timeout=90,check=True)
  with (out/(name+'-spread-bypass-count.log')).open('w') as f:subprocess.run(['/tmp/train-plus-spread-'+name],stdout=f,stderr=subprocess.STDOUT,timeout=30,check=True)
finally:p.write_text(original)
