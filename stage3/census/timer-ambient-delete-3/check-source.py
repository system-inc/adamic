from pathlib import Path
import hashlib,json,subprocess
p=Path(__file__).parent;repo=Path('/workspace/timer-ambient-delete');before=Path('/workspace/scratch/timer-road5-before');after=Path('/workspace/scratch/timer-ambient-delete-3-after')
ref='refs/stage3/99-wasm-crypto-import-after'
files=subprocess.check_output(['git','-C',str(before),'ls-tree','-r','--name-only',ref,'src'],text=True).splitlines()
changed=[]
for file in files:
 b=subprocess.check_output(['git','-C',str(before),'show',ref+':'+file]);a=(after/file).read_bytes()
 if b!=a:changed.append(file)
 if file=='src/compiler/sys.ts':sys_before=b;sys_after=a
assert changed==['src/compiler/sys.ts'],changed
expected=sys_before
for line in [b'declare function setTimeout(handler: (label?: unknown) => void, timeout: number, label?: unknown): unknown;',b'declare function clearTimeout(handle: unknown): void;']:
 assert expected.count(line)==1;expected=expected.replace(line+b'\r\n',b'')
old=b'clearTimeout: (timeoutId: unknown): void => clearTimeout(timeoutId)';new=b'clearTimeout: (timeoutId: Parameters<typeof clearTimeout>[0]): void => { clearTimeout(timeoutId); }';assert expected.count(old)==2;expected=expected.replace(old,new);assert expected==sys_after
p.joinpath('sys.before.ts.txt').write_bytes(sys_before);p.joinpath('sys.after.ts.txt').write_bytes(sys_after)
p.joinpath('source-scope.json').write_text(json.dumps({'base':'f37253b347c4f7e7bc223ed05fe04edb8cef0cd4','head':subprocess.check_output(['git','-C',str(repo),'rev-parse','HEAD'],text=True).strip(),'source_files':len(files),'changed_files':changed,'exact_reviewed_edit':True,'CRLF_preserved':True,'sys_before_sha256':hashlib.sha256(sys_before).hexdigest(),'sys_after_sha256':hashlib.sha256(sys_after).hexdigest()},indent=2)+'\n')
print('Only sys.ts changes, by exactly the ambient pair deletion and two block-bodied cancellation forwarders; CRLF preserved.')
