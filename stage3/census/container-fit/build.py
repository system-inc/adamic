import pathlib,subprocess,json,hashlib,difflib,os
O=pathlib.Path(__file__).resolve().parent
E=O/'evidence'; S=pathlib.Path('/workspace/container-fit-build'); S.mkdir(exist_ok=True)
def run(cmd,cwd,log):
 with (E/log).open('w') as f: subprocess.run(['timeout','--kill-after=10s','1200',*cmd],cwd=cwd,stdout=f,stderr=subprocess.STDOUT,check=True)
for label in ['base','log','head']:
 src=pathlib.Path('/workspace/container-fit-'+label)
 if label!='base':
  (src/'cohere').rmdir(); (src/'cohere').symlink_to('/workspace/container-fit-base/cohere',target_is_directory=True)
 dest=S/label
 run(['python3','/workspace/container-fit-base/stage3/census/latent/make_overlay.py',str(src),str(dest)],src,label+'-overlay.log')
 hook=dest/'internal_lower_latent_hook.go'; original=hook.read_text(); start=original.index('func latentRun('); a,b=original[:start],original[start:]
 for needle,extra in [('\tfor _, file := range program.Files() {\n','\t\tif !anyMembersCensusSelected(program.FileName(file), "") { continue }\n'),('\t\tfor _, node := range file.Statements.Nodes {\n','\t\t\tif !anyMembersCensusSelected(program.FileName(file), program.Where(node)) { continue }\n')]:
  assert b.count(needle)==1; b=b.replace(needle,needle+extra,1)
 hook.write_text(a+b)
 helper=dest/'first-error-piece.go.txt'; helper.write_bytes((O/'tooling/first-error-piece.go.txt').read_bytes())
 mapping=json.loads((dest/'overlay.json').read_text())
 mapping['Replace'][str(src/'internal/lower/any_members_census_piece.go')]=str(helper)
 # Preserve generated Go artifacts as .go.txt, including all overlay targets.
 for orig,repl in list(mapping['Replace'].items()):
  p=pathlib.Path(repl)
  if p.suffix=='.go':
   q=p.with_suffix('.go.txt'); p.rename(q); mapping['Replace'][orig]=str(q)
 hook=dest/'internal_lower_latent_hook.go.txt'
 (dest/'overlay.json').write_text(json.dumps(mapping,indent=2))
 run(['gofmt','-w',*[v for v in mapping['Replace'].values()]],src,label+'-gofmt.log')
 (E/(label+'-selector.patch.txt')).write_text(''.join(difflib.unified_diff(original.splitlines(True),hook.read_text().splitlines(True))))
 run(['go','build','-buildvcs=false','-overlay='+str(dest/'overlay.json'),'-o',str(dest/'latent'),'./stage3/census/latent/tool'],src,label+'-build.log')
 run(['go','build','-buildvcs=false','-o',str(dest/'census'),'./stage3/census/tool'],src,label+'-census-build.log')
 print(label,'built',flush=True)
