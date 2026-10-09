from pathlib import Path
import subprocess,json,time
root=Path('/workspace/scratch/native3-records-probes');root.mkdir(exist_ok=True)
programs={
'01-index': 'export interface MapLike<T> { [index: string]: T; }\nconsole.log("ok");\n',
'02-namespace': 'export namespace Debug { export function fail(): never { throw new Error("failure"); } }\nconsole.log("ok");\n',
'03-nullable': 'export type Value = string | number | boolean | string[] | null | undefined;\nconsole.log("ok");\n',
'04-cast': 'interface Message { flag?: boolean; }\nfunction diag(flag?: boolean) { return { flag }; }\nconst value = diag() as Message;\nconsole.log("ok");\n',
'05-comma': 'let x = 0;\nconst value = (x++, x);\nconsole.log(`${value}`);\n',
'06-shebang': 'function scan(text: string): number { return /#!/.exec(text)![0].length; }\nconsole.log(`${scan("#!")}`);\n',
'07-initializer': 'function make(): string { let text: string = undefined!; text = "ok"; return text; }\nconsole.log(make());\n',
'08-codepoint': 'function point(text: string): number { return text.codePointAt(0)!; }\nconsole.log(`${point("A")}`);\n',
'09-string-cast': 'const worker: (codePoint: number) => string = (String as any).fromCodePoint ? codePoint => (String as any).fromCodePoint(codePoint) : codePoint => String.fromCharCode(codePoint);\nconsole.log(worker(65));\n',
'10-property': 'const values = { Script_Extensions: undefined! as Set<string> };\nconsole.log("ok");\n',
'11-array': 'function make(length: number): number { const values = new Array(length); return values.length; }\nconsole.log(`${make(2)}`);\n',
'12-generic-return': 'export function each<K, V, U>(map: ReadonlyMap<K, V>, callback: (value: V, key: K) => U | undefined): U | undefined { for (const [key, value] of map) { const result = callback(value, key); if (result !== undefined) return result; } return undefined; }\nconsole.log(each(new Map<string, number>([["x", 1]]), (value, key) => key) ?? "missing");\n',
'13-function-union': 'const callback = (value: string | number): string => `${value}`;\nconsole.log(callback(1));\n',
'14-nullish': 'const Debug = { fail: (message: string): never => { throw new Error(message); } };\nfunction read(values: readonly number[]): number { return values[0] ?? Debug.fail("missing"); }\nconsole.log(`${read([1])}`);\n',
'15-condition': 'const match = /ok/.exec("ok");\nif (match) console.log("ok");\n'}
rows=[]
for name,source in programs.items():
 file=root/(name+'.a');file.write_text(source)
 row={'name':name,'source':source}
 for kind,command in [('node',['node','--disable-warning=ExperimentalWarning','/workspace/scanner-native3-scratch/oracle/node.mjs',str(file)]),('build',['/workspace/scratch/scanner-native3-records-adamic','build',str(file),'-o',str(root/(name+'.native'))])]:
  start=time.monotonic()
  with (root/(name+'.'+kind+'.stdout')).open('wb') as out,(root/(name+'.'+kind+'.stderr')).open('wb') as err:
   row[kind+'_exit']=subprocess.run(command,cwd='/workspace/scanner-native3-scratch',stdout=out,stderr=err).returncode
  row[kind+'_seconds']=time.monotonic()-start
  row[kind+'_stdout']=(root/(name+'.'+kind+'.stdout')).read_text();row[kind+'_stderr']=(root/(name+'.'+kind+'.stderr')).read_text()
 rows.append(row);print(json.dumps(row),flush=True)
(root/'results.json').write_text(json.dumps(rows,indent=2)+'\n')
