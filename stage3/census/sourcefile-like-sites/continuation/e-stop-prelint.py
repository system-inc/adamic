import pathlib,time,os,signal,json,sys
p=pathlib.Path(__file__).resolve().parent;seen=set();start=time.monotonic()
while True:
 for proc in pathlib.Path('/proc').iterdir():
  if not proc.name.isdigit():continue
  try:args=(proc/'cmdline').read_bytes().split(b'\0')
  except OSError:continue
  if args and args[0].endswith(b'baobab-latent') and any(b'/continuation/e/raw/' in a for a in args):
   try:os.killpg(int(proc.name),signal.SIGTERM);seen.add(int(proc.name))
   except ProcessLookupError:pass
 if 'e census workflow child exit=0' in pathlib.Path(sys.argv[1] if len(sys.argv)>1 else '/tmp/baobab-e-controller.log').read_text():break
 assert time.monotonic()-start<180,'prelint controller did not finish'
 time.sleep(.05)
(p/'e/ABANDONED.json').write_text(json.dumps({'reason':(sys.argv[2] if len(sys.argv)>2 else 'Full pinned ESLint config requires the authorized unified-signatures suppression.')+' Census children were terminated and reaped by their original bounded wrappers before source changes. These are not final variant (e) results.','terminated_groups':sorted(seen),'child_status_files_preserve_own_exits':True},indent=2)+'\n')
print(json.dumps({'prelint_attempt_stopped':True,'terminated_groups':sorted(seen)}))
