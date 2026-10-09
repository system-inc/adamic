import json,pathlib,subprocess,os
r=pathlib.Path(__file__).resolve().parents[3];d=pathlib.Path(os.environ.get('CONSTANT_MUTANTS','/tmp/adamic-constant-mutants'));d.mkdir(exist_ok=True)
p=r/'internal/native/constant_tables.go';original=p.read_text()
mutants=[
 ('number-element-order','\tdata := strings.Join(values, ",\\n\\t")','\tif len(values) >= 2 { values[0], values[1] = values[1], values[0] }\n\tdata := strings.Join(values, ",\\n\\t")','native'),
 ('record-element-order','\tfirst := rows[0]','\trows[2], rows[3] = rows[3], rows[2]\n\tfirst := rows[0]','native'),
 ('runtime-expression-as-constant','\treturn 0, false\n}\n\nfunc (e *emitter) numberTableCopy','\tif expression.Type() == ir.Number { return 0, true }; return 0, false\n}\n\nfunc (e *emitter) numberTableCopy','native'),
 ('lose-negative-zero','value = -value','value = -value; if value == 0 { value = 0 }','native'),
]
env=os.environ.copy();env.update(ADAMIC_GATE_UNCACHED='1',ADAMIC_NATIVE_SPLIT='1',ADAMIC_NATIVE_JOBS='3')
results=[]
for name,anchor,replacement,expected in mutants:
 assert original.count(anchor)==1,(name,original.count(anchor))
 path=d/(name+'.go');path.write_text(original.replace(anchor,replacement))
 overlay=d/(name+'.json');overlay.write_text(json.dumps({'Replace':{str(p):str(path)}}))
 with (d/(name+'.log')).open('wb') as log:
  result=subprocess.run(['go','test','-overlay='+str(overlay),'./internal/native','-run','^TestConstantTablesMatchNode$','-v','-count=1','-timeout=10m'],cwd=r,env=env,stdout=log,stderr=subprocess.STDOUT)
 output=(d/(name+'.log')).read_text()
 assert result.returncode!=0 and 'native ' in output and 'Node ' in output and 'error:' not in output and 'build failed' not in output,(name,output)
 print(name,'caught by Node output comparison',flush=True)
 results.append({'mutant':name,'exit':result.returncode,'caught_by':'TestConstantTablesMatchNode: native stdout differs from Node; compiled and ran'})
(d/'results.json').write_text(json.dumps(results,indent=2)+'\n')
