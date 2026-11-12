"""Check shard composition against an unsharded first-error fixture on each side."""
import os,sys,pathlib,subprocess,json
from merge import merge,load,key
out=pathlib.Path(__file__).resolve().parent
label=sys.argv[1];variant='-scan' if len(sys.argv)>2 and sys.argv[2]=='scan' else '';build=pathlib.Path('/workspace/oph-build-'+label);runs=[]
for tag,shards,shard in [('whole',1,0),('s0',2,0),('s1',2,1)]:
 env={k:v for k,v in os.environ.items() if not k.startswith('LATENT_') and not k.startswith('CENSUS_SHARD')}
 env.update(CENSUS_SKIP_SCAN='1' if variant and tag=='s1' else '0',LATENT_FULL='0',LATENT_ASSERT_NO_OUTPUT='1',CENSUS_SHARDS=str(shards),CENSUS_SHARD=str(shard))
 prefix=build/('fixture'+variant+'-'+tag)
 cmd=[sys.executable,str(out/'run.py'),str(prefix)+'.log',str(prefix)+'.time.json',str(build/('entry-census'+variant)),'/workspace/oph-fixture/main.ts',str(prefix)+'.jsonl']
 p=subprocess.run(cmd,env=env);print('fixture '+tag+' child_exit='+str(p.returncode),flush=True)
 if p.returncode:sys.exit(p.returncode)
 runs.append(load(str(prefix)+'.jsonl'))
whole=runs[0];merged=merge(runs[1:]);
if variant: assert whole[:-1]==load(str(build/'fixture-whole.jsonl'))[:-1]
assert whole[0]==merged[0] and not whole[0]['checker_rejected']
def canonical(rows):
 return [{**r,'findings':sorted(r['findings'],key=key)} for r in rows[1:-1]]
assert canonical(whole)==canonical(merged)
units=[u for r in merged for u in r.get('units',[])]
witnesses=[u for u in units if u['status']=='deferred_generic_instantiation' and u['instantiated']]
assert any(len(u.get('instantiated_by',[]))>=2 for u in witnesses)
findings=[f for r in whole for f in r.get('findings',[])];assert findings
result=dict(side=label,headers_equal=True,checker_clean=True,file_unit_finding_ledgers_equal=True,cross_shard_generic_witnesses_merged=True,units=len(units),findings=len(findings),relation_report='Retained separately for each shard; cache counters are process-scoped and not additive.')
(out/(label+variant+'-fixture-validation.json')).write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result))
