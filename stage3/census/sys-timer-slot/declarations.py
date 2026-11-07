"""Compare built artifacts and draft corresponding reference baseline edits."""
import pathlib, json, difflib, hashlib
scratch=pathlib.Path('/workspace/scratch/sys-timer-slot');out=pathlib.Path('/workspace/adamic/stage3/census/sys-timer-slot'); data={}
def diff(a,b,old,new):return ''.join(difflib.unified_diff(a.splitlines(True),b.splitlines(True),fromfile=old,tofile=new))
for label in ['A','B']:
    row={}; chunks=[]
    for name in ['typescript.d.ts','typescript.internal.d.ts']:
        a=scratch/'main/built/local'/name;b=scratch/label/'built/local'/name
        if not a.exists() or not b.exists():row[name]={'missing':True};continue
        aa=a.read_text();bb=b.read_text();d=diff(aa,bb,'a/built/local/'+name,'b/built/local/'+name)
        row[name]={'identical':aa==bb,'diff':d,'mainSha256':hashlib.sha256(a.read_bytes()).hexdigest(),'candidateSha256':hashlib.sha256(b.read_bytes()).hexdigest()};chunks.append(d)
    source=(scratch/label/'src/compiler/sys.ts').read_text()
    newSet=next(l.strip() for l in source.splitlines() if l.strip().startswith('setTimeout?<') or l.strip().startswith('setTimeout?(callback:'))
    newClear='clearTimeout?(timeoutId: unknown): void;'
    ref=scratch/'main/tests/baselines/reference/api/typescript.d.ts';aa=ref.read_text()
    oldSet='setTimeout?(callback: (...args: any[]) => void, ms: number, ...args: any[]): any;';oldClear='clearTimeout?(timeoutId: any): void;'
    assert aa.count(oldSet)==aa.count(oldClear)==1
    bb=aa.replace(oldSet,newSet).replace(oldClear,newClear)
    d=diff(aa,bb,'a/tests/baselines/reference/api/typescript.d.ts','b/tests/baselines/reference/api/typescript.d.ts')
    row['referenceBaseline']={'draft':True,'diff':d};chunks.append(d)
    (out/(label+'-declarations.diff.txt')).write_text('\n'.join(chunks))
    ja=json.loads((scratch/'main-compiler-js.json').read_text());jb=json.loads((scratch/(label+'-compiler-js.json')).read_text())
    row['javascript']={'mainCount':len(ja),'candidateCount':len(jb),'identical':ja==jb,'main':ja,'candidate':jb}
    data[label]=row
(out/'declarations.json').write_text(json.dumps(data,indent=2)+'\n')
