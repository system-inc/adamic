from pathlib import Path
import json, os, subprocess, hashlib
root=Path('/workspace/markdown-fresh-zubr_3s9');repository=root/'repository'
env=dict(os.environ, ADAMIC_TOOLS=str(root/'tools'), GOPATH=str(root/'gopath'), GOCACHE=str(root/'gocache'), PATH='/usr/local/bin:/usr/bin:/bin', GOTOOLCHAIN='auto')
env.pop('ADAMIC_MARKDOWNWIDTH_DEPS',None);env.pop('ADAMIC_GATE_UNCACHED',None)
def run(name,args):
    with (root/(name+'.log')).open('wb') as output:
        result=subprocess.run(args,cwd=repository,env=env,stdout=output,stderr=subprocess.STDOUT)
    print(name,'exit=',result.returncode,flush=True)
    if result.returncode:raise SystemExit(result.returncode)
run('fetch-final',['git','fetch','/workspace/adamic','devtools/setup-fast'])
run('checkout-final',['git','checkout','FETCH_HEAD'])
# Retain the old successful installation and begin with no installed npm dependencies.
previous=root/'tools/markdown-width-before-final'
(root/'tools/markdown-width').rename(previous)
run('setup-final',['bash','cloud/setup.sh'])
def tree(path):
    return {str(file.relative_to(path)):hashlib.sha256(file.read_bytes()).hexdigest() for file in path.rglob('*') if file.is_file() and file.name!='.adamic-stamp'}
assert tree(previous)==tree(root/'tools/markdown-width'),'installed bytes changed'
(root/'final-proof.json').write_text(json.dumps(dict(commit=subprocess.check_output(['git','rev-parse','HEAD'],cwd=repository).decode().strip(),final_dependency_install_initially_absent=True,old_and_final_installed_bytes_identical=True,command=['bash','cloud/setup.sh']),indent=2)+'\n')
command='source "'+str(root/'tools/env.sh')+'"\nADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/markdownblocks -run "^TestMarkdownUnicodeWidths$" -count=1 -timeout 30m -v'
run('width-final',['bash','-c',command])
