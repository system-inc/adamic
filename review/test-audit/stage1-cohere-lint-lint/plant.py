from pathlib import Path
import subprocess,json
R=Path('/workspace/adamic');E=R/'review/test-audit/stage1-cohere-lint-lint';BASE='1f34d0d300301faebc94d397adee1a090923d2c7'
def base(p):return subprocess.check_output(['git','show',BASE+':'+p],cwd=R,text=True)
sources={}
for m in json.loads((E/'witness-plan.json').read_text()):
 p=m['file'];s=sources.get(p,base(p));start=s.index('func '+m['function']+'(');end=s.find('\nfunc ',start+1);end=len(s) if end<0 else end;part=s[start:end];old=m['before']
 if m['id'] in ['W5','W6']:new='if diff := difference(got, want); os.Getenv("ADAMIC_U110_WITNESS") != "'+m['id']+'" && diff != "" {'
 else:new='if os.Getenv("ADAMIC_U110_WITNESS") == "'+m['id']+'" || '+old[3:]
 assert part.count(old)==1
 sources[p]=s[:start]+part.replace(old,new,1)+s[end:]
for m in json.loads((E/'production-plan.json').read_text()):
 p=m['file'];s=sources.get(p,base(p));old=m['before']
 if m['id']=='M1':new="        if(auditMutant !== 'M1') rules.visit(index, parent);"
 elif m['id']=='M2':new="    return auditMutant === 'M2' ? right.start - left.start : left.start - right.start;"
 elif m['id']=='M3':new="        if(auditMutant === 'M3' ? !this.junkRows : this.junkRows) {"
 else:new=old+"\n    if(auditMutant === 'E1') return 0;"
 assert s.count(old)==1
 sources[p]=s.replace(old,new,1)
for p in ['stage1/cohere/lint/main.ts','stage1/cohere/lint/lint.ts']:sources[p]="import { auditMutant } from './audit_u110.ts';\n"+sources[p]
for p,s in sources.items():(R/p).write_text(s)
(R/'stage1/cohere/lint/audit_u110.ts').write_text("import { readTextFile } from 'adamic';\nconst selected = readTextFile('/tmp/adamic-u110-mutant');\nexport const auditMutant = selected.kind === 'Ok' ? selected.text : '';\n")
Path('/tmp/adamic-u110-mutant').write_text('')
subprocess.run(['gofmt','-w',*[p for p in sources if p.endswith('.go')]],cwd=R,check=True)
