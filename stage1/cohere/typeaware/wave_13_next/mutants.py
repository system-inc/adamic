"""Run compiling semantic mutants, accepting only normal exit and byte-only catches."""
import json, pathlib, re, subprocess, sys
repo, scratch, controls, baseline = map(lambda s:pathlib.Path(s).resolve(),sys.argv[1:5])
source = repo/'stage1/cohere/typeaware/wave_13_next'
stage0, archive = map(pathlib.Path,sys.argv[5:7])
scratch.mkdir(parents=True,exist_ok=True)
changes = [
 ('process-state','no_process_exit_after_output.a','if(state.chains.length > 0)','if(state.chains.length === 0)'),
 ('race-handle','no_uncleared_race_timeout.a','this.timer(index) && this.lost(executor, index)','this.timer(index) && !this.lost(executor, index)'),
 ('blocking-order','require_blocking_standard_streams.a','const ordered = this.index.canBlock.has(c.rules().path);','const ordered = false;'),
]
results=[]
for name,file,before,after in changes:
    directory=scratch/name;directory.mkdir(exist_ok=True)
    for original in source.glob('*.a'):
        text=original.read_text()
        if original.name==file:
            assert text.count(before)==1,(name,text.count(before))
            text=text.replace(before,after)
        def absolute(match):
            path=(original.parent/match.group(1)).resolve()
            return "'"+str(path)+"'"
        text=re.sub(r"'([^']+\.ts)'",absolute,text)
        (directory/original.name).write_text(text)
    binary=directory/'native'
    with (directory/'build.log').open('wb') as log:
        result=subprocess.run([str(stage0),'build',str(directory/'suite.a'),'-o',str(binary),'--tsgo',str(archive)],stdout=log,stderr=subprocess.STDOUT)
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
