#!/usr/bin/env python3
import subprocess,sys,os,pathlib,shutil,time
args=sys.argv[1:];real='/workspace/adamic-tools/bin/clang'
if args==['--version']:
 r=subprocess.run([real,*args]);print('CSS defense C coverage instrumentation');sys.exit(r.returncode)
compile='-c' in args
extra=[]
if not compile or any(pathlib.Path(a).name in ['string_share.c','string_append.c'] for a in args):extra=['-fprofile-instr-generate','-fcoverage-mapping']
r=subprocess.run([real,*args,*extra])
if r.returncode==0 and not compile and '-o' in args:
 out=pathlib.Path(args[args.index('-o')+1]);dest=pathlib.Path('/workspace/defend-css-tmp/cov-bin-'+os.environ['CSS_COVERAGE_ROW']);dest.mkdir(exist_ok=True);shutil.copy2(out,dest/str(time.time_ns()))
sys.exit(r.returncode)
