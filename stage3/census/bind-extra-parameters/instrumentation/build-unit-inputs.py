from pathlib import Path
import json,subprocess,shutil
s=Path('/workspace/bind-extra-parameters-scratch');old=s/'head-overlay';new=s/'head-unit-inputs-overlay';shutil.copytree(old,new)
data=json.loads((new/'overlay.json').read_text());data['Replace']={k:v.replace(str(old)+'/',str(new)+'/') for k,v in data['Replace'].items()}
p=new/'internal_lower_latent_scoped_tail.go';text=p.read_text();needle='func (l *lowering) latentFinishUnit(unit *LatentUnit, node *ast.Node) {';assert text.count(needle)==1;text=text.replace(needle,needle+'\n l.bindCollectUnit(unit,node)\n');p.write_text(text)
p=new/'unit_inputs.go';p.write_text((s/'unit_inputs.go.txt').read_text());data['Replace'][str(s/'head/internal/lower/unit_inputs.go')]=str(p)
(new/'overlay.json').write_text(json.dumps(data,indent=2)+'\n');subprocess.run(['gofmt','-w',str(new/'unit_inputs.go'),str(new/'internal_lower_latent_scoped_tail.go')],check=True)
with (s/'head-unit-inputs-build.log').open('w') as log:subprocess.run(['go','build','-buildvcs=false','-overlay='+str(new/'overlay.json'),'-o',str(s/'head-unit-inputs-census'),'./stage3/census/latent/tool'],cwd=s/'head',stdout=log,stderr=subprocess.STDOUT,check=True)
