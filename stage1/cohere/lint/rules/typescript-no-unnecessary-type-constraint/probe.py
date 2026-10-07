#!/usr/bin/env python3
"""Observe unchanged Go findings and the existing serializer's explicit refusals."""
import argparse,json,subprocess
from pathlib import Path
owned=Path(__file__).resolve().parent
repository=owned.parents[4]
p=argparse.ArgumentParser();p.add_argument('--scratch',type=Path,required=True);args=p.parse_args()
scratch=args.scratch.resolve();scratch.mkdir(parents=True,exist_ok=True)
source=(repository/'stage1/cohere/lint/testdata/oracle.go').read_text()
# Keep collection and serialization byte-for-byte; replace only the unused main.
source=source.replace('func main() {','func serializerMain() {')
(scratch/'oracle.go').write_text(source)
virtual=repository/'cohere/wave15_third_oracle.go'
probe=repository/'cohere/wave15_third_probe.go'
overlay=scratch/'overlay.json'
overlay.write_text(json.dumps({'Replace':{str(virtual):str(scratch/'oracle.go'),str(probe):str(owned/'probe.go.txt')}}))
with (scratch/'build.log').open('w') as log:
 subprocess.run(['go','build','-overlay='+str(overlay),'-o',str(scratch/'oracle'),str(virtual),str(probe)],cwd=repository/'cohere',stdout=log,stderr=subprocess.STDOUT,check=True)
rows=[('typescript-no-unnecessary-type-constraint','@typescript-eslint/no-unnecessary-type-constraint','unexpected suggestion shape'),('typescript-prefer-as-const','@typescript-eslint/prefer-as-const','unexpected fix shape'),('typescript-prefer-enum-initializers','@typescript-eslint/prefer-enum-initializers','unexpected suggestion shape')]
for slug,name,error in rows:
 witness=repository/'stage1/cohere/lint/rules'/slug/'testdata/blocked.ts.txt'
 with (scratch/(slug+'-shape.log')).open('w') as log:
  subprocess.run([str(scratch/'oracle'),'--shape',name,str(witness)],stdout=log,stderr=subprocess.STDOUT,check=True)
 with (scratch/(slug+'-serializer.log')).open('w') as log:
  result=subprocess.run([str(scratch/'oracle'),name,str(witness)],stdout=log,stderr=subprocess.STDOUT)
 output=(scratch/(slug+'-serializer.log')).read_text()
 if result.returncode!=2 or 'panic: '+error not in output:raise SystemExit('missing expected serializer refusal: '+name)
 print(name+': shape exit=0; serializer exit=2 '+error,flush=True)

# These are Go-only witness sensitivity checks, not port/backend mutants.
mutations=[
 ('typescript-no-unnecessary-type-constraint','no_unnecessary_type_constraint.go','core.NewTextRange(name.End(), typeParameter.Constraint.End())','core.NewTextRange(name.End(), name.End())'),
 ('typescript-prefer-as-const','prefer_as_const.go','ctx.InsertAfter(initializer, " as const")','ctx.InsertAfter(initializer, " as mutable")'),
 ('typescript-prefer-enum-initializers','prefer_enum_initializers.go','strconv.Itoa(index+1)','strconv.Itoa(index+2)'),
]
for (slug,name,error),(mutantSlug,file,anchor,replacement) in zip(rows,mutations):
 original=repository/'cohere/internal/lint/rules/typescript'/file
 body=original.read_text()
 if body.count(anchor)!=1:raise SystemExit('mutant anchor not unique: '+file)
 mutated=scratch/file;mutated.write_text(body.replace(anchor,replacement))
 mutantOverlay=scratch/(slug+'-mutant-overlay.json')
 mapping=json.loads(overlay.read_text());mapping['Replace'][str(original)]=str(mutated)
 mutantOverlay.write_text(json.dumps(mapping))
 binary=scratch/(slug+'-mutant')
 with (scratch/(slug+'-mutant-build.log')).open('w') as log:
  subprocess.run(['go','build','-overlay='+str(mutantOverlay),'-o',str(binary),str(virtual),str(probe)],cwd=repository/'cohere',stdout=log,stderr=subprocess.STDOUT,check=True)
 witness=repository/'stage1/cohere/lint/rules'/slug/'testdata/blocked.ts.txt'
 with (scratch/(slug+'-mutant-shape.log')).open('w') as log:
  subprocess.run([str(binary),'--shape',name,str(witness)],stdout=log,stderr=subprocess.STDOUT,check=True)
 baseline=(scratch/(slug+'-shape.log')).read_bytes()
 changed=(scratch/(slug+'-mutant-shape.log')).read_bytes()
 if baseline==changed:raise SystemExit('Go probe mutant escaped output comparison: '+name)
 print(name+': Go-only repair mutant compiled, exited 0, caught by shape-output comparison',flush=True)
