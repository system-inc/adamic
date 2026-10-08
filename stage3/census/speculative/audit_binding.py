"""The hidden-parameter census must reject bypassing the real local lowerer."""
import json, os, shutil, subprocess, sys
from pathlib import Path
repo,overlay,binary,out=map(lambda x:Path(x).resolve(),sys.argv[1:5])
out.mkdir(parents=True,exist_ok=True)
folder=out/'mutant-overlay'
shutil.copytree(overlay,folder,dirs_exist_ok=True)
source=folder/'internal_lower_latent_speculative.go'
text=source.read_text()
needle="\n\treturn l.latentDeclareLocalRaw(name)\n}"
assert text.count(needle)==1
replacement="\n local = len(l.result.Locals)\n l.result.Locals = append(l.result.Locals,ir.Local{Name:name.Text(),Type:l.latentPlaceholder(name).Type(),Function:l.functionIndex})\n if l.locals == nil {l.locals=map[*ast.Symbol]int{}}\n l.locals[l.symbol(name)]=local\n return local,nil\n}"
source.write_text(text.replace(needle,replacement))
config=json.loads((overlay/'overlay.json').read_text())
config['Replace']={k:str(folder/Path(v).name) for k,v in config['Replace'].items()}
(folder/'overlay.json').write_text(json.dumps(config))
mutant=folder/'census'
with (out/'build.log').open('w') as log:
 subprocess.run(['go','build','-buildvcs=false','-overlay='+str(folder/'overlay.json'),'-o',str(mutant),'./stage3/census/latent/tool'],cwd=repo,stdout=log,stderr=subprocess.STDOUT,check=True)
for name,exe in [('baseline',binary),('bypass-binding-mutant',mutant)]:
 with (out/(name+'.log')).open('w') as log:
  result=subprocess.run(['python3',str(repo/'stage3/census/speculative/audit_signature.py'),str(exe),str(out/name)],stdout=log,stderr=subprocess.STDOUT)
 text=(out/(name+'.log')).read_text()
 if name=='baseline':assert result.returncode==0,text
 else:assert result.returncode!=0 and 'AssertionError' in text,text
 print(name+': '+('PASS' if name=='baseline' else 'caught by hidden parameter failure census'))
