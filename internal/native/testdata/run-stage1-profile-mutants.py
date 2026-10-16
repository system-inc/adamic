#!/usr/bin/env python3
"""Each profile-binding guard is removed in a compiling overlay, one test per guard."""
import json, os, subprocess, tempfile
from pathlib import Path
root=Path(__file__).resolve().parents[3]
source=root/'internal/native/stage1_profile.go'
cases=[('version','record.Version == current.Version','TestStage1ProfileVersion'),('units','slices.Equal(record.Units, current.Units)','TestStage1ProfileUnitNames'),('runtime','record.Runtime == current.Runtime','TestStage1ProfileRuntime'),('profile','record.Profile == current.Profile','TestStage1ProfileText'),('compiler','record.CompilerSHA256 == current.CompilerSHA256','TestStage1ProfileCompiler'),('compiler-version','record.Compiler == current.Compiler','TestStage1ProfileCompilerVersion'),('compile','slices.Equal(record.Flags, current.Flags)','TestStage1ProfileCompileFlags'),('link','slices.Equal(record.LinkFlags, current.LinkFlags)','TestStage1ProfileLinkFlags'),('training-compile','slices.Equal(record.TrainingFlags, current.TrainingFlags)','TestStage1ProfileTrainingFlags'),('training-link','slices.Equal(record.TrainingLinkFlags, current.TrainingLinkFlags)','TestStage1ProfileTrainingLinkFlags'),('target','record.Target == current.Target','TestStage1ProfileTarget')]
cases += [('unit-inventory','for _, unit := range parts {','TestStage1ProfileAllUnits'),('fingerprint-content','h.Write(f.contents)','TestRuntimeFingerprintCoversEveryFile'),('shard-content','if profileHash(content) != part.SHA256 {','TestStage1ProfileShardBytes')]
with tempfile.TemporaryDirectory(prefix='stage1-profile-mutants-') as work:
 for name,before,test in cases:
  if os.environ.get('ADAMIC_PROFILE_MUTANTS') and name not in os.environ['ADAMIC_PROFILE_MUTANTS'].split(','):continue
  text=source.read_text();assert text.count(before)==1
  changed=Path(work)/(name+'.go.txt');changed.write_text(text.replace(before,{'unit-inventory':'for _, unit := range parts[:0] {','fingerprint-content':'h.Write(nil)','shard-content':'if false && profileHash(content) != part.SHA256 {'}.get(name,'true'),1))
  overlay=Path(work)/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(source):str(changed)}}))
  log=Path(tempfile.gettempdir())/('stage1-profile-mutant-'+name+'.log')
  with log.open('w') as output:
   result=subprocess.run(['go','test','-p','1','-overlay='+str(overlay),'./internal/native','-run','^'+test+'$','-count=1','-v','-timeout=55s'],cwd=root,stdout=output,stderr=subprocess.STDOUT,timeout=60)
  body=log.read_text()
  if result.returncode==0 or '--- FAIL: '+test not in body or (name not in ['unit-inventory','fingerprint-content','shard-content'] and 'stale profile accepted' not in body):raise RuntimeError(name+' survived or failed for the wrong reason: '+body)
  print(name+' caught by '+test,flush=True)
