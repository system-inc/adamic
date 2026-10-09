from pathlib import Path
import subprocess,time,json
out=Path('review/test-audit/internal-lower-enums'); rows=out.joinpath('rows.txt').read_text().splitlines(); env=dict(__import__('os').environ);env['ADAMIC_CYCLE_LEDGER_ROOT']='/tmp/u032/typescript'
# Only actual Go panic frames trigger isolated matrix recovery.
panics=[]
for p in sorted(out.glob('M[0-9][0-9].log'))+sorted(out.glob('P*.log')):
 if '-Test' in p.stem or p.stem.endswith('-vet'):continue
 events=[]
 for line in p.read_text().splitlines():
  try:events.append(json.loads(line))
  except ValueError:pass
 if any(e.get('Output','').startswith(('panic: ','fatal error: ')) for e in events):panics.append(p.stem)
for mid in panics:
 for row in rows:
  env.update(ADAMIC_MUTANT=mid,ADAMIC_BUILD_CACHE_DIR='/tmp/u032/cache/'+mid,ADAMIC_CYCLE_LEDGER_OUTPUT='/tmp/u032/ledger-'+mid+'-'+row+'.json')
  start=time.time()
  with out.joinpath(mid+'-'+row+'.log').open('w') as log:
   result=subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','^'+row+'$'],env=env,stdout=log,stderr=subprocess.STDOUT)
  with out.joinpath('recovery-times.txt').open('a') as log:log.write(f'{mid} {row} {result.returncode} {time.time()-start:.9f}\n')
files=json.loads(out.joinpath('base-files.json').read_text())
def restore():
 for f in files:Path(f).write_bytes(subprocess.check_output(['git','show','origin/main:'+f]))
restore();Path('internal/lower/audit_switch.go').unlink()
for diff in sorted(out.joinpath('diffs').glob('*.diff')):
 subprocess.run(['git','apply','--check',str(diff)],check=True);subprocess.run(['git','apply',str(diff)],check=True)
 start=time.time()
 with out.joinpath(diff.stem+'-vet.log').open('w') as log:r=subprocess.run(['timeout','90','go','vet','./internal/lower/'],stdout=log,stderr=subprocess.STDOUT)
 with out.joinpath('vet-times.txt').open('a') as log:log.write(f'{diff.stem} {r.returncode} {time.time()-start:.9f}\n')
 restore()
 if r.returncode:print('VET FAILURE',diff.stem);raise SystemExit(1)
print('isolated recovery:',panics,'all standalone diffs applied and compiled')
