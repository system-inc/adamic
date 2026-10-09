#!/usr/bin/python3
import subprocess,sys,time,os,json
start=time.monotonic();rc=subprocess.call(['/workspace/adamic-tools/bin/clang',*sys.argv[1:]])
if os.environ.get('AUDIT_CLANG_LOG'):
 with open(os.environ['AUDIT_CLANG_LOG'],'a') as f:f.write(json.dumps(dict(arguments=sys.argv[1:],seconds=time.monotonic()-start,returncode=rc))+'\n')
sys.exit(rc)
