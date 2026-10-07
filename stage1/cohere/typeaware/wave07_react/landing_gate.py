#!/usr/bin/env python3
"""Rerun wave-07's existing isolated gates after a main rebase."""
import json
import os
import pathlib
import subprocess
import time

ROOT = pathlib.Path(__file__).resolve().parents[4]
UNIT = pathlib.Path(__file__).resolve().parent
OUT = pathlib.Path(os.environ.get('ADAMIC_WAVE07_LANDING', '/workspace/wave-07-landing'))
OUT.mkdir(exist_ok=True)
RESUME = os.environ.get('ADAMIC_WAVE07_RESUME', '')
records = json.loads((OUT/'commands.json').read_text()) if RESUME else []
resume_pending = bool(RESUME)

def run(name, command, extra=None):
    global resume_pending
    if resume_pending:
        if name != RESUME:
            previous = [row for row in records if row['name'] == name]
            assert previous and previous[-1]['exit'] == 0, name+' has no passing prior run'
            print(name+': prior PASS (resumed run)', flush=True)
            return
        resume_pending = False
    environment = dict(os.environ)
    environment.update(extra or {})
    started = time.monotonic_ns()
    log = OUT / (name + '.log')
    with log.open('wb') as output:
        process = subprocess.run([str(value) for value in command], cwd=ROOT,
                                 env=environment, stdout=output, stderr=subprocess.STDOUT)
    records.append(dict(name=name, command=[str(value) for value in command],
                        environment=extra or {}, exit=process.returncode,
                        elapsed_ns=time.monotonic_ns()-started))
    (OUT / 'commands.json').write_text(json.dumps(records, indent=2)+'\n')
    if process.returncode:
        raise RuntimeError(name+' failed; see '+str(log))
    print(name+': PASS', flush=True)

run('original-trio', ['go','test','./stage1/cohere/typeaware','-run',
    '^TestWave07AgreementAndMutants$','-count=1','-v','-timeout','30m'],
    {'ADAMIC_WAVE07_ARTIFACTS':str(OUT/'original-trio'),
     'ADAMIC_WAVE07_COMPILER_MANIFEST':'/workspace/wave-07-artifacts/compiler.manifest',
     'ADAMIC_WAVE07_REPOSITORY_MANIFEST':'/workspace/wave-07-artifacts/repository.manifest',
     'ADAMIC_TYPESCRIPT_SOURCE':'/workspace/wave-07-typescript'})
for name, directory, script, artifacts in [
    ('timer','wave07_continuation','validate_timer.py','/workspace/wave-07-timer'),
    ('process','wave07_continuation','validate_process.py','/workspace/wave-07-process'),
    ('streams','wave07_continuation','validate_streams.py','/workspace/wave-07-streams-final'),
    ('rest','wave07_next','validate_rest.py','/workspace/wave-07-next-rest'),
    ('promise','wave07_next','validate_promise.py','/workspace/wave-07-next-promise'),
    ('regex','wave07_next','validate_regex.py','/workspace/wave-07-next-regex'),
]:
    run(name,['python3',ROOT/'stage1/cohere/typeaware'/directory/script,
              artifacts,'--compiler','/workspace/wave-07-typescript'])
run('continuation-questions',['python3',ROOT/'stage1/cohere/typeaware/wave07_continuation/validate_questions.py',
                              '/workspace/wave-07-questions'])
run('regex-question',['python3',ROOT/'stage1/cohere/typeaware/wave07_next/validate_regex_question.py',
                     '/workspace/wave-07-next-questions'])
run('bridge',['go','test','./bridge/tsgo/...','-count=1','-v','-timeout','15m'],
    {'ADAMIC_TSGO_CORPUS':'/workspace/wave-07-typescript'})
run('fact-guards',['go','test','./stage1/cohere/typeaware','-run',
    '^(TestFactsDecoderGuards|TestInspectRequestRefusals|TestPinnedTypeFlags)$',
    '-count=1','-v','-timeout','30m'])
run('node-oracle',['go','test','./internal/oracle','-run',
    '^TestTheOracleCatchesOneByte$|^TestLibraryMapSetIteratorCopiesRefused$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures|devirtualize|call_targets_.*|047cb0d_n_.*|library_map_set_iterator_(number_hash|exhausted)|override_same_representation)[.]a$',
    '-count=1','-v','-timeout','30m'])
run('vet',['go','vet','./...'])
run('gofmt',['gofmt','-l','cmd','internal','bridge/tsgo','stage1/cohere/typeaware'])
assert not (OUT/'gofmt.log').read_text().strip(), 'Go formatting differs'
run('react-branch-blocker',['python3',UNIT/'probe.py'])
run('jsx-dependency',['python3',UNIT/'dependency_probe.py'])
assert not resume_pending, 'Unknown resume step: '+RESUME
print('PASS wave-07 landing gates; React ports and shared dispatch integration remain incomplete.',flush=True)
