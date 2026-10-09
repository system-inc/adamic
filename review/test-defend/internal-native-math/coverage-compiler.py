#!/usr/bin/env python3
import subprocess,sys,os,pathlib,hashlib,shutil
args=sys.argv[1:]
real='/workspace/adamic-tools/llvm/bin/clang'
if '--version' in args:sys.exit(subprocess.call([real]+args))
result=subprocess.run([real,'-fprofile-instr-generate','-fcoverage-mapping']+args)
if result.returncode==0 and '-c' not in args and '-o' in args:
 output=pathlib.Path(args[args.index('-o')+1]);directory=pathlib.Path(os.environ['DEFENSE_COVERAGE_DIR']);directory.mkdir(exist_ok=True)
 if output.is_file():shutil.copy2(output,directory/('binary-'+hashlib.sha256(output.read_bytes()).hexdigest()[:12]))
sys.exit(result.returncode)
