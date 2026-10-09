"""Require the real admission gate to catch an emitted-JavaScript wrong result."""
import json,os,pathlib,shutil,subprocess
out=pathlib.Path('review/compiler/generic-values-next/lint-chain-refresh')
roles={}
for binary in json.loads((out/'observed-proof-compilers.json').read_text()):
    p=subprocess.run(['timeout','45',binary,'admission-lower','stage3/fixtures/generic-values/01_comparer.a'],capture_output=True,text=True)
    assert p.returncode in [0,1]
    roles['head' if p.returncode==0 else 'base']=binary
assert set(roles)=={'head','base'}
(out/'mutant-compilers.json').write_text(json.dumps(roles,indent=2)+'\n')
wrapper='/workspace/generic-values-proof-tmp/lint-refresh-admission-wrong-result.py'
shutil.copyfile('review/compiler/generic-values/admission_wrong_result.py',wrapper)
os.chmod(wrapper,0o755)
env=dict(os.environ,GENERIC_VALUES_HEAD_COMPILER=roles['head'])
args=['timeout','180','/tmp/generic-values-admission-delta','--base','f5236b48','--head','e96fd5c3','--manifest',str(out/'admission-manifest.json'),'--manifest-generator','cloud/admission-corpus/manifest.py','--manifest-generator-revision','543925aa','--json','--workers','4','--compile-timeout','45s','--timeout','10s','--corpus-filter','diff','--base-binary',roles['base'],'--base-lower-binary',roles['base'],'--head-binary',wrapper,'--head-lower-binary',roles['head']]
with (out/'admission-mutant.json').open('w') as stdout,(out/'admission-mutant.log').open('w') as stderr:
    p=subprocess.run(args,env=env,stdout=stdout,stderr=stderr)
assert p.returncode==1,p.returncode
j=json.loads((out/'admission-mutant.json').read_text()); assert j['verdict']=='fail'
r=next(r for r in j['programs'] if r['path'].endswith('/01_comparer.a'))
assert r['agree'] is False
assert r['node']['stdout']==r['native']['stdout']!=r['javascript']['stdout']
assert all(r[b]['exit']==0 and not r[b]['stderr'] for b in ['node','native','javascript'])
print('admission mutant caught',j['phase_seconds']['total'])
