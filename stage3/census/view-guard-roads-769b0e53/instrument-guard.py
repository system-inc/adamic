#!/usr/bin/env python3
"""Install identical measurement code on either pinned compiler, in scratch only."""
import json,sys,shutil
from pathlib import Path
root,base,destination=map(lambda x:Path(x).resolve(),sys.argv[1:4]);evidence=Path(__file__).resolve().parent
destination.mkdir(parents=True,exist_ok=True);data=json.loads((base/'overlay.json').read_text());replacements={}
for original,replacement in data['Replace'].items():
 src=Path(replacement);dst=destination/src.name;shutil.copyfile(src,dst);replacements[original]=str(dst)
def edit(relative,old,new):
 p=Path(replacements[str(root/relative)]);s=p.read_text();assert s.count(old)==1,(relative,s.count(old));p.write_text(s.replace(old,new))
edit('internal/lower/latent_hook.go','l := fresh()','l := fresh()\n defer l.roadsGuardObserve(unit.Where, record.File)')
edit('internal/lower/latent_units.go','l := base.latentSnapshot()','l := base.latentSnapshot()\n defer l.roadsGuardObserve(unit.Where, record.File)')
hook=destination/'roads_guard_hook.go';shutil.copyfile(evidence/'guard-hook.go.txt',hook);replacements[str(root/'internal/lower/roads_guard_hook.go')]=str(hook)
source=(root/'internal/lower/graph_types.go').read_text();needle='func (f *cycleFinder) viewPairGuard() error {';assert source.count(needle)==1
source=source.replace(needle,needle+'\n if roadsGuardForceAll() && len(f.l.viewSources)>0 { return f.l.notYet(f.l.viewSources[0].node, "scratch mutant: every view pair refuses") }',1)
graph=destination/'roads_graph_types.go';graph.write_text(source);replacements[str(root/'internal/lower/graph_types.go')]=str(graph)
(destination/'overlay.json').write_text(json.dumps({'Replace':replacements},indent=2)+'\n');print(destination/'overlay.json')
