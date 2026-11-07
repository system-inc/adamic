from run import *
import sys
(S/'bin').mkdir(exist_ok=True)
assert run('native-build',['go','build','-o',str(S/'bin/adamic'),'./cmd/adamic'])==0
assert run('overlay',['python3','stage3/census/latent/make_overlay.py',str(ROOT),str(S/'overlay')])==0
p=S/'overlay/overlay.json';data=json.loads(p.read_text())
for virtual,physical in list(data['Replace'].items()):
 old=pathlib.Path(physical)
 if old.suffix=='.go':new=old.with_name(old.name+'.txt');old.rename(new);data['Replace'][virtual]=str(new)
p.write_text(json.dumps(data,indent=2)+'\n')
assert run('latent-build',['go','build','-buildvcs=false','-overlay='+str(p),'-o',str(S/'bin/census'),'./stage3/census/latent/tool'])==0
