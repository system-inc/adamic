from pathlib import Path
import shutil,subprocess,json,re,time
root=Path('/workspace/scratch/native3-next-walk')
if not root.exists():
 root.mkdir();shutil.copytree('/workspace/scratch/native3-slice',root/'adapted')
 for name in ['main.a','token-names.a']:shutil.copyfile('/workspace/scratch/native3-combined-unsplit/'+name,root/name)
rows=json.loads((root/'stops.json').read_text()) if (root/'stops.json').exists() else []
for order in range(len(rows)+1,16):
 start=time.monotonic()
 with (root/f'stop-{order:02}.stdout').open('wb') as out,(root/f'stop-{order:02}.stderr').open('wb') as err:
  code=subprocess.run(['/workspace/scratch/scanner-next-adamic','build',str(root/'main.a'),'-o',str(root/'scanner')],cwd='/workspace/scanner-native3-next',stdout=out,stderr=err).returncode
 diagnostic=(root/f'stop-{order:02}.stderr').read_text();m=re.search(r'adamic: (.+?):(\d+):(\d+): (.*)',diagnostic)
 if code==0:print('DISCOVERY BUILD SUCCEEDED: throwing placeholders, no scanner proof',flush=True);break
 assert code==1 and m,(order,code,diagnostic)
 file,line,col,message=m.groups();row={'order':order,'file':file.split('/src/compiler/')[-1],'line':int(line),'column':int(col),'message':message,'build_exit':code,'wall_seconds':time.monotonic()-start};rows.append(row)
 (root/'stops.json').write_text(json.dumps(rows,indent=2)+'\n');print(json.dumps(row),flush=True)
 if order==15:break
 p=Path(file);s=p.read_text()
 if 'index signature' in message and '[index: string]: T;' in s:
  s=s.replace('[index: string]: T;','')+'\nexport function discoveryMapLike(): never { throw new Error("discovery placeholder: MapLike"); }\n';p.write_text(s)
 elif 'both null and undefined' in message and 'export type CompilerOptionsValue =' in s:
  original=next(x for x in s.splitlines() if x.startswith('export type CompilerOptionsValue ='));s=s.replace(original,'export type CompilerOptionsValue = string;\nexport function discoveryCompilerOptionsValue(): never { throw new Error("discovery placeholder: CompilerOptionsValue"); }');p.write_text(s)
 elif 'utf16EncodeAsStringWorker' in s.splitlines()[int(line)-1]:
  original=next(x for x in s.splitlines() if x.startswith('const utf16EncodeAsStringWorker:'));s=s.replace(original,'const utf16EncodeAsStringWorker: (codePoint: number) => string = (codePoint: number): string => { throw new Error("discovery placeholder: UTF16 worker"); };');p.write_text(s)
 elif 'Script_Extensions: undefined! as Set<string>' in s.splitlines()[int(line)-1]:
  s=s.replace('Script_Extensions: undefined! as Set<string>,','Script_Extensions: (() : Set<string> => { throw new Error("discovery placeholder: Script_Extensions"); })(),');p.write_text(s)
 else:
  command=['node','/tmp/scanner-next-stub.cjs',file,line,col]
  if 'function returning' in message:command.append('never')
  result=subprocess.run(command,capture_output=True,text=True)
  (root/f'stub-{order:02}.json').write_text(result.stdout)
  if result.returncode:print('NEEDS PLACEHOLDER: '+result.stderr,flush=True);break
print('WALK RECORDED '+str(len(rows))+' stops.',flush=True)
