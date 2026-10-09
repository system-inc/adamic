from pathlib import Path
import subprocess,shutil,json,re,time,os
root=Path('/workspace/scratch/scanner-any-sites'); scanner=root/'adapted/src/compiler/scanner.ts'; original=scanner.read_text(); results=[]
def build(name):
 start=time.monotonic()
 with (root/(name+'.stdout')).open('wb') as out,(root/(name+'.stderr')).open('wb') as err:
  code=subprocess.run(['/workspace/scratch/scanner-any-next-adamic','build',str(root/'main.a'),'-o',str(root/'scanner')],cwd='/workspace/scanner-native3-next',stdout=out,stderr=err).returncode
 diagnostic=(root/(name+'.stderr')).read_text(); row={'name':name,'exit':code,'wall_seconds':time.monotonic()-start,'diagnostic':diagnostic};results.append(row);print(json.dumps(row),flush=True)
 (root/'checkpoints.json').write_text(json.dumps(results,indent=2)+'\n')
scanner.write_text(original.replace('arg0?: string | number','arg0?: any'));build('private-restored-any');scanner.write_text(original)
# Suppress the containing factory only for discovery of later map/driver sites.
s=original;m=re.search(r'export function createScanner\(',s);line=s[:m.start()].count('\n')+1
r=subprocess.run(['node','/tmp/scanner-next-stub.cjs',str(scanner),str(line),'1'],capture_output=True,text=True);assert r.returncode==0,r.stderr;(root/'factory-placeholder.json').write_text(r.stdout)
s=scanner.read_text();start=s.index('export const textToKeywordObj:');end=s.index('\nconst textToKeyword =',start)
s=s[:start]+'export const textToKeywordObj: MapLike<KeywordSyntaxKind> = ((): MapLike<KeywordSyntaxKind> => { throw new Error("discovery placeholder: keyword object"); })();\n'+s[end:];scanner.write_text(s);build('map-concrete')
scanner.write_text(s.replace('new Map<string, KeywordSyntaxKind>(Object.entries(textToKeywordObj))','new Map(Object.entries(textToKeywordObj))'));build('map-restored-inference');scanner.write_text(s)
# Remove only unrelated unsupported Object.entries operations, keeping driver callback.
s=s.replace('const textToKeyword = new Map<string, KeywordSyntaxKind>(Object.entries(textToKeywordObj));','const textToKeyword: Map<string, KeywordSyntaxKind> = ((): Map<string, KeywordSyntaxKind> => { throw new Error("discovery placeholder: keyword map"); })();')
start=s.index('const nonBinaryUnicodeProperties = ');end=s.index('\nconst ',start+1)
s=s[:start]+'const nonBinaryUnicodeProperties: Map<string, string> = ((): Map<string, string> => { throw new Error("discovery placeholder: Unicode aliases"); })();\n'+s[end:];scanner.write_text(s);build('driver-concrete')
p=root/'main.a';driver=p.read_text();p.write_text(driver.replace('arg0: string | number | undefined','arg0'));build('driver-restored-any');p.write_text(driver)
(root/'final-placeholder-scanner.ts.txt').write_text(s)
