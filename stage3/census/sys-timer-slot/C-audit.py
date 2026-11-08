"""Cross-check C while proving that all pre-C A/B evidence remains unchanged."""
import pathlib,json,subprocess,hashlib,os
repo=pathlib.Path('/workspace/adamic');out=repo/'stage3/census/sys-timer-slot';s=pathlib.Path('/workspace/scratch/sys-timer-slot');parent='ef28f442f7cbc83f725da967532f5b06dfc8599e'
oldPaths=subprocess.check_output(['git','ls-tree','-r','--name-only',parent,'--','stage3/census/sys-timer-slot'],cwd=repo,text=True).splitlines();preserved=0
for name in oldPaths:
 if pathlib.Path(name).name in ['REPORT.md','REPORT.json']:continue
 assert (repo/name).read_bytes()==subprocess.check_output(['git','show',parent+':'+name],cwd=repo),name;preserved+=1
oldMD=subprocess.check_output(['git','show',parent+':stage3/census/sys-timer-slot/REPORT.md'],cwd=repo);assert (out/'REPORT.md').read_bytes().startswith(oldMD)
oldJSON=json.loads(subprocess.check_output(['git','show',parent+':stage3/census/sys-timer-slot/REPORT.json'],cwd=repo));newJSON=json.loads((out/'REPORT.json').read_text());assert {k:v for k,v in newJSON.items() if k!='C'}==oldJSON
C=json.loads((out/'C.json').read_text());assert newJSON['C']==C
for label,count in [('C-a',2),('C-b',2),('C-noinfer-a',2),('C-noinfer-b',0)]:
 row=C['variants'][label];assert len(row['checks']['adamic'])==len(row['checks']['solution'])==count
 assert row['sourceFilesCompared']==746 and row['changedSources']==['src/compiler/sys.ts']
 assert row['javascript']['identical'] and row['javascript']['count']==9
 assert len(row['checks']['headProbes'])==4
 expectedSet=label=='C-noinfer-b'
 for probe in row['checks']['headProbes']:assert probe['fits']==(expectedSet if probe['name']=='setTimeout' else True)
 for name in ['typescript.d.ts','typescript.internal.d.ts']:assert not row['declarations'][name]['identical']
 assert (s/label/'tests/baselines/reference/api/typescript.d.ts').read_bytes()==(s/'main/tests/baselines/reference/api/typescript.d.ts').read_bytes()
 assert len(row['latent']['kept'])+len(row['latent']['new'])==row['latent']['distinctSites']
 assert len(row['latent']['kept'])+len(row['latent']['gone'])==C['mainCensus']['distinctSites']
for loc in ['1501','1502','1599','1600']:
 rows=C['variants']['C-noinfer-b']['queriedMainLocations'][loc];assert any(x['kind']=='Refused' for x in rows),loc
assert all(not x['threw'] for x in C['nodeWitness']['results']) and len(C['nodeWitness']['results'])==4
assert not C['isolatedCancellationProbe']['header']['checker_rejected']
assert hashlib.sha256((s/'census').read_bytes()).hexdigest()==oldJSON['provenance']['censusSha256']
for p in out.rglob('*'):
 assert p.suffix not in ['.go','.pyc'],str(p)
 if p.suffix in ['.md','.txt','.py','.cjs']:
  t=p.read_text();assert '\u2014' not in t and '\u2013' not in t,str(p)
ancestors=set();pid=os.getpid()
while pid>1 and pid not in ancestors:
 ancestors.add(pid)
 try:pid=int(pathlib.Path(f'/proc/{pid}/stat').read_text().split(') ')[1].split()[1])
 except (FileNotFoundError,ValueError):break
live=[]
for p in pathlib.Path('/proc').iterdir():
 if not p.name.isdigit() or int(p.name) in ancestors:continue
 try:stat=(p/'stat').read_text().split(') ')[1].split();cmd=(p/'cmdline').read_bytes().replace(b'\0',b' ').decode(errors='replace');cwd=os.readlink(p/'cwd')
 except (PermissionError,FileNotFoundError,ProcessLookupError):continue
 if stat[0]!='Z' and (str(s) in cmd or cwd.startswith(str(s))):live.append({'pid':int(p.name),'command':cmd,'cwd':cwd})
assert not live,live
result={'priorEvidenceFilesPreserved':preserved,'priorReportPrefixAndJsonPreserved':True,'CSourceScope':True,'pairedDiagnostics':True,'allNineJavascriptFilesIdentical':True,'exactDeclarationDiffs':True,'fourRuntimeWitnessesPassed':True,'fourProviderVarianceRefusalsInCheckerCleanAlternative':True,'sameCensusBinaryAsAB':True,'liveTaskProcesses':live}
(out/'C-audit.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result))
