from pathlib import Path
import subprocess, os, json
root=Path('/workspace/utf16-measure/mutants'); scratch=root.parent;env=os.environ.copy()
index=root/'internal/native/runtime/string_index.c';append=root/'internal/native/runtime/string_append.c';index_good=index.read_text();append_good=append.read_text()
mutants=[('append-keeps-cursor',append,append_good.replace('\t\tadamic_string_free_index(string);\n\t\tstring->index = NULL;','\t\t// Mutant: retain the stale index and its cursor.'),'TestStringViewAfterAppendMatchesNode','heap-buffer-overflow'),('supplementary-one-unit',index,index_good.replace('units += size == 4 ? 2 : 1;', 'units += 1;',1),'TestStringIndexMatchesNode/mixed','native lines'),('backward-return-cursor',index,index_good.replace('while (start > unit) {','while (start > unit) {\n\t\t\t\t\t*low = false;\n\t\t\t\t\treturn offset;',1),'TestStringIndexMatchesNode/mixed','mismatches')]
rows=[]
try:
 for name,path,source,test,expected in mutants:
  index.write_text(index_good);append.write_text(append_good);path.write_text(source)
  log_path=scratch/('final-mutant-'+name+'.log')
  with log_path.open('w') as log:
   result=subprocess.run(['go','test','-count=1','-v','-run',test,'./internal/native'],cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT)
  output=log_path.read_text()
  if result.returncode==0 or expected not in output: raise RuntimeError(name+' was not caught for the expected reason')
  rows.append(dict(mutant=name,exit=result.returncode,caught_by=expected));print(rows[-1],flush=True)
finally:
 index.write_text(index_good);append.write_text(append_good)
(scratch/'mutants.json').write_text(json.dumps(rows,indent=2)+'\n')
