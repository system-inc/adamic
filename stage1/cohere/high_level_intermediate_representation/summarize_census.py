#!/usr/bin/env python3
"""Refresh checked-in evidence only after the full Node/native comparison test passes."""
import collections
import json
import pathlib
import re
import sys

lane = pathlib.Path(__file__).resolve().parent
census = pathlib.Path(sys.argv[1])
path = lane / 'testdata/construction-summary.json'
summary = json.loads(path.read_text())
records = json.loads((census / 'records.json').read_text())
def probe(record):
    return all(call.startswith('probe:') for call in record['calls'])
corpus = [r for r in records['Records'] if not probe(r)]
probes = [r for r in records['Records'] if probe(r)]
matched = [r for r in corpus if r['eligible']]
excluded = [r for r in corpus if r.get('excluded')]
summary['constructionCorpusFunctions'] = sum(r['functions'] for r in corpus)
summary['constructionCorpusFunctionsMatched'] = sum(r['functions'] for r in matched)
summary['constructionActiveCorpusFunctions'] = sum(r['functions'] for r in corpus if not r.get('excluded'))
summary['pathProbes'] = sum(r['functions'] for r in probes)
summary['pathProbesMatched'] = sum(r['functions'] for r in probes if r['eligible'])
summary['upstreamFixtures'] = records['Fixtures']
summary['flowExcludedFixtures'] = sorted(records['Flow'])
summary['flowExcludedCorpusGraphs'] = [{'key':r['key'], 'start':r['start'], 'end':r['end'], 'calls':r['calls'], 'reason':r['excluded']} for r in excluded]
summary['matchedFunctions'] = [{key:r[key] for key in ('key','start','end','checker','calls')} for r in matched]
skips = re.findall(r'--- SKIP: ([^ ]+)', (census / 'go-tests.log').read_text())
assert set(skips) == set(summary['skippedGoTests']), 'Reclassify changed original Go skips before regenerating'
assert sum(len(g['tests']) for g in summary['skippedGoTestGroups']) == len(skips)
kinds = collections.Counter()
for r in corpus:
    kinds.update(re.findall(r'^instruction \d+ \d+ \S+ \S+ (\S+)', r['dump'], re.M))
summary['instructionKinds'] = dict(sorted(kinds.items()))
path.write_text(json.dumps(summary, indent=2) + '\n')
print(f"{summary['constructionCorpusFunctionsMatched']}/{summary['constructionCorpusFunctions']} corpus; "
      f"{summary['pathProbesMatched']}/{summary['pathProbes']} probes; {len(excluded)} excluded Flow graphs")
