import os,time,json,shutil,subprocess
from pathlib import Path
repo=Path('/workspace/adamic');work=Path('/workspace/scratch/brands3');tree=work/'retyped-tree';old=work/'without91';evidence=repo/'stage3/adapt/40-explicit-any/evidence/brands-undefined-3';proof=evidence/'adapt91';proof.mkdir(exist_ok=True);logs=work/'proof-logs';logs.mkdir(exist_ok=True)
while Path('/proc/12053').exists():time.sleep(2)
for n in ['90','91','92']:
 r=json.loads((work/('adapt'+n+'-rerun.log')).read_text());assert r['files']==0,r
 (evidence/('adapt'+n+'-rerun.json')).write_text(json.dumps(r,indent=2)+'\n')
assert (tree/'patch-set.md').is_file()
shutil.copytree(tree,old,ignore=shutil.ignore_patterns('built','node_modules'))
for file,body in [('src/compiler/factory/nodeFactory.ts','factory'),('src/compiler/utilities.ts','objectAllocator')]:
 p=old/file;s=p.read_bytes();before=('fn => { fn('+body+'); }').encode();after=('fn => fn('+body+')').encode();assert s.count(before)==1;p.write_bytes(s.replace(before,after))
assert (tree/'package-lock.json').read_bytes()==(repo/'stage3/typescript/package-lock.json').read_bytes()
for root in [tree,old]:(root/'node_modules').symlink_to(repo/'stage3/typescript/node_modules',target_is_directory=True)
env=dict(os.environ,NODE_PATH=str(repo/'stage3/api/node_modules'),PYTHONDONTWRITEBYTECODE='1');rows=[]
def run(name,args,cwd=repo,expected=0,limit=600):
 start=time.monotonic()
 with (logs/(name+'.txt')).open('w') as f:code=subprocess.run(list(map(str,args)),cwd=cwd,env=env,stdout=f,stderr=subprocess.STDOUT,timeout=limit).returncode
 rows.append(dict(name=name,command=list(map(str,args)),cwd=str(cwd),expected=expected,exit=code,seconds=time.monotonic()-start));(evidence/'preparation-commands.json').write_text(json.dumps(rows,indent=2)+'\n')
 assert code==expected,(name,code)
p=repo/'stage3/adapt/91-patcher-discard/proofs'
run('91-no-registrant',['node',p/'references.cjs',tree,proof/'references.json'])
run('91-body-only',['node',p/'scope.cjs',old,tree,proof/'scope.json'])
run('retyped-build',['npm','run','build'],tree)
run('without91-build',['npm','run','build'],old)
run('91-original-node-witness',['node',p/'hooks.cjs',old/'built/local/typescript.js',proof/'hooks-original.json'])
run('91-before-node-witness',['node',p/'hooks-check.cjs',old/'built/local/typescript.js','[1,0,1,0]',proof/'hooks-before.json'])
run('91-after-node-witness',['node',p/'hooks-check.cjs',tree/'built/local/typescript.js','[1,1,1,1]',proof/'hooks-after.json'])
run('91-main-node-witness',['node',p/'hooks-check.cjs',Path('/workspace/scratch/brands2/control-tree/built/local/typescript.js'),'[1,1,1,1]',proof/'hooks-main.json'])
run('91-original-witness-still-red-after',['node',p/'hooks.cjs',tree/'built/local/typescript.js',proof/'unexpected-original-hooks.json'],expected=1)
run('brand-member-audit',['node',repo/'stage3/adapt/40-explicit-any/brands-undefined-proof.cjs','audit',tree,evidence])
run('retype-only-source-api',['node',repo/'stage3/adapt/40-explicit-any/retype-check.cjs','/workspace/scratch/brands2/control-tree',tree,evidence])
(tree/'node_modules').unlink()
(evidence/'preparation.json').write_text(json.dumps({'status':'pass','apply_exit':0,'91_no_registrant':True,'91_body_only':True,'91_node_witnesses_unchanged':True,'own_checks_90_91_92_idempotent':True,'brand_audit_reads_writes_tests':[0,0,0]},indent=2)+'\n')
(work/'preparation-done.json').write_text('{"status":"pass"}\n')
print('PASS full apply and preparation proofs')
