import json,pathlib,subprocess,sys,time,os
out=pathlib.Path(__file__).resolve().parent
root=out.parents[2]
name,*args=sys.argv[1:]
started=time.monotonic()
with (out/(name+'.log')).open('w') as log:
    try: code=subprocess.run(args,cwd=root,stdout=log,stderr=subprocess.STDOUT,timeout=600).returncode
    except subprocess.TimeoutExpired: code=124
row={'name':name,'command':args,'exit':code,'seconds':round(time.monotonic()-started,3)}
with (out/'commands.jsonl').open('a') as stream:stream.write(json.dumps(row)+'\n')
print(json.dumps(row),flush=True)
sys.exit(code)
