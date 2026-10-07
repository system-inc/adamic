#!/usr/bin/env python3
"""Run external correctness comparisons with pinned inputs and reject skips."""
import json
import os
import pathlib
import subprocess

ROOT = pathlib.Path(__file__).resolve().parents[4]
OUT = pathlib.Path('/workspace/wave-07-required-inputs')
OUT.mkdir(exist_ok=True)
LIBRARY = pathlib.Path('/workspace/wave-07-required-oracles')
SOURCE = pathlib.Path('/workspace/wave-07-typescript')
versions = {'prettier':'3.9.6','postcss':'8.5.16','postcss-scss':'4.0.9',
    'graphql':'17.0.2','postcss-selector-parser':'2.2.3',
    'postcss-values-parser':'2.0.1','postcss-media-query-parser':'0.2.3'}
for name, expected in versions.items():
    actual = json.loads((LIBRARY/'node_modules'/name/'package.json').read_text())['version']
    assert actual == expected, (name, actual, expected)
tag = subprocess.check_output(['git','-C',str(SOURCE),'describe','--tags','--exact-match'],text=True).strip()
assert tag == 'v6.0.3', tag
revision = subprocess.check_output(['git','-C',str(SOURCE),'rev-parse','HEAD'],text=True).strip()
environment = dict(os.environ)
settings = {'ADAMIC_TYPESCRIPT_SOURCE':str(SOURCE)}
for name in ['CSS_LIBRARY','CSS_PRINTER_LIBRARY','CSSSTRINGS_LIBRARY','CSSNUMBERS_LIBRARY',
    'GRAPHQL_LIBRARY','SELECTOR_LIBRARY','VALUES_LIBRARY','MEDIA_QUERY_LIBRARY','JSON_PRETTIER']:
    settings['ADAMIC_'+name] = str(LIBRARY)
environment.update(settings)
packages = ['stage1/typescript/parser','stage1/typescript/scanner',
    'stage1/cohere/lint','stage1/cohere/css','stage1/cohere/cssstrings',
    'stage1/cohere/cssnumbers','stage1/cohere/graphql','stage1/cohere/selector',
    'stage1/cohere/values','stage1/cohere/mediaquery','stage1/cohere/json']
names = ['TestCompilerExpressionsAgree','TestScannerAgreesWithTypescriptGo',
    'TestCompilerAndStage1Agree','TestThePortParsesAsGoCohereDoes',
    'TestCSSPrinterAgreesWithGo','TestCSSPrinterBoundaryProofs','TestCSSStrings','TestCSSNumbers',
    'TestTheLibraryDoesNotReturnOnUnconsumedNamespaceBars','TestUpstreamNumericSeparatorGap',
    'TestUpstreamRepositoryCorpusParity','TestExternalComparisonCatchesThreePrinterMutants']
command = ['go','test','-p','2','-json','-count=1','-timeout','30m',
    '-run','^('+'|'.join(names)+')$', *['./'+package for package in packages]]
receipt = {'command':command,'environment':settings,'versions':versions,
    'typescript_revision':revision,'typescript_tag':tag}
(OUT/'receipt.json').write_text(json.dumps(receipt,indent=2)+'\n')
with (OUT/'events.jsonl').open('wb') as output, (OUT/'stderr.log').open('wb') as error:
    result = subprocess.run(command,cwd=ROOT,env=environment,stdout=output,stderr=error)
events = [json.loads(line) for line in (OUT/'events.jsonl').read_text().splitlines()]
skips = [event for event in events if event.get('Action') == 'skip']
failures = [event for event in events if event.get('Action') == 'fail']
passed = [event for event in events if event.get('Action') == 'pass' and event.get('Test')]
receipt.update(exit=result.returncode,skips=skips,failures=failures,
    passed=[{'package':event['Package'],'test':event['Test']} for event in passed])
(OUT/'receipt.json').write_text(json.dumps(receipt,indent=2)+'\n')
print('External correctness gate: exit',result.returncode,'skips',len(skips),
      'failed events',len(failures),'passed test events',len(passed),flush=True)
assert result.returncode == 0 and not skips and not failures, 'see '+str(OUT/'events.jsonl')
for package in packages:
    assert any(row['package'].endswith('/'+package) for row in receipt['passed']), package+' ran no selected check'
print('PASS pinned external correctness inputs; no selected test skipped',flush=True)
