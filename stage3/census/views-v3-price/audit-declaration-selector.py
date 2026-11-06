#!/usr/bin/env python3
"""Prove scheduling off identity on both compilers and selected fixture union."""
import json,os,pathlib,re,subprocess
P=pathlib.Path(__file__).resolve().parent;S=pathlib.Path('/workspace/scratch/v3-selector-audit');S.mkdir(exist_ok=True)
proof=[]
for side in ['main','head']:
 for fixture in ['per-read','map-good','for-of-good']:
  source=pathlib.Path('/workspace/v3-head/internal/oracle/testdata/views_v3')/(fixture+'.ts');root=str(source.parent)
  env=dict(os.environ,LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1');env.pop('PRICE_DECLARATION_PIECE',None);env.pop('PRICE_INVENTORY',None)
  def run(binary,name,selection=None):
   e=dict(env)
   if selection:
    path=S/(name+'.selection.json');path.write_text(json.dumps(selection));e.update(PRICE_DECLARATION_PIECE=str(path),PRICE_DECLARATION_ROOT=root,LATENT_ONLY_FILE=str(source))
   output=S/(name+'.jsonl')
   with (S/(name+'.log')).open('w') as log:subprocess.run([str(binary),str(source),str(output)],env=e,stdout=log,stderr=log,check=True,timeout=90)
   return output.read_bytes()
  name=side+'-'+fixture;new=run('/workspace/scratch/v3-'+side+'-declaration-bin',name+'-off')
  old=pathlib.Path('/workspace/scratch/v3-ir-proof')/(name+'-count.jsonl')
  assert new==old.read_bytes(),(name,'selector off changed counted meter bytes')
  stripped=re.sub(rb'"(?:ir_measured|view_elements|view_element_checks)":(?:true|false|[0-9]+),',b'',new)
  assert stripped==old.with_name(name+'-meter.jsonl').read_bytes(),(name,'original meter byte identity')
  rows=list(map(json.loads,new.splitlines()));whole=rows[1];tops=[]
  for u in whole['units']:
   if u['top'] not in tops:tops.append(u['top'])
  divided=[];finding_units={}
  for i,part in enumerate([tops[:len(tops)//2],tops[len(tops)//2:]]):
   selection=dict(id=name+'-'+str(i),files=[dict(file=source.name,top_units=[v.replace(root+'/','') for v in part])])
   raw=run('/workspace/scratch/v3-'+side+'-declaration-bin',name+'-'+str(i),selection)
   selected=list(map(json.loads,raw.splitlines()))[1]
   assert {u['top'] for u in selected['units']}==set(part)
   divided.extend(selected['units'])
   for f in selected['findings']:
    if f['phase']=='lowering':assert f['unit'] in {u['where'] for u in selected['units']};finding_units.setdefault(f['unit'],[]).append(f)
  assert divided==whole['units'],(name,'fixture unit union moved or lost')
  before={}
  for f in whole['findings']:
   if f['phase']=='lowering':before.setdefault(f['unit'],[]).append(f)
  assert finding_units==before,(name,'fixture lowering findings changed')
  proof.append(dict(side=side,fixture=fixture,mode_off_byte_identical=True,unit_union_exact=True,lowering_findings_exact=True))
(S/'proof.json').write_text(json.dumps(proof,indent=2)+'\n');print(json.dumps(proof,indent=2))
