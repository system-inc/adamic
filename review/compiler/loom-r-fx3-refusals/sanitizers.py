from pathlib import Path
import subprocess,json,time,os
out=Path('review/compiler/loom-r-fx3-refusals')
exe=(out/'adamic').resolve(); results=[]
fixtures=[Path('internal/lower/testdata/array_narrowing')/(name+'.a') for name in ['optional_number','optional_string','some_negated','find_last_index','preserved_union','boolean_callback']]
mixed=out/'mixed-literal.a'; mixed.write_text("const items: (string | number)[] = [1, 'a'];\nconsole.log(`${typeof items[0]} ${items[0]}`);\nconsole.log(`${typeof items[1]} ${items[1]}`);\n")
fixtures.append(mixed)
for fixture in fixtures:
    began=time.monotonic(); name=fixture.stem
    expected=subprocess.run(['node','--disable-warning=ExperimentalWarning','oracle/node.mjs',str(fixture.resolve())],capture_output=True,timeout=15)
    (out/(name+'-node.log')).write_bytes(expected.stdout+expected.stderr)
    assert expected.returncode==0 and not expected.stderr
    js=out/(name+'.mjs')
    with js.open('wb') as f, (out/(name+'-js-build.log')).open('wb') as log:
        assert subprocess.run([str(exe),'js',str(fixture)],stdout=f,stderr=log,timeout=90).returncode==0
    gotjs=subprocess.run(['node','--disable-warning=ExperimentalWarning','oracle/node.mjs',str(js.resolve())],capture_output=True,timeout=15)
    (out/(name+'-js.log')).write_bytes(gotjs.stdout+gotjs.stderr)
    assert (gotjs.returncode,gotjs.stdout,gotjs.stderr)==(expected.returncode,expected.stdout,expected.stderr)
    binary=Path('/tmp/adamic-gate')/('fx3-'+name)
    command=[str(exe),'build',str(fixture),'-o',str(binary),'--sanitize']
    with (out/(name+'-sanitize-build.log')).open('wb') as log:
        assert subprocess.run(command,stdout=log,stderr=subprocess.STDOUT,timeout=90).returncode==0
    got=subprocess.run([str(binary)],capture_output=True,timeout=15,env={**os.environ,'ASAN_OPTIONS':'detect_leaks=1:halt_on_error=1','UBSAN_OPTIONS':'halt_on_error=1'})
    (out/(name+'-sanitize.log')).write_bytes(got.stdout+got.stderr)
    assert (got.returncode,got.stdout,got.stderr)==(expected.returncode,expected.stdout,expected.stderr),(name,got)
    results.append(dict(fixture=str(fixture),build=command,stdout=got.stdout.decode(),exit=got.returncode,seconds=round(time.monotonic()-began,3)))
    (out/'sanitizers.json').write_text(json.dumps(results,indent=2)+'\n');print(name+': source Node, generated JS, ASan/UBSan/LSan agree',flush=True)
