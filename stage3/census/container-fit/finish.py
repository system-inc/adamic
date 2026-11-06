"""Check no census child remains, render results and validate the saved evidence."""
import datetime,hashlib,json,pathlib,subprocess,sys
O=pathlib.Path(__file__).resolve().parent
active=[]
for line in subprocess.check_output(['ps','-eo','pid=,args='],text=True).splitlines():
 fields=line.strip().split(None,1)
 if len(fields)<2:continue
 command=fields[1].split()[0]
 if command.startswith('/workspace/container-fit-build/') and command.rsplit('/',1)[-1] in ['latent','census']:
  active.append(dict(pid=int(fields[0]),command=command))
(O/'evidence/process-cleanup.json').write_text(json.dumps(dict(utc=datetime.datetime.now(datetime.timezone.utc).isoformat(),active_census_processes=active),indent=2)+'\n')
assert not active,active
subprocess.run([sys.executable,str(O/'summarize.py')],check=True)
with (O/'REPORT.md').open('a') as f:f.write('\n## Validation and cleanup\n\nSource immutability, raw archive hashes, root coverage, process limits, report-only branch scope and explicit unknown latent outcomes are checked by validate.py; see validation.txt. All launched census children exited and were reaped. The final process check found no active census process.\n')
with (O/'validation.txt').open('w') as f:subprocess.run([sys.executable,str(O/'validate.py')],stdout=f,stderr=subprocess.STDOUT,check=True)
manifest={str(p.relative_to(O)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(O.rglob('*')) if p.is_file() and p.name!='artifacts.sha256.json' and '__pycache__' not in str(p)}
(O/'artifacts.sha256.json').write_text(json.dumps(manifest,indent=2)+'\n')
print((O/'validation.txt').read_text())
