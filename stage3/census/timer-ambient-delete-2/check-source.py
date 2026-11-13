from pathlib import Path
import hashlib,json,subprocess
b=Path('/workspace/scratch/timer-road5-before');a=Path('/workspace/scratch/timer-road5-after');g=Path('/workspace/scratch/timer-ambient-base');e=Path(__file__).parent
sys=Path('src/compiler/sys.ts')
raw=subprocess.check_output(['git','-C',str(b),'show','refs/stage3/99-wasm-crypto-import-after:'+str(sys)])
(b/sys).write_bytes(raw)
files=sorted(p.relative_to(a) for p in (a/'src').rglob('*') if p.is_file());changes=[]
for f in files:
 if (a/f).read_bytes()!=(b/f).read_bytes():changes.append(str(f))
 # The exploratory source gate used universal-newline text I/O on sys.ts.
 if f==sys:assert (a/f).read_text()==(g/f).read_text(),str(f)
 else:assert (a/f).read_bytes()==(g/f).read_bytes(),str(f)
assert changes==['src/compiler/sys.ts'],changes
before=(b/sys).read_bytes();after=(a/sys).read_bytes();expected=before
for line in [b'declare function setTimeout(handler: (label?: unknown) => void, timeout: number, label?: unknown): unknown;',b'declare function clearTimeout(handle: unknown): void;']:
 assert expected.count(line)==1;expected=expected.replace(line+b'\r\n',b'')
old=b'clearTimeout: (timeoutId: unknown): void => clearTimeout(timeoutId)';new=b'clearTimeout: (timeoutId: Parameters<typeof clearTimeout>[0]): void => clearTimeout(timeoutId)';assert expected.count(old)==2;expected=expected.replace(old,new);assert expected==after
result={'base':subprocess.check_output(['git','rev-parse','HEAD^'],text=True).strip(),'head':subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip(),'source_files':len(files),'changed_files':changes,'actual_apply_equals_gate_counterfactual_modulo_sys_line_endings':True,'final_sys_preserves_CRLF':True,'sys_before_sha256':hashlib.sha256(before).hexdigest(),'sys_after_sha256':hashlib.sha256(after).hexdigest()}
e.joinpath('source-scope.json').write_text(json.dumps(result,indent=2)+'\n');e.joinpath('sys.before.ts.txt').write_bytes(before);e.joinpath('sys.after.ts.txt').write_bytes(after);print(json.dumps(result))
