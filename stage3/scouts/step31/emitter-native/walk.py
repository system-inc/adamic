#!/usr/bin/env python3
"""Walk actual first diagnostics in a disposable clone, never an oracle tree."""
import argparse,json,os,re,shutil,subprocess,time
from pathlib import Path
parser=argparse.ArgumentParser()
parser.add_argument('compiler');parser.add_argument('candidate',type=Path);parser.add_argument('slice',type=Path);parser.add_argument('entry',type=Path);parser.add_argument('output',type=Path)
args=parser.parse_args();args.output.mkdir()
shutil.copytree(args.slice,args.output/'slice')
entry=args.output/'replay.a';entry.write_text(args.entry.read_text().replace(str(args.slice),str(args.output/'slice')))
root=Path(__file__).resolve().parent
rows=[]
for order in range(1,16):
    prefix=args.output/f'stop-{order:02}'
    started=time.monotonic()
    for split in ['0','1']:
        mode=args.output/f'stop-{order:02}-split-{split}'
        with mode.with_suffix('.stdout').open('wb') as out,mode.with_suffix('.stderr').open('wb') as err:
            result=subprocess.run([args.compiler,'build',str(entry),'-o',str(args.output/f'emitter-{split}')],cwd=args.candidate,env=dict(os.environ,ADAMIC_NATIVE_SPLIT=split),stdout=out,stderr=err)
        mode.with_suffix('.exit').write_text(str(result.returncode)+'\n')
    assert (args.output/f'stop-{order:02}-split-0.stderr').read_bytes()==(args.output/f'stop-{order:02}-split-1.stderr').read_bytes(), 'split diagnostics differ'
    shutil.copyfile(args.output/f'stop-{order:02}-split-0.stderr',prefix.with_suffix('.stderr'))
    diagnostic=prefix.with_suffix('.stderr').read_text()
    match=re.search(r'^(?:adamic: )?(.+?):(\d+):(\d+): (.+)$',diagnostic,re.M)
    if not match:
        raise RuntimeError(f'unlocated stop {order}, exit {result.returncode}: {diagnostic[:1000]}')
    file,line,column,message=match.groups()
    row=dict(order=order,file=file,line=int(line),column=int(column),message=message,exit=result.returncode,seconds=time.monotonic()-started)
    rows.append(row)
    (args.output/'stops.json').write_text(json.dumps(rows,indent=2)+'\n')
    print(json.dumps(row),flush=True)
    if result.returncode==0 or order==15:break
    stub=subprocess.run(['node',str(root/'stub.cjs'),file,line,column],text=True,capture_output=True)
    (args.output/f'stub-{order:02}.stdout').write_text(stub.stdout)
    (args.output/f'stub-{order:02}.stderr').write_text(stub.stderr)
    if stub.returncode:
        print('Stopped honestly: no valid discovery edit. '+stub.stderr,flush=True);break
