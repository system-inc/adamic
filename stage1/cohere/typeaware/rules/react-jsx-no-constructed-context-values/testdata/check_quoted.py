from pathlib import Path
import os, subprocess, time
rule=Path(__file__).resolve().parents[1]
out=Path(os.environ.get('ADAMIC_WAVE15_QUOTED_ARTIFACTS','/tmp/wave15-quoted'));out.mkdir(parents=True,exist_ok=True)
compiler=os.environ.get('ADAMIC_WAVE15_SIXTH_COMPILER','/tmp/wave15-d65-adamic')
def run(args,label):
 start=time.perf_counter_ns(); p=subprocess.run(args,capture_output=True)
 (out/(label+'.stdout')).write_bytes(p.stdout);(out/(label+'.stderr')).write_bytes(p.stderr)
 assert p.returncode==0,(label,p.returncode,p.stderr[:2000]);return p.stdout,time.perf_counter_ns()-start
truth,_=run(['go','run',str(rule/'testdata/quoted_oracle.go')],'go')
assert len(truth.splitlines())==1112071
for sanitize in [False,True]:
 label='quoted-asan' if sanitize else 'quoted'
 run([compiler,'build',str(rule/'quoted_controls.a'),'-o',str(out/label)]+(['--sanitize'] if sanitize else []),label+'-build')
 actual,elapsed=run([str(out/label)],label)
 assert actual==truth,(label,'Go quoted bytes differ')
 print(label,'PASS',len(actual),'bytes, 1112064 scalar values and seven strings; ns',elapsed,flush=True)
javascript,_=run([compiler,'js',str(rule/'quoted_controls.a')],'js-build')
(out/'quoted.mjs').write_bytes(javascript)
package=out/'node_modules/adamic';package.mkdir(parents=True,exist_ok=True)
(package/'package.json').write_text('{"type":"module","exports":"./index.mjs"}')
(package/'index.mjs').write_bytes((rule.parents[4]/'oracle/adamic.mjs').read_bytes())
actual,_=run(['node',str(out/'quoted.mjs')],'node');assert actual==truth
print('Emitted JavaScript on Node matches independent Go bytes',flush=True)
p=subprocess.run([str(out/'quoted'),'--surrogate'],capture_output=True)
assert p.returncode==70 and b'unpaired surrogate' in p.stderr
(out/'surrogate.stderr').write_bytes(p.stderr)
original=(rule/'quoted_name.a').read_text()
(rule/'quoted_mutant.a').write_text(original.replace(r"result += '\\u' + hex(point, 4)",r"result += '\\x' + hex(point, 4)",1))
assert (rule/'quoted_mutant.a').read_text()!=original
(rule/'quoted_mutant_controls.a').write_text((rule/'quoted_controls.a').read_text().replace("'./quoted_name.a'","'./quoted_mutant.a'"))
try:
 run([compiler,'build',str(rule/'quoted_mutant_controls.a'),'-o',str(out/'mutant')],'mutant-build')
 actual,_=run([str(out/'mutant')],'mutant');assert actual!=truth
 print('Escape-prefix mutant compiled and exited zero; independent Go bytes caught it',flush=True)
finally:
 (rule/'quoted_mutant.a').unlink();(rule/'quoted_mutant_controls.a').unlink()

(rule/'quoted_guard_mutant.a').write_text(original.replace('point >= 55296 && point <= 57343','point < 0',1))
(rule/'quoted_guard_mutant_controls.a').write_text((rule/'quoted_controls.a').read_text().replace("'./quoted_name.a'","'./quoted_guard_mutant.a'"))
try:
 run([compiler,'build',str(rule/'quoted_guard_mutant_controls.a'),'-o',str(out/'guard-mutant')],'guard-mutant-build')
 actual,_=run([str(out/'guard-mutant'),'--surrogate'],'guard-mutant')
 print('Surrogate guard mutant compiled and exited zero; required refusal 70 caught it',flush=True)
finally:
 (rule/'quoted_guard_mutant.a').unlink();(rule/'quoted_guard_mutant_controls.a').unlink()
