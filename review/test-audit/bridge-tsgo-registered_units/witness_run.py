from audit_run import *
import subprocess
w=json.load(open(P/'witness-plan.json'))
for item in w:
 id=item['id'];cache=str(S/'cache'/id);overlay=item.get('overlay');physical=id in ['W_STALE','W_LINK']
 if physical:
  # These harness sources are passed to a separate compiler, outside the outer Go overlay.
  file=Path(item['file']);old=file.read_text();text=Path('/tmp/u004/witness')/id/file.name;file.write_text(text.read_text());overlay=None
 if id=='W_LSAN':os.environ['ASAN_OPTIONS']='detect_leaks=0';os.environ['LSAN_OPTIONS']='detect_leaks=0';cache=None
 try:
  for row in item['rows']:
   corpus=any(row.endswith(x) for x in ['Checker','Parser','Types','Utilities']);d=run(id+'-'+row,'^'+row+'$',corpus,overlay=overlay,cache=cache if id!='W_EQUAL' else None)
 finally:
  if physical:file.write_text(old)
  if id=='W_LSAN':os.environ.pop('ASAN_OPTIONS');os.environ.pop('LSAN_OPTIONS')
