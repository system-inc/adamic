import pathlib,subprocess,difflib,json,time,os
D=pathlib.Path('review/test-audit/stage1-cohere-estree-gaps');os.environ['ADAMIC_ESTREE_LIBRARY']='/tmp/u086/library'
helper='stage1/cohere/estree/audit_selector.ts';ht="import { readTextFile } from 'adamic';\nconst file = readTextFile('/tmp/u086/selector');\nconst selected = file.kind === 'Error' ? '' : file.text.trim();\nexport function active(id: string): boolean { return selected === id; }\n"
probes=[('PMain','stage1/cohere/estree/main.ts',"function run(path: string): void {","function run(path: string): void {\n    if(active('PMain')) { return; }",'main.ts'),('PParser','stage1/typescript/parser/parser.ts','file(): number {',"file(): number {\n        if(active('PParser')) { return 0; }",'gaps/parserRecovery.ts'),('PUtf8','internal/native/runtime/utf8.c','return (double)text->length;','if (getenv("ADAMIC_MUTANT") != NULL && strcmp(getenv("ADAMIC_MUTANT"), "PUtf8") == 0) return 0;\n\treturn (double)text->length;','gaps/rawInput.ts'),('PLower','internal/lower/lower.go','func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {','func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {\nif os.Getenv("ADAMIC_MUTANT") == "PLower" { return nil,nil }',None)]
records=[]
for id,f,old,new,entry in probes:
 s=pathlib.Path(f).read_text();n=s.replace(old,new);extra={}
 if id in ['PMain','PParser']:
  path='./audit_selector.ts' if id=='PMain' else '../../cohere/estree/audit_selector.ts';n="import { active } from '"+path+"';\n"+n;extra[helper]=ht
 if id=='PLower':n=n.replace('"context"','"context"\n"os"')
 patch=''.join(difflib.unified_diff(s.splitlines(True),n.splitlines(True),fromfile='a/'+f,tofile='b/'+f))
 for ef,et in extra.items():patch+=''.join(difflib.unified_diff([],et.splitlines(True),fromfile='/dev/null',tofile='b/'+ef))
 (D/(id+'.diff')).write_text(patch)
 pathlib.Path(f).write_text(n)
 for ef,et in extra.items():pathlib.Path(ef).write_text(et)
 cmd=['timeout','90','go','vet','./internal/lower/'] if id=='PLower' else ['timeout','90','go','run','./cmd/adamic','build','stage1/cohere/estree/'+entry,'-o','/tmp/u086/validate-'+id,'--sanitize']
 env=os.environ.copy();env.update(ADAMIC_NATIVE_SPLIT='1',ADAMIC_BUILD_CACHE_DIR='/tmp/u086/cache/validate-'+id)
 t=time.monotonic()
 try:
  with (D/('validate-'+id+'.log')).open('w') as out:rc=subprocess.call(cmd,stdout=out,stderr=subprocess.STDOUT,env=env)
 finally:
  pathlib.Path(f).write_text(s)
  for ef in extra:pathlib.Path(ef).unlink()
 records.append(dict(id=id,command=' '.join(cmd),wall_seconds=round(time.monotonic()-t,3),exit=rc));(D/'probe-builds.json').write_text(json.dumps(records,indent=2));print(id,rc,flush=True)
 if rc:raise RuntimeError(id)
