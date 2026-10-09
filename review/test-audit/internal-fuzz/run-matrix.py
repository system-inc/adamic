import pathlib,json,subprocess,time,os,difflib
P=pathlib.Path('review/test-audit/internal-fuzz')
plan=json.loads((P/'mutation-plan.json').read_text())
original={m['file']:pathlib.Path(m['file']).read_text() for m in plan}
wrapped={
'M01':'rand.NewPCG(auditUint("M01", seed, 0), 0x61646d6963)',
'M02':'return auditBool("M02", !g.without[feature], g.without[feature])',
'M03':'}, 1, auditInt("M03", 3000, 0))',
'M04':'return []namedList{{node.ArgumentList(), auditBool("M04", true, false)}}',
'M05':'if auditBool("M05", s.Exact, !s.Exact) {',
'M06':'outcome.Native.ExitCode == auditInt("M06", 70, 71)',
'M07':'auditBool("M07", status.Signal() == syscall.SIGXCPU, status.Signal() == syscall.SIGTERM)',
'M08':'case auditBool("M08", !bytes.Equal(expected.Stdout, actual.Stdout), bytes.Equal(expected.Stdout, actual.Stdout)):',
'M09':'\t\t\tif !auditMutant("M09") { cut.start = items.ends[start-1] }\n',
'M10':'\tif !auditMutant("M10") { g.names++ }\n',
'M11':'switch g.seed % auditUint("M11", 8, 7) {',
'M12':'return text(Number, auditString("M12", "0", "1"))',
'M13':'switch (g.seed / 10) % auditUint("M13", 5, 4) {',
'M14':'return sliceBytes >= auditInt("M14", 64, 63) && sliceBytes >= ownerBytes/4',
'M15':'\t\t\t\tif !auditMutant("M15") { s.current = candidates[index] }\n',
'M16':'&& auditBool("M16", candidate.Has(signature), true)',
}
probes=[
('PGenerate','internal/fuzz/generate.go','func Generate(seed uint64) *Program {','return nil'),
('PWithout','internal/fuzz/generate.go','func GenerateWithout(seed uint64, without []string) *Program {','return nil'),
('PFeatures','internal/fuzz/generate.go','func GenerateFeatures(seed uint64, without []string, with []string) *Program {','return nil'),
('PShrink','internal/fuzz/shrink.go','func Shrink(program *Program, key string, try func(*Program) Outcome) *Program {','return &Program{}'),
('PSource','internal/fuzz/source.go','func newSourceTree(path string, text string) *sourceTree {','return nil'),
('PReduce','internal/fuzz/reduce.go','func Reduce(path string, source string, signature Signature, original Observation, observe func(source string, slot int) Observation, workers int, budget int) Reduction {','return Reduction{}'),
('PJudge','internal/fuzz/run.go','func (c *Checkout) judge(outcome Outcome, binary string, directory string) Outcome {','return Outcome{}'),
('PExecute','internal/fuzz/run.go','func execute(directory string, environment []string, limit time.Duration, name string, arguments ...string) Run {','return Run{}'),
('PTry','internal/fuzz/run.go','func (c *Checkout) Try(source string, directory string) Outcome {','return Outcome{}'),
('PShareCuts','internal/fuzz/generate.go','func allShareCuts() []shareCut {','return nil'),
]
texts=original.copy()
for m in plan: texts[m['file']]=texts[m['file']].replace(m['old'],wrapped[m['id']])
for id,file,entry,ret in probes:
 old=original[file]; new=old.replace(entry,entry+'\n\t'+ret)
 (P/(id+'.diff')).write_text(''.join(difflib.unified_diff(old.splitlines(True),new.splitlines(True),fromfile='a/'+file,tofile='b/'+file)))
 texts[file]=texts[file].replace(entry,entry+'\n\tif auditMutant("'+id+'") { '+ret+' }')
helper=pathlib.Path('internal/fuzz/audit_switch.go')
helper.write_text('package fuzz\nimport "os"\nfunc auditMutant(id string) bool { return os.Getenv("ADAMIC_MUTANT")==id }\nfunc auditBool(id string,a,b bool) bool { if auditMutant(id) {return b};return a }\nfunc auditInt(id string,a,b int) int {if auditMutant(id){return b};return a}\nfunc auditUint(id string,a,b uint64) uint64 {if auditMutant(id){return b};return a}\nfunc auditString(id string,a,b string) string {if auditMutant(id){return b};return a}\n')
results=[]
rows=[s for s in (P/'list.log').read_text().splitlines() if s.startswith('Test')]
def run(id,regex='.',suffix=''):
 env=os.environ.copy();env['ADAMIC_MUTANT']=id;env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u020/cache/'+id
 log=P/(id+suffix+'.log'); start=time.monotonic()
 with log.open('w') as f: r=subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/fuzz/','-run',regex],stdout=f,stderr=subprocess.STDOUT,env=env)
 result=dict(id=id,regex=regex,log=str(log),status=r.returncode,wall=time.monotonic()-start)
 results.append(result);(P/'runs.json').write_text(json.dumps(results,indent=2)+'\n')
 return log.read_text(),r.returncode
try:
 for file,text in texts.items():pathlib.Path(file).write_text(text)
 subprocess.run(['gofmt','-w',*texts, str(helper)],check=True)
 with (P/'switch-vet.log').open('w') as f: subprocess.run(['go','vet','./internal/fuzz/'],stdout=f,stderr=subprocess.STDOUT,check=True)
 for id in [m['id'] for m in plan]+[p[0] for p in probes]:
  text,status=run(id)
  if 'panic:' in text or status==124:
   for row in rows:
    run(id,'^'+row+'$','-'+row)
finally:
 for file,text in original.items():pathlib.Path(file).write_text(text)
 helper.unlink(missing_ok=True)
