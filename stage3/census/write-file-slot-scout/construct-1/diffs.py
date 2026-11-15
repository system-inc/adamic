from pathlib import Path
import difflib,json,hashlib
r=Path('/workspace/scratch/write-file-construct-run');out=r/'diffs';out.mkdir(exist_ok=True);base=r/'control-adapted';tree=r/'lane/adapted-tree';receipt={}
def diff(a,b,output):
 lines=''.join(difflib.unified_diff(a.read_text().splitlines(True),b.read_text().splitlines(True),fromfile=str(a),tofile=str(b)));output.write_text(lines);return len(lines.encode())
for p in ['src/compiler/sys.ts','src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts']:
 diff(base/p,tree/p,out/(Path(p).name+'.source.patch'))
 for mode in ['true','false']:diff(r/'emits'/('main-'+Path(p).name+'.comments-'+mode+'.js'),r/'emits'/('candidate-'+Path(p).name+'.comments-'+mode+'.js'),out/(Path(p).name+'.comments-'+mode+'.js.diff'))
candidate=r/'declarations/candidate/typescript.d.ts'
reference=r/'declarations/control/typescript.d.ts';delta=diff(reference,candidate,out/'public-vs-control.diff');assert delta==0 and candidate.read_bytes()==reference.read_bytes();receipt['control']={'byteIdentical':True,'diffBytes':delta,'sha256':hashlib.sha256(reference.read_bytes()).hexdigest()}
upstream=Path('/workspace/scratch/write-file-slot-typescript/tests/baselines/reference/api/typescript.d.ts');stockDelta=diff(upstream,candidate,out/'public-vs-upstream.diff');beforeDelta=diff(upstream,reference,out/'control-public-vs-upstream.diff');assert (out/'public-vs-upstream.diff').read_text().splitlines()[2:]==(out/'control-public-vs-upstream.diff').read_text().splitlines()[2:];receipt['upstream']={'byteIdentical':False,'inheritedDiffUnchanged':True,'newDeltaBytes':0,'controlDiffBytes':beforeDelta,'candidateDiffBytes':stockDelta,'sha256':hashlib.sha256(upstream.read_bytes()).hexdigest()}
for name in ['typescript.internal.d.ts','typescript.js','run.js','_tsc.js','tsc.js']:
 a=base/'built/local'/name;b=tree/'built/local'/name
 if a.exists() and b.exists():receipt[name]={'diffBytes':diff(a,b,out/(name+'.bundle.diff'))}
(r/'diffs.json').write_text(json.dumps(receipt,indent=2)+'\n');print(json.dumps(receipt,indent=2))
