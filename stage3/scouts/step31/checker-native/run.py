#!/usr/bin/env python3
"""Measure the actual pinned CommonJS observer, without inventing a native driver."""
import gzip,hashlib,json,os,re,subprocess,time
from pathlib import Path
here=Path(__file__).resolve().parent
out=here/'evidence';out.mkdir(exist_ok=True)
compiler=Path('/tmp/step31-native-adamic')
assert compiler.is_file(), 'build scratch compiler first'
def run(stem,command,split=None):
    env=dict(os.environ)
    if split is not None: env['ADAMIC_NATIVE_SPLIT']=str(split)
    start=time.perf_counter()
    with (out/(stem+'.stdout')).open('wb') as stdout,(out/(stem+'.stderr')).open('wb') as stderr:
        result=subprocess.run(command,env=env,stdout=stdout,stderr=stderr,cwd="/tmp/step31-native-compiler" if command[0]==str(compiler) else None)
    (out/(stem+'.exit')).write_text(str(result.returncode)+'\n')
    return {'exit':result.returncode,'seconds':time.perf_counter()-start,'stdout':(out/(stem+'.stdout')).read_text(),'stderr':(out/(stem+'.stderr')).read_text(),'command':command}
rows=[]
for profile,file in [('raw','checker-entry.a'),('node-declarations','checker-node-declarations.a')]:
    pair=[run(f'{profile}-{s}',[str(compiler),'build','/tmp/step31-driver/'+file,'-o',f'/tmp/step31-checker-{profile}-{s}'],s) for s in (0,1)]
    assert pair[0]['exit']==pair[1]['exit']
    assert pair[0]['stderr']==pair[1]['stderr'] and pair[0]['stdout']==pair[1]['stdout']
    first=pair[0]['stderr'].splitlines()[0] if pair[0]['stderr'] else None
    rows.append({'profile':profile,'builds':pair,'first_stop':first,'binary_exists':Path(f'/tmp/step31-checker-{profile}-0').exists()})
probe=here/'probes/commonjs.a'
check=run('probe-types',[str(compiler),'types',str(probe)])
code=re.search(r'error (TS\d+):',check['stderr'])
assert code
text=probe.read_text();header='// a-check: type error '+code.group(1)+'\n'
if not text.startswith('// a-check:'): probe.write_text(header+text)
check=run('probe-types',[str(compiler),'types',str(probe)])
build=run('probe-build',[str(compiler),'build',str(probe),'-o','/tmp/step31-commonjs'])
node_code="require('node:vm').runInNewContext(require('node:fs').readFileSync(process.argv[1],'utf8'),{require,console})"
node=run('probe-node',['node','-e',node_code,str(probe)])
mutant=Path('/tmp/step31-commonjs-mutant.a');mutant.write_text(probe.read_text().replace('/tmp/leaf','/tmp/changed'))
changed=run('probe-mutant',['node','-e',node_code,str(mutant)])
assert node['exit']==changed['exit']==0 and node['stderr']==changed['stderr']==''
assert node['stdout']=='leaf\n' and changed['stdout']=='changed\n'
ranking=json.loads(Path('/tmp/step31-ranking.json').read_text())
reasons=[{'rank':i+1,**{key:r.get(key) for key in ('kind','reason','bytes_revealed_if_fixed_alone')}} for i,r in enumerate(ranking['ranked_reasons'])]
(here/'ranking-reasons.json').write_text(json.dumps(reasons,indent=2)+'\n')
for row in rows:
    text=re.sub(r'^.*?:\d+:\d+: ','',row['first_stop'] or '')
    if text.startswith("stage 0 can't lower ") and text.endswith(' yet'): text='NotYet: '+text[len("stage 0 can't lower "):-len(' yet')]
    row['reason']=text
    row['ranking_matches']=[r for r in reasons if text==r['kind']+': '+r['reason']]
    row['owner']='library' if text.startswith('NotYet: node:module.') else 'compiler'
    row['walk_termination']='The CommonJS entry call is at module scope. There is no enclosing function body to replace; removing it would remove the observer rather than expose its implementation.'
typed=here/'probes/node-require.a'
typed_build=run('typed-probe-build',[str(compiler),'build',str(typed),'-o','/tmp/step31-node-require'])
assert 'stage 0 can' in typed_build['stderr'] and 'node:module.require' in typed_build['stderr']
loader='/workspace/adamic/stage3/drivers/tsc-entry/node.mjs'
node_env=os.environ.get('SCANNER_TYPESCRIPT');os.environ['SCANNER_TYPESCRIPT']='/home/agent/.cache/adamic-stage3/api/node_modules/typescript/lib/typescript.js'
typed_node=run('typed-probe-node',['node',loader,str(typed)])
typed_mutant=Path('/tmp/step31-node-require-mutant.a');typed_mutant.write_text(typed.read_text().replace('/tmp/leaf','/tmp/changed'))
typed_changed=run('typed-probe-mutant',['node',loader,str(typed_mutant)])
assert typed_node['exit']==typed_changed['exit']==0 and typed_node['stderr']==typed_changed['stderr']==''
assert typed_node['stdout']=='leaf\n' and typed_changed['stdout']=='changed\n'
result={'typed_probe':{'build':typed_build,'node':typed_node,'mutant':typed_changed},'candidate' :'79dd1abae85972b8123fb813d2d98e49ef612730','stricter':'8f32e51e8fc41b8f1177453213ca5453ce764486','scratch_merge':subprocess.check_output(['git','-C','/tmp/step31-native-compiler','rev-parse','HEAD'],text=True).strip(),'scout':'2de9fc1b01d3da3beb35c554e70d98cd6d9fe536','ranking_pin':'ec0b16c04f3bbdfbaf3932ab01307f90b08156fd','ranking_sha256':hashlib.sha256(Path('/tmp/step31-ranking.json').read_bytes()).hexdigest(),'driver_bytes_unchanged':Path('/tmp/step31-driver/checker-entry.a').read_bytes()==Path('/tmp/step31-driver/stage3/scouts/step31/checker-dump.cjs').read_bytes(),'profiles':rows,'probe':{'types':check,'build':build,'node':node,'mutant':changed},'native_comparator_run':False,'node_manifest':json.loads(Path('/tmp/step31-native-node-cache/manifest.json').read_text())}
(here/'result.json').write_text(json.dumps(result,indent=2)+'\n')
metadata=subprocess.check_output(['go','version','-m',str(compiler)])
(out/'binary-raw.txt.gz').write_bytes(gzip.compress(metadata,mtime=0))
(here/'binary.txt').write_text('\n'.join(line.rstrip() for line in metadata.decode().splitlines())+'\n')
for name,path in [('compiler-build','/tmp/step31-native-build.log'),('node','/tmp/step31-native-node.log'),('resolution','/tmp/step31-native-resolution.log')]:
 (out/(name+'.txt.gz')).write_bytes(gzip.compress(Path(path).read_bytes(),mtime=0))
print(json.dumps({r['profile']:r['first_stop'] for r in rows}),flush=True)
print('Node 301 baseline retained; CommonJS witness and computed input mutant verified')
