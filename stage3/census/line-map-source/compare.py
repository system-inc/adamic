import difflib, hashlib, importlib.util, json, pathlib, shutil, subprocess, sys
sys.dont_write_bytecode = True
out = pathlib.Path(__file__).resolve().parent
repo = pathlib.Path('/workspace/adamic')
before, after = out/'before', out/'after'
result = {}
base_tree = pathlib.Path('/workspace/scratch/line-map-before')
final_tree = pathlib.Path('/workspace/scratch/line-map-final')
base_files = {p.relative_to(base_tree): p for p in (base_tree/'src').rglob('*') if p.is_file()}
final_files = {p.relative_to(final_tree): p for p in (final_tree/'src').rglob('*') if p.is_file()}
assert base_files.keys() == final_files.keys()
changed = [str(name) for name in sorted(base_files) if base_files[name].read_bytes() != final_files[name].read_bytes()]
assert changed == ['src/compiler/scanner.ts','src/compiler/types.ts'], changed
pins = json.loads((repo/'stage3/adapt/71-writable-views/line-map-source.json').read_text())
old = (base_tree/'src/compiler/types.ts').read_bytes().decode()
expected = old.replace(pins['sourceFileLike'], pins['lineMapSource']+'\r\n\r\n'+pins['sourceFileLike'])
assert (final_tree/'src/compiler/types.ts').read_bytes().decode() == expected
old = (base_tree/'src/compiler/scanner.ts').read_bytes().decode()
expected = old.replace('    type LineAndCharacter,','    type LineAndCharacter, type LineMapSource,').replace(pins['starts']['before'],pins['starts']['after']).replace(pins['position']['before'],pins['position']['after'])
assert (final_tree/'src/compiler/scanner.ts').read_bytes().decode() == expected
result['exact_measured_e'] = dict(changed_source_files=changed,source_file_like_and_documentation_unchanged=True,bodies_unchanged=True,exact_bytes=True)

for name in ['javascript.json','runtime-exports.json','witness.json']:
    a, b = (before/name).read_bytes(), (after/name).read_bytes()
    result[name] = dict(identical=a==b, before_sha256=hashlib.sha256(a).hexdigest(), after_sha256=hashlib.sha256(b).hexdigest())
    assert a == b, name
for name in json.loads((before/'javascript.json').read_text()):
    relative = pathlib.Path('javascript')/name['name'].replace('.ts','.js')
    assert (before/relative).read_bytes() == (after/relative).read_bytes(), str(relative)
a, b = (before/'typescript.d.ts').read_bytes(), (after/'typescript.d.ts').read_bytes()
(out/'public.diff').write_text(''.join(difflib.unified_diff(a.decode().splitlines(True),b.decode().splitlines(True),fromfile='before/typescript.d.ts',tofile='after/typescript.d.ts')))
child = subprocess.run(['node',str(repo/'stage3/lane/normalize-api.cjs'),str(before/'typescript.d.ts'),str(before/'typescript.d.ts'),str(after/'typescript.d.ts')],capture_output=True)
(out/'normalize-api.status.json').write_text(json.dumps(dict(child_exit=child.returncode,stderr=child.stderr.decode()),indent=2)+'\n')
assert child.returncode == 0
normalized = json.loads(child.stdout)
(out/'normalized-public-api.json').write_bytes(child.stdout)
assert normalized['composed_changes'] == [], normalized
result['public_api'] = dict(byte_identical=a==b,normalized_changes=len(normalized['composed_changes']))
source = pathlib.Path('/workspace/scratch/line-map-typescript')
lane = out/'lane'
tree = lane/'adapted-tree'
tree.mkdir(parents=True,exist_ok=True)
commands=[]
def run(args):
    c = subprocess.run(args,capture_output=True)
    commands.append(dict(command=args,child_exit=c.returncode,stderr=c.stderr.decode()))
    assert c.returncode == 0, (args,c.stderr)
    return c.stdout
run(['git','init','-q',str(tree)])
objects=run(['git','-C',str(source),'rev-parse','--path-format=absolute','--git-path','objects']).decode().strip()
(tree/'.git/objects/info/alternates').write_text(objects+'\n')
pin='050880ce59e30b356b686bd3144efe24f875ebc8'
run(['git','-C',str(tree),'update-ref','HEAD',pin])
original=run(['git','-C',str(tree),'show',pin+':tests/baselines/reference/api/typescript.d.ts'])
for folder,text in [('reference',original),('local',b)]:
    p=tree/'tests/baselines'/folder/'api/typescript.d.ts';p.parent.mkdir(parents=True,exist_ok=True);p.write_bytes(text)
(lane/'oracle').mkdir(exist_ok=True)
(lane/'oracle/baseline.diff').write_text(''.join(difflib.unified_diff(original.decode().splitlines(True),b.decode().splitlines(True),fromfile='reference/api/typescript.d.ts',tofile='local/api/typescript.d.ts')))
sys.path.insert(0,str(repo/'stage3/lane'))
import check
sanction_path=repo/'stage3/lane/sanctioned-api.json'
sanction_before=hashlib.sha256(sanction_path.read_bytes()).hexdigest()
errors,facts=check.api_check(lane,json.loads(sanction_path.read_text()))
result['lane_api']=dict(errors=errors,facts=facts,setup_commands=commands,sanctioned_api_sha256=sanction_before,sanctioned_api_unchanged=hashlib.sha256(sanction_path.read_bytes()).hexdigest()==sanction_before)
(out/'comparison.json').write_text(json.dumps(result,indent=2)+'\n')
assert not errors, errors
assert result['lane_api']['sanctioned_api_unchanged']
shutil.rmtree(tree/'.git')
print(json.dumps(result,indent=2))
