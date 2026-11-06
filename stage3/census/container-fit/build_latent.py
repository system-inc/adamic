"""One pinned label, original and scheduling-only selector binaries."""
import difflib,gzip,hashlib,json,pathlib,subprocess,sys
O=pathlib.Path(__file__).resolve().parent;E=O/'evidence/latent-build';E.mkdir(exist_ok=True)
label=sys.argv[1];src=pathlib.Path('/workspace/container-fit-'+label);D=pathlib.Path('/workspace/container-fit-build-v2')/label;D.mkdir(parents=True,exist_ok=True)
meter=pathlib.Path('/workspace/container-fit-head/stage3/census/latent/make_overlay.py')
def run(cmd,name,cwd=src):
 with (E/(label+'-'+name+'.log')).open('w') as f:subprocess.run(cmd,cwd=cwd,stdout=f,stderr=subprocess.STDOUT,check=True)
run(['python3',str(meter),str(src),str(D)],'overlay')
run(['gofmt','-w',*[str(p) for p in D.glob('*.go')]],'format')
run(['go','build','-buildvcs=false','-overlay='+str(D/'overlay.json'),'-o',str(D/'original'),'./stage3/census/latent/tool'],'original')
hook=D/'internal_lower_latent_hook.go';old=hook.read_text();start=old.index('func latentRun(');a,b=old[:start],old[start:]
for needle,extra in [('\tfor _, file := range program.Files() {\n','\t\tif !anyMembersCensusSelected(program.FileName(file), "") { continue }\n'),('\t\tfor _, node := range file.Statements.Nodes {\n','\t\t\tif !anyMembersCensusSelected(program.FileName(file), program.Where(node)) { continue }\n')]:
 assert b.count(needle)==1;b=b.replace(needle,needle+extra,1)
hook.write_text(a+b)
helper=D/'first-error-piece.go.txt';helper.write_bytes((O/'tooling/first-error-piece.go.txt').read_bytes())
mapping=json.loads((D/'overlay.json').read_text());mapping['Replace'][str(src/'internal/lower/any_members_census_piece.go')]=str(helper)
for original,replacement in list(mapping['Replace'].items()):
 p=pathlib.Path(replacement)
 if p.suffix=='.go':q=p.with_suffix('.go.txt');p.rename(q);mapping['Replace'][original]=str(q)
(D/'overlay.json').write_text(json.dumps(mapping,indent=2)+'\n')
run(['gofmt','-w',*mapping['Replace'].values()],'selector-format')
with gzip.open(E/(label+'-selector.patch.gz'),'wt') as f:f.write(''.join(difflib.unified_diff(old.splitlines(True),(D/'internal_lower_latent_hook.go.txt').read_text().splitlines(True),fromfile='original',tofile='selector',n=0)))
run(['go','build','-buildvcs=false','-overlay='+str(D/'overlay.json'),'-o',str(D/'latent'),'./stage3/census/latent/tool'],'selector')
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
(E/(label+'.json')).write_text(json.dumps(dict(compiler=subprocess.check_output(['git','-C',str(src),'rev-parse','HEAD'],text=True).strip(),meter='62ddd47969e5168e5a0e74a23498b85d144b2361',original_sha256=sha(D/'original'),selector_sha256=sha(D/'latent'),helper_sha256=sha(helper)),indent=2)+'\n')
print(label,'BUILT',flush=True)
