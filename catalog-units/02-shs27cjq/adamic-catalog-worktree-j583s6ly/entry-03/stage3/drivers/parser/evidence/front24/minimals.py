from pathlib import Path
import subprocess,json
root=Path('/workspace/adamic/stage3/drivers/parser');out=Path('/workspace/scratch/parser-front24-minimals');out.mkdir();rows=[]
for name in ['native-array-is-array-any.a','native-predicate-callback-parameter.a','native-namespace-object-receiver.a','native-namespace-class.a','native-callable-namespace.a','native-partial-record-view.a','native-function-any-view.a','native-debugger-statement.a','native-generic-assert-non-nullable.a']:
 row={'file':name}
 for mode,cmd in [('node',['node','--disable-warning=ExperimentalWarning','oracle/node.mjs',str(root/name)]),('emit',['/workspace/scratch/parser-front24-adamic','c',str(root/name)])]:
  with (out/(name+'.'+mode+'.stdout')).open('wb') as stdout,(out/(name+'.'+mode+'.stderr')).open('wb') as stderr:p=subprocess.run(cmd,cwd='/tmp/parser-front7-scratch',stdout=stdout,stderr=stderr)
  row[mode]={'exit':p.returncode,'stdout':(out/(name+'.'+mode+'.stdout')).read_text() if mode=='node' else 'C retained in scratch only','stderr':(out/(name+'.'+mode+'.stderr')).read_text()}
 rows.append(row)
(out/'report.json').write_text(json.dumps(rows,indent=2)+'\n')
