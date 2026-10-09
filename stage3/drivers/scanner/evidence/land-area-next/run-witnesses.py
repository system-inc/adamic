from pathlib import Path
import json,subprocess,time
r=Path('/workspace/scratch/native3-next-probes');r.mkdir(exist_ok=True)
previous=json.loads(Path('/workspace/adamic/stage3/drivers/scanner/evidence/native3-records/witnesses/results.json').read_text());old={x['name']:x['source'] for x in previous}
programs={
'01-index':old['01-index'],
'02-mutable-namespace':'export namespace Debug { export let isDebugging = false; }\nconsole.log("ok");\n',
'03-error-cast':'const error = new Error("failure");\nif ((Error as any).captureStackTrace) (Error as any).captureStackTrace(error);\nconsole.log("ok");\n',
'04-diagnostic-cast':old['04-cast'],
'05-uint16':'const values = new Uint16Array(2);\nconsole.log(String(values.length));\n',
'06-enum-slot':'enum SyntaxKind { Identifier = 1, Keyword = 2 }\nfunction numberValue(): number { return 1; }\nconst token: SyntaxKind.Identifier = numberValue();\nconsole.log(String(token));\n',
'07-string-cast':old['09-string-cast'],
'08-array':old['11-array'],
'09-generic-return':old['12-generic-return'],
'10-any':'function report(value: any): void { console.log("ok"); }\nreport("value");\n',
'11-computed-field':'const keywords = { ["" + "constructor"]: 1 };\nconsole.log("ok");\n',
'12-map-any':'const values = new Map<string, any>();\nconsole.log("ok");\n',
'13-entries':'function count(value: { x: number }): number { return Object.entries(value).length; }\nconsole.log(String(count({ x: 1 })));\n',
'14-any-callback':'const callback = (value: any): void => { console.log("ok"); };\ncallback("value");\n',
'15-never-string':'export function stopped(): never { throw new Error("placeholder"); }\nexport function name(): string | undefined { return stopped(); }\nconsole.log("ok");\n',
}
rows=[]
for name,source in programs.items():
 f=r/(name+'.a');f.write_text(source);row={'name':name,'source':source}
 for kind,command in [('node',['node','--disable-warning=ExperimentalWarning','/workspace/scanner-native3-next/oracle/node.mjs',str(f)]),('build',['/workspace/scratch/scanner-next-adamic','build',str(f),'-o',str(r/(name+'.native'))])]:
  start=time.monotonic()
  with (r/(name+'.'+kind+'.stdout')).open('wb') as out,(r/(name+'.'+kind+'.stderr')).open('wb') as err:row[kind+'_exit']=subprocess.run(command,cwd='/workspace/scanner-native3-next',stdout=out,stderr=err).returncode
  row[kind+'_seconds']=time.monotonic()-start;row[kind+'_stdout']=(r/(name+'.'+kind+'.stdout')).read_text();row[kind+'_stderr']=(r/(name+'.'+kind+'.stderr')).read_text()
 rows.append(row);print(json.dumps(row),flush=True)
(r/'results.json').write_text(json.dumps(rows,indent=2)+'\n')
