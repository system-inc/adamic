"""Independent numeric subscriptions and fail-closed snapshot witnesses."""
import argparse,json,pathlib,subprocess
p=argparse.ArgumentParser();p.add_argument('--artifacts',required=True);p.add_argument('--stage0',required=True);p.add_argument('--archive',required=True);p.add_argument('--fixtures',required=True);a=p.parse_args()
s=pathlib.Path(__file__).resolve().parent;r=s.parents[3];out=pathlib.Path(a.artifacts).resolve();out.mkdir(parents=True,exist_ok=True)
def run(name,cmd,expected=0):
 with (out/(name+'.stdout')).open('wb') as stdout,(out/(name+'.stderr')).open('wb') as stderr:proc=subprocess.run([str(x) for x in cmd],cwd=r,stdout=stdout,stderr=stderr)
 data=(out/(name+'.stdout')).read_bytes();errors=(out/(name+'.stderr')).read_bytes();assert proc.returncode==expected,(name,proc.returncode,errors);return data,errors
run('listener-build',[a.stage0,'build',s/'testdata/listener_probe.a','-o',out/'listener'])
native,errors=run('listener-native',[out/'listener']);truth,_=run('listener-go',['go','run',s/'testdata/listener_oracle.go']);assert native==truth and errors==b''
# Names come from the independent Go AST enums, rather than from native declarations.
go_names,_=run('listener-go-names',['go','run',s/'testdata/listener_oracle.go','--json']);name_sets=json.loads(go_names)
for name,kinds in name_sets.items():
 manifest=json.loads((s/'listeners'/name/'rule.json').read_text());assert manifest==dict(name=name,kinds=kinds)
mutant=out/'numeric-mutant';mutant.mkdir(exist_ok=True)
for file in s.glob('*.a'):
 text=file.read_text().replace("'../","'"+str(s.parent)+"/")
 if file.name=='symbol_description.a':before='readonly number[] = [214]';assert text.count(before)==1;text=text.replace(before,'readonly number[] = [215]')
 (mutant/file.name).write_text(text)
(mutant/'listener_probe.a').write_text((s/'testdata/listener_probe.a').read_text().replace("'../","'./"))
run('numeric-mutant-build',[a.stage0,'build',mutant/'listener_probe.a','-o',out/'numeric-mutant-native'])
changed,errors=run('numeric-mutant-run',[out/'numeric-mutant-native']);assert changed!=truth and errors==b''
print('three manifests and numeric exports agree; 214-to-215 mutant compiles, exits zero, differs only by numeric bytes',flush=True)
# Simulate absent compiler metadata. The supported rule must refuse, not go silent.
missing=out/'missing-snapshot';missing.mkdir(exist_ok=True)
for file in s.glob('*.a'):
 text=file.read_text().replace("'../","'"+str(s.parent)+"/")
 if file.name=='syntax_snapshot.a':before='node.present = true';assert text.count(before)==1;text=text.replace(before,'node.present = false')
 (missing/file.name).write_text(text)
run('snapshot-mutant-build',[a.stage0,'build',missing/'suite.a','-o',out/'snapshot-mutant-native','--tsgo',a.archive])
fixture=next(path for path in sorted((pathlib.Path(a.fixtures)/'cases').iterdir()) if json.loads((path/'metadata.json').read_text())['Rule']=='require-await')
data,errors=run('snapshot-mutant-run',[out/'snapshot-mutant-native',fixture/'tsconfig.json',fixture/'roots.manifest','--await'],70)
assert data==b'' and b'function missing from checker syntax snapshot' in errors
print('missing function metadata mutant compiles and is refused before output, exit 70',flush=True)
print('PASS',flush=True)
