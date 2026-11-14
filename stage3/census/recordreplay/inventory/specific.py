from pathlib import Path
import json,ast,subprocess as s,hashlib,shutil,os
D=Path('/workspace/scratch/recordreplay');R=D/'head';O=D/'inventory-audit';result={}
def run(name,argv,expected):
 p=s.run(['python3',str(D/'run.py'),'inventory-'+name,*map(str,argv)],cwd=R)
 assert p.returncode==expected,(name,p.returncode,expected)
 resource=json.loads((D/('inventory-'+name+'.json')).read_text());log=(D/('inventory-'+name+'.log')).read_text();result[name]={'exit':p.returncode,'command':list(map(str,argv)),'resource':resource,'diagnostic':log[:3000] if expected else 'pass'}
# Capture the original size guard failing even without the new declaration.
old=O/'old-99-check.cjs';old.write_bytes(s.check_output(['git','show','e00ce47c6:stage3/adapt/99-wasm-crypto-import/proofs/check.cjs'],cwd=R))
run('99-old-on-control',['node',old,O/'before99',D/'control',O/'unused.json'],1)
run('93-source-scope',['node','-e',"require(process.argv[1]).TestBuiltinSourceScope(process.argv[2],process.argv[3])",R/'stage3/adapt/93-builtin-imports/proofs/check.cjs',D/'control',D/'candidate2'],1)
run('92-source-scope',['node','-e',"require(process.argv[1]).TestTTYSourceScope(process.argv[2],process.argv[3])",R/'stage3/adapt/92-tty-host-results/proofs/check.cjs',D/'control',D/'candidate2'],1)
run('91-source-scope',['node',R/'stage3/adapt/91-patcher-discard/proofs/scope.cjs',D/'control',D/'candidate2',O/'91.json'],1)
# Positive and negative source-scope controls for the repaired current reader.
run('99-current',['node',R/'stage3/adapt/99-wasm-crypto-import/proofs/check.cjs',O/'before99',D/'candidate2',O/'99-current.json'],0)
mutant=O/'mutant';shutil.copytree(D/'candidate2/src',mutant/'src',copy_function=os.link)
f=mutant/'src/recordreplay.d.ts';original=f.read_bytes();f.unlink();f.write_bytes(original.replace(b'unknown',b'number'))
run('99-declaration-mutant',['node',R/'stage3/adapt/99-wasm-crypto-import/proofs/check.cjs',O/'before99',mutant,O/'mutant.json'],1);f.write_bytes(original)
extra=mutant/'src/unreviewed.d.ts';extra.write_text('interface Unreviewed {}\n')
run('99-extra-file-mutant',['node',R/'stage3/adapt/99-wasm-crypto-import/proofs/check.cjs',O/'before99',mutant,O/'mutant.json'],1);extra.unlink()
# Execute scoped.py's actual planning/inventory segment, omitting its build/oracle calls.
code=(R/'stage3/adapt/99-wasm-crypto-import/proofs/scoped.py').read_text();a=code.index("main=D/'main';");b=code.index("file=scoped/'src/compiler/sys.ts'",a);segment=code[a:b];tree=ast.parse(segment);tree.body=[n for n in tree.body if not isinstance(n,ast.Expr) or not isinstance(n.value,ast.Call) or not isinstance(n.value.func,ast.Name) or n.value.func.id!='run']
workspace=O/'scoped-segment';workspace.mkdir();(workspace/'main').symlink_to(D/'control',target_is_directory=True);(workspace/'scoped').symlink_to(D/'candidate2',target_is_directory=True)
ns={'D':workspace,'P':R/'stage3/adapt/99-wasm-crypto-import/proofs','s':s,'h':lambda b:hashlib.sha256(b).hexdigest(),'json':json}
exec(compile(tree,'scoped.py inventory segment','exec'),ns)
assert ns['actualSources']==ns['expectedSources'];assert len(ns['actualSources'])==748
result['99-scoped-inventory-segment']={'status':'pass','source_files':748,'exact_source_plan_matches':True,'method':'actual planning and inventory statements; historical apply/build/lane calls omitted'}
# Frozen source-list validator: use the actual measured tree, including its built/test files.
text=(R/'stage3/census/speculative/server-run.sh').read_text();a=text.index('def source_checks(');b=text.index('\n\ndef ',a+1);ns={'digest':lambda b:hashlib.sha256(b).hexdigest()};exec(text[a:b],ns)
root=D/'candidate2';files=[p for p in root.rglob('*') if p.is_file() and p.suffix in ('.ts','.a')]
identity={'files':[{'file':str(p.relative_to(root)),'bytes':p.stat().st_size,'sha256':hashlib.sha256(p.read_bytes()).hexdigest()} for p in files]};ns['source_checks'](root,identity)
result['speculative-source-checks-fresh']={'status':'pass','files':len(files),'includes_added':any(r['file']=='src/recordreplay.d.ts' for r in identity['files'])}
identity['files']=[r for r in identity['files'] if r['file']!='src/recordreplay.d.ts']
try:ns['source_checks'](root,identity);raise RuntimeError('stale inventory unexpectedly admitted')
except AssertionError as e:result['speculative-source-checks-stale']={'status':'reject','reason':str(e),'only_omitted_file':'src/recordreplay.d.ts'}
(O/'specific.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps({k:v.get('status',v.get('exit')) for k,v in result.items()}))
