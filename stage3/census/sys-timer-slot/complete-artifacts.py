"""Finish emission evidence after typed builds fail, without claiming a clean build."""
import sys,pathlib,json,hashlib,shutil,os
sys.dont_write_bytecode=True
sys.path.insert(0,str(pathlib.Path(__file__).parent));from run import run,scratch,repo,evidence
baseline=json.loads((scratch/'main-compiler-js.json').read_text());assert len(baseline)==9
for label in ['A','B']:
    shutil.copy2(scratch/(label+'-compiler-js.json'),scratch/(label+'-partial-compiler-js.json'))
    # The normal failed build emitted the other eight JS bundles. Only the lssl
    # shim task remains; invoke only the existing upstream writer callback.
    assert run(label+'-finish-shim',['node',str(evidence/'finish-shim.mjs'),str(scratch/label)],scratch/label)==0
    assert run(label+'-bundle-declarations',['node','scripts/dtsBundler.mjs','--entrypoint','./built/local/typescript/typescript.d.ts','--output','./built/local/typescript.d.ts'],scratch/label)==0
    hashes={name:hashlib.sha256((scratch/label/'built/local'/name).read_bytes()).hexdigest() for name in baseline}
    (scratch/(label+'-compiler-js.json')).write_text(json.dumps(hashes,indent=2)+'\n')
for label in ['main','A','B']:
    assert run(label+'-public-timer-api',['node',str(evidence/'artifact-api.cjs'),str(scratch/label),str(scratch/(label+'-public-timers.json'))])==0
