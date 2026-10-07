#!/usr/bin/env python3
import os,pathlib,subprocess,tempfile
root=pathlib.Path(__file__).resolve().parents[3]; source=root/'cloud'; reports=pathlib.Path(__file__).resolve().parent
mutants=[('leak-archive','setup-gate-inputs.py',"selected = phase == 'env' or (phase == 'env-archive') == archive_input","selected = True",'ADAMIC_GATE_INPUTS_MODULE','test_environment_modes'),('leave-stale-variable','setup-gate-inputs.py',"print('unset ' + name)","print('export ' + name + '=stale')",'ADAMIC_GATE_INPUTS_MODULE','test_environment_modes'),('archive-only-ignored','setup-darwin.py','gate_archive = flags.gate_inputs or flags.gate_archive','gate_archive = flags.gate_inputs','ADAMIC_ARCHIVE_DARWIN_MODULE','test_darwin_selection_calls_and_exports'),('default-archive-dropped','setup-darwin.py','gate_archive = flags.gate_inputs or flags.gate_archive','gate_archive = flags.gate_archive','ADAMIC_ARCHIVE_DARWIN_MODULE','test_darwin_selection_calls_and_exports'),('no-archive-builds-archive','setup-darwin.py','gate_archive = flags.gate_inputs or flags.gate_archive','gate_archive = flags.gate_inputs or flags.gate_archive or flags.gate_inputs_no_archive','ADAMIC_ARCHIVE_DARWIN_MODULE','test_darwin_selection_calls_and_exports')]
for name,file,old,new,variable,test in mutants:
    with tempfile.TemporaryDirectory(prefix='archive-mode-mutant-') as temporary:
        directory=pathlib.Path(temporary)
        for helper in list(source.glob('*.py'))+[source/'node-pin.json']:(directory/helper.name).symlink_to(helper)
        target=directory/file
        if target.is_symlink():target.unlink()
        text=(source/file).read_text();assert old in text,(name,old)
        target.write_text(text.replace(old,new,1))
        env=os.environ.copy();env[variable]=str(target)
        with (reports/f'mutant-{name}.log').open('w') as log:result=subprocess.run(['python3','-m','unittest',f'cloud.test_archive_modes.ArchiveModes.{test}'],cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT)
        assert result.returncode!=0,name
        print(f'{name}: caught by {test}',flush=True)
import sys,unittest
from unittest.mock import patch
sys.path.insert(0,str(root))
from cloud import test_archive_modes as modes
for name,old,new in [('conflict-exits-success', ' >&2\n exit 2\nfi', ' >&2\n exit 0\nfi'),('conflict-diagnostic-lost', ' conflicts with ', ' accepts ')]:
    with tempfile.TemporaryDirectory(prefix='archive-conflict-mutant-') as temporary:
        directory=pathlib.Path(temporary);(directory/'cloud').mkdir();(directory/'internal').symlink_to(root/'internal')
        text=(source/'setup.sh').read_text();assert old in text
        (directory/'cloud/setup.sh').write_text(text.replace(old,new,1))
        with (reports/f'mutant-{name}.log').open('w') as log,patch.object(modes,'SOURCE',directory/'cloud'):
            result=unittest.TextTestRunner(stream=log).run(unittest.TestSuite([modes.ArchiveModes('test_conflicting_flags_fail_before_preparation')]))
        assert not result.wasSuccessful(),name
        print(f'{name}: caught by test_conflicting_flags_fail_before_preparation',flush=True)
