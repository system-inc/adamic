import pathlib,subprocess,time,json,os,signal,shutil
root=pathlib.Path.cwd();out=root/'review/test-audit/stage1-cohere-markdownblocks-structure_layout_shards';meta=[]
scratch=pathlib.Path('/tmp/u130-standalone');scratch.mkdir(exist_ok=True)
for directory in ['stage1/cohere/markdownblocks','stage1/cohere/markdowninline']:
 for name in subprocess.check_output(['git','ls-files',directory],text=True).splitlines():
  if name.endswith('.ts'):
   target=scratch/name;target.parent.mkdir(parents=True,exist_ok=True);target.write_bytes(subprocess.check_output(['git','show','d054e3578d4e53bef8cb8c9ab82d6ace69cdafeb:'+name]))
def execute(id,cmd,cwd=None):
 start=time.monotonic()
 with(out/(id+'.log')).open('w')as f:
  proc=subprocess.Popen(cmd,stdout=f,stderr=subprocess.STDOUT,start_new_session=True,cwd=cwd)
  try:code=proc.wait(timeout=90)
  except subprocess.TimeoutExpired:
   os.killpg(proc.pid,signal.SIGKILL);proc.wait();code=124
 m=dict(id=id,command=cmd,seconds=time.monotonic()-start,exit=code);meta.append(m);(out/'standalone-builds.json').write_text(json.dumps(meta,indent=2));print(m,flush=True);return code
compiler='/tmp/u130-adamic'
if execute('compiler-build',['go','build','-o',compiler,'./cmd/adamic']):raise SystemExit(1)
entries={'M1':'width_probe.ts','M2':'u130_heading.ts','M3':'u130_whitespace.ts','M4':'text_probe.ts','P1':'width_probe.ts','P2':'text_probe.ts','P3':'u130_document.ts','P4':'u130_whitespace.ts'}
drivers={
'u130_heading.ts': "import { DocumentArena, printDocument } from '../document.ts';\nimport { printHeading } from '../structure.ts';\nconst a = new DocumentArena();\nconsole.log(printDocument(a, printHeading(a, [a.text('x')], 1, false, '# x'), 120, 4));\n",
'u130_document.ts': "import { DocumentArena, printDocument } from '../document.ts';\nconst a = new DocumentArena();\nconsole.log(printDocument(a, a.text('x'), 120, 4));\n",
'u130_whitespace.ts': "import { DocumentArena } from '../document.ts';\nimport { printWhitespace } from '../whitespace.ts';\nconst a = new DocumentArena();\nconst token = { present: true, type: 'word', value: 'x', kind: 'non-cjk', cj: false, leading: false, trailing: false };\nconsole.log(a.node(printWhitespace(a, { value: ' ', proseWrap: 'always', link: false, previous: token, next: token, afterNext: token, ancestorKinds: [], ancestorSetext: [], samples: [] })).kind);\n"
}
for name,source in drivers.items():
 (out/name).write_text(source)
 (scratch/'stage1/cohere/markdownblocks/testdata'/name).write_text(source)
for id,entry in entries.items():
 patch=out/(id+'.diff')
 if execute(id+'-apply-check',['git','apply','--check',str(patch)],scratch):continue
 changed=subprocess.check_output(['git','apply','--numstat',str(patch)],text=True,cwd=scratch).strip().split('\t')[-1]
 before=(scratch/changed).read_bytes()
 try:
  subprocess.run(['git','apply',str(patch)],check=True,cwd=scratch)
  execute(id+'-standalone-build',[compiler,'build',str(scratch/'stage1/cohere/markdownblocks/testdata'/entry),'-o','/tmp/u130-'+id,'--sanitize'])
 finally:(scratch/changed).write_bytes(before)

for name in drivers:(scratch/'stage1/cohere/markdownblocks/testdata'/name).unlink()
