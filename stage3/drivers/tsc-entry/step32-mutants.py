#!/usr/bin/env python3
"""Mutation-test the step 32 evidence verifier on disposable copies."""
import gzip,json,shutil,subprocess,sys,tempfile
from pathlib import Path
here=Path(__file__).resolve().parent
original=here/'evidence/step32'
subprocess.run([sys.executable,str(here/'step32-verify.py'),str(original)],check=True)
cases={
 'invent-combined-mode': lambda d: d.update(combined_mode_measured=True),
 'invent-compiler-pin': lambda d: d.update(compiler_commit='0'*40),
 'invent-first-stop': lambda d: d.update(current_first_stop='invented'),
 'invent-split-success': lambda d: d.update(current_exits=[1,0]),
}
for name,change in cases.items():
    with tempfile.TemporaryDirectory(prefix='step32-mutant-') as tmp:
        out=Path(tmp)/'evidence';shutil.copytree(original,out)
        result=json.loads((out/'result.json').read_text());change(result)
        (out/'result.json').write_text(json.dumps(result))
        check=subprocess.run([sys.executable,str(here/'step32-verify.py'),str(out)],capture_output=True)
        (original/(name+'.stderr.gz')).write_bytes(gzip.compress(check.stderr,mtime=0))
        assert check.returncode!=0,name
        assert b'AssertionError' in check.stderr,name
        print(name+': KILLED by evidence assertion')
for name in ('drop-saved-stop','corrupt-node-stdout','invent-owner','admit-saved-stop'):
    with tempfile.TemporaryDirectory(prefix='step32-mutant-') as tmp:
        out=Path(tmp)/'evidence';shutil.copytree(original,out)
        file=out/('comparison.json' if name in ('drop-saved-stop','admit-saved-stop') else 'walk.json')
        rows=json.loads(file.read_text())
        if name=='drop-saved-stop': rows['stops'].pop()
        elif name=='admit-saved-stop': rows['stops'][0]['status']='disappear'
        elif name=='corrupt-node-stdout': rows[0]['node_stdout']='wrong\n'
        else: rows[0]['owner']='unowned'
        file.write_text(json.dumps(rows))
        check=subprocess.run([sys.executable,str(here/'step32-verify.py'),str(out)],capture_output=True)
        (original/(name+'.stderr.gz')).write_bytes(gzip.compress(check.stderr,mtime=0))
        assert check.returncode!=0,name
        assert b'AssertionError' in check.stderr,name
        print(name+': KILLED by evidence assertion')
for name,file in [('corrupt-split-stream','current-1.stderr.gz'),('dirty-binary','binary.txt.gz'),('corrupt-entry-node','node-entry.stdout.gz')]:
    with tempfile.TemporaryDirectory(prefix='step32-mutant-') as tmp:
        out=Path(tmp)/'evidence';shutil.copytree(original,out)
        text=gzip.decompress((out/file).read_bytes())
        text=text.replace(b'vcs.modified=false',b'vcs.modified=true') if name=='dirty-binary' else text+b'changed stream\n'
        (out/file).write_bytes(gzip.compress(text,mtime=0))
        check=subprocess.run([sys.executable,str(here/'step32-verify.py'),str(out)],capture_output=True)
        (original/(name+'.stderr.gz')).write_bytes(gzip.compress(check.stderr,mtime=0))
        assert check.returncode!=0,name
        assert b'AssertionError' in check.stderr,name
        print(name+': KILLED by evidence assertion')
print('11/11 evidence mutants killed')
