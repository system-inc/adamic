from pathlib import Path
import subprocess,json,concurrent.futures
root = Path(__file__).resolve().parents[2]
scratch = Path('/tmp/adamic-error-coverage')
scratch.mkdir(exist_ok=True)
dest=root/'notes/error-classes/unsupported'
files=sorted((root/'internal/oracle/testdata').glob('coverage_error_*.a'))+sorted(dest.glob('*.a'))+sorted((root/'notes/error-classes').glob('*.a'))
def execute(path):
    name=path.stem; out=scratch/name; result={'program':str(path.relative_to(root))}
    commands=[('node',['node','--disable-warning=ExperimentalWarning','oracle/node.mjs',str(path)]),('build',['go','run','./cmd/adamic','build',str(path.relative_to(root)),'-o',str(out)]),('js_build',['go','run','./cmd/adamic','js',str(path.relative_to(root))])]
    for label,cmd in commands:
        p=subprocess.run(cmd,cwd=root,capture_output=True,text=True,timeout=120)
        result[label]={'exit':p.returncode,'stdout':p.stdout,'stderr':p.stderr}
        (scratch/f'{name}.{label}.stdout').write_text(p.stdout);(scratch/f'{name}.{label}.stderr').write_text(p.stderr)
    if result['build']['exit']==0:
        p=subprocess.run([str(out)],capture_output=True,text=True,timeout=60)
        result['native']={'exit':p.returncode,'stdout':p.stdout,'stderr':p.stderr}
    if result['js_build']['exit']==0:
        js=scratch/f'{name}.mjs';js.write_text(result['js_build']['stdout'])
        p=subprocess.run(['node','--disable-warning=ExperimentalWarning',str(root/'oracle/node.mjs'),str(js)],capture_output=True,text=True,timeout=60)
        result['backend']={'exit':p.returncode,'stdout':p.stdout,'stderr':p.stderr}
    (scratch/f'{name}.json').write_text(json.dumps(result,indent=2))
    print(name, 'build',result['build']['exit'], 'native agrees',result.get('native')==result['node'], 'JS agrees',result.get('backend')==result['node'],flush=True)
    if path.parent.name == 'testdata':
        assert result['build']['exit'] == 0 and result['native'] == result['backend'] == result['node'], name
    elif path.parent.name == 'unsupported':
        assert result['build']['exit'] != 0, name
    else:
        assert result['build']['exit'] == 0 and result['native'] != result['node'], name
with concurrent.futures.ThreadPoolExecutor(max_workers=5) as pool: list(pool.map(execute,files))
