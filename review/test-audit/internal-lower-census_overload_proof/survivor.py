import pathlib,subprocess,os,json,time
p=pathlib.Path('review/test-audit/internal-lower-census_overload_proof');(p/'probes').mkdir(exist_ok=True)
sources={'dead':pathlib.Path('internal/oracle/testdata/census_boolean_dead_branch.a').read_text(),'contradiction':"function dead(flag: boolean): void { if (flag && !flag) { console.log(flag && flag); } } dead(false);\n",'undefined':"function mix(x: boolean | undefined): void { if (x === undefined) { console.log((x && x) || x); } } mix(undefined);\n"};records=[]
for name,s in sources.items():
 source=p/'probes'/(name+'.a.txt');source.write_text(s);actual=pathlib.Path('/tmp/u027-'+name+'.a');actual.write_text(s)
 outputs=[]
 for id in ['', 'M11']:
  env=os.environ.copy();env['ADAMIC_MUTANT']=id;label=name+'.'+(id or 'baseline');cmd=['timeout','90','/tmp/u027-adamic','c',str(actual)];r=subprocess.run(cmd,env=env,capture_output=True)
  (p/'probes'/(label+'.stdout')).write_bytes(r.stdout);(p/'probes'/(label+'.stderr')).write_bytes(r.stderr);outputs.append((r.returncode,r.stdout,r.stderr));records.append(dict(probe=name,mutant=id or 'baseline',command='ADAMIC_MUTANT='+id+' '+' '.join(cmd),exit=r.returncode,stdout_bytes=len(r.stdout),stderr=r.stderr.decode(errors='replace')))
 print(name,'changed',outputs[0]!=outputs[1],'exit',outputs[0][0],outputs[1][0],flush=True)
(p/'survivor-probes.json').write_text(json.dumps(records,indent=2))
