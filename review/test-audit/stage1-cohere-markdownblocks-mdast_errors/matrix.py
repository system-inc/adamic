import pathlib,json,subprocess,time,os
root=pathlib.Path('/workspace/adamic');ev=root/'review/test-audit/stage1-cohere-markdownblocks-mdast_errors';menu=json.loads((ev/'menu.json').read_text());originals=json.loads((ev/'originals.json').read_text());records=[];valid=[]
for m in menu:
 path=root/m['file'];original=originals[m['file']];changed=original.replace(m['old'],m['new']);label=m['id'];env=dict(os.environ)
 if label.startswith('G') or label=='P5':env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u128/cache/'+label
 start=time.monotonic()
 apply=subprocess.run(['git','apply','--check',str(ev/(label+'.diff'))],cwd=root,capture_output=True);assert apply.returncode==0,(label,apply.stderr)
 overlay=None
 try:
  if m['file'].endswith('.go'):
   scratch=pathlib.Path('/tmp/u128/'+label+'.go');scratch.write_text(changed);overlay=pathlib.Path('/tmp/u128/'+label+'-overlay.json');overlay.write_text(json.dumps({'Replace':{str(path):str(scratch)}}));extra=['-overlay='+str(overlay)]
   with(ev/(label+'-vet.log')).open('w')as log:v=subprocess.run(['go','vet']+extra+(['./internal/lower/'] if label=='P5' else ['./stage1/cohere/markdownblocks/']),cwd=root,stdout=log,stderr=subprocess.STDOUT)
   valid.append(dict(id=label,apply=0,vet=v.returncode,wall=time.monotonic()-start));assert v.returncode==0,label
  else:path.write_text(changed);extra=[]
  cmd=['timeout','120','go','test']+extra+['-json','-count=1','-timeout','90s','./stage1/cohere/markdownblocks/','-run','^('+'|'.join(m['members'])+')$'];start=time.monotonic()
  with(ev/(label+'.log')).open('w')as log:r=subprocess.run(cmd,cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT)
  records.append(dict(id=label,command=' '.join(cmd),exit=r.returncode,wall=time.monotonic()-start,matrix_rows=m['members'],cache=env.get('ADAMIC_BUILD_CACHE_DIR')));(ev/'matrix-commands.json').write_text(json.dumps(records,indent=2));(ev/'validation.json').write_text(json.dumps(valid,indent=2));print(label,r.returncode,round(records[-1]['wall'],2),flush=True)
 finally:
  if not m['file'].endswith('.go'):path.write_text(original)
