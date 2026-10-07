#!/usr/bin/env python3
from pathlib import Path
import subprocess,json,re
own=Path(__file__).resolve().parent;r=own.parents[4];out=Path('/workspace/wave04-context/unicode');out.mkdir(exist_ok=True);compiler=Path('/workspace/typeaware-wave-04-landing-d65/adamic')
def run(label,args,expected=0):
 with (out/(label+'.stdout')).open('wb') as so,(out/(label+'.stderr')).open('wb') as se:code=subprocess.run(list(map(str,args)),cwd=r,stdout=so,stderr=se).returncode
 assert code==expected,(label,code);return (out/(label+'.stdout')).read_bytes(),(out/(label+'.stderr')).read_bytes()
truth,_=run('go',['go','run',own/'unicode_oracle.go',out/'inputs.json']);values=json.loads((out/'inputs.json').read_text())
driver=f"import {{ upper }} from '{own}/unicode_regex.a';\nimport {{ quote }} from '{own}/quote.a';\nimport {{ written }} from '{r}/stage1/typescript/parser/nodes.ts';\n"
driver+='const values: readonly string[] = '+json.dumps(values,ensure_ascii=True)+';\nfor(const value of values){ console.log(`${upper(value)}\\t${written(quote(value))}`); }\n'
entry=out/'probe.a';entry.write_text(driver)
for variant in ['normal','sanitize']:
 cmd=[compiler,'build',entry,'-o',out/variant]
 if variant=='sanitize':cmd.append('--sanitize')
 run(variant+'-build',cmd);actual,error=run(variant+'-run',[out/variant]);assert actual==truth and not error,variant
actual,error=run('source-node',['node','--disable-warning=ExperimentalWarning',r/'oracle/node.mjs',entry]);assert actual==truth and not error
js,_=run('emit',[compiler,'js',entry]);(out/'probe.js').write_bytes(js);actual,error=run('emitted-js',['node','--disable-warning=ExperimentalWarning',r/'oracle/node.mjs',out/'probe.js']);assert actual==truth and not error
for label,module,before,after in [('uppercase','unicode_regex.a',r'\u{41}-\u{5a}',r'\u{42}-\u{5a}'),('quote','quote.a','code === 127','code === 128')]:
 path=own/module;text=path.read_text();assert text.count(before)==1;text=text.replace(before,after);text=re.sub(r"from '(\.[^']+)'",lambda m:"from '"+str((path.parent/m[1]).resolve())+"'",text);mutant=out/(label+'-mutant.a');mutant.write_text(text);probe=out/(label+'-probe.a');probe.write_text(driver.replace(str(path),str(mutant)))
 run(label+'-build',[compiler,'build',probe,'-o',out/label,'--sanitize']);actual,error=run(label+'-run',[out/label]);assert actual!=truth and not error
print(f'PASS Unicode/quote: {len(values)} inputs / {len(truth)} bytes, Go/native/sanitizers/source Node/emitted JS, two comparison-only mutants.')
