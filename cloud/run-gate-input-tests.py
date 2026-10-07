#!/usr/bin/env python3
"""Run exactly the requested optional correctness checks and retain every JSON event."""
import concurrent.futures
import json
import os
from pathlib import Path
import re
import subprocess
import sys
import time

ROWS = [
 ('lint', 'stage1/cohere/lint', 'TestCompilerAndStage1Agree'),
 ('parser-expressions', 'stage1/typescript/parser', 'TestCompilerExpressionsAgree'),
 ('parser-whole', 'stage1/typescript/parser', 'TestWholeCompilerAgrees'),
 ('postcss', 'stage1/cohere/css', 'TestThePortParsesAsGoCohereDoes/PostCSS'),
 ('graphql', 'stage1/cohere/graphql', 'TestThePortParsesAsGoCohereDoes/as_graphql-js'),
 ('media-query', 'stage1/cohere/mediaquery', 'TestThePortParsesAsGoCohereDoes/as_postcss-media-query-parser'),
 ('selector', 'stage1/cohere/selector', 'TestThePortParsesAsGoCohereDoes/as_postcss-selector-parser'),
 ('selector-nontermination', 'stage1/cohere/selector', 'TestTheLibraryDoesNotReturnOnUnconsumedNamespaceBars'),
 ('values', 'stage1/cohere/values', 'TestThePortParsesAsGoCohereDoes/as_postcss-values-parser'),
 ('json-numeric', 'stage1/cohere/json', 'TestUpstreamNumericSeparatorGap'),
 ('json-corpus', 'stage1/cohere/json', 'TestUpstreamRepositoryCorpusParity'),
 ('json-mutants', 'stage1/cohere/json', 'TestExternalComparisonCatchesThreePrinterMutants'),
 ('css-printer-default', 'stage1/cohere/css', 'TestCSSPrinterAgreesWithGo/default'),
 ('css-printer-narrow', 'stage1/cohere/css', 'TestCSSPrinterAgreesWithGo/narrow'),
 ('css-printer-boundaries', 'stage1/cohere/css', 'TestCSSPrinterBoundaryProofs'),
 ('gitignore-100MiB', 'stage1/cohere/gitignore', 'TestThePortAnswersAsGoCohereAndGitDo/catches_R2_the_size_limit_one_byte_lower'),
 ('checker-split', 'internal/native', 'TestSplitTSGoAgrees')]
VARIABLES = ['ADAMIC_TYPESCRIPT_SOURCE','ADAMIC_CSS_LIBRARY','ADAMIC_GRAPHQL_LIBRARY',
             'ADAMIC_MEDIA_QUERY_LIBRARY','ADAMIC_SELECTOR_LIBRARY','ADAMIC_VALUES_LIBRARY',
             'ADAMIC_JSON_PRETTIER','ADAMIC_CSS_PRINTER_LIBRARY','ADAMIC_GITIGNORE_LARGEST','ADAMIC_CLANG_TSGO_ARCHIVE']


def main():
    repository, output, mode = sys.argv[1:]
    output = Path(output); output.mkdir(parents=True, exist_ok=True)
    environment = dict(os.environ, ADAMIC_GATE_UNCACHED='1')
    if mode == 'without':
        # These were removed by ordinary setup's env.sh. Refuse to fake that proof.
        assert not any(environment.get(name) for name in VARIABLES), 'source ordinary setup env.sh first'
    else:
        assert all(environment.get(name) for name in VARIABLES), 'source --gate-inputs env.sh first'
    def command(args):return subprocess.check_output(args,cwd=repository,timeout=60).decode().strip()
    flags=dict(commit=command(['git','rev-parse','HEAD']), nproc=command(['nproc']),
               cpu_max=Path('/sys/fs/cgroup/cpu.max').read_text().strip(),go=command(['go','version']),
               clang=command(['clang','--version']).splitlines()[0],node=command(['node','--version']),cached=False,
               inputs={name:environment.get(name) for name in VARIABLES})
    def run(row):
        identifier,package,test=row
        expression='/'.join('^'+re.escape(part)+'$' for part in test.split('/'))
        args=['go','test','./'+package,'-run',expression,'-count=1','-timeout=30m','-json']
        before=Path('/proc/loadavg').read_text().strip(); started=time.monotonic()
        # JSON corpus report retains the actual differing answers, not just identities.
        env=dict(environment, ADAMIC_JSON_REPORT=str(output.resolve()/(identifier+'-upstream-differences.txt')))
        log=output/(identifier+'.jsonl')
        with log.open('wb') as stream:
            result=subprocess.run(args,cwd=repository,env=env,stdout=stream,stderr=subprocess.STDOUT,timeout=1860)
        seconds=time.monotonic()-started; after=Path('/proc/loadavg').read_text().strip()
        events=[]
        for line in log.read_text().splitlines():
            try:events.append(json.loads(line))
            except ValueError:pass
        actions=[event['Action'] for event in events if event.get('Test')==test]
        outputs=[event.get('Output','') for event in events]
        (output/(identifier+'.log')).write_text(''.join(outputs))
        verdict=next((x.upper() for x in reversed(actions) if x in ['skip','pass','fail']),
                     'ABSENT' if result.returncode==0 else 'BLOCKED')
        failed_tests={event.get('Test') for event in events if event['Action']=='fail'}
        diagnostic_outputs=[event.get('Output','') for event in events if event.get('Test') in ({test} | failed_tests)]
        diagnostics=[line.strip() for text in diagnostic_outputs for line in text.splitlines()
                     if re.search(r'\.go:\d+:',line) and re.search(r'differ|exit|error|failed|cannot|changed|panic|lost|survived|refused',line,re.I)]
        record=dict(id=identifier,package=package,test=test,verdict=verdict,exit=result.returncode,seconds=seconds,
                    first_diagnostic=diagnostics[0] if diagnostics else None,command=args,
                    build_flags=flags,load_before=before,load_after=after)
        (output/(identifier+'.json')).write_text(json.dumps(record,indent=2)+'\n')
        print(identifier,verdict,f'{seconds:.3f}s',flush=True)
        return record
    # Large sanitized/compiler corpora have a high peak; cap concurrency explicitly.
    with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:
        results=list(pool.map(run,ROWS))
    (output/'verdicts.json').write_text(json.dumps(results,indent=2)+'\n')


if __name__=='__main__':main()
