import pathlib,json,subprocess,time
repo=pathlib.Path('/workspace/adamic');out=pathlib.Path(__file__).parent;base=pathlib.Path('/workspace/scratch');rows=[]
def run(args):
 with (out/'census2-build-current.log').open('a') as log:child=subprocess.run(args,cwd=repo,stdout=log,stderr=subprocess.STDOUT)
 rows.append(dict(argv=args,exit=child.returncode));(out/'census2-build-commands.json').write_text(json.dumps(rows,indent=2)+'\n');print('census build child exit='+str(child.returncode),flush=True)
 if child.returncode:raise SystemExit(child.returncode)
run(['python3',str(out/'make_loader_overlay.py'),str(base/'tracing2-census-loader')]);run(['python3','stage3/census/latent/make_overlay.py',str(repo),str(base/'tracing2-census-lower')])
p=base/'tracing2-census-lower/internal_load_load.go';s=p.read_text();original=(repo/'internal/load/load.go').read_text();a=original.index('func compilerOptions() *core.CompilerOptions {');b=original.index('\n// Load checks',a);s=s.replace(original[a:b],'func compilerOptions() *core.CompilerOptions { return censusOptions() }\n');s=s.replace('roots = append(roots, preludePath)','ownedRoots := append([]tspath.RootedFilePath{}, roots...)\n roots = censusRoots()').replace('roots[:len(paths)]','ownedRoots');p.write_text(s)
d=json.loads((base/'tracing2-census-loader/overlay.json').read_text());d['Replace'].update(json.loads((base/'tracing2-census-lower/overlay.json').read_text())['Replace']);d['Replace'][str(repo/'internal/load/census_config_hook.go')]=str(out/'config_hook.go.txt');p=base/'tracing2-census-overlay.json';p.write_text(json.dumps(d,indent=2)+'\n')
run(['go','build','-overlay',str(p),'-o',str(base/'tracing2-latent'),'./stage3/census/latent/tool']);run(['go','build','-o',str(base/'tracing2-adamic'),'./cmd/adamic']);run(['go','build','-o',str(base/'tracing2-first-stop'),'./stage3/census/tool'])
