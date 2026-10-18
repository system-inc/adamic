from pathlib import Path
import subprocess,json,re,time
root=Path.cwd();ev=root/'review/compiler/loom-r-views-fx-a';cli='/tmp/r-views-adamic';rows=[]
def run(name,args,limit=60):
 start=time.monotonic();r=subprocess.run(list(map(str,args)),cwd=root,capture_output=True,text=True,timeout=limit)
 row=dict(name=name,command=list(map(str,args)),exit=r.returncode,stdout=r.stdout,stderr=r.stderr,seconds=round(time.monotonic()-start,3));rows.append(row);return row
def node(path):return run('source-node',['node','--disable-warning=ExperimentalWarning','oracle/node.mjs',path])
def program_body(path):
 source=path.read_text()
 return source.split('\n',1)[1] if source.startswith('// a-check:') else source
try:
 tests=(root/'internal/lower/loom_views_a_test.go').read_text()
 cases=re.findall(r'requireLoomViewRefusal\(t, "([^"]+)", ("(?:\\.|[^"\\])*"), ("(?:\\.|[^"\\])*"), ("(?:\\.|[^"\\])*")\)',tests)
 assert len(cases)==11,len(cases)
 for name,output,diagnostic,fix in cases:
  output,diagnostic,fix=map(json.loads,[output,diagnostic,fix]);path=root/'internal/lower/testdata/loom_views_a'/f'{name}.a'
  truth=node(path);assert (truth['exit'],truth['stdout'],truth['stderr'])==(0,output,''),truth
  for backend in ['c','js']:
   got=run(name+'-'+backend,[cli,backend,path])
   assert got['exit']!=0 and diagnostic in got['stderr'] and fix in got['stderr'] and '.a:' in got['stderr'],got
 controls=ev/'controls';controls.mkdir(exist_ok=True)
 key=program_body(root/'internal/lower/testdata/loom_views_a/key.a').replace("view['count']",'view.count').replace("count: 'seven'",'count: 7',1)
 unread=program_body(root/'internal/lower/testdata/loom_views_a/unsupported.a').replace('console.log(`${view.label} / ${view.count}`);',"console.log('created');",1)
 for name,source,expected in [('key_fix',key,'14\ntrue\n'),('unsupported_unread',unread,'created\n')]:
  path=controls/f'{name}.a';path.write_text(source)
  truth=node(path);assert (truth['exit'],truth['stdout'],truth['stderr'])==(0,expected,''),truth
  js=run(name+'-js-compile',[cli,'js',path]);assert js['exit']==0,js
  jsfile=Path('/tmp')/f'r-views-{name}.mjs';jsfile.write_text(js['stdout']);js.pop('stdout')
  got=run(name+'-javascript',['node','--disable-warning=ExperimentalWarning','oracle/node.mjs',jsfile]);assert (got['exit'],got['stdout'],got['stderr'])==(0,expected,''),got
  for variant,flags in [('release',[]),('sanitized',['--sanitize'])]:
   binary=Path('/tmp')/f'r-views-{name}-{variant}'
   built=run(name+'-'+variant+'-build',[cli,'build',path,'-o',binary,*flags]);assert built['exit']==0,built
   got=run(name+'-'+variant,[binary]);assert (got['exit'],got['stdout'],got['stderr'])==(0,expected,''),got
 print('11 Node-pinned probes refused by both CLI backends; two controls agree in JavaScript, release native and sanitized native.')
finally:
 (ev/'entrypoints.json').write_text(json.dumps(rows,indent=2)+'\n')
