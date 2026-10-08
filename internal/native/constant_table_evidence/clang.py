#!/usr/bin/env python3
"""Transparent timing wrapper for scratch measurements, never a compiler default."""
import json, os, pathlib, subprocess, sys, time
real = '/workspace/adamic-tools/llvm/bin/clang'
args = sys.argv[1:]
line_tables = pathlib.Path(sys.argv[0]).parent.name == 'line'
if line_tables:
    args = ['-gline-tables-only' if x == '-g' else x for x in args]
start = time.monotonic()
result = subprocess.run([real, *args])
seconds = time.monotonic() - start
trace = os.environ.get('OUTLINE_TRACE')
if trace and '-c' in args and not pathlib.Path(args[args.index('-c') + 1]).is_absolute():
    source = args[args.index('-c') + 1]
    target = pathlib.Path(trace + '-units')
    target.mkdir(exist_ok=True)
    (target / source).write_bytes(pathlib.Path(source).read_bytes())
if trace:
    row = {'seconds':seconds, 'args':args, 'cwd':os.getcwd(), 'exit':result.returncode}
    fd=os.open(trace+'.jsonl', os.O_CREAT|os.O_WRONLY|os.O_APPEND,0o644)
    os.write(fd,(json.dumps(row)+'\n').encode());os.close(fd)
if '--version' in args and line_tables:
    print('outline scratch line tables variant')
sys.exit(result.returncode)
