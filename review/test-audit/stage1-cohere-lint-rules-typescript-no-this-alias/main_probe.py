import pathlib,subprocess,json,time,os,difflib
root=pathlib.Path('/workspace/adamic'); ev=pathlib.Path('/tmp/u124/evidence'); file='stage1/cohere/lint/rules/typescript-no-this-alias/profile.a'; source=root/file; base=source.read_text(); anchor='const args = programArguments();'; assert base.count(anchor)==1; changed=base[:base.index(anchor)]
(ev/'P2.diff').write_text(''.join(difflib.unified_diff(base.splitlines(True),changed.splitlines(True),fromfile='a/'+file,tofile='b/'+file)))
catalog=json.loads((ev/'catalog.json').read_text()); catalog.append({'id':'P2','file':file,'line':base[:base.index(anchor)].count('\n')+1,'before':'top-level manifest driver from const args through the end of the module','after':'empty top-level driver (implicit void answer)','kind':'empty executable entry probe'}); (ev/'catalog.json').write_text(json.dumps(catalog,indent=2))
runs=json.loads((ev/'runs.json').read_text())
def run(id,c):
 start=time.monotonic()
 with (ev/(id+'.log')).open('w') as f:p=subprocess.run(['timeout','120',*c],cwd=root,env={**os.environ,'ADAMIC_BUILD_CACHE_DIR':'/tmp/u124/cache/P2'},stdout=f,stderr=subprocess.STDOUT)
 runs.append({'id':id,'command':['timeout','120',*c],'environment':{'ADAMIC_BUILD_CACHE_DIR':'/tmp/u124/cache/P2'},'wall':time.monotonic()-start,'exit':p.returncode}); (ev/'runs.json').write_text(json.dumps(runs,indent=2)); print(id,p.returncode,flush=True); assert p.returncode==0
source.write_text(changed)
try:
 run('P2',['go','test','-json','-count=1','-timeout','90s','./stage1/cohere/lint/rules/typescript-no-this-alias/','-run','.'])
 run('P2-build',['/tmp/u124/adamic','build',file,'-o','/tmp/u124/P2-native'])
 run('P2-output',['/tmp/u124/P2-native',str(ev/'manifest.txt')])
finally:source.write_text(base)
