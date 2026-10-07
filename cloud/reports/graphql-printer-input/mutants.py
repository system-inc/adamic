import contextlib,importlib.util,io,json,pathlib,shutil,tempfile,unittest
repo=pathlib.Path('/workspace/adamic');reports=repo/'cloud/reports/graphql-printer-input'
s=importlib.util.spec_from_file_location('npm',repo/'cloud/setup-gate-npm.py');npm=importlib.util.module_from_spec(s);s.loader.exec_module(npm)
inputs=dict(lock='lock',manifest='manifest',bootstrap='bootstrap',helper='helper',node='node')
for dropped in inputs:
 def mutant(**values):
  values.pop(dropped);return npm.digest(json.dumps(values,sort_keys=True).encode())
 try:assert mutant(**inputs)!=mutant(**dict(inputs,**{dropped:'changed'})),dropped
 except AssertionError as failure:print('drop-'+dropped+': key assertion caught mutant')
 else:raise AssertionError('mutant survived')
s=importlib.util.spec_from_file_location('tests',repo/'cloud/test_gate_inputs.py');m=importlib.util.module_from_spec(s);s.loader.exec_module(m)
with (reports/'missing-export.log').open('w') as log:
 value=m.gate.VARIABLES.pop('ADAMIC_GRAPHQL_PRETTIER');result=unittest.TextTestRunner(stream=log).run(unittest.TestSuite([m.Inputs('test_shared_prettier_exports_and_typescript_pin')]));assert len(result.failures)==1;m.gate.VARIABLES['ADAMIC_GRAPHQL_PRETTIER']=value
print('missing GraphQL export: caught')
with tempfile.TemporaryDirectory(dir='/tmp/adamic-gate') as temporary:
 root=pathlib.Path(temporary);dest=root/'gate-inputs/css-printer';shutil.copytree(repo/'cloud/gate-inputs/css-printer',dest);file=dest/'package.json';data=json.loads(file.read_text());data['dependencies']['graphql']='17.0.1';file.write_text(json.dumps(data));m.SOURCE=root
 with (reports/'wrong-pin.log').open('w') as log:
  result=unittest.TextTestRunner(stream=log).run(unittest.TestSuite([m.Inputs('test_shared_formatter_pins_and_integrity')]));assert len(result.failures)==1
print('graphql 17.0.1: exact pin assertion caught mutant')
