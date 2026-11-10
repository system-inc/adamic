from control import *
name=sys.argv[1]
tree=root/name
if name=='parent':
 dep=tree/'cohere'
 if dep.is_dir() and not dep.is_symlink():
  dep.rmdir(); dep.symlink_to(root/'head/cohere',target_is_directory=True)
# Use canonical dependency paths for Go's build cache.
mod=json.loads(subprocess.check_output(['go','mod','edit','-json'],cwd=tree,text=True))
work=root/(name+'.go.work')
repls=[r for r in mod['Replace'] if r['New']['Path'].startswith('./cohere/') and r['Old']['Path']!='github.com/microsoft/TypeScript/tsc']
work.write_text('go 1.27\nuse (\n'+str(tree)+'\n'+str(root/'head/cohere/TypeScript/tsc')+'\n)\nreplace (\n'+''.join(r['Old']['Path']+' => '+str(root/'head'/r['New']['Path'])+'\n' for r in repls)+')\n')
extra={'GOWORK':str(work)}
steps=[('overlay',['python3','stage3/census/latent/make_overlay.py',tree,root/(name+'-overlay')]),('entry-overlay',['python3','stage3/meter/entry_overlay.py',root/(name+'-overlay'),root/(name+'-entry-overlay')]),('build',['go','build','-buildvcs=false','-overlay='+str(root/(name+'-entry-overlay/overlay.json')),'-o',root/(name+'-census'),'./stage3/census/latent/tool'])]
for phase,args in steps:
 code=run(name+'-'+phase,args,tree,extra)
 if code: sys.exit(code)
