import pathlib,os,time,shutil,json
root=pathlib.Path('/workspace/deletion-replay/tmp')
while not pathlib.Path('/workspace/adamic/review/test-defend/deletion-set/internal-oracle/replay-status.json').exists():
 active=''
 for p in pathlib.Path('/proc').glob('[0-9]*'):
  try:active+=(p/'cmdline').read_bytes().decode(errors='replace')+str((p/'exe').resolve())+'\n'
  except OSError:pass
 for p in root.glob('go-build*'):
  try:
   if time.time()-p.stat().st_mtime>120 and str(p) not in active:
    shutil.rmtree(p);print(json.dumps({'deleted_inactive_replay_build':str(p),'time':time.time()}),flush=True)
  except OSError:pass
 time.sleep(30)
