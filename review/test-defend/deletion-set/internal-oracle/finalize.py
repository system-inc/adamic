import datetime,gzip,json,pathlib,re,shutil,subprocess
P=pathlib.Path(__file__).resolve().parent
m=json.loads((P/'matrix.json').read_text());assert len(m)==77
old=json.loads((P/'witness-only-attempts.json').read_text());r=next(x for x in m if x['replay_index']==55)
# Preserve the initially rejected sole-witness attempt as well as its qualified reruns.
if not any(x['log']=='logs/055-D5-r0.log' for x in r['runs']):
 r['runs']=old[0]['runs']+r['runs'];r['witness_failures']=sorted(set(r['witness_failures'])|{'TestInterfaceCastScalarTags'});r['wall_seconds']=round(sum(x['wall_seconds'] for x in r['runs']),3)
(P/'matrix.json').write_text(json.dumps(m,indent=2)+'\n')
subprocess.run(['python3',str(P/'make-report.py')],check=True)
report=json.loads((P/'report.json').read_text())
report['notes'].update({'family_skip_correction':'Supplied prefixes did not match five actual family members; expanded-skipped.json records the exact 22 skipped top-level tests.', 'retention':'The two bounds-diagnostic guards jointly catch both lost diffs. At least one must remain; listing both keep rows does not claim individual uniqueness.', 'invalid_ref':'test-defend/internal-oracle-private_generic_mutant/row.cover is a file-shaped non-existent ref; parent branch fetched.', 'interruption':'Environment interruption and a premature resume required matrix recovery. Every recovered catch was verified against retained JSON logs; overwritten incomplete runs were repeated.', 'cleanup_rejection':'Automatic approval review rejected further deletion from /home/agent/.cache/go-build as outside authorized /tmp cleanup scope. That deletion was not performed.'})
start=datetime.datetime.fromisoformat('2026-10-09T18:03:20.747038+00:00');now=datetime.datetime.now(datetime.timezone.utc)
for row in report['keep']:
 if row['test'] in ('TestCheckedViewV2ArrayArmBoundary','TestInterfaceCastImportedConstruction'):
  row['because']='D5 (diffs/b15-055-D5.diff) has no clean ordinary catcher after 25 panic exclusions; retain conservatively, not a clean last-catcher proof'
for mutant in report['mutants']:
 if mutant['diff']=='diffs/b15-055-D5.diff':mutant['qualification']='Completed with only a witness failure after 25 panic exclusions; ordinary panics are listed separately'
report['notes']['timing']={'baseline_wall_seconds':277.272,'replay_elapsed_seconds_including_interruption':round((now-start).total_seconds(),3),'replay_sum_wall_seconds':round(sum(r.get('wall_seconds',0) for r in m),3),'replay_concurrency':2,'ended_utc':now.isoformat()}
(P/'report.json').write_text(json.dumps(report,indent=2)+'\n')
refs={x['log'] for r in m for x in r['runs']}
# Explicitly label abandoned/unreferenced replay logs so none is mistaken for accepted evidence.
orphans=[]
for f in (P/'logs').glob('*.log'):
 if re.fullmatch(r'\d{3}-.*-r\d+\.log',f.name) and str(f.relative_to(P)) not in refs:
  orphan=str(f.relative_to(P));orphans.append(orphan);f.rename(f.with_name(f.name+'.unreferenced'))
(P/'unreferenced-logs.json').write_text(json.dumps(orphans,indent=2)+'\n')
compressed={}
for f in P.rglob('*'):
 if f.is_file() and ('.log' in f.name) and not f.name.endswith('.gz'):
  target=f.with_name(f.name+'.gz')
  with f.open('rb') as source,gzip.open(target,'wb',compresslevel=9) as out:shutil.copyfileobj(source,out)
  compressed[str(f.relative_to(P))]=str(target.relative_to(P));f.unlink()
def replace(value):
 if isinstance(value,dict):return {k:replace(v) for k,v in value.items()}
 if isinstance(value,list):return [replace(v) for v in value]
 if isinstance(value,str):return compressed.get(value,value)
 return value
for f in P.glob('*.json'):
 try:value=json.loads(f.read_text())
 except ValueError:continue
 f.write_text(json.dumps(replace(value),indent=2)+'\n')
(P/'compressed-logs.json').write_text(json.dumps(compressed,indent=2)+'\n')
# Verify accepted evidence survived compression, including every candidate skip and ordinary catch.
w=set(json.loads((P/'witnesses.json').read_text()));skip=set(json.loads((P/'expanded-skipped.json').read_text()))
for r in json.loads((P/'matrix.json').read_text()):
 assert not set(r.get('still_caught_by',[]))&w
 assert not set(r.get('still_caught_by',[]))&skip
 for run in r['runs']:
  with gzip.open(P/run['log'],'rt') as f:
   for line in f:
    try:e=json.loads(line)
    except ValueError:continue
    if e.get('Action')=='run':assert e.get('Test','').split('/')[0] not in skip
print(json.dumps({'keep':report['keep'],'deletable':report['deletable'],'timing':report['notes']['timing']},indent=2))
