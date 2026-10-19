import json, os, signal, subprocess, time
from pathlib import Path
out=Path('review/compiler/loom-slice-2-2')
env=dict(os.environ, GOPROXY='https://proxy.golang.org|direct', GOMAXPROCS='4', ADAMIC_GATE_UNCACHED='1')
excluded = {'TestCallbackWideningNumberEvery', 'TestCallbackWideningNullableObjectMap', 'TestWeakDifferentialHeldUnion', 'TestWeakDifferentialBooleanFilter'}
names = []
import re
for stem in ['miscompile_2a', 'weak_differential', 'self_compare', 'callback_widening']:
    names.extend(re.findall(r'^func (Test\w+)\(', Path('internal/oracle/' + stem + '_test.go').read_text(), re.M))
oracle_pattern = '^(' + '|'.join(name for name in names if name not in excluded) + ')$'
checks=[('flow', ['go', 'test', './internal/flow', '-run', '^(TestCallbackWideningNumberCapture|TestCallbackWideningNumberEvery|TestCallbackWideningNumberFilter|TestCallbackWideningNumberFind|TestCallbackWideningNumberFindIndex|TestCallbackWideningNumberFindLast|TestCallbackWideningNumberForEach|TestCallbackWideningNumberGeneric|TestCallbackWideningNumberIndex|TestCallbackWideningNumberMap|TestCallbackWideningNumberNamedSort|TestCallbackWideningNumberReduceAccumulator|TestCallbackWideningNumberReduce|TestCallbackWideningNumberSome|TestCallbackWideningNumberSort|TestCallbackWideningNumberThrow|TestCallbackWideningNumberToSorted|TestCallbackWideningObjectEvery|TestCallbackWideningObjectFilter|TestCallbackWideningObjectFind|TestCallbackWideningObjectForEach|TestCallbackWideningObjectGeneric|TestCallbackWideningObjectMap|TestCallbackWideningObjectReduce|TestCallbackWideningObjectSome|TestCallbackWideningObjectSort|TestCallbackWideningNumberGenericMap|TestCallbackWideningObjectGenericMap|TestCallbackWideningNumberArguments|TestCallbackWideningBooleanMap|TestCallbackWideningStringMap|TestCallbackWideningOptionalNumberMap)$', '-count=1', '-json', '-timeout=90s'], 240), ('fresh', ['go', 'test', './internal/fresh', '-run', '^(TestCallbackWideningNumberCapture|TestCallbackWideningNumberEvery|TestCallbackWideningNumberFilter|TestCallbackWideningNumberFind|TestCallbackWideningNumberFindIndex|TestCallbackWideningNumberFindLast|TestCallbackWideningNumberForEach|TestCallbackWideningNumberGeneric|TestCallbackWideningNumberIndex|TestCallbackWideningNumberMap|TestCallbackWideningNumberNamedSort|TestCallbackWideningNumberReduceAccumulator|TestCallbackWideningNumberReduce|TestCallbackWideningNumberSome|TestCallbackWideningNumberSort|TestCallbackWideningNumberToSorted|TestCallbackWideningObjectEvery)$', '-count=1', '-json', '-timeout=90s'], 240), ('fixtures', ['go', 'test', './stage3/fixtures', '-run', '^(TestFixturesSelfCompare|TestFixtureDirectoriesHaveTopLevelTests)$', '-count=1', '-json', '-timeout=90s'], 240), ('vet', ['go', 'vet', './internal/lower', './internal/native', './internal/ir', './internal/flow', './internal/fresh', './internal/oracle', './stage3/fixtures'], 240), ('registry', ['go', 'test', './internal/oracle', '-run', '^TestFamiliesOwnEveryCaseOnce$', '-count=1', '-json', '-timeout=90s'], 120), ('counts', ['go', 'test', './internal/oracle', '-run', 'TestCountsAreRecorded', '-count=1', '-v', '-timeout=600s', '-args', '-update-counts'], 660)]
for name,command,limit in checks:
 start=time.monotonic()
 with (out/(name+'.log')).open('w') as log:
  p=subprocess.Popen(command,stdout=log,stderr=subprocess.STDOUT,env=env,start_new_session=True)
  try: code=p.wait(timeout=limit)
  except subprocess.TimeoutExpired: os.killpg(p.pid,signal.SIGKILL);code=p.wait()
 row=dict(name=name,command=command,exit=code,seconds=round(time.monotonic()-start,3))
 with (out/'checks.jsonl').open('a') as log:log.write(json.dumps(row)+'\n')
 print(json.dumps(row),flush=True)
 if code: raise SystemExit(code)
