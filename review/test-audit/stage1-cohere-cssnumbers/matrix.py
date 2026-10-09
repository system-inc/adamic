import pathlib,subprocess,json,time,os,difflib
root=pathlib.Path('/tmp/u082');out=pathlib.Path('review/test-audit/stage1-cohere-cssnumbers')
while not (root/'timings-done').exists():time.sleep(1)
base=(out/'base.txt').read_text().strip();files=['stage1/cohere/cssnumbers/numbers.ts','stage1/cohere/cssnumbers/port_test.go','stage1/cohere/cssnumbers/shards_test.go'];orig={f:subprocess.check_output(['git','show',base+':'+f],text=True) for f in files};menu=json.loads((out/'menu.json').read_text());env=os.environ.copy();env['ADAMIC_CSSNUMBERS_LIBRARY']='/tmp/u082/library'
def invoke(name,cmd,env=env):
 start=time.monotonic()
 with (root/(name+'.log')).open('w') as log:p=subprocess.run(cmd,env=env,stdout=log,stderr=subprocess.STDOUT)
 (root/(name+'.meta')).write_text(json.dumps(dict(exit=p.returncode,wall=time.monotonic()-start,command=cmd,cache=env.get('ADAMIC_BUILD_CACHE_DIR'),mode=env.get('ADAMIC_AUDIT_MODE'))))
 return p.returncode
try:
 for m in menu:
  f=m['file'];pathlib.Path(f).write_text(orig[f].replace(m['old'],m['new'],1));(root/'standalone').mkdir(exist_ok=True)
  rc=invoke(m['id']+'-build',['timeout','90','go','run','./cmd/adamic','build','stage1/cohere/cssnumbers/main.ts','-o',str(root/'standalone'/m['id'])])
  pathlib.Path(f).write_text(orig[f])
  if rc:raise SystemExit(m['id']+' standalone build failed')
 scratch=orig.copy();f=files[0]
 scratch[f]=scratch[f].replace("import { adjustStrings }", "import { readTextFile } from 'adamic';\nconst auditRead = readTextFile('/tmp/u082/selector');\nconst auditChoice = auditRead.kind === 'Error' ? '' : auditRead.text;\nimport { adjustStrings }",1)
 replacements={ 'M1':"return auditChoice === 'M1' ? 'q' : 'Q';",'M2':"(auditChoice === 'M2' ? code < 57 : code <= 57)",'M3':"end > dot + (auditChoice === 'M3' ? 2 : 1)",'M4':"(auditChoice === 'M4' ? code > 128 : code >= 128)"}
 for m in menu:scratch[f]=scratch[f].replace(m['old'],replacements[m['id']],1)
 scratch[f]=scratch[f].replace('export function formatValue(value: string, single: boolean): string {',"export function formatValue(value: string, single: boolean): string {\n    if(auditChoice === 'P1') return '';",1)
 p=files[1];old='\t\tcommand := numbersSetupCommand(t, "go", "build", "-overlay="+overlayPath, "-o", filepath.Join(dir, "go-printer"), filepath.Join(cohere, "cmd/adamic_stage_one/main.go"))\n\t\tcommand.Dir = cohere\n\t\tif output, err := childguard.CombinedOutput(command, childguard.Options{}); err != nil {\n\t\t\treturn fmt.Errorf("Go bridge: %w\\n%s", err, output)\n\t\t}'
 assert old in scratch[p]
 scratch[p]=scratch[p].replace(old,'\t\tif os.Getenv("ADAMIC_AUDIT_MODE") != "S1" {\n'+old+'\n\t\t}',1)
 scratch[p]=scratch[p].replace('if bytes.Equal(r.stdout, want.stdout) {','if os.Getenv("ADAMIC_AUDIT_MODE") == "W2" || bytes.Equal(r.stdout, want.stdout) {',1)
 p=files[2]
 scratch[p]=scratch[p].replace('\treturn units\n}', '\tif os.Getenv("ADAMIC_AUDIT_MODE") == "S4" { return units[1:] }; return units\n}',1)
 scratch[p]=scratch[p].replace('func numbersOutputError(unit numbersUnit, side string, got, want []byte) error {','func numbersOutputError(unit numbersUnit, side string, got, want []byte) error {\n\tif os.Getenv("ADAMIC_AUDIT_MODE") == "W1" { return nil }',1)
 marker='}, func(dir string) error {\n\t\tloaded, err := load.Load([]string{main})'
 assert marker in scratch[p]
 scratch[p]=scratch[p].replace(marker,'}, func(dir string) error {\n\t\tif os.Getenv("ADAMIC_AUDIT_MODE") == "S2" { return nil }\n\t\tloaded, err := load.Load([]string{main})',1)
 marker='}, func(dir string) error {\n\t\tsource, err := os.ReadFile(program.c)'
 assert marker in scratch[p]
 scratch[p]=scratch[p].replace(marker,'}, func(dir string) error {\n\t\tif os.Getenv("ADAMIC_AUDIT_MODE") == "S3" { return nil }\n\t\tsource, err := os.ReadFile(program.c)',1)
 for p,s in scratch.items():pathlib.Path(p).write_text(s)
 subprocess.run(['gofmt','-w',files[1],files[2]],check=True)
 with (out/'scratch-switch.diff').open('w') as log:subprocess.run(['git','diff','--',*files],stdout=log,check=True)
 env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u082/cache/switch'
 def matrix(mid,pattern='.',mode='',suffix=''):
  (root/'selector').write_text(mid if mid in ['M1','M2','M3','M4','P1'] else '')
  selected=env.copy();selected['ADAMIC_AUDIT_MODE']=mode
  if mode in ['S1','S2','S3']:selected['ADAMIC_BUILD_CACHE_DIR']='/tmp/u082/cache/'+mode
  return invoke(mid+suffix,['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/cssnumbers/','-run',pattern],selected)
 if matrix('clean-switch'):raise SystemExit('clean switched baseline failed')
 for mid in ['M1','M2','M3','M4','P1']:matrix(mid)
 matrix('W1','^TestCSSNumbersPlantedDisagreement$','W1')
 matrix('W2','^TestCSSNumbers_(350|351|352)$','W2')
 matrix('S1','^TestProduct_CSSNumbersGo(Oracle|Answers)$','S1')
 matrix('S2','^TestProduct_CSSNumbers(Port|Mutant[012])Lowered$','S2')
 matrix('S3','^TestProduct_CSSNumbers((Port|Mutant[012])Native|FastNative)$','S3')
 matrix('S4','^(TestCSSNumbersUnion|TestCSSNumbers_Setup)$','S4')
 for mid in ['S1','S2','S3']:
  products=list((root/'cache'/mid).rglob('go-printer' if mid=='S1' else 'program.c' if mid=='S2' else 'port'))
  (root/(mid+'-artifacts.json')).write_text(json.dumps([str(p) for p in products],indent=2))
finally:
 for f,s in orig.items():pathlib.Path(f).write_text(s)
(root/'matrix-done').write_text('done')
