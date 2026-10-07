import json,os,pathlib,subprocess,tempfile
repo=pathlib.Path('/workspace/adamic');source=repo/'internal/nodepin/nodepin.go';original=source.read_text();reports=repo/'cloud/reports/node-pin'
for name,mutant,test in [('pin-drift',original.replace('const Version = "v24.19.0"','const Version = "v24.21.0"'),'TestSetupVersion'),('missing-oracle-guard',original.replace('if version != Version {','if false {'),'TestCheckUsesPATH')]:
 with tempfile.TemporaryDirectory(prefix='node-go-mutant-',dir='/tmp/adamic-gate') as temporary:
  root=pathlib.Path(temporary);file=root/'nodepin.go';file.write_text(mutant);overlay=root/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(source):str(file)}}))
  with (reports/(name+'.log')).open('wb') as log:
   result=subprocess.run(['go','test','-overlay',str(overlay),'./internal/nodepin','-run','^'+test+'$','-count=1','-v'],cwd=repo,stdout=log,stderr=subprocess.STDOUT,timeout=120)
  text=(reports/(name+'.log')).read_text();assert result.returncode!=0 and ('setup pins v24.19.0, nodepin.Version is v24.21.0' in text if name=='pin-drift' else 'mismatch must name both versions' in text),text
  print(name+': intended assertion caught mutant')
