"""Build disposable Go overlays, retaining all production code unchanged on disk."""
import json,pathlib,subprocess,sys
repo=pathlib.Path.cwd()
art=repo/'stage3/census/relation-overflow-real'
scratch=pathlib.Path(sys.argv[1]).resolve()
scratch.mkdir(parents=True,exist_ok=True)
subprocess.run(['python3','stage3/census/latent/make_overlay.py',str(repo),str(scratch/'latent')],check=True)
subprocess.run(['python3','stage3/meter/entry_overlay.py',str(scratch/'latent'),str(scratch/'entry')],check=True)
base=json.loads((scratch/'entry/overlay.json').read_text())['Replace']
for mode in ['real','census']:
 d=dict(base)
 def put(name,text):
  dest=scratch/(mode+'_'+name.replace('/','_'))
  dest.write_text(text)
  d[str(repo/name)]=str(dest)
 if mode=='real':
  text=(repo/'internal/load/load.go').read_text().replace('type Program struct {','type Program struct {\n latentSites []*ast.Diagnostic\n latentDiagnostics []string')
  text=text.replace('formatted := make([]string, 0, len(all))','p.latentSites = all\n formatted := make([]string, 0, len(all))')
  text+='\nfunc LatentLoad(paths []string) (*Program,error) { return Load(paths) }\n'
  put('internal/load/load.go',text)
 for source in sorted((repo/'internal/lower').glob('*.go')):
  if not source.name.endswith('_test.go'):
   put('internal/relationreal/'+source.name,source.read_text())
 checker='cohere/TypeScript/tsc/internal/checker/'
 text=(repo/(checker+'checker.go')).read_text()
 assert text.count('type Checker struct {')==1
 put(checker+'checker.go',text.replace('type Checker struct {','type Checker struct {\n measureOverflowPairs []measureOverflowPair'))
 text=(repo/(checker+'relater.go')).read_text()
 needle='c.setRelationResult(relation, id, RelationComparisonResultFailed|RelationComparisonResultComplexityOverflow)'
 assert text.count(needle)==1
 put(checker+'relater.go',text.replace(needle,'c.measureOverflow(source,target,relation,id,errorNode)\n\t\t'+needle))
 put(checker+'measure_overflow.go',(art/'checker-hook.go.txt').read_text())
 put('internal/load/measure_relations.go',(art/'snapshot.go.txt').read_text())
 put('stage3/census/relation-overflow-real/tool/main.go',(art/'main.go.txt').read_text())
 (scratch/(mode+'.json')).write_text(json.dumps({'Replace':d},indent=2)+'\n')
