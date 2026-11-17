import os,sys,pathlib,subprocess,json,time,hashlib
scratch=pathlib.Path(sys.argv[1]).resolve()
label=sys.argv[2];source=scratch/label;dest=scratch/(label+'-tools');dest.mkdir(exist_ok=True)
rows=[]
def run(name,args):
 start=time.time()
 with (dest/(name+'.log')).open('w') as log:
  p=subprocess.run(args,cwd=source,stdout=log,stderr=subprocess.STDOUT)
 rows.append(dict(name=name,args=args,exit=p.returncode,seconds=time.time()-start))
 (dest/'steps.json').write_text(json.dumps(rows,indent=2)+'\n')
 print(name,p.returncode,round(time.time()-start,2),flush=True)
 if p.returncode:sys.exit(p.returncode)
run('latent-overlay',['python3',str(source/'stage3/census/latent/make_overlay.py'),str(source),str(dest/'latent-overlay')])
run('latent-gofmt',['gofmt','-w',*map(str,(dest/'latent-overlay').glob('*.go'))])
run('entry-overlay',['python3',str(source/'stage3/meter/entry_overlay.py'),str(dest/'latent-overlay'),str(dest/'entry-overlay')])
run('entry-gofmt',['gofmt','-w',*map(str,(dest/'entry-overlay').glob('*.go'))])
run('entry-build',['go','build','-buildvcs=false','-overlay='+str(dest/'entry-overlay/overlay.json'),'-o',str(dest/'entry-census'),'./stage3/census/latent/tool'])
run('latent-build',['go','build','-buildvcs=false','-overlay='+str(dest/'latent-overlay/overlay.json'),'-o',str(dest/'latent-census'),'./stage3/census/latent/tool'])
run('scan-overlay',['python3',str(source/'stage3/census/scan-all/make_overlay.py'),str(dest/'scan-overlay')])
run('scan-build',['go','build','-trimpath','-ldflags=-buildid=','-buildvcs=false','-overlay='+str(dest/'scan-overlay/overlay.json'),'-o',str(dest/'scan-all'),'./stage3/census/scan-all/tool/main.go'])
meta={'commit':subprocess.check_output(['git','rev-parse','HEAD'],cwd=source,text=True).strip(),'submodules':subprocess.check_output(['git','submodule','status','--recursive'],cwd=source,text=True),'go':subprocess.check_output(['go','version'],text=True).strip(),'binaries':{n:hashlib.sha256((dest/n).read_bytes()).hexdigest() for n in ['entry-census','latent-census','scan-all']}}
(dest/'provenance.json').write_text(json.dumps(meta,indent=2)+'\n')
