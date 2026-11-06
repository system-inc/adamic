#!/usr/bin/env python3
"""Add a top-level scheduling guard to the already proven full counted overlay."""
import json,pathlib,sys
source,out=map(lambda s:pathlib.Path(s).resolve(),sys.argv[1:3]);out.mkdir(parents=True,exist_ok=True)
mapping={}
for logical,physical in json.loads((source/'overlay.json').read_text())['Replace'].items():
 target=out/(pathlib.Path(physical).name.removesuffix('.go')+'.txt' if pathlib.Path(physical).suffix=='.go' else pathlib.Path(physical).name)
 target.write_bytes(pathlib.Path(physical).read_bytes());mapping[logical]=str(target)
def target(suffix):
 rows=[v for k,v in mapping.items() if k.endswith('/'+suffix)];assert len(rows)==1,suffix
 return pathlib.Path(rows[0])
f=target('internal/lower/latent_units.go');text=f.read_text();needle='for _, candidate := range latentCandidates(program, file) {\n\t\t\tnode := candidate.node';assert text.count(needle)==1
text=text.replace(needle,needle+'\n if !priceTopSelected(candidate.top) { continue }',1);f.write_text(text)
f=target('stage3/census/latent/tool/main.go');text=f.read_text();needle='lower.Latent(context.Background(), program, func(record lower.LatentFile) {';assert text.count(needle)==1
text=text.replace(needle,'if os.Getenv("PRICE_INVENTORY") == "1" { if err := encoder.Encode(lower.PriceInventory(program)); err != nil { panic(err) }; return }\n'+needle,1);f.write_text(text)
repository=pathlib.Path(next(k for k in mapping if k.endswith('/internal/lower/lower.go'))).parents[2]
f=out/'price_declaration_select.go.txt';f.write_bytes(pathlib.Path(__file__).with_name('declaration-select.go.txt').read_bytes());mapping[str(repository/'internal/lower/price_declaration_select.go')]=str(f)
(out/'overlay.json').write_text(json.dumps({'Replace':mapping},indent=2)+'\n')
