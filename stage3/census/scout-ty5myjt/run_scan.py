"""Run the existing scan-only wrapper; no body lowering or usable output."""
import pathlib, subprocess, os, sys, json, time, signal, hashlib
root=pathlib.Path('/workspace/scratch/scout-ty5myjt')
filename=os.environ.get('SCAN_FILE','program.ts')
for name in sys.argv[1:]:
 tree=root/('adapted' if name=='baseline' else name)
 out=root/('scan-results' if filename=='program.ts' else 'scan-utilities-results')/name;out.mkdir(parents=True,exist_ok=False)
 env=os.environ|{'LATENT_FULL':'1','LATENT_ASSERT_NO_OUTPUT':'1','LATENT_ONLY_FILE':str(tree/'src/compiler'/filename),'CENSUS_MODULE_FORMAT':'CommonJSBundle'}
 command=[str(root/'scan-only'),str(tree/'src/compiler'),str(out/(filename+'.jsonl'))]
 start=time.monotonic();timeout=False
 with (out/'run.log.txt').open('w') as log:
  child=subprocess.Popen(command,env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
  try: status=child.wait(timeout=1500)
  except subprocess.TimeoutExpired:
   timeout=True;os.killpg(child.pid,signal.SIGKILL);status=child.wait()
 result={'candidate':name,'file':filename,'mode':'scan only','command':command,'seconds':time.monotonic()-start,'exit':status,'timeout':timeout,'process_group_bound_seconds':1500,'binary_sha256':hashlib.sha256((root/'scan-only').read_bytes()).hexdigest()}
 (out/'run.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result),flush=True)
 if status:sys.exit(1)
