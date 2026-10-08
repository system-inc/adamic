#!/usr/bin/env python3
"""Query the minimal census reproducer with both checkers and option variants."""
import argparse
import json
from pathlib import Path
import subprocess

parser = argparse.ArgumentParser()
parser.add_argument('checker', type=Path)
parser.add_argument('output', type=Path)
args = parser.parse_args()
unit = Path(__file__).resolve().parent
out = args.output.resolve()
out.mkdir(parents=True, exist_ok=False)
(out / 'src/compiler').mkdir(parents=True)
source = (unit / 'reduced-union.a').read_text()
(out / 'src/compiler/probe.ts').write_text(source)
line = source.splitlines()[2]
sites = [{'File':'probe.ts','Line':3,'Column':line.index('narrow')+1,'Kind':'KindIdentifier','Text':'narrow','Source':'never','Target':'{ id?: number }'}]
(out / 'sites.json').write_text(json.dumps(sites,indent=2)+'\n')
options = {'strict':True,'exactOptionalPropertyTypes':True,'noUncheckedIndexedAccess':True,'verbatimModuleSyntax':True,'erasableSyntaxOnly':False,'target':'es2024','module':'esnext','moduleResolution':'bundler','types':[],'noEmit':True,'allowImportingTsExtensions':True,'moduleDetection':'force','lib':['es2024']}
variants = {'adamic':{},'no-exact':{'exactOptionalPropertyTypes':False},'no-indexed':{'noUncheckedIndexedAccess':False},'no-extra':{'exactOptionalPropertyTypes':False,'noUncheckedIndexedAccess':False,'verbatimModuleSyntax':False,'erasableSyntaxOnly':False},'non-strict':{'strict':False,'exactOptionalPropertyTypes':False,'noUncheckedIndexedAccess':False}}
report = {}
for name, override in variants.items():
    config = out / (name + '.json')
    config.write_text(json.dumps({'files':['src/compiler/probe.ts'],'compilerOptions':options|override},indent=2)+'\n')
    commands = [(str(args.checker.resolve()),str(config),str(out/'sites.json'),str(out/(name+'-result.json')),'--diagnostics'),('node',str(unit/'stock.cjs'),str(config),str(out/'sites.json'),str(out/(name+'-stock.json')))]
    for index, command in enumerate(commands):
        with (out / (name + '-' + str(index) + '.log')).open('w') as log:
            subprocess.run(command, stdout=log, stderr=subprocess.STDOUT, check=True)
    adamic = json.loads((out/(name+'-result.json')).read_text())[0]
    stock = json.loads((out/(name+'-stock.json')).read_text())
    assert adamic['type']=='A' and stock['rows'][0]['type']=='A'
    assert adamic['diagnostics']==0 and stock['diagnostics']==[]
    assert adamic['relation']['source']=='never' and adamic['relation']['reduced_never_intersection']
    assert not adamic['never'] and not stock['rows'][0]['never']
    assert any(m['type']=='never' for m in stock['rows'][0]['members'])
    report[name] = {'options':options|override,'adamic':adamic,'stock':stock}
(out/'report.json').write_text(json.dumps(report,indent=2)+'\n')
print(json.dumps({'variants':len(report),'status':'pass','whole_value':'A','census_component':'never','both_checkers_agree':True}))
