import pathlib,sys,fcntl,subprocess
p=pathlib.Path(__file__).resolve().parents[1]
with open('/tmp/baobab-checkpoint.lock','a') as lock:
 fcntl.flock(lock,fcntl.LOCK_EX)
 child=subprocess.run(['python3',str(p/'checkpoint.py'),sys.argv[1]])
 print('checkpoint child exit='+str(child.returncode),flush=True)
 raise SystemExit(child.returncode)
