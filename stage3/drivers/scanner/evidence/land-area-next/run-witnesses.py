from pathlib import Path
import argparse,hashlib,json,subprocess,time
repository=Path(__file__).resolve().parents[5]
parser=argparse.ArgumentParser(description='Refresh selected minimal scanner witnesses with an explicit compiler.')
parser.add_argument('--compiler',type=Path,required=True)
parser.add_argument('--output',type=Path,required=True)
parser.add_argument('--case',action='append',default=[])
args=parser.parse_args();compiler=args.compiler.resolve();r=args.output.resolve();r.mkdir(parents=True,exist_ok=True)
previous=json.loads((repository/'stage3/drivers/scanner/evidence/native3-records/witnesses/results.json').read_text());old={x['name']:x['source'] for x in previous}
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
assert not set(args.case)-programs.keys(), 'unknown witness case'
for name,source in programs.items():
 if args.case and name not in args.case:continue
 f=Path(__file__).resolve().parent/'witnesses'/(name+'.a')
 assert f.read_text()==source, f'generator and tracked witness differ: {name}'
 row={'name':name,'source':source,'source_sha256':hashlib.sha256(f.read_bytes()).hexdigest(),'compiler_sha256':hashlib.sha256(compiler.read_bytes()).hexdigest()}
 for kind,command in [('node',['node','--disable-warning=ExperimentalWarning',str(repository/'oracle/node.mjs'),str(f)]),('build',[str(compiler),'build',str(f),'-o',str(r/(name+'.native'))])]:
  start=time.monotonic()
  with (r/(name+'.'+kind+'.stdout')).open('wb') as out,(r/(name+'.'+kind+'.stderr')).open('wb') as err:row[kind+'_exit']=subprocess.run(command,cwd=repository,stdout=out,stderr=err).returncode
  row[kind+'_seconds']=time.monotonic()-start;row[kind+'_stdout']=(r/(name+'.'+kind+'.stdout')).read_text();row[kind+'_stderr']=(r/(name+'.'+kind+'.stderr')).read_text()
 if name in ['10-any','14-any-callback']:
  assert row['node_exit']==0 and row['node_stdout']=='ok\n' and row['build_exit']==1
  assert 'Adamic 0.1 refuses a value of type any: any has no proven type' in row['build_stderr']
  row['classification']='Refused'
 rows.append(row);print(json.dumps(row),flush=True)
(r/'results.json').write_text(json.dumps(rows,indent=2)+'\n')
