import pathlib,subprocess,json,re,collections,shutil
base=pathlib.Path('/workspace/scratch')
exec((base/'check-types.py').read_text().split('before,n=run(')[0])
tree=base/'mutant-final'
shutil.copytree(base/'after-final/src',tree/'src')
(tree/'node_modules').symlink_to(base/'after-final/node_modules',target_is_directory=True)
control,_=run('mutant-control',str(tree))
expected=json.loads((base/'diagnostics-after.json').read_text())
assert collections.Counter(control)==collections.Counter(expected)
originals={f:(tree/'src/compiler'/f).read_bytes() for f in ['types.ts','commandLineParser.ts']}
sites=[('CompilerOptions',n,'boolean') for n in ['showConfig','help','init','listFiles','listEmittedFiles','build','version']]+[('CompilerOptions','configFile','TsConfigSourceFile'),('CompilerOptions','project','string'),('ProjectReference','originalPath','string'),('CompilerOptions','configFilePath','string'),('TSConfig','include','readonly string[]'),('TSConfig','exclude','readonly string[]')]
report=[]
try:
 for owner,member,typ in sites:
  fp=tree/'src/compiler'/('commandLineParser.ts' if owner=='TSConfig' else 'types.ts')
  original=originals[fp.name];text=original.decode()
  needle=f'{member}?: {typ} | undefined;'
  assert text.count(needle)==1,(member,text.count(needle))
  fp.write_bytes(text.replace(needle,f'{member}?: {typ};').encode())
  rows,_=run('mutant-'+member,str(tree))
  added=list((collections.Counter(rows)-collections.Counter(control)).elements());removed=list((collections.Counter(control)-collections.Counter(rows)).elements())
  assert len(added)==(2 if member=='configFilePath' else 1) and not removed,(member,added,removed)
  if member=='configFilePath':assert 'TS2412:' in added[1] and 'commandLineParser.ts:3092:5:' in added[1]
  code='2322' if member=='originalPath' else '2375'
  assert 'commandLineParser.ts:' in added[0] and 'TS'+code+':' in added[0],(member,added)
  report.append({'owner':owner,'member':member,'diagnostics':len(rows),'restored':added[0],'allAdded':added})
  (base/'mutants.json').write_text(json.dumps(report,indent=2)+'\n')
  fp.write_bytes(original)
finally:
 for f,data in originals.items():(tree/'src/compiler'/f).write_bytes(data)
for f,data in originals.items():assert (tree/'src/compiler'/f).read_bytes()==data
print(json.dumps({'mutantsKilled':len(report)},indent=2))
