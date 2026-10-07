from pathlib import Path
import os, subprocess, time
rule = Path(__file__).resolve().parents[1]
out = Path(os.environ.get('ADAMIC_WAVE15_UNICODE_ARTIFACTS', '/tmp/wave15-unicode-sweep'))
out.mkdir(parents=True, exist_ok=True)
compiler = os.environ.get('ADAMIC_WAVE15_SIXTH_COMPILER', '/tmp/wave15-b8-adamic')
def run(command, label):
    start = time.perf_counter_ns(); result = subprocess.run(command, capture_output=True)
    (out/(label+'.stdout')).write_bytes(result.stdout); (out/(label+'.stderr')).write_bytes(result.stderr)
    assert result.returncode == 0, (label, result.returncode, result.stderr.decode()[:2000])
    return result.stdout, time.perf_counter_ns()-start
truth, elapsed = run(['go','run',str(rule/'testdata/unicode_oracle.go')], 'go')
print('Go all scalar values IsUpper', len(truth.splitlines()), 'uppercase points', flush=True)
node, elapsed = run(['node',str(rule/'testdata/unicode_oracle.mjs')], 'node')
assert node == truth, 'Node property version differs from Go Unicode table'
print('Node literal matches Go across 1114112 code points',flush=True)
for sanitize in [False, True]:
    label = 'unicode-asan' if sanitize else 'unicode'
    run([compiler,'build',str(rule/'unicode_controls.a'),'-o',str(out/label)] + (['--sanitize'] if sanitize else []), label+'-build')
    actual, elapsed = run([str(out/label)], label)
    assert actual == truth, (label, 'native property differs from Go')
    print(label,'PASS all 1114112 code points',len(actual),'bytes; ns',elapsed,flush=True)
original = (rule/'rule.a').read_text(); needle = '/^\\p{Lu}/u'; assert needle in original
(rule/'unicode_mutant.a').write_text(original.replace(needle,'/^\\p{Ll}/u',1))
(rule/'unicode_mutant_controls.a').write_text((rule/'unicode_controls.a').read_text().replace("'./rule.a'","'./unicode_mutant.a'"))
try:
    run([compiler,'build',str(rule/'unicode_mutant_controls.a'),'-o',str(out/'mutant')],'mutant-build')
    actual, elapsed = run([str(out/'mutant')],'mutant')
    assert actual != truth, 'Unicode category mutant escaped'
    print('Unicode Lu to Ll mutant compiled exit 0; external Go and Node bytes catch it',flush=True)
finally:
    (rule/'unicode_mutant.a').unlink(); (rule/'unicode_mutant_controls.a').unlink()
