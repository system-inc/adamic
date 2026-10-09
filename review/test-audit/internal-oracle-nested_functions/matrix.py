import json,subprocess,os,time
from pathlib import Path
p=Path(__file__).resolve().parent
rows=json.loads((p/'rows.json').read_text())
extra=['TestNativeAgreesWithNode','TestScannerNestedReferences','TestScannerNestedReferenceMutants','TestNestedReferenceIdentityMutant','TestTheOracleCatchesOneByte']
pattern='^('+'|'.join(rows+extra)+')$/^(internal|stage3|nested)/^(oracle|drivers)$/^(testdata|scanner)$/^(nested_|new_expression_|typed_arrays|probes)'
(p/'matrix-scope.json').write_text(json.dumps({'rows':rows+extra,'pattern':pattern,'limitation':'TestNativeAgreesWithNode is restricted to named fixture subtests. All other package kills unknown.'},indent=2))
for id in ['control']+[x['id'] for x in json.loads((p/'plan.json').read_text())]:
 env=dict(os.environ,ADAMIC_GATE_UNCACHED='1',ADAMIC_MUTANT='' if id=='control' else id)
 subprocess.run(['python3',str(p/'run.py'),id,'-run',pattern],env=env,check=True)
 result=json.loads((p/(id+'.json')).read_text())
 if id=='control' and result['exit']!=0:raise SystemExit('RED CONTROL, stopping')
