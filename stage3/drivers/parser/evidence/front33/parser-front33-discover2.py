from pathlib import Path
import subprocess,json,re
root=Path('/workspace/scratch/parser-front33-discovery2-slice');out=Path('/workspace/scratch/parser-front33-discovery2');out.mkdir();rows=[]
compiler='/workspace/scratch/parser-front33-adamic';scratch='/workspace/scratch/parser-front33-main'
def run(cmd,stem):
 with (out/(stem+'.stdout')).open('wb') as a,(out/(stem+'.stderr')).open('wb') as b: code=subprocess.run(cmd,cwd=scratch,stdout=a,stderr=b).returncode
 return code,(out/(stem+'.stderr')).read_text()
for i in range(15):
 code,text=run([compiler,'c',str(root/'parser-proof-main.a')],f'{i+1:02}')
 row={'order':i+1,'exit':code,'all_diagnostics':text};rows.append(row)
 if code==0:row['c_emitted']=True;break
 m=re.search(r'(?:adamic: )?(/.*?):(\d+):(\d+): (.*)',text)
 if not m:row['reason']='No diagnostic coordinate';break
 file,line,col,message=m.groups();p=Path(file);row.update(file=str(p.relative_to(root)),line=int(line),column=int(col),message=message,found_behind_stubs=i>0)
 if 'before all runtime namespaces' in message:
  code,err=run(['node','/workspace/scratch/parser-front33-variable-placeholder.cjs',file,line,col],f'{i+1:02}-stub');row['placeholder']='Typed let declaration and unconditional arithmetic guard throwing before it; no calls. Discovery only.'
 elif p.name=='corePublic.ts' and 'index signature' in message:
  text_source=p.read_text();before='export interface MapLike<T> {\n    [index: string]: T;\n}';assert text_source.count(before)==1
  replacement='export type MapLike<T> = Record<string, T>;\nfunction parserProofMissingMapLike<T>(): Record<string, T> { throw \"discovery-only dictionary declaration\"; }\n';p.write_text(text_source.replace(before,replacement));code=0;err='';row['placeholder']=replacement
 elif 'class inside a namespace' in message:
  code,err=run(['node','/workspace/scratch/parser-front30-class-placeholder.cjs',file],f'{i+1:02}-stub');row['placeholder']='Namespace class replaced by interface and throwing typed value; discovery only.'
 elif 'namespace merged with a function' in message:
  s=p.read_text();start=s.index('export function log(');end=s.index('\n}',start)+2;old=s[start:end];replacement='export function parserProofLogPlaceholder(s: string): void { throw "discovery-only merged log"; }'+'\n'*old.count('\n');p.write_text(s[:start]+replacement+s[end:]);code=0;err='';row['placeholder']=replacement
 elif 'there is no body proving this parameter' in message and p.name=='core.ts':
  code,err=run(['node','/workspace/scratch/parser-front33-overload-placeholder.cjs',file,line,col],f'{i+1:02}-stub');row['placeholder']='All predicates in the containing overload group become boolean; implementation throws. Exact edits retained; discovery only.'
 else:
  if p.name=='performanceCore.ts' and 'enum initialization' in message:
   text_source=p.read_text();text_source=text_source.replace('function tryGetPerformance() {','function tryGetPerformance(): { shouldWriteNativeEvents: boolean; performance: Performance } | undefined {');p.write_text(text_source)
  code,err=run(['node','/workspace/scratch/parser-front33-stub.cjs',file,line,col],f'{i+1:02}-stub')
  if code==0: row['stub']=json.loads((out/f'{i+1:02}-stub.stdout').read_text());row['placeholder']='Throwing function body; exact predicate callback edit retained if applicable.'
 row['stub_exit']=code
 if code:row['stub_error']=err;break
 (out/'report.json').write_text(json.dumps({'warning':'Discovery only. No Node identity/native acceptance claim.','rows':rows},indent=2)+'\n')
(out/'report.json').write_text(json.dumps({'warning':'Discovery only. No Node identity/native acceptance claim.','rows':rows},indent=2)+'\n')
