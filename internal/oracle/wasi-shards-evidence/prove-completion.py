import importlib.util, json
from pathlib import Path
spec=importlib.util.spec_from_file_location('verify','internal/oracle/verify_wasi_shards.py')
v=importlib.util.module_from_spec(spec); spec.loader.exec_module(v)
root=Path('/workspace/scratch/oracle-split')
expected=v.manifest(v.events(root/'proof.json'))
logs=[v.events(root/f'shard-{i:03}.json') for i in range(8)]
print(v.verify(expected,logs))
for i in range(8):
 try: v.verify(expected,logs[:i]+logs[i+1:])
 except ValueError as error:
  assert f'missing shard TestWASIAgreesWithNode/shard-{i:03}' in str(error),str(error)
  print('omitted receipt:',error)
 else: raise AssertionError('absent shard accepted')
v.self_test()
