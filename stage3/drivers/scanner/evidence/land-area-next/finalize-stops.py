from pathlib import Path
import subprocess,json,re,time
r=Path('/workspace/scratch/native3-next-walk');rows=json.loads((r/'stops.json').read_text());assert len(rows)==14
raw=(r/'stop-15.stderr').read_text();m=re.search(r'main.c:(\d+):(\d+): error: (.*)',raw);assert m
line,col,msg=m.groups();rows.append({'order':15,'file':'generated/main.c','line':int(line),'column':int(col),'message':msg,'build_exit':1,'phase':'native clang','source_context':'utilities.ts:getNameOfScriptTarget return, after forEachEntry throwing never-return placeholder','qualification':'Discovered after a never-return placeholder; the same clang error also reproduces independently without scanner stubs.','probe':'15-never-string.a'})
for s in rows[:14]:
 s['probe']={1:'01-index.a',2:'02-mutable-namespace.a',3:'03-error-cast.a',4:'04-diagnostic-cast.a',5:'05-uint16.a',6:'06-enum-slot.a',7:'07-string-cast.a',8:'08-array.a',9:'09-generic-return.a',10:'10-any.a',11:'11-computed-field.a',12:'12-map-any.a',13:'13-entries.a',14:'14-any-callback.a'}[s['order']]
 if s['file'].endswith('/main.a'):s['file']='driver/main.a'
 if s['order']==12:s['qualification']='Map<any> depends on the index-signature declaration removal and keyword-object initializer placeholder.'
(r/'stops.json').write_text(json.dumps(rows,indent=2)+'\n')
p=Path('/workspace/scratch/native3-next-probes');f=p/'error-capture-control.a';row={'source':f.read_text(),'scope':'supplementary typed Error capture control, not an extra ordered scanner stop'}
for kind,cmd in [('node',['node','--disable-warning=ExperimentalWarning','/workspace/scanner-native3-next/oracle/node.mjs',str(f)]),('build',['/workspace/scratch/scanner-next-adamic','build',str(f),'-o',str(p/'error-capture-control.native')])]:
 start=time.monotonic()
 with (p/('error-capture-control.'+kind+'.stdout')).open('wb') as out,(p/('error-capture-control.'+kind+'.stderr')).open('wb') as err:row[kind+'_exit']=subprocess.run(cmd,cwd='/workspace/scanner-native3-next',stdout=out,stderr=err).returncode
 row[kind+'_seconds']=time.monotonic()-start;row[kind+'_stdout']=(p/('error-capture-control.'+kind+'.stdout')).read_text();row[kind+'_stderr']=(p/('error-capture-control.'+kind+'.stderr')).read_text()
assert row['node_exit']==0 and row['build_exit']==1 and 'node:globals.ErrorConstructor.captureStackTrace' in row['build_stderr'];(p/'error-capture-control.json').write_text(json.dumps(row,indent=2)+'\n')
print('PASS: fifteenth clang stop recorded and supplementary ErrorConstructor witness confirmed.')
