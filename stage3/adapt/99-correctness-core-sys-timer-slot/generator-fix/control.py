import sys
sys.dont_write_bytecode=True
from run import *
assert run('head1-regenerate',['python3',S/'head1-control/stage3/lane/generate-api.py','/workspace/typescript-source','--output',S/'head1-regenerated.json'])==0
old=json.loads((S/'head1-api.log').read_text());new=json.loads((S/'head1-regenerated.json').read_text());fields=[k for k in old if old[k]!=new[k]]
assert fields==['proofs'],fields
changes=[{'file':a['file'],'before':a['sha256'],'after':b['sha256']} for a,b in zip(old['proofs'],new['proofs']) if a!=b];assert len(changes)==2
(S/'head1-generator-drift.json').write_text(json.dumps({'changedFields':fields,'changes':changes,'timerAdaptationAbsent':True},indent=2)+'\n');print(json.dumps(changes,indent=2))
