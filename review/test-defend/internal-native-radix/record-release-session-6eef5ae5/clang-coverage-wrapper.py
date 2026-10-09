#!/usr/bin/env python3
import os,sys,subprocess,pathlib,hashlib
args=sys.argv[1:];cmd=['/workspace/adamic-tools/bin/clang']+args
if '--version' not in args:cmd+=['-fprofile-instr-generate','-fcoverage-mapping']
r=subprocess.run(cmd)
if r.returncode==0 and '-o' in args and '-c' not in args:
 out=pathlib.Path(args[args.index('-o')+1]);dest=pathlib.Path(os.environ['DEFENSE_C_BIN']);dest.mkdir(parents=True,exist_ok=True)
 if out.exists():(dest/(out.name+'-'+hashlib.sha256(out.read_bytes()).hexdigest()[:12])).write_bytes(out.read_bytes())
sys.exit(r.returncode)
