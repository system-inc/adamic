import hashlib, json, pathlib, shutil
out = pathlib.Path(__file__).resolve().parent
source = pathlib.Path('/workspace/scratch/line-map-final')
target = pathlib.Path('/workspace/scratch/line-map-mutant')
assert json.loads((out/'final-apply.status.json').read_text())['child_exit'] == 0
shutil.copytree(source,target)
p = target/'src/compiler/scanner.ts'
before = p.read_bytes()
needle = b'/** @internal */\r\nexport function getLineAndCharacterOfPosition(sourceFile: LineMapSource, position: number): LineAndCharacter;\r\n'
assert before.count(needle) == 1
after = before.replace(needle,b'',1)
p.write_bytes(after)
remaining = after[after.index(b'export function getLineAndCharacterOfPosition'):]
assert remaining.startswith(b'export function getLineAndCharacterOfPosition(sourceFile: SourceFileLike, position: number): LineAndCharacter;')
changed=[]
for p in (source/'src').rglob('*'):
    if p.is_file() and p.read_bytes() != (target/p.relative_to(source)).read_bytes():changed.append(str(p.relative_to(source)))
assert changed == ['src/compiler/scanner.ts']
(out/'mutant.json').write_text(json.dumps(dict(source_tree=str(source),mutant_tree=str(target),changed_files=changed,removed=needle.decode(),before_sha256=hashlib.sha256(before).hexdigest(),after_sha256=hashlib.sha256(after).hexdigest(),public_overload_first=True,implementation_and_getLineStarts_still_LineMapSource=True),indent=2)+'\n')
print(json.dumps(dict(changed_files=changed,public_overload_first=True)))
