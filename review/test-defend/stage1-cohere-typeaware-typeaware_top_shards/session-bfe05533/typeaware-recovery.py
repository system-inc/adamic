import pathlib,json
code=pathlib.Path('/workspace/typeaware-mutant.py').read_text().split('try:\n f.write_text')[0]
exec(code,globals())
runs=json.loads((P/'mutant-runs.json').read_text())
try:
 f.write_text(original.replace(plan['old'],plan['new']))
 for id,name in [('D1-matrix-recovery-003','TestTypeAwareAgreementAndMutants_003'),('D1-matrix-recovery-010','TestTypeAwareAgreementAndMutants_010')]:
  r=run(id,[name])
  if len(r['states'])!=1 or r['exit']!=0:raise RuntimeError('incomplete or red recovery '+id)
finally:f.write_text(original)
