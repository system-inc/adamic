"""Run compiling semantic mutants, accepting only normal exit and byte-only catches."""
import json, pathlib, re, subprocess, sys
repo, scratch, controls, baseline = map(lambda s:pathlib.Path(s).resolve(),sys.argv[1:5])
source = repo/'stage1/cohere/typeaware/wave_13_next'
stage0, archive = map(pathlib.Path,sys.argv[5:7])
scratch.mkdir(parents=True,exist_ok=True)
results=[]
changes=[
 ('platform-declaration','platform_symbol.go','out.yes(source.IsDeclarationFile)','out.yes(false)'),
 ('resolved-return','call_declaration.go','flags = uint64(t.Flags())','flags = uint64(262144)'),
 ('graph-reachability','output_flow.go','out.yes(block.Reachable)','out.yes(!block.Reachable)'),
 ('module-resolution','program_modules.go','target = file.FileName()','target = ""'),
]
for name,file,before,after in changes:
    directory=scratch/name;directory.mkdir(exist_ok=True)
    original=repo/'bridge/tsgo/checker'/file
    text=original.read_text();assert text.count(before)==1,name
    changed=directory/file;changed.write_text(text.replace(before,after))
    overlay=directory/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(original):str(changed)}}))
    mutant_archive=directory/'checker.a'
    with (directory/'archive.log').open('wb') as log:
        result=subprocess.run(['go','build','-overlay',str(overlay),'-buildmode=c-archive','-o',str(mutant_archive),'./bridge/tsgo/archive'],cwd=repo,stdout=log,stderr=subprocess.STDOUT)
    assert result.returncode==0,name+' archive failed'
    binary=directory/'native'
    with (directory/'build.log').open('wb') as log:
        result=subprocess.run([str(stage0),'build',str(source/'suite.a'),'-o',str(binary),'--tsgo',str(mutant_archive)],stdout=log,stderr=subprocess.STDOUT)
    assert result.returncode==0,name+' failed compilation'
    caught=None
    for case in sorted(controls.glob('*/case.json')):
        out,err=directory/(case.parent.name+'.stdout'),directory/(case.parent.name+'.stderr')
        with out.open('wb') as stdout,err.open('wb') as stderr:
            result=subprocess.run([str(binary),str(case.parent/'tsconfig.json'),str(case.parent/'roots.manifest')],stdout=stdout,stderr=stderr)
        assert result.returncode==0 and not err.read_bytes(),(name,result.returncode,err.read_text())
        wanted=(baseline/(case.parent.name+'-go.stdout')).read_bytes();actual=out.read_bytes()
        if wanted!=actual:
            difference=next((at for at,(a,b) in enumerate(zip(wanted,actual)) if a!=b),min(len(wanted),len(actual)))
            caught=dict(mutant=name,test=json.loads(case.read_text())['test'],case=case.parent.name,exit=0,stderr_bytes=0,first_difference=difference)
            break
    assert caught is not None,name+' survived'
    results.append(caught);print(json.dumps(caught),flush=True)
(scratch/'results.json').write_text(json.dumps(results,indent=2)+'\n')
