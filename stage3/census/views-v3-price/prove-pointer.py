#!/usr/bin/env python3
"""Prove exact counter source visits a shared ElementCheck pointer once."""
import json,os,pathlib,subprocess
p=pathlib.Path(__file__).resolve().parent
scratch=pathlib.Path('/workspace/scratch/v3-pointer-proof');scratch.mkdir(exist_ok=True)
s=(p/'ir-count.go.txt').read_text().split('func priceObserve(')[0].replace('package lower','package main',1)
s+='''
func main(){
 check:=&ir.ViewElementCheck{}
 a:=ir.ForOf{ElementCheck:check}
 b:=ir.ForOf{ElementCheck:check}
 elements,checks:=priceCount(a,b,check)
 if elements!=0||checks!=1{panic("distinct ElementCheck count failed")}
}
'''
driver=scratch/'driver.go.txt';driver.write_text(s)
meta={'Replace':{'/workspace/v3-head/stage3/census/latent/tool/main.go':str(driver)}}
(scratch/'overlay.json').write_text(json.dumps(meta)+'\n')
results=[]
for mutant in (False,True):
 env=dict(os.environ);env.pop('V3_MUTANT_SKIP_NODE',None)
 if mutant:env['V3_MUTANT_SKIP_NODE']='ViewElementCheck'
 with (scratch/('mutant.log' if mutant else 'normal.log')).open('w') as log:
  run=subprocess.run(['go','run','-overlay='+str(scratch/'overlay.json'),'./stage3/census/latent/tool'],cwd='/workspace/v3-head',env=env,stdout=log,stderr=log)
 assert (run.returncode!=0)==mutant
 results.append(dict(mutant=mutant,returncode=run.returncode,distinct_checks_expected=1))
(scratch/'proof.json').write_text(json.dumps(results,indent=2)+'\n')
print(json.dumps(results))
