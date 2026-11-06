#!/usr/bin/env python3
"""Extend a finished meter overlay without editing meter or compiler sources."""
import json,pathlib,sys
source,out=map(lambda x:pathlib.Path(x).resolve(),sys.argv[1:3]);out.mkdir(parents=True,exist_ok=True)
replacements={}
for name,target in json.loads((source/'overlay.json').read_text())['Replace'].items():
 p=out/(pathlib.Path(target).name.removesuffix('.go')+'.go.txt');p.write_bytes(pathlib.Path(target).read_bytes());replacements[name]=str(p)
def target(suffix):
 matches=[p for n,p in replacements.items() if n.endswith('/'+suffix)];assert len(matches)==1,suffix
 return pathlib.Path(matches[0])
hook=target('internal/lower/latent_hook.go');s=hook.read_text();needle='type LatentUnit struct {';assert s.count(needle)==1
s=s.replace(needle,needle+'\n ViewElements int `json:"view_elements"`\n ViewElementChecks int `json:"view_element_checks"`\n IRMeasured bool `json:"ir_measured"`',1);hook.write_text(s)
units=target('internal/lower/latent_units.go');s=units.read_text();needle='l := base.latentSnapshot()';assert s.count(needle)==1
s=s.replace(needle,needle+'\n var priceStatements []ir.Statement\n defer func(){priceObserve(&unit,l,priceStatements)}()',1)
needle='_, err := l.statement(node)';assert s.count(needle)==1
s=s.replace(needle,'var err error\n priceStatements, err = l.statement(node)',1);units.write_text(s)
repository=pathlib.Path(next(n for n in replacements if n.endswith('/internal/lower/lower.go'))).parents[2]
p=out/'price_count.go.txt';p.write_bytes(pathlib.Path(__file__).with_name('ir-count.go.txt').read_bytes());replacements[str(repository/'internal/lower/price_count.go')]=str(p)
(out/'overlay.json').write_text(json.dumps({'Replace':replacements},indent=2)+'\n')
