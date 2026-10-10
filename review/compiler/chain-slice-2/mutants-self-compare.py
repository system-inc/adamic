import json,os,signal,subprocess,time
from pathlib import Path
root=Path.cwd();out=root/'review/compiler/chain-slice-2';results=[]
for name,path,before,after,marker in [
 ('static-inline-definition','internal/native/units.go','if d.function && d.tokens[0].text == "static" && d.tokens[1].text == "inline" {','if false {',"inline function 'adamic_unit_adamic_same_number' is not defined"),
 ('split-numeric-fold','internal/native/self_compare.go','comparison := fmt.Sprintf("%s(%s, %s)", helper, left, right)','comparison := fmt.Sprintf("%s(%s, %s)", helper, left, right)\n if operandType == ir.Number { comparison = fmt.Sprintf("((void)(%s), (void)(%s), true)", left, right) }','native: <nil>')]:
 source=(root/path).read_text();assert source.count(before)==1
 directory=out/('overlay-'+name);directory.mkdir(exist_ok=True);target=directory/'source.go.txt';target.write_text(source.replace(before,after))
 overlay=directory/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(root/path):str(target)}}))
 command=['go','test','./internal/native','-overlay='+str(overlay),'-run','^TestSplitSelfCompareAgreesWithNode$','-count=1','-v','-timeout=90s']
 log=directory/'test.log';start=time.monotonic()
 with log.open('w') as f:
  p=subprocess.Popen(command,stdout=f,stderr=subprocess.STDOUT,start_new_session=True)
  try:code=p.wait(timeout=110)
  except subprocess.TimeoutExpired:os.killpg(p.pid,signal.SIGKILL);code=p.wait()
 text=log.read_text();caught=code==1 and marker in text
 if name=='split-numeric-fold':caught=caught and 'got ' in text and 'Node ' in text and '[build failed]' not in text
 row={'mutant':name,'caught':caught,'exit':code,'seconds':round(time.monotonic()-start,3),'structural_only':name=='static-inline-definition','command':command}
 results.append(row);(out/'mutants-self-compare-results.json').write_text(json.dumps(results,indent=2)+'\n');print(json.dumps(row),flush=True);assert caught,name
