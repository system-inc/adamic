#!/usr/bin/env python3
"""Drop cache components individually and require the intended assertions to fail."""
import os
from pathlib import Path
import shutil
import subprocess
import tempfile

source=Path(__file__).resolve().parent
scratch=Path(tempfile.mkdtemp(prefix='gate-input-mutants-',dir='/tmp/adamic-gate'))
print('mutant logs:',scratch,flush=True)
shutil.copyfile(source/'setup-gate-npm.py',scratch/'setup-gate-npm.py')
text=(source/'setup-gate-inputs.py').read_text()
variants={('drop-'+name):text.replace('def cache_key(**inputs):','def cache_key(**inputs):\n    inputs.pop('+repr(name)+', None)')
          for name in ['kind','head','packages','environment','version','cc_version','flags','validation_flags','helper','commit','url','size','content']}
variants['drop-artifact-root-mode']=text.replace('root_mode=directory.stat().st_mode & 0o777', 'root_mode=0o755')
variants['drop-artifact-entries']=text.replace('entries=entries)', 'entries=[])')
variants.update({'drop-artifact-bytes':text.replace('file_hash(path)', "'ignored'"),
                 'drop-artifact-modes':text.replace('mode = path.lstat().st_mode & 0o777','mode = 0o644'),
                 'drop-artifact-names':text.replace('str(relative), mode', "'ignored', mode")})
# Don't mutate the file_hash function definition: only actual entry comparisons.
variants['drop-artifact-bytes']=text.replace('mode, file_hash(path)', "mode, 'ignored'")
for name,variant in variants.items():
    assert variant!=text
    module=scratch/(name+'.py');module.write_text(variant)
    environment=dict(os.environ,ADAMIC_GATE_INPUTS_MODULE=str(module));environment.pop('ADAMIC_SETUP_INTEGRATION',None)
    log=scratch/(name+'.log')
    with log.open('wb') as output:
        result=subprocess.run(['python3',str(source/'test_gate_inputs.py'),'Inputs.test_every_key_component','Inputs.test_actual_content_modes_and_names'],env=environment,stdout=output,stderr=subprocess.STDOUT,timeout=60)
    answer=log.read_text();assert result.returncode==1 and 'AssertionError' in answer and 'FAILED (failures=' in answer,answer
    print(name,'caught by intended input/byte assertion',flush=True)
    if name=='drop-packages':
        with (scratch/'real-dirty-source.log').open('wb') as output:
            result=subprocess.run(['python3',str(source/'test_gate_inputs.py'),'Archive'],env=dict(environment,ADAMIC_SETUP_INTEGRATION='1'),stdout=output,stderr=subprocess.STDOUT,timeout=300)
        answer=(scratch/'real-dirty-source.log').read_text();assert result.returncode==1 and 'AssertionError' in answer and 'skipped (validated Go actions' in answer,answer
        print(name,'also caught by real dirty-source archive rebuild assertion',flush=True)
# The seven npm seats use the same five-component key and installed tree validator.
text=(source/'setup-gate-npm.py').read_text()
npm_variants={name:text.replace(name+'='+name,name+"='ignored'") for name in ['lock','manifest','bootstrap','helper','node']}
npm_variants.update({'bytes':text.replace('digest(path.read_bytes())',"'ignored'"),
                     'modes':text.replace('mode = path.lstat().st_mode & 0o777','mode = 0o644'),
                     'names':text.replace('relative, mode',"'ignored', mode"),
                     'root-mode':text.replace('directory.stat().st_mode & 0o777','0o755'),
                     'integrity':text.replace("algorithm != 'sha512' or base64.b64encode(hashlib.sha512(data).digest()).decode() != expected",'False')})
for component,variant in npm_variants.items():
    assert variant!=text
    module=scratch/('npm-drop-'+component+'.py');module.write_text(variant)
    log=scratch/('npm-drop-'+component+'.log')
    with log.open('wb') as output:
        result=subprocess.run(['python3',str(source/'test_markdown_setup.py'),'InstallationKey'],env=dict(os.environ,ADAMIC_MARKDOWN_SETUP_MODULE=str(module)),stdout=output,stderr=subprocess.STDOUT,timeout=60)
    answer=log.read_text();assert result.returncode==1 and 'AssertionError' in answer and 'FAILED (failures=' in answer,answer
    print('npm-drop-'+component,'caught by npm key assertion',flush=True)

# Go list's generated header must not dirty the checkout being audited.
text=(source/'setup-gate-inputs.py').read_text()
variant=text.replace("VALIDATION_FLAGS = ['-trimpath', '-buildvcs=false', '-buildmode=exe']",
                     "VALIDATION_FLAGS = ['-trimpath', '-buildvcs=false', '-buildmode=c-archive']")
assert variant!=text
module=scratch/'list-emits-header.py';module.write_text(variant)
with (scratch/'list-emits-header.log').open('wb') as output:
    result=subprocess.run(['python3',str(source/'test_gate_inputs.py'),'Archive'],
                          env=dict(os.environ,ADAMIC_GATE_INPUTS_MODULE=str(module),ADAMIC_SETUP_INTEGRATION='1',
                                   ADAMIC_GATE_ARCHIVE_PROOF_MODULE='gate-proof-header-'+str(os.getpid())),
                          stdout=output,stderr=subprocess.STDOUT,timeout=300)
answer=(scratch/'list-emits-header.log').read_text()
assert result.returncode==1 and 'Go list must not emit .h' in answer and 'AssertionError' in answer,answer
print('list-emits-header caught by clean-checkout assertion',flush=True)
