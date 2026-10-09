from pathlib import Path
import shutil,subprocess,json,re,time
root=Path('/workspace/scratch/native3-records-walk');root.mkdir();shutil.copytree('/workspace/scratch/native3-slice',root/'adapted')
for name in ['main.a','token-names.a']:shutil.copyfile('/workspace/scratch/native3-unsplit-proof/'+name,root/name)
expected=json.loads(Path('/workspace/adamic/stage3/drivers/scanner/evidence/native3/stops.json').read_text());rows=[]
for order in range(1,16):
 start=time.monotonic()
 with (root/f'stop-{order:02}.stdout').open('wb') as out,(root/f'stop-{order:02}.stderr').open('wb') as err:
  code=subprocess.run(['/workspace/scratch/scanner-native3-records-adamic','build',str(root/'main.a'),'-o',str(root/'scanner')],cwd='/workspace/scanner-native3-scratch',stdout=out,stderr=err).returncode
 diagnostic=(root/f'stop-{order:02}.stderr').read_text();m=re.search(r'adamic: (.+?):(\d+):(\d+): (.*)',diagnostic)
 assert code==1 and m,(order,code,diagnostic)
 file,line,col,message=m.groups();assert message==expected[order-1]['message'],(order,message)
 row={'order':order,'file':file.split('/src/compiler/')[-1],'line':int(line),'column':int(col),'message':message,'build_exit':code,'wall_seconds':time.monotonic()-start,'probe':expected[order-1]['probe']};rows.append(row);print(json.dumps(row),flush=True)
 if order==15:break
 p=Path(file);s=p.read_text()
 if order==1:s=s.replace('[index: string]: T;','')+'\nexport function discoveryMapLikePlaceholder(): never { throw new Error("discovery placeholder 1: MapLike"); }\n'
 elif order==2:
  pos=s.index('export namespace Debug');s=s[:pos]+'''export const Debug = {
 isDebugging: false,
 fail: (message?: string, stackCrawlMark?: {}): never => { throw new Error("discovery placeholder 2: Debug.fail"); },
 assertEqual: <T>(a: T, b: T, msg?: string, msg2?: string, stackCrawlMark?: {}): void => { throw new Error("discovery placeholder 2: Debug.assertEqual"); }
};\n'''
 elif order==3:
  original=next(x for x in s.splitlines() if x.startswith('export type CompilerOptionsValue ='));s=s.replace(original,'export type CompilerOptionsValue = string;\nexport function discoveryCompilerOptionsValue(): never { throw new Error("discovery placeholder 3"); }')
 elif order==4:
  a=s.index('function diag(');b=s.index('\n}',a)+2;signature=s[a:s.index(' {\n',a)];s=s[:a]+signature+': DiagnosticMessage { throw new Error("discovery placeholder 4: diag"); }'+s[b:]
 elif order==9:
  original=next(x for x in s.splitlines() if x.startswith('const utf16EncodeAsStringWorker:'));s=s.replace(original,'const utf16EncodeAsStringWorker: (codePoint: number) => string = (codePoint: number): string => { throw new Error("discovery placeholder 9: UTF16 worker"); };')
 elif order==10:s=s.replace('Script_Extensions: undefined! as Set<string>,','Script_Extensions: (() : Set<string> => { throw new Error("discovery placeholder 10: Script_Extensions"); })(),')
 else:
  result=subprocess.run(['node','/tmp/scanner-native3-stub.cjs',file,line,col],capture_output=True,text=True);assert result.returncode==0,result.stderr
  (root/f'stub-{order:02}.json').write_text(result.stdout);s=p.read_text()
  if order==12:s=s.replace('): U | undefined { throw new Error("discovery placeholder: forEachEntry"); }','): never { throw new Error("discovery placeholder: forEachEntry"); }')
 p.write_text(s)
(root/'stops.json').write_text(json.dumps(rows,indent=2)+'\n')
print('PASS: fifteen fresh ordered stops match the prior diagnostic messages.',flush=True)
