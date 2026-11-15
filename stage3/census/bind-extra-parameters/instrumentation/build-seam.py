from pathlib import Path
import json,subprocess,shutil
s=Path('/workspace/bind-extra-parameters-scratch')
for side in ['base','head']:
 old=s/(side+'-overlay');new=s/(side+'-seam-overlay');shutil.copytree(old,new,dirs_exist_ok=True)
 data=json.loads((new/'overlay.json').read_text());data['Replace']={k:v.replace(str(old)+'/',str(new)+'/') for k,v in data['Replace'].items()}
 p=new/'internal_lower_lower.go';text=(s/side/'internal/lower/lower.go').read_text()
 start=text.index('\tfiles := program.Files()');end=text.index('\n}\n',start)
 text=text[:start]+'\treturn nil, fmt.Errorf("latent census: measurement only; no IR output")'+text[end:]
 text=text.replace('\n\t"github.com/system-inc/adamic/internal/buildlog"','').replace('\n\t"time"','')
 needle='type lowering struct {';assert text.count(needle)==1;text=text.replace(needle,needle+'\n latentDeclarations map[int]*ast.Node\n latentReady map[int]bool\n latentAncestors map[*ast.Node]bool\n latentProgramRecordsPending bool\n bindSeamReach func(ir.Expression)([]int,bool)\n')
 needle='\tif err := lowering.checkBoundArguments(); err != nil {';assert text.count(needle)==1;text=text.replace(needle,'\tlowering.bindSeamObserve()\n'+needle);p.write_text(text)
 hook=new/'bind_seam.go';hook.write_text((s/'bind_seam.go.txt').read_text());data['Replace'][str(s/side/'internal/lower/bind_seam.go')]=str(hook)
 main=new/'stage3_census_latent_tool_main.go';main.write_text((s/'bind_seam_main.go.txt').read_text());data['Replace'][str(s/side/'stage3/census/latent/tool/main.go')]=str(main)
 if side=='head':
  source=(s/side/'internal/lower/function_bind.go').read_text();needle='func (l *lowering) boundValueReach() func(ir.Expression) ([]int, bool) {';assert source.count(needle)==1;source=source.replace(needle,needle+'\n if l.bindSeamReach != nil { return l.bindSeamReach }\n');p=new/'internal_lower_function_bind.go';p.write_text(source);data['Replace'][str(s/side/'internal/lower/function_bind.go')]=str(p)
 (new/'overlay.json').write_text(json.dumps(data,indent=2)+'\n')
 subprocess.run(['gofmt','-w']+list(map(str,new.glob('*.go'))),check=True)
 with (s/(side+'-seam-build.log')).open('w') as f:subprocess.run(['go','build','-buildvcs=false','-overlay='+str(new/'overlay.json'),'-o',str(s/(side+'-bind-seam')),'./stage3/census/latent/tool'],cwd=s/side,stdout=f,stderr=subprocess.STDOUT,check=True)
 print('built bind seam '+side,flush=True)
