import hashlib,json,pathlib,subprocess
s=pathlib.Path('/workspace/scratch/path-mint-618'); w=s/'work'
def git(*a,cwd=w):return subprocess.check_output(['git',*a],cwd=cwd,text=True).strip()
def hashfile(p):return hashlib.sha256(p.read_bytes()).hexdigest()
p={'base':'e531d51d66c45368afa7a9598bc6baa8602b4b41','scout':'13b63e39bb0b304c83d4ee78ed5dd0ed9f8e3e1e','typescript':git('rev-parse','HEAD',cwd=s/'TypeScript'),'cohere':git('rev-parse','HEAD',cwd=w/'cohere'),'checker':git('rev-parse','HEAD',cwd=w/'cohere/TypeScript'),'node':subprocess.check_output(['node','--version'],text=True).strip(),'go':subprocess.check_output([str(s/'go/bin/go'),'version'],text=True).strip(),'files':{}}
for name in ['census','adamic']:p['files'][name]=hashfile(s/name)
for tree in ['base','lane/adapted-tree']:
 t=s/tree;p['files'][tree]={str(f.relative_to(t)):hashfile(f) for f in sorted((t/'src').rglob('*')) if f.is_file()}
(s/'evidence/provenance.json').write_text(json.dumps(p,indent=2)+'\n')
