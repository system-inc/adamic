import pathlib,subprocess,sys,time,json,os
out=pathlib.Path('/workspace/scratch/overload-transformers/stage3/census/overload-transformers/landing-932a6344')
name=sys.argv[1];cmd=sys.argv[2:];start=time.time()
with (out/(name+'.stdout')).open('w') as stdout,(out/(name+'.stderr')).open('w') as stderr:
 p=subprocess.run(cmd,stdout=stdout,stderr=stderr)
r={'command':cmd,'cwd':os.getcwd(),'exit':p.returncode,'seconds':time.time()-start}
(out/(name+'.json')).write_text(json.dumps(r,indent=2)+'\n')
print(('PASS ' if p.returncode==0 else 'FAIL ')+name+' '+json.dumps(r),flush=True)
sys.exit(p.returncode)
