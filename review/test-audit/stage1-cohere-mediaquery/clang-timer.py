#!/usr/bin/python3
import os,sys,subprocess,time,json,pathlib
args=['/workspace/adamic-tools/bin/clang']+sys.argv[1:]
started=time.monotonic();r=subprocess.run(args);elapsed=time.monotonic()-started
p=pathlib.Path('/workspace/adamic/review/test-audit/stage1-cohere-mediaquery/clang-times');p.mkdir(exist_ok=True)
label=os.environ.get('ADAMIC_AUDIT_ID','unknown')
(p/(label+'-'+str(os.getpid())+'.json')).write_text(json.dumps(dict(id=label,seconds=elapsed,exit=r.returncode,command=args))+'\n')
sys.exit(r.returncode)
