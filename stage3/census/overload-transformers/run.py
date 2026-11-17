import subprocess,sys,pathlib,time,json,os
name=sys.argv[1]; cmd=sys.argv[2:]
e=pathlib.Path('/workspace/scratch/overload-transformers/stage3/census/overload-transformers/logs'); e.mkdir(parents=True,exist_ok=True)
start=time.time()
with (e/(name+'.stdout')).open('w') as out,(e/(name+'.stderr')).open('w') as err:
 p=subprocess.run(cmd,stdout=out,stderr=err)
r={'command':cmd,'cwd':os.getcwd(),'exit':p.returncode,'seconds':time.time()-start}
(e/(name+'.json')).write_text(json.dumps(r,indent=2)+'\n')
print(json.dumps(r),flush=True)
sys.exit(p.returncode)
