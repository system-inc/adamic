import json,subprocess,sys
from pathlib import Path
repo=Path('/workspace/adamic');old='fb1c8d3b72de2686c486cf97685079589308dafd';new=subprocess.check_output(['git','rev-parse','origin/main'],cwd=repo,text=True).strip()
changed=subprocess.check_output(['git','diff','--name-only',old,new],cwd=repo,text=True).splitlines();assert changed==['bridge/tsgo/units_test.go'],changed
cmd=['/workspace/adamic-tools/go/bin/go','list','-json','./bridge/tsgo'];r=subprocess.run(cmd,cwd=repo,capture_output=True,text=True);print('go list child_exit='+str(r.returncode));
if r.returncode:sys.exit(r.returncode)
p=json.loads(r.stdout);tests=p.get('TestGoFiles',[])+p.get('XTestGoFiles',[]);assert 'units_test.go' in tests;assert 'units_test.go' not in p['GoFiles'];assert 'units_test.go' not in p.get('EmbedFiles',[])
result=dict(appliedMain=old,finalMain=new,changed=changed,compilerAndAdaptationInputsIdentical=True,testGoFiles=tests,productionGoFiles=p['GoFiles'],explanation='Only a non-embedded Go test changed. Full apply, stock tsc, and emission inputs are byte-identical. Compiler builds, types, and censuses are refreshed after rebase.')
(Path(__file__).parent/'tip-equivalence.json').write_text(json.dumps(result,indent=2)+'\n');print('production and adaptation inputs identical')
