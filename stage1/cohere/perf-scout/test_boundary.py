import json, os, re, unittest
from pathlib import Path
import boundary

class BoundaryAcceptance(unittest.TestCase):
    def test_counts_and_bytes(self):
        root=Path(os.environ['ADAMIC_SCOUT_BOUNDARY'])
        report=json.loads((root/'measurements.json').read_text())
        expected={'scanner':1786701,'public23':15953,'compiler77':68639633}
        self.assertEqual(set(report['workloads']),set(expected))
        for name,w in report['workloads'].items():
            for side in ('before','borrow','encoder'):
                self.assertEqual(len(w['samples'][side]),3)
                for s in w['samples'][side]:
                    self.assertEqual(s['sha256'],w['sha256']);self.assertEqual(s['bytes'],w['bytes'])
            a,b=[w['sites'][s] for s in ('before','borrow')]
            count=lambda r:dict((k,int(v)) for k,v in re.findall(r'(allocations|frees|retains|releases|peak|regions) (\d+)',r['global_counts']))
            x,y=count(a),count(b)
            for k in ('allocations','frees','peak','regions'):self.assertEqual(x[k],y[k])
            for k in ('retains','releases'):self.assertEqual(x[k]-y[k],expected[name])
            selected=[s for s in a['sites'] if s['eligible']]
            self.assertEqual(sum(s['calls'] for s in selected),expected[name])
            newer={s['key']:s for s in b['sites']}
            for s in selected:
                self.assertEqual(s['retains'],s['calls']);self.assertEqual(s['paired_releases'],s['calls'])
                self.assertEqual(newer[s['key']]['calls'],s['calls'])
                self.assertEqual(newer[s['key']]['retains'],0);self.assertEqual(newer[s['key']]['paired_releases'],0)
        self.assertNotEqual(report['mutant_exit'],0)
        self.assertIn('AddressSanitizer: heap-use-after-free',(root/'mutant.stderr').read_text())

    def test_vocabulary(self):
        d=json.loads((Path(os.environ['ADAMIC_SCOUT_BOUNDARY'])/'vocabulary.json').read_text())
        self.assertEqual(len(d),3)
        for r in d.values():
            walk,member=r['sites'];self.assertGreater(walk['calls'],0)
            self.assertEqual(walk['calls'],member['calls']);self.assertGreater(walk['retains'],0)
            self.assertEqual(member['retains'],0);self.assertEqual(member['releases'],0)

    def test_escape_mutants(self):
        safe='adamic_object * node = lookup();\ndouble kind = node->kind;\nadamic_release(node);'
        self.assertEqual(boundary.aliases_and_reads(safe,'node')[2],[])
        for extra in ('return node;','store(node);','adamic_value * slot = adamic_object_data_field(node, 0);'):
            self.assertTrue(boundary.aliases_and_reads(safe+'\n'+extra,'node')[2])
        self.assertTrue(boundary.aliases_and_reads(safe+'\nadamic_object * alias = node;\nreturn alias;','node')[2])
