from pathlib import Path
import difflib
r=Path('/workspace/scratch/write-file-slot-measure');out=r/'review';out.mkdir(exist_ok=True)
def diff(a,b,labela,labelb,dest):
 dest.write_text(''.join(difflib.unified_diff(a.read_text().splitlines(True),b.read_text().splitlines(True),fromfile=labela,tofile=labelb)))
for n in ['a-construction','b-slot']:
 for suffix in ['comments-true.js','comments-false.js','sys.d.ts']:
  diff(r/('main.'+suffix),r/(n+'.'+suffix),'main/'+suffix,n+'/'+suffix,out/(n+'.'+suffix+'.diff'))
 diff(r/'main/src/compiler/sys.ts',r/n/'src/compiler/sys.ts','main/src/compiler/sys.ts',n+'/src/compiler/sys.ts',out/(n+'.source.patch'))
 diff(r/'declarations/main/typescript.d.ts',r/'declarations'/n/'typescript.d.ts','main/typescript.d.ts',n+'/typescript.d.ts',out/(n+'.public.d.ts.diff'))
base=Path('/workspace/scratch/write-file-slot-typescript/tests/baselines/reference/api/typescript.d.ts')
for n in ['main','a-construction','b-slot']:
 diff(base,r/'declarations'/n/'typescript.d.ts','tests/baselines/reference/api/typescript.d.ts',n+' bundled/typescript.d.ts',out/(n+'-vs-upstream-api.diff'))
for p in out.glob('*.comments-true.js.diff'): print(p.name,p.read_text())
