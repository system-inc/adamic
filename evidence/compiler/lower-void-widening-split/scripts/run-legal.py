from pathlib import Path
import subprocess,json
root=Path('/workspace/adamic');out=Path('/tmp/void-split-evidence')
source='internal/oracle/testdata/void_widening_legal.a'
commands=[('native-build',['/tmp/adamic-void-split','build',source,'-o',str(out/'legal-native'),'--sanitize']),('js-build',['/tmp/adamic-void-split','js',source]),('node',['node','--disable-warning=ExperimentalWarning',str(root/'oracle/node.mjs'),str(root/source)]),('native',[str(out/'legal-native')]),('js',['node','--disable-warning=ExperimentalWarning',str(root/'oracle/node.mjs'),str(out/'legal-backend.mjs')])]
exits={}
for name,cmd in commands:
 p=subprocess.run(cmd,cwd=root,capture_output=True);exits[name]=p.returncode
 (out/('legal-'+name+'.stdout')).write_bytes(p.stdout);(out/('legal-'+name+'.stderr')).write_bytes(p.stderr)
 assert p.returncode==0,(name,p.stderr)
 if name=='js-build':(out/'legal-backend.mjs').write_bytes(p.stdout)
(out/'legal-runs.json').write_text(json.dumps(exits,indent=2)+'\n')
assert (out/'legal-native.stdout').read_bytes()==(out/'legal-js.stdout').read_bytes()==(out/'legal-node.stdout').read_bytes()
assert not any((out/('legal-'+name+'.stderr')).read_bytes() for name in ['native','js','node'])
print((out/'legal-native.stdout').read_text(),end='')
