import json
import os
from pathlib import Path
import tempfile
import unittest
import vocabulary_accounting as accounting
import vocabulary_wall as wall

class VocabularyProfile(unittest.TestCase):
    def test_cache_accounting_mutant(self):
        with tempfile.TemporaryDirectory() as tmp:
            p=Path(tmp)/'profile'
            text='events: Ir D1mr\nfn=(1) adamic_retain\n1 7 2\ncfn=(2) adamic_release\ncalls=1 1\n1 4 1\nfn=(2)\n1 4 1\nsummary: 11 3\n'
            p.write_text(text);r=accounting.summarize(p)
            self.assertEqual(r['total'],{'Ir':11,'D1mr':3})
            self.assertEqual(r['categories']['retain']['Ir'],7)
            self.assertEqual(r['sites'][0]['count'],1)
            self.assertEqual(r['sites'][0]['line'],1)
            self.assertEqual(r['sites'][0]['callee'],'adamic_release')
            p.write_text(text.replace('summary: 11 3','summary: 15 4'))
            with self.assertRaises(ValueError):accounting.summarize(p)
            p.write_text(text)
            combined=accounting.summarize_parts([p,p])
            self.assertEqual(combined['total'],{'Ir':22,'D1mr':6})

    def test_entry_fixture_mutant(self):
        source='func (vocabulary *abbreviationVocabulary) find(name string) (abbreviationFinding, bool) { return abbreviationFinding{}, false }'
        with self.assertRaises(ValueError):wall.go_source(source,entries=True)

    def test_compiler_handoff_exact_targets(self):
        import re
        text=Path(__file__).with_name('COMPILER-HANDOFF.md').read_text()
        rows=re.findall(r'^\| \d+ / (?:\d+|-) \| `[^`]+` / \d+ \| (yes|no) \| (\d+) / (\d+) / (\d+) \|',text,re.M)
        self.assertEqual(len(rows),643)
        selected=[row for row in rows if row[0]=='yes']
        self.assertEqual(len(selected),562)
        self.assertEqual([sum(int(row[i]) for row in selected) for i in (1,2,3)],[68639633,1786701,15953])

    def test_complete_real_profiles(self):
        root=Path(os.environ['ADAMIC_SCOUT_VOCABULARY'])
        r=json.loads((root/'wall.json').read_text())
        self.assertEqual(r['rounds'],3)
        import hashlib
        fixtures=json.loads((root/'fixtures.json').read_text())
        self.assertEqual(fixtures['source_count'],77)
        self.assertEqual(len(fixtures['files']),77)
        self.assertEqual(len(fixtures['rules']),93)
        for source in fixtures['files']:
            self.assertEqual(source['manifest_row'].split('\t')[1],'all')
            data=Path(source['path']).read_bytes()
            self.assertEqual(hashlib.sha256(data).hexdigest(),source['sha256'])
        for samples in r['release'].values():
            self.assertEqual(len(samples),3)
        self.assertGreater(r['bytes'],0)
        hashes={sample['sha256'] for samples in r['release'].values() for sample in samples}
        self.assertEqual(len(hashes),1)
        for samples in r['release'].values():
            for sample in samples:self.assertEqual(sample['bytes'],r['bytes'])
        gate=json.loads((root/'gate.json').read_text())
        self.assertEqual(gate['calls'],r['scope']['native-wall']['calls'])
        self.assertLess(gate['matched'],gate['calls'])
        self.assertGreater(gate['matched'],r['scope']['Go-wall']['calls'])
        for name,scope in r['scope'].items():
            self.assertGreater(scope['calls'],0);self.assertGreater(scope['timed_calls'],0)
            self.assertGreater(scope['estimated_ns'],0)
            if name.endswith('entries'):self.assertGreater(sum(scope['entries']),0)
        self.assertGreater(r['native_pc']['samples'],100)
        self.assertEqual(r['scope']['native-wall']['overflow'],0)
        for side in ('native','Go'):
            profile=accounting.summarize_parts([root/'native-checkpoints.callgrind.out',*sorted(root.glob('native-checkpoints.callgrind.out.*'))]) if side=='native' else accounting.summarize(root/(side+'.callgrind.out'))
            self.assertGreater(profile['total']['Ir'],0)
            self.assertIn('D1mr',profile['total'])
            self.assertGreater(profile['total']['D1mr']+profile['total']['D1mw'],0)
            function='adamic_function_164_abbreviation' if side=='native' else 'github.com/system-inc/cohere/internal/lint/rules/nexus.(*abbreviationVocabulary).find'
            calls=sum(edge['count'] for edge in profile['calls'] if edge['callee']==function)
            self.assertEqual(calls,r['scope'][side+'-wall']['calls'])
            if side=='native':
                for callee,target in (('adamic_retain',980189039),('adamic_release',975750750)):
                    self.assertEqual(sum(site['count'] for site in profile['sites'] if site['callee']==callee),target)
