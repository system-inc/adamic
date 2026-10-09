import pathlib,subprocess,json,time,os
root=pathlib.Path('/tmp/u082');groups=json.loads(pathlib.Path('review/test-audit/stage1-cohere-cssnumbers/scope.json').read_text());env=os.environ.copy();env['ADAMIC_CSSNUMBERS_LIBRARY']='/tmp/u082/library'
for i,g in enumerate(groups):
 for n in range(1,4):
  cmd=['timeout','120','go','test','-count=1','-timeout','90s','./stage1/cohere/cssnumbers/','-run',g['pattern']]
  start=time.monotonic()
  with (root/f'time-{i}-{n}.log').open('w') as log:p=subprocess.run(cmd,env=env,stdout=log,stderr=subprocess.STDOUT)
  (root/f'time-{i}-{n}.meta').write_text(json.dumps(dict(exit=p.returncode,wall=time.monotonic()-start,command=cmd)))
  if p.returncode:break
(root/'timings-done').write_text('done')
