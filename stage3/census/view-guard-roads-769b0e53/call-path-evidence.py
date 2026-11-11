#!/usr/bin/env python3
import json,re,subprocess
from pathlib import Path
out=Path(__file__).resolve().parent
production={
 'internal/lower/lower.go':['lowering.findCycles(modules)'],
 'internal/lower/cycles.go':['finder.graphTypes(modules)','l.selectProgramMembership(finder, modules)'],
 'internal/lower/graph_types.go':['f.viewPairGuard()'],
 'internal/lower/program_region.go':['f.viewPairGuard()'],
}
result={}
for label,root in [('parent',Path('/workspace/roads-parent')),('head',Path('/workspace/roads-head'))]:
 destination=out/'call-path'/label;destination.mkdir(parents=True,exist_ok=True)
 rows=[]
 for name,needles in production.items():
  source=(root/name).read_text();(destination/(Path(name).name+'.txt')).write_text(source)
  for needle in needles:
   locations=[n for n,line in enumerate(source.splitlines(),1) if needle in line]
   assert len(locations)==1,(label,name,needle,locations)
   rows.append({'file':name,'call':needle,'line':locations[0]})
 generated=Path('/workspace/roads-'+label+'-overlay/internal_lower_lower.go').read_text()
 assert 'findCycles(' not in generated
 assert 'lowerChecked(' not in generated
 assert 'lowerWithElementKey(' not in generated
 assert 'return nil, fmt.Errorf("latent census: measurement only; no IR output")' in generated
 latent=[]
 for p in Path('/workspace/roads-'+label+'-overlay').glob('*.go'):
  # All guard and cycle-pass calls would have an ordinary Go call spelling.
  assert not re.search(r'\b(?:findCycles|viewPairGuard|graphTypes|selectProgramMembership)\s*\(',p.read_text()),p
  latent.append(p.name)
 result[label]={'production_call_path':rows,'generated_lower_pipeline_removed':True,'guard_or_cycle_calls_in_generated_overlay':0,'overlay_go_files_checked':sorted(latent)}
# The entire measured production delta is two files. No latent runner changes exist.
result['changed_files']=subprocess.check_output(['git','-C','/workspace/roads-head','diff','--name-only','16429a30','769b0e53'],text=True).splitlines()
assert result['changed_files']==['internal/lower/graph_types.go','internal/lower/graph_view_guard_test.go']
(out/'call-path-evidence.json').write_text(json.dumps(result,indent=2)+'\n')
print('call-path evidence assertions passed')
