"""Production-profile guard evidence: a scheduled/constructed check is not emitted IR."""
import json,os,subprocess,tempfile
from pathlib import Path
import sys
binary=Path(sys.argv[1]).resolve()
with tempfile.TemporaryDirectory(prefix='latent-indexed-audit-') as tmp:
 root=Path(tmp);source=root/'source';source.mkdir()
 # .a is the authored fixture; .ts is a scratch checker-profile materialization.
 (source/'input.ts').write_bytes((Path(__file__).parent/'fixtures/guard-attribution.a').read_bytes())
 (source/'tsconfig.json').write_text(json.dumps({'compilerOptions':{'strict':True,'noUncheckedIndexedAccess':False,'target':'ES2022'}}))
 with (root/'run.log').open('w') as log:subprocess.run([str(binary),str(source),str(root/'raw.jsonl')],env={**os.environ,'LATENT_ASSERT_NO_OUTPUT':'1'},stdout=log,stderr=subprocess.STDOUT,check=True)
 header,record=[json.loads(x) for x in (root/'raw.jsonl').read_text().splitlines()]
 assert header['project_options'] and header['requires_indexed_presence']
 assert not header['diagnostic_sites'],header
 jsonsites=[x for x in header['option_dispositions'] if x.get('kind')=='json-stringify-defined']
 assert len(jsonsites)==1 and jsonsites[0]['state']=='scheduled-check',header
 assert record['units'][2]['status']=='attempted',record
 sites=[x for x in header['option_dispositions'] if 'noUncheckedIndexedAccess' in x['site']['options']]
 assert len(sites)==2 and all(x['state']=='scheduled-check' for x in sites)
 checks=[x for x in record['checks'] or [] if x['kind']=='indexed-presence']
 reads=record['reads'] or []
 assert len(checks)==1 and ':1:' in checks[0]['where'],record
 assert all(len(x['sites'])==1 for x in reads),record
 assert {int(x['where'].rsplit(':',2)[1]) for x in reads}=={1,2},record
 failures=[x for x in record['function_attempts'] if x.get('failure') and ':2:' in x['where']]
 assert len(failures)==1 and failures[0]['failure']['kind']=='NotYet' and failures[0]['failure']['reason']=='assigning a field of a value' and ':2:' in failures[0]['where'],record
 # The constructed line-2 guard is absent from IR after the enclosing refusal.
 assert not any(':2:' in x['where'] for x in checks)
 # Attribution/scheduling mutants must fail the independent guard witness.
 for name,forged in [('schedule-as-IR',checks+[{'where':reads[-1]['where']}]),('drop-IR-guard',[])]:
  try:assert len(forged)==1 and ':1:' in forged[0]['where']
  except AssertionError:print(name+' mutant caught')
  else:raise AssertionError(name+' survived')
 print('production profile: 2 scheduled; 2 constructed; 1 survives IR; blocked body discarded; scheduled JSON eligible; pass')
