import pathlib,json,subprocess,time,re
p=pathlib.Path('review/test-audit/stage1-cohere-formatfiles-formatfiles_shards');names=['TestFormatfilesShardPlantedDisagreement']+[f'TestThePortParsesAsGoCohereDoes_{i:03d}' for i in range(32)]+['TestThePortParsesAsGoCohereDoesUnion'];listed=(p/'logs'/'list.log').read_text().splitlines();assert all(n in listed for n in names);(p/'scope.json').write_text(json.dumps(dict(base=subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip(),tests=names,moved=[],vanished=[]),indent=2))
results=[]
for label,regex in [('witness','^TestFormatfilesShardPlantedDisagreement$'),('family','^TestThePortParsesAsGoCohereDoes')]:
 for i in range(3):
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/formatfiles/','-run',regex];start=time.monotonic()
  with (p/'logs'/f'timing-{label}-{i}.log').open('w') as out:r=subprocess.run(cmd,stdout=out,stderr=subprocess.STDOUT)
  es=[json.loads(l) for l in (p/'logs'/f'timing-{label}-{i}.log').read_text().splitlines() if l.startswith('{')];assert r.returncode==0;seconds=None
  for e in es:
   m=re.search(r'\bok\s+\S+\s+([0-9.]+)s',e.get('Output',''))
   if m:seconds=float(m[1])
  results.append(dict(label=label,command=' '.join(cmd),seconds=seconds,wall_seconds=time.monotonic()-start));(p/'timings.json').write_text(json.dumps(results,indent=2));print(label,i,seconds,flush=True)
