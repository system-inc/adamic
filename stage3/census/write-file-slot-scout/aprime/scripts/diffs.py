from pathlib import Path
import difflib,json,hashlib
r=Path('/workspace/scratch/write-file-slot-aprime');out=r/'diffs';out.mkdir(exist_ok=True);base=Path('/workspace/scratch/write-file-slot-measure/main');tree=r/'candidate';baseline=r/'baseline-lane/adapted-tree'
def diff(a,b,output):
 lines=''.join(difflib.unified_diff(a.read_text().splitlines(True),b.read_text().splitlines(True),fromfile=str(a),tofile=str(b)));output.write_text(lines);return len(lines.encode())
for p in ['src/compiler/sys.ts','src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts']:
 diff(base/p,tree/p,out/(Path(p).name+'.source.patch'))
 for mode in ['true','false']:diff(r/'emits'/('main-'+Path(p).name+'.comments-'+mode+'.js'),r/'emits'/('candidate-'+Path(p).name+'.comments-'+mode+'.js'),out/(Path(p).name+'.comments-'+mode+'.js.diff'))
main=Path('/workspace/scratch/write-file-slot-measure/declarations/main/typescript.d.ts');candidate=r/'declarations/candidate/typescript.d.ts';public=diff(main,candidate,out/'public-vs-main.diff');assert public==0
upstream=Path('/workspace/scratch/write-file-slot-typescript/tests/baselines/reference/api/typescript.d.ts');diff(upstream,candidate,out/'public-vs-upstream.diff')
internal=diff(baseline/'built/local/typescript.internal.d.ts',tree/'built/local/typescript.internal.d.ts',out/'internal-declarations.diff')
for name in ['typescript.js','run.js','_tsc.js','tsc.js']:
 a=baseline/'built/local'/name;b=tree/'built/local'/name
 if a.exists() and b.exists():diff(a,b,out/(name+'.bundle.diff'))
print('public delta',public,'internal diff bytes',internal);print((out/'internal-declarations.diff').read_text()[:3000]);
