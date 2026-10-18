from pathlib import Path
import os,json,subprocess,time
root=Path.cwd();out=root/'review/compiler/loom-r-records-ac-m';logs=out/'source-mutants';logs.mkdir(exist_ok=True)
script=root/'stage3/records-maplike-rebuild/run-mutants.py'
prefix=script.read_text().split('for name, files, test, witness in mutants:')[0]
# Main now preserves nullable references with an explicit dictionary metadata stop.
# Disable that same stop, rather than transplant the old representation expression.
prefix=prefix.replace("nullable = replace('expression.go', ' || of == ir.Record || of == ir.String', ' || of == ir.String')", "nullable = replace('expression.go', 'if of == ir.Record {', 'if false {')")
os.environ['RECORDS_REBUILD_MUTANT_LOGS']=str(logs)
space={'__file__':str(script)};exec(compile(prefix,str(script),'exec'),space)
mutants=space['mutants'];early=space['early'];text=space['text']
for name,function,statement,test,witness in [
 ('index-read-type','recordReadType','return read, nil','TestNamedRecordReadTypes','named read used index type'),
 ('unrestricted-write','recordNamedWrite','return nil','TestNamedRecordRefusals','want NotYet named property named'),
 ('erase-named-contract','sameRecordNamedContracts','return true','TestNamedRecordRefusals','want NotYet seen as')]:
 mutants.append((name,{'records_named.go':early('records_named.go',function,statement)},test,witness))
before='if l.includesNull(t) && !dynamicObjectType(t) {'
mutants.append(('nullable-slot',{'records_named.go':text('records_named.go').replace(before,'if false {',1)},'TestNamedRecordRefusals','slot representation'))
results=[]
for name,files,test,witness in mutants:
 replacements={}
 for file,source in files.items():
  assert source!=text(file),name
  target=logs/(name+'-'+file+'.txt');target.write_text(source);replacements[str(root/'internal/lower'/file)]=str(target)
 overlay=logs/(name+'.json');overlay.write_text(json.dumps({'Replace':replacements})+'\n')
 command=['go','test','-overlay',str(overlay),'./internal/lower','-run','^'+test+'$','-count=1','-v','-timeout','90s','-parallel','4'];start=time.monotonic()
 with (logs/(name+'.log')).open('w') as log:r=subprocess.run(command,stdout=log,stderr=subprocess.STDOUT,timeout=180)
 observed=(logs/(name+'.log')).read_text();caught=r.returncode==1 and witness in observed
 result=dict(name=name,command=command,exit=r.returncode,catcher=test,witness=witness,caught=caught,seconds=round(time.monotonic()-start,3));results.append(result);(out/'source-mutants.json').write_text(json.dumps(results,indent=2)+'\n');print(json.dumps(result),flush=True)
 assert caught,(name,observed)
