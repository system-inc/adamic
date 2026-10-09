import pathlib,subprocess,os,time,json,re
p=pathlib.Path('review/test-audit/stage1-cohere-formatfiles-formatfiles_products');src=pathlib.Path('stage1/cohere/formatfiles/formatfiles_products_test.go').read_text();names=re.findall(r'func (Test\w+)\(',src);groups=[[names[0]],[names[1]],[names[2]],names[3:]];rs=[]
for group in groups:
 name=group[0] if len(group)==1 else 'TestProduct_FormatfilesNativeMutants family';pat='^('+'|'.join(group)+')$'
 for i in range(3):
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/formatfiles/','-run',pat];t=time.monotonic();label=('family' if len(group)>1 else name)+'-'+str(i)
  with open(p/('timing-'+label+'.log'),'w') as f:r=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT)
  rs.append(dict(row=name,members=group,trial=i,command=cmd,exit=r.returncode,wall_seconds=time.monotonic()-t));(p/'timing.json').write_text(json.dumps(rs,indent=2));print(label,r.returncode,round(rs[-1]['wall_seconds'],3),flush=True)
