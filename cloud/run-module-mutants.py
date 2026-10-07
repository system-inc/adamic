#!/usr/bin/env python3
"""Run source mutants against the closure's invalidation and integrity tests."""
from pathlib import Path
import os, subprocess, tempfile, json
root=Path(__file__).resolve().parent
source=(root/'setup-modules.py').read_text()
mutants={}
for component in ['manifests','version','environment','helper']:
 old='dict(manifests=manifests, version=version, environment=environment, helper=helper)'
 replacement='dict('+', '.join(f'{name}={name}' for name in ['manifests','version','environment','helper'] if name!=component)+')'
 mutants['key-'+component]=(source.replace(old,replacement), 'Modules.test_key_components')
for field in ['st_mode','st_size','st_mtime_ns','st_ctime_ns','st_ino']:
 mutants['cache-'+field]=(source.replace('info.'+field,'0'), 'Modules.test_every_cache_field')
mutants['cache-path']=(source.replace('[str(file), info.st_mode','["", info.st_mode'),'Modules.test_every_cache_field')
mutants['zip-integrity']=(source.replace("zipped != record['Sum'] or ",''),'ClosureChecks.test_archive_integrity')
mutants['extracted-integrity']=(source.replace(" or content_sum(files) != record['Sum']",''),'ClosureChecks.test_extracted_integrity')
mutants['manifest-integrity']=(source.replace("if content_sum([('go.mod', Path(record['GoMod']).read_bytes())]) != record['GoModSum']:","if False:"),'ClosureChecks.test_go_mod_integrity')
mutants['mandatory-all-modules']=(source.replace('if all_modules:\n        return prepare_all','if True:\n        return prepare_all'),'ClosureIntegration.test_unreachable_require_cannot_fail_mandatory_setup')
report=root/'reports/setup-closure';results=[]
with tempfile.TemporaryDirectory(prefix='module-mutants-',dir='/tmp/adamic-gate') as temporary:
 for name,(mutant,test) in mutants.items():
  assert mutant!=source,name
  file=Path(temporary)/(name+'.py');file.write_text(mutant)
  log=report/('mutant-'+name+'.log')
  with log.open('w') as output:
   result=subprocess.run(['python3','-m','unittest','test_setup_modules.'+test],cwd=root,env=dict(os.environ,ADAMIC_MODULES_HELPER=str(file),ADAMIC_SETUP_INTEGRATION='1'),stdout=output,stderr=subprocess.STDOUT)
  text=log.read_text()
  assert result.returncode and ('FAIL:' in text or 'ERROR:' in text),name
  assert 'SyntaxError' not in text,name
  results.append(dict(mutant=name,test=test,exit=result.returncode));print(name,'caught',flush=True)
(report/'mutants.json').write_text(json.dumps(results,indent=2)+'\n')
