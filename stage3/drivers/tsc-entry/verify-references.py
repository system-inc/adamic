#!/usr/bin/env python3
"""Check project-reference evidence against Node, stock checking and native stops."""
import argparse,gzip,json,re
from pathlib import Path
p=argparse.ArgumentParser();p.add_argument('evidence',nargs='?',type=Path,default=Path(__file__).resolve().parent/'evidence/project-references');e=p.parse_args().evidence
def raw(name):return gzip.decompress((e/(name+'.gz')).read_bytes()).decode()
def code(name):return int((e/(name+'.exit')).read_text())
def observation(name):return json.loads(raw(name+'.stdout'))
assert raw('node-version.stdout')=='v24.19.0\n' and raw('tsc-version.stdout')=='Version 6.0.3\n'
assert code('adamic-before')==1 and 'app/main.ts:1:23: error TS6305:' in raw('adamic-before.stderr')
before=observation('loader-before');assert not before['loaded'] and 'TS6305:' in before['error']
for name in ['tsc-single','tsc-build','node-emitted']:assert code(name)==0 and not raw(name+'.stderr'),name
build=raw('tsc-build.stdout');assert build.index("/dependency/tsconfig.json'...")<build.index("/app/tsconfig.json'...")
assert raw('node-emitted.stdout')=='7\n'
outputs=json.loads((e/'emitted-files.json').read_text());assert 'out/dependency/value.d.ts' in outputs and 'out/app/main.js' in outputs
assert observation('loader-after')['loaded'] and not observation('loader-after')['error']
assert observation('flat-minimal-load')['loaded'] and not observation('flat-minimal-load')['error']
assert 'stage 0 can\'t lower the library method log yet' in raw('flat-minimal.stderr') and 'TS6305' not in raw('flat-minimal.stderr')
minimal=json.loads((e/'stock-minimal-source.json').read_text());assert len(minimal['roots'])==len(minimal['sources'])==2 and not minimal['diagnostics']
for prefix in ['entry','entry-types']:
 for suffix in ['stdout','stderr']:assert raw(prefix+'-0.'+suffix)==raw(prefix+'-1.'+suffix)
 assert code(prefix+'-0')==code(prefix+'-1')==1
 assert 'TS6305' not in raw(prefix+'-0.stderr')
assert re.match(r'.*/src/compiler/debug.ts:200:19: error TS2339: Property \'captureStackTrace\' does not exist on type \'ErrorConstructor\'\.',raw('entry-0.stderr'))
assert re.match(r"adamic: .*/src/compiler/sys.ts:1502:51: stage 0 can't lower a namespace read before runtime initialization",raw('entry-types-0.stderr'))
without=json.loads((e/'stock-entry-without-types.json').read_text());with_types=json.loads((e/'stock-entry-with-types.json').read_text())
assert len(with_types['roots'])==len(with_types['sources'])==81 and with_types['roots']==without['roots'] and not with_types['diagnostics']
assert without['options']['types']==[] and with_types['options']['types']==['node']
assert len(without['diagnostics'])==66 and without['diagnostics'][0]['code']==2339
locations={(d['file'],d['line'],d['column'],d['code']) for d in without['diagnostics']}
actual={(f,int(l),int(c),int(n)) for f,l,c,n in re.findall(r'([^\n]+):(\d+):(\d+): error TS(\d+):',raw('entry-0.stderr'))};# Adamic supplies console through its prelude; stock's es2020-only profile does not.
console_location=(without['diagnostics'][0]['file'],865,16,2584)
assert console_location in locations and actual==locations-{console_location}
assert len(actual)==65
assert code('host-node')==0 and raw('host-node.stdout')=='true\n' and not raw('host-node.stderr')
assert code('host-adamic')==1 and 'main.ts:2:7: error TS2339:' in raw('host-adamic.stderr')
assert code('host-stock')==2 and 'error TS2339:' in raw('host-stock.stdout')
assert code('mutant-tsc-build')!=0 and 'unimported.ts(1,14): error TS2322:' in raw('mutant-tsc-build.stdout')
mutant=observation('mutant-flat-load');assert not mutant['loaded'] and 'unimported.ts:1:14: error TS2322:' in mutant['error']
assert code('mutant-stock-source')==1 and any(d['code']==2322 and d['file'].endswith('/unimported.ts') for d in json.loads((e/'mutant-stock-source.json').read_text())['diagnostics'])
assert code('entry-node')==0 and raw('entry-node.stdout')=='Version 6.0.3\n' and not raw('entry-node.stderr')
i=json.loads((e/'identity.json').read_text());assert len(i['after'])==746 and i['previous']==i['after'] and not i['built_compiler_declarations_exist'] and not i['publication_compiler_diff']
trace=json.loads((e/'loader-trace.json').read_text());assert 'projectConfig.ProjectReferences()' in trace['loadInput_references'] and 'config.ProjectReferences()' in trace['auditProjectOptions_references']
assert 'UseSourceOfProjectReference' not in trace['loadInput_references']
assert 'ok  ' in raw('loader-tests.log') and 'FAIL' not in raw('loader-tests.log')
print('PASS: Node build order/output, missing-output diagnostic, explicit source roots, both flattening profiles, 81-source stock oracle, real-input mutant, source identity and focused loader tests')
