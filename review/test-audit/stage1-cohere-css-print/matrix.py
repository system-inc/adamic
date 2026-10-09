import os,pathlib,json,subprocess,time
out=pathlib.Path('review/test-audit/stage1-cohere-css-print')
base=os.environ.copy(); base['ADAMIC_BUILD_CACHE_DIR']='/tmp/u079/cache/switched'; base['ADAMIC_CSS_PROFILE_DIR']='/tmp/u079/switched-profile'; base['ADAMIC_CSS_PROFILE_SNAPSHOTS']='/tmp/u079/switched-profile'; base['ADAMIC_BUILD_LOG']='/tmp/u079/builds.log'
records=json.loads((out/'matrix-runs.json').read_text()) if (out/'matrix-runs.json').exists() else []
def run(ident,label,pattern,env=None):
 e=base.copy(); e.update(env or {}); e['ADAMIC_MUTANT']=ident; pathlib.Path('/tmp/u079/selector').write_text(ident)
 p=out/f'{ident or "clean"}-{label}.log'; start=time.time()
 with p.open('w') as f: code=subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/css/','-run',pattern],stdout=f,stderr=subprocess.STDOUT,env=e).returncode
 events=[]
 for l in p.read_text().splitlines():
  try: events.append(json.loads(l))
  except: pass
 events=[e for e in events if e.get('Package')=='github.com/system-inc/adamic/stage1/cohere/css']
 record=dict(id=ident,label=label,command='timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/css/ -run '+pattern,pattern=pattern,wall=time.time()-start,code=code,log=str(p),outcomes={e['Test']:e['Action'] for e in events if e.get('Test') and e.get('Action') in ['pass','fail','skip']},ran=[e['Test'] for e in events if e.get('Test') and e['Action']=='run'],panic=any(e.get('Output','').startswith('panic:') for e in events),over_budget='test timed out after' in p.read_text() or code==124)
 records.append(record); (out/'matrix-runs.json').write_text(json.dumps(records,indent=2)); return record
fast='^Test(ClosedPrinterRegexGap|CSSPrinterBoundaryProofs|OptionalBooleanPrinterMatchesGo|SharedSliceAppendAgreesWithNode)$'
callers='^TestCSSPrinterAgreesWithGo_0(0[0-9]|1[0-5])$'
# One switched source builds once per native product. Cache reuse is content keyed;
# each fresh executable reads the selector file, and runtime probes use ADAMIC_MUTANT.
# The broad caller family cooked at 90 seconds. The resumed matrix is bounded
# to the six assigned semantic rows; omitted callers are unknown.
if '--resume' not in __import__('sys').argv:
 r=run('','profile-build','^TestCSSProfileArtifacts$')
 if r['code']!=0: raise SystemExit('switched native profile build failed')
 r=run('','fast-baseline',fast)
 if r['code']!=0: raise SystemExit('switch altered clean fast baseline')
 r=run('','caller-baseline',callers)
 if r['code']!=0: raise SystemExit('clean caller family red or over budget; narrow it before mutation')
else:
 r=run('','snapshot-baseline','^TestCSSProfileSnapshotsAgree$')
 if r['code']!=0: raise SystemExit('switch altered snapshot baseline')
# Timed independently because this opt-in row occupies most of the binary budget.
for ident in ['M1','M2','M3','M4']:
 run(ident,'fast',fast)
 run(ident,'throughput','^TestCSSPrinterThroughput$')
 run(ident,'snapshots','^TestCSSProfileSnapshotsAgree$')
for ident,pattern in [('P1','^Test(CSSPrinterBoundaryProofs|OptionalBooleanPrinterMatchesGo)$'),('P2','^TestClosedPrinterRegexGap$'),('P3','^TestSharedSliceAppendAgreesWithNode$'),('P4','^TestClosedPrinterRegexGap$')]:
 run(ident,'probe',pattern)
 if ident=='P1':
  run(ident,'throughput-probe','^TestCSSPrinterThroughput$'); run(ident,'snapshot-probe','^TestCSSProfileSnapshotsAgree$')
for ident,pattern in [('W1','^TestCSSPrinterShardingCatchesDisagreement$'),('S4','^TestCSSPrinterShardingCatchesDisagreement$'),('S1','^TestProduct_CSSPrinter(SanitizedAndLowered|SemicolonMutant|IndentMutant|WidthMutant)$'),('S2','^TestProduct_CSSPrinter(ParserOracle|Oracle)$'),('S3','^TestCSSProfileArtifacts$')]:
 extra={'ADAMIC_BUILD_CACHE_DIR':'/tmp/u079/cache/'+ident,'ADAMIC_CSS_PROFILE_DIR':'/tmp/u079/'+ident+'-profile'}
 run(ident,'construction',pattern,extra)
pathlib.Path('/tmp/u079/selector').write_text('')
