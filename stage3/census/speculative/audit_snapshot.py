"""Prove optimized speculative snapshots retain legacy values and mutable ownership."""
import json, os, subprocess, sys
from pathlib import Path
repo, overlay, out = map(lambda x: Path(x).resolve(), sys.argv[1:4])
out.mkdir(parents=True, exist_ok=True)
config=json.loads((overlay/'overlay.json').read_text())
config['Replace'][str(repo/'internal/lower/latent_scalar_copy_test.go')]=str(repo/'stage3/census/latent/scalar_copy_test.go.txt')
(out/'overlay.json').write_text(json.dumps(config))
for name,mutant in [('baseline','0'),('shared-slice-mutant','1')]:
 env=dict(os.environ,LATENT_MUTANT_SHARE_SCALAR_SLICE=mutant)
 with (out/(name+'.log')).open('w') as log:
  result=subprocess.run(['go','test','-overlay='+str(out/'overlay.json'),'./internal/lower','-run','^TestLatentSpecSnapshotScalarCopy$','-count=1'],cwd=repo,env=env,stdout=log,stderr=subprocess.STDOUT)
 text=(out/(name+'.log')).read_text()
 if name=='baseline': assert result.returncode==0,text
 else: assert result.returncode!=0 and 'snapshot aliases mutable source state' in text,text
 print(name+': '+('PASS' if name=='baseline' else 'caught by snapshot ownership check'))

for name,virtual,needle,replacement,message in [
 ('scalar-value-mutant','latent_full.go','result.Set(v)\n\t\t\t\tfor _, index','for _, index','optimized snapshot differs from legacy copy'),
 ('function-cursor-mutant','latent_full.go','copy.function = &copy.result.Functions[i]','copy.function = l.function','snapshot function cursor not rebound'),
 ('private-field-mutant','latent_speculative.go','field.PkgPath != "" || !latentSpecImmutable(field.Type)','!latentSpecImmutable(field.Type)','private field copied silently'),
 ('mapper-identity-mutant','latent_state.go','func latentCopy_lowering(value lowering, seen map[any]any) lowering {\n\tresult := value','func latentCopy_lowering(value lowering, seen map[any]any) lowering {\n\tresult := value\n result.typeMapper = new(typeMapper)','checker mapper identity changed'),
]:
 mutated=json.loads(json.dumps(config))
 key=str(repo/'internal/lower'/virtual)
 text=Path(mutated['Replace'][key]).read_text()
 assert text.count(needle)>=1,(name,needle)
 destination=out/(name+'.go')
 destination.write_text(text.replace(needle,replacement,1))
 mutated['Replace'][key]=str(destination)
 settings=out/(name+'.json')
 settings.write_text(json.dumps(mutated))
 with (out/(name+'.log')).open('w') as log:
  result=subprocess.run(['go','test','-overlay='+str(settings),'./internal/lower','-run','^TestLatentSpecSnapshotScalarCopy$','-count=1'],cwd=repo,env=dict(os.environ,LATENT_MUTANT_SHARE_SCALAR_SLICE='0'),stdout=log,stderr=subprocess.STDOUT)
 text=(out/(name+'.log')).read_text()
 assert result.returncode!=0 and message in text,text
 print(name+': caught by '+message)

with (out/'function-schema-mutant.log').open('w') as log:
 result=subprocess.run(['go','test','-overlay='+str(out/'overlay.json'),'./internal/lower','-run','^TestLatentSpecSnapshotScalarCopy$','-count=1'],cwd=repo,env=dict(os.environ,LATENT_MUTANT_SKIP_FUNCTION_SCHEMA='1',LATENT_MUTANT_SHARE_SCALAR_SLICE='0'),stdout=log,stderr=subprocess.STDOUT)
text=(out/'function-schema-mutant.log').read_text()
assert result.returncode!=0 and 'function schema drift copied silently' in text,text
print('function-schema-mutant: caught by unsupported mutable function-field guard')
