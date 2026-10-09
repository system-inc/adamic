import pathlib,json,re,statistics,difflib,shutil,subprocess,hashlib
r=pathlib.Path('/workspace/adamic/review/test-audit/stage3-fixtures-fixtures');scope=json.loads((r/'scope.json').read_text());times=json.loads((r/'seconds.json').read_text());members=scope['family_members'];plan=json.loads((r/'plan.json').read_text());runs=json.loads((r/'matrix-runs.json').read_text());setup=json.loads((r/'setup-runs.json').read_text())
def events(mid):
 result=[]
 for line in (r/'logs'/f'{mid}.log').read_text().splitlines():
  try:result.append(json.loads(line))
  except:pass
 return result
matrix={};probes={};fails={}
for mid in ['M1','M2','M3','P1','P2']:
 es=events(mid);top={e['Test']:e['Action'] for e in es if e.get('Test') and '/' not in e['Test'] and e.get('Action') in ('pass','fail','skip')};matrix[mid]=top;fails[mid]=[n for n,v in top.items() if v=='fail']
(r/'matrix.json').write_text(json.dumps(matrix,indent=2));rows=[]
configs=[('TestFixtures family','family',['external-run','self'],'Fresh Node stdout/stderr/exit versus native, recorded Node bytes, and self-written stage0 diagnostic/admission snapshots','M3',['M1','M2','M3'],['P1','P2']),('TestPrepareFixtureOracleHook','prepare','self','Build recipe and content-addressed hook integrity; returned path is ignored','Kprepare',[],['Pprepare']),('TestFixturePaths','paths','self','Hand-written portable path acceptance table','Kpaths',[],['Ppaths']),('TestFixtureShardManifest','manifest','self','Fixture identity uniqueness, path validity and nonempty enumeration; shard assignments only logged','Kmanifest',[],['Pmanifest'])]
for name,key,kind,oracle,proof,ks,ps in configs:
 ms=members if key=='family' else [name];es=events(proof);err=next(e['Output'].strip() for e in es if e.get('OutputType')=='error' and e.get('Test','').split('/')[0] in ms);command=' '.join(next(x['command'] for x in (runs if proof.startswith('M') else setup) if x['id']==proof));pk=[];probeactions={}
 for pid in ps:
  actions=[e.get('Action') for e in events(pid) if e.get('Test') in ms and e.get('Action') in ('pass','fail')];probeactions[pid]=actions
  if 'fail' in actions:pk.append(pid)
 vacuous=all('pass' in a and 'fail' not in a for a in probeactions.values());row=dict(test=name,package='stage3/fixtures',file='stage3/fixtures/fixtures_test.go',seconds=times[key]['median'],oracle=oracle,oracle_kind=kind,kills=ks,unique_kills=ks,last_proven_fail=proof+': '+err,verdict='sacred' if key=='family' else 'setup-check',subsumed_by=[],mutants_in_matrix=3 if key=='family' else 0,probe_kills=pk,subsumer_seconds=None,vacuous=vacuous,bounded=False,matrix_rows=['TestFixtures family','TestPrepareFixtureOracleHook','TestFixturePaths','TestFixtureShardManifest','TestTransformedNodeRunnerGuard','TestTransformedNodeRunnerGuardHook','TestFixtureDirectoriesHaveTopLevelTests'],evidence=command+'; '+err,members=ms,seconds_samples=times[key]['samples'])
 if key=='family':
  # Positive stage0 observations whose own Lower empty probe still permits Compiles. Exclude source-Node checks and checker failures before entry.
  compiling=set()
  for status in (r.parents[2]/'stage3/fixtures').glob('*/status.json'):
   for entry in json.loads(status.read_text()):
    if entry['stage0']['outcome']=='Compiles':compiling.add(status.parent.name+'/'+entry['file'])
  row['vacuous_subcases']=[e['Test'] for e in events('P1') if e.get('Action')=='pass' and e.get('Test','').endswith('/stage0') and '/'.join(e['Test'].split('/')[1:-1]) in compiling]
  row['probe_note']='P1 positive stage0 checks may accept nil IR; native subprocess failure still kills their containing family. P2 is independently empty native.C.'
 if key=='paths':row['vacuous_subcases']=[e['Test'] for e in events('Ppaths') if e.get('Action')=='pass' and '/' in e.get('Test','')]
 rows.append(row)
(r/'rows.json').write_text(json.dumps(rows,indent=2));print([(x['test'],x['verdict'],x['vacuous']) for x in rows]);print('positive admission subcases',len(rows[0].get('vacuous_subcases',[])))
# Verify against the unchanged starting source tree.
checks=[]
for f in r.glob('*.diff'):
 if f.name=='switch.diff':continue
 z=subprocess.run(['git','apply','--check',str(f)],cwd=r.parents[2],capture_output=True,text=True);checks.append(dict(diff=f.name,rc=z.returncode,output=z.stdout+z.stderr))
(r/'apply-checks.json').write_text(json.dumps(checks,indent=2));print('apply failures',[x for x in checks if x['rc']])
