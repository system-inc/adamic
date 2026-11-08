import json, pathlib, re, subprocess
p=pathlib.Path('evidence'); source=p/'admission-delta-final.json'
r=json.loads(source.read_text())
rows=[x for x in r.get('programs',[]) if x['class']=='newly-refused']
lines=['# New refusals','','Comparison: `ff41896cee93f832a4a8fa68c4c1822ddef5885e` to `550acd97678cf5433a29649ded1110b6a92b4c08`.',f"\nCount: **{len(rows)}**. Admission tool verdict: **{r.get('verdict')}**. Complete corpus: `{r.get('complete_corpus')}`.",'','Flags identify override diagnostics originating in `internal/lower/class_inheritance.go`, and constrained generic method views that require inspection as potentially honest programs. The complete diagnostic is retained for each refusal.','']
for x in rows:
    msg=x['compile'].get('stderr','')
    code=subprocess.run(['git','show','550acd97678cf5433a29649ded1110b6a92b4c08:'+x['path']],capture_output=True,text=True).stdout
    flags=[]
    if 'override' in msg:
        flags.append('OVERRIDE REFUSAL (class_inheritance.go)'+(' — source has a generic method' if re.search(r'\w+\s*<[^>]+>\s*\(',code) else ''))
    if re.search(r'handle\s*<[^>]*extends\s+string',code):
        flags.append('CONSTRAINED HANDLE VIEW — inspect as honest handle<T extends string>(x: Animal) seen as handle<T>(x: Animal)')
    lines += ['## '+x['path'],'']
    if flags: lines += ['; '.join(flags),'']
    lines += ['```text',msg.rstrip(),'```','']
(p/'new-refusals.md').write_text('\n'.join(lines)+'\n')
print(json.dumps({'count':len(rows),'verdict':r.get('verdict'),'complete_corpus':r.get('complete_corpus'),'programs':len(r.get('programs',[]))}))
