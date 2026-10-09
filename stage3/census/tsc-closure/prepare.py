#!/usr/bin/env python3
"""Build pinned measurement hooks against untouched compiler main in scratch."""
import argparse,io,json,os,subprocess,tarfile
from pathlib import Path
p=argparse.ArgumentParser();p.add_argument('repository',type=Path);p.add_argument('output',type=Path);a=p.parse_args();repo=a.repository.resolve();out=a.output.resolve();out.mkdir();tree=out/'tree';pin='f1502d130bc0b4440f29b93b76b7d18bff3f6a60'
def run(cmd,cwd,name,env=None):
 with (out/(name+'.log')).open('wb') as log:subprocess.run(cmd,cwd=cwd,stdout=log,stderr=subprocess.STDOUT,env=env,check=True)
run(['git','worktree','add','--detach',str(tree),'HEAD'],repo,'worktree')
archive=subprocess.check_output(['git','archive',pin,'stage3/census/latent','stage3/census/hidden'],cwd=repo)
with tarfile.open(fileobj=io.BytesIO(archive)) as tar:tar.extractall(tree,filter='data')
(tree/'cohere').rmdir();(tree/'cohere').symlink_to(repo/'cohere',target_is_directory=True)
module=json.loads(subprocess.check_output(['go','mod','edit','-json'],cwd=tree,text=True,env={**os.environ,'GOWORK':'off'}));reps=[r for r in module['Replace'] if r['New']['Path'].startswith('./cohere') and r['Old']['Path']!='github.com/microsoft/TypeScript/tsc']
work=out/'go.work';work.write_text('go 1.27\nuse (\n'+str(tree)+'\n'+str(repo/'cohere/TypeScript/tsc')+'\n)\nreplace (\n'+''.join(r['Old']['Path']+' => '+str(repo/r['New']['Path'])+'\n' for r in reps)+')\n');env={**os.environ,'GOWORK':str(work)}
run(['python3',str(tree/'stage3/census/latent/make_overlay.py'),str(tree),str(out/'overlay')],tree,'overlay',env)
main=out/'overlay/stage3_census_latent_tool_main.go';s=main.read_text();needle='\tsort.Strings(paths)';assert s.count(needle)==1
s=s.replace(needle,'''\tif name := os.Getenv("LATENT_ROOT_MANIFEST"); name != "" {
\t\tdata, err := os.ReadFile(name); if err != nil { panic(err) }
\t\tvar manifest struct { Root string `json:"root"`; Files []struct { File string `json:"file"` } `json:"files"` }
\t\tif err := json.Unmarshal(data, &manifest); err != nil { panic(err) }
\t\tpaths = nil
\t\tfor _, file := range manifest.Files { paths = append(paths, filepath.Join(manifest.Root, file.File)) }
\t\tif len(paths) == 0 { panic("empty closure roots") }
\t}
'''+needle).replace('"status": "measurement",','"status": "measurement", "root_files": paths,');main.write_text(s)
run(['gofmt','-w',*[str(p) for p in (out/'overlay').glob('*.go')]],tree,'gofmt',env)
run(['go','build','-buildvcs=false','-overlay='+str(out/'overlay/overlay.json'),'-o',str(out/'census'),'./stage3/census/latent/tool'],tree,'build',env)
print(out/'census')
