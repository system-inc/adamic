from pathlib import Path
import os,subprocess,json
base=Path('/workspace/scratch/scanner-main-next-records');out=base/'forof-candidates';out.mkdir()
sources=['function consume(values: (string | number)[]): void { for (const value of values) console.log("ok"); } consume(["s", 1]);\n','function consume(values: string | string[]): void { for (const value of values) console.log(value); } consume("ok");\n']
rows=[]
for i,source in enumerate(sources):
 file=out/(str(i)+'.a');file.write_text(source);row={'source':source}
 for phase,cmd in [('node',['node','/workspace/adamic/stage3/drivers/scanner/scratch-witness.cjs',file]),('build',[base/'adamic','build',file,'-o',out/(str(i)+'-native')])]:
  with (out/(str(i)+'-'+phase+'.stdout')).open('wb') as stdout,(out/(str(i)+'-'+phase+'.stderr')).open('wb') as stderr:
   code=subprocess.run(['bash','-c','source /workspace/adamic-tools/env.sh; exec "$@"','tools',*map(str,cmd)],cwd=base/'compiler-tree',env=dict(os.environ,SCANNER_TYPESCRIPT='/workspace/scratch/native3-cache/api/node_modules/typescript/lib/typescript.js',ADAMIC_NATIVE_SPLIT='0'),stdout=stdout,stderr=stderr).returncode
  row[phase+'_exit']=code;row[phase+'_stderr']=(out/(str(i)+'-'+phase+'.stderr')).read_text()
 rows.append(row)
(out/'results.json').write_text(json.dumps(rows,indent=2)+'\n');print(json.dumps(rows,indent=2))
