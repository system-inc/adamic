"""Bounded parser probes with Go-positive fixtures, Node/native refusals and a clean control."""
import json, os, pathlib, subprocess, sys
here = pathlib.Path(__file__).resolve().parent
root = here.parents[6]
work = pathlib.Path(sys.argv[1]).resolve()
corpus = json.loads((work/'corpus.json').read_text())
binary = work/'parser-probe'
with (work/'parser-build.log').open('w') as log:
    subprocess.run(['go','run',str(here/'build.go'),str(here/'parser.ts'),str(binary)],cwd=root,stdout=log,stderr=log,check=True)
sides = {'Node':['node','--disable-warning=ExperimentalWarning',str(root/'oracle/node.mjs'),str(here/'parser.ts')], 'native':[str(binary)]}
env = dict(os.environ,ASAN_OPTIONS='detect_leaks=1',UBSAN_OPTIONS='halt_on_error=1')
control = work/'control.ts'
control.write_text("const image = 'img';\n")
for side, command in sides.items():
    with (work/(side+'-control.stdout')).open('w') as out, (work/(side+'-control.stderr')).open('w') as err:
        run = subprocess.run(command+[str(control)],cwd=root,env=env,stdout=out,stderr=err,timeout=10)
    assert run.returncode==0, side+' control failed'
    assert (work/(side+'-control.stdout')).read_text().strip().isdigit()
    assert not (work/(side+'-control.stderr')).read_text()
blocked = ['no-head-element','no-html-link-for-pages','no-img-element','no-page-custom-font','no-styled-jsx-in-document','no-sync-scripts','no-unwanted-polyfillio']
ledger=[]
for name in blocked:
    positive = [r for r in corpus if r['rule']=='@next/next/'+name and r.get('findings')]
    assert positive, name+' lacks Go positive'
    case = min(positive,key=lambda r:len(r['source']))
    path = work/(name+'.tsx')
    path.write_text(case['source'])
    observed={}
    for side, command in sides.items():
        with (work/(name+'-'+side+'.stdout')).open('w') as out, (work/(name+'-'+side+'.stderr')).open('w') as err:
            run=subprocess.run(command+[str(path)],cwd=root,env=env,stdout=out,stderr=err,timeout=10)
        stdout=(work/(name+'-'+side+'.stdout')).read_text()
        stderr=(work/(name+'-'+side+'.stderr')).read_text()
        assert run.returncode==70, (name,side,run.returncode,stderr)
        assert not stdout
        assert 'parser slice' in stderr or 'expected' in stderr, stderr
        assert 'AddressSanitizer' not in stderr and 'runtime error:' not in stderr
        observed[side]={'exit':run.returncode,'stderr':stderr}
    assert observed['Node']==observed['native'], name+' refusals differ'
    ledger.append({'rule':case['rule'],'file':case['file'],'source':case['source'],'goFindings':case['findings'],'otherFiles':case.get('otherFiles',0),'observations':observed})
    print(case['rule'],len(case['findings']),'Go findings; both exit 70:',observed['Node']['stderr'].strip())
(work/'parser-gaps.json').write_text(json.dumps(ledger,indent=2)+'\n')
print('clean control passes on both backends; seven positive JSX sources refused explicitly')
