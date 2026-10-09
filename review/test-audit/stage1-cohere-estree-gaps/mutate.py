import pathlib,json,subprocess,time,os,difflib
D=pathlib.Path('review/test-audit/stage1-cohere-estree-gaps');plan=json.loads((D/'plan.json').read_text());original={m['file']:pathlib.Path(m['file']).read_text() for m in plan}; original['stage1/cohere/estree/main.ts']=pathlib.Path('stage1/cohere/estree/main.ts').read_text(); original['internal/lower/lower.go']=pathlib.Path('internal/lower/lower.go').read_text();commands=[];os.environ['ADAMIC_ESTREE_LIBRARY']='/tmp/u086/library';selector=pathlib.Path('/tmp/u086/selector');selector.write_text('')
def run(label,args,env=None):
 e=os.environ.copy();e.update(env or {});t=time.monotonic()
 with (D/(label+'.log')).open('w') as out:rc=subprocess.call(args,stdout=out,stderr=subprocess.STDOUT,env=e)
 commands.append(dict(label=label,command=' '.join(args),env=env or {},wall_seconds=round(time.monotonic()-t,3),exit=rc));(D/'mutant-commands.json').write_text(json.dumps(commands,indent=2));print(label,rc,flush=True);return rc

def diff(id,file,old,new):
 (D/(id+'.diff')).write_text(''.join(difflib.unified_diff(old.splitlines(True),new.splitlines(True),fromfile='a/'+file,tofile='b/'+file)))
for m in plan:
 f=m['file'];s=original[f];new=s.replace(m['old'],m['new']);diff(m['id'],f,s,new)
 if m['kind']!='production':continue
 pathlib.Path(f).write_text(new)
 if m['id'] in ['M01','M02','M03','M04']:
  path='stage1/cohere/estree/main.ts' if m['id'] in ['M01','M04'] else 'stage1/cohere/estree/gaps/parserRecovery.ts' if m['id']=='M02' else 'stage1/cohere/estree/gaps/rawInput.ts'
  args=['timeout','90','go','run','./cmd/adamic','build',path,'-o','/tmp/u086/validate-'+m['id'],'--sanitize'];env={'ADAMIC_NATIVE_SPLIT':'1','ADAMIC_BUILD_CACHE_DIR':'/tmp/u086/cache/validate-'+m['id']}
 else:args=['timeout','90','go','vet','./internal/lower/'];env={}
 rc=run('validate-'+m['id'],args,env);pathlib.Path(f).write_text(s)
 if rc:raise RuntimeError('standalone mutant did not build '+m['id'])
# Compile fixed instrumentation once per product; selectors alter only existing code choices.
helper='stage1/cohere/estree/audit_selector.ts';helperText="import { readTextFile } from 'adamic';\nconst file = readTextFile('/tmp/u086/selector');\nconst selected = file.kind === 'Error' ? '' : file.text.trim();\nexport function active(id: string): boolean { return selected === id; }\n";pathlib.Path(helper).write_text(helperText)
for f,s in original.items():
 if f=='stage1/cohere/estree/jsxConvert.ts':s="import { active } from './audit_selector.ts';\n"+s.replace("'JSXIdentifier'","(active('M01') ? 'JSXName' : 'JSXIdentifier')")
 if f=='stage1/cohere/estree/protocol.ts':s="import { active } from './audit_selector.ts';\n"+s.replace('new DumpFrame(property.value.node, level + 1, -1)',"new DumpFrame(property.value.node, active('M04') ? level + 2 : level + 1, -1)")
 if f=='stage1/typescript/parser/parser.ts':
  s="import { active } from '../../cohere/estree/audit_selector.ts';\n"+s.replace("const members: number[] = [];\n        while(this.kind() !== 'CloseBraceToken')", "const members: number[] = [];\n        while(active('M02') ? this.kind() === 'CloseBraceToken' : this.kind() !== 'CloseBraceToken')")
  s=s.replace('file(): number {',"file(): number {\n        if(active('PParser')) { return 0; }")
 if f=='internal/native/runtime/utf8.c':s=s.replace('return (double)text->length;', 'const char *audit = getenv("ADAMIC_MUTANT");\n\tif (audit != NULL && strcmp(audit, "PUtf8") == 0) return 0;\n\treturn (double)text->length + (audit != NULL && strcmp(audit, "M03") == 0 ? 1 : 0);')
 if f=='internal/lower/diagnostics.go':
  for m in plan:
   if m['file']==f:s=s.replace(m['old'],'if auditSelected("'+m['id']+'") { '+m['new']+' }; '+m['old'] if m['id']!='M06' else 'Where: auditWhere(l.program.Where(node)), What: what')
 if f=='stage1/cohere/estree/main.ts':s="import { active } from './audit_selector.ts';\n"+s.replace('function run(path: string): void {',"function run(path: string): void {\n    if(active('PMain')) { return; }")
 if f=='internal/lower/lower.go':s=s.replace('func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {','func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {\nif auditSelected("PLower") {return nil,nil}')
 pathlib.Path(f).write_text(s)
gohelper=pathlib.Path('internal/lower/audit_selector.go');gohelper.write_text('package lower\nimport "os"\nfunc auditSelected(id string) bool{return os.Getenv("ADAMIC_MUTANT")==id}\nfunc auditWhere(s string) string{if auditSelected("M06"){return s+":0"};return s}\n')
# Preserve the combined instrumentation for transparent reproduction, then remove it.
switch=''
for f,s in original.items():switch+=''.join(difflib.unified_diff(s.splitlines(True),pathlib.Path(f).read_text().splitlines(True),fromfile='a/'+f,tofile='b/'+f))
for f in [helper,str(gohelper)]:switch+=''.join(difflib.unified_diff([],pathlib.Path(f).read_text().splitlines(True),fromfile='/dev/null',tofile='b/'+f))
(D/'switch.diff').write_text(switch)
assert run('switch-vet',['timeout','90','go','vet','./internal/lower/','./stage1/cohere/estree/'])==0
# Bounded clean execution verifies the switch itself before selectors are changed.
for label,regex in [('switch-clean-jsx','^TestJSXAgreement(_[0-9]{3})?$'),('switch-clean-parser','^TestParserRecoveryGap$'),('switch-clean-raw','^TestRawInputGap$')]:
 assert run(label,['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/estree/','-run',regex],{'ADAMIC_BUILD_CACHE_DIR':'/tmp/u086/cache/switch'})==0,label
for m in plan:
 if m['kind']!='production':continue
 id=m['id'];selector.write_text(id)
 rows=m['matrix_rows'];parts=['TestJSXAgreement(_[0-9]{3})?' if r=='TestJSXAgreement family' else r for r in rows];regex='^('+'|'.join(parts)+')$'
 env={'ADAMIC_MUTANT':id,'ADAMIC_BUILD_CACHE_DIR':'/tmp/u086/cache/'+(id if id in ['M05','M06','M07'] else 'switch')}
 run(id,['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/estree/','-run',regex],env)
for id,rows in [('PMain',['TestJSXAgreement(_[0-9]{3})?']),('PParser',['TestParserRecoveryGap']),('PUtf8',['TestRawInputGap']),('PLower',['TestPostfixValueGap','TestMethodReplacementGap','TestInterfaceDefaultGap','TestInterfaceTypeMethodGap'])]:
 selector.write_text(id);run(id,['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/estree/','-run','^('+'|'.join(rows)+')$'],{'ADAMIC_MUTANT':id,'ADAMIC_BUILD_CACHE_DIR':'/tmp/u086/cache/'+('PLower' if id=='PLower' else 'switch')})
selector.write_text('')
for f,s in original.items():pathlib.Path(f).write_text(s)
pathlib.Path(helper).unlink();gohelper.unlink()
# Construction and witness checks use the explicitly allowed scratch harness changes.
for m in plan:
 if m['kind']=='production':continue
 f=m['file'];s=original[f];pathlib.Path(f).write_text(s.replace(m['old'],m['new']));assert run('vet-'+m['id'],['timeout','90','go','vet','./stage1/cohere/estree/'])==0
 regex='^('+'|'.join(m['matrix_rows'])+')$';run(m['id'],['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/estree/','-run',regex],{'ADAMIC_BUILD_CACHE_DIR':'/tmp/u086/cache/'+m['id']});pathlib.Path(f).write_text(s)
