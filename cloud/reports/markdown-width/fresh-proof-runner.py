from pathlib import Path
import json, os, subprocess, tempfile
root=Path(tempfile.mkdtemp(prefix='markdown-fresh-', dir='/workspace'))
repository=root/'repository'
def run(name, args, env, cwd=None):
    with (root/(name+'.log')).open('wb') as output:
        result=subprocess.run(args, cwd=cwd, env=env, stdout=output, stderr=subprocess.STDOUT)
    print(name, 'exit=', result.returncode, flush=True)
    if result.returncode: raise SystemExit(result.returncode)
base=dict(os.environ)
run('clone', ['git','clone','--shared','--no-checkout','/workspace/adamic',str(repository)],base)
run('checkout',['git','checkout','87a9dbe'],base,repository)
env=dict(base, ADAMIC_TOOLS=str(root/'tools'), GOPATH=str(root/'gopath'), GOCACHE=str(root/'gocache'), PATH='/usr/local/bin:/usr/bin:/bin', GOTOOLCHAIN='auto')
env.pop('ADAMIC_MARKDOWNWIDTH_DEPS',None)
env.pop('ADAMIC_GATE_UNCACHED',None)
(root/'proof.json').write_text(json.dumps(dict(root=str(root), commit=subprocess.check_output(['git','rev-parse','HEAD'],cwd=repository).decode().strip(),initial_tools_present=(root/'tools').exists(),initial_gopath_present=(root/'gopath').exists(),initial_gocache_present=(root/'gocache').exists()),indent=2))
print('fresh proof:',root,flush=True)
run('setup',['bash','cloud/setup.sh'],env,repository)
# Source exactly the env.sh emitted by this isolated setup.
command='source "'+str(root/'tools/env.sh')+'"\nADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/markdownblocks -run "^TestMarkdownUnicodeWidths$" -count=1 -timeout 30m -v'
run('width',['bash','-c',command],env,repository)
