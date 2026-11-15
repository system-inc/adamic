import json,os,pathlib,resource,subprocess,sys,time
root=pathlib.Path('/workspace/scratch/record-cycle');name=sys.argv[1];args=sys.argv[2:];started=time.monotonic()
with (root/(name+'.log')).open('wb') as log:
 child=subprocess.Popen(args,stdout=log,stderr=log,start_new_session=True)
 try: rc=child.wait(timeout=3500)
 except subprocess.TimeoutExpired:
  import signal
  os.killpg(child.pid,signal.SIGTERM)
  try: rc=child.wait(timeout=10)
  except subprocess.TimeoutExpired: os.killpg(child.pid,signal.SIGKILL);rc=child.wait()
receipt=dict(command=args,exit=rc,seconds=time.monotonic()-started,peak_rss_kib=resource.getrusage(resource.RUSAGE_CHILDREN).ru_maxrss)
(root/(name+'.receipt.json')).write_text(json.dumps(receipt,indent=2)+'\n');print(json.dumps(receipt));sys.exit(rc)
