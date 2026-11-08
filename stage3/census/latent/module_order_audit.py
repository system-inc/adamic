"""Test real production module-order refusals in a separately built guarded overlay.

Usage: python3 module_order_audit.py BINARY [--predicates]
No overlay or compiler is built by this test.
"""
import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest

from stream_records import census_records

BINARY = str(Path(sys.argv.pop(1)).resolve())
PREDICATES = '--predicates' in sys.argv
if PREDICATES:
    sys.argv.remove('--predicates')

class ModuleOrderOutput(unittest.TestCase):
    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory(prefix='latent-order-error-')
        self.addCleanup(self.temporary.cleanup)
        self.root = Path(self.temporary.name)
        self.source = self.root / 'source'
        self.source.mkdir()
        (self.source/'a-clean.ts').write_text('export const answer: number = 1;\nexport function safe(): number { return answer; }\n')
        (self.source/'a-deferred.ts').write_text('export function identity<T>(value: T): T { return value; }\n')
        (self.source/'b-dynamic.ts').write_text('export const pending = import("./c-leaf");\n')
        (self.source/'c-leaf.ts').write_text('export const leaf: number = 2;\n')

    def run_mode(self, mode, only=None):
        env = {k:v for k,v in os.environ.items() if k not in ['LATENT_FULL','LATENT_SPECULATIVE','LATENT_ONLY_FILE']}
        env['LATENT_ASSERT_NO_OUTPUT']='1'
        if not PREDICATES: env['LATENT_RELATION_PER_UNIT']='1'
        if mode=='full': env['LATENT_FULL']='1'
        if mode=='speculative': env['LATENT_SPECULATIVE']='1'
        if only is not None: env['LATENT_ONLY_FILE']=str(self.source/only)
        output=self.root/(mode+('-selected' if only else '')+'.jsonl')
        log=self.root/(output.stem+'.log')
        with log.open('w') as stream:
            result=subprocess.run([BINARY,str(self.source),str(output)],env=env,stdout=stream,stderr=subprocess.STDOUT,timeout=30)
        self.assertEqual(result.returncode,0,log.read_text())
        rows=[json.loads(line) for line in output.read_text().splitlines()]
        self.assertEqual(rows[0]['diagnostics'],[])
        if not PREDICATES:
            self.assertEqual(rows[-1]['status'],'relation_report')
            relation=rows[-1]['relation_overflow']
            self.assertEqual(relation['after_load'],'')
            self.assertEqual(relation['end'],'')
            self.assertTrue(relation['per_unit'])
            rows=census_records(rows)
        return {Path(row['file']).name:row for row in rows[1:]}

    def test_order_error_is_a_lowering_finding_for_every_eligible_unit(self):
        # The first root succeeds before b-dynamic fails. No accumulated partial
        # order may reach prepasses or make a-clean's units look finding-free.
        for mode in (['first-error'] if PREDICATES else ['first-error','full','speculative']):
            with self.subTest(mode=mode):
                row=self.run_mode(mode)['a-clean.ts']
                self.assertEqual(len(row['units']),2)
                for unit in row['units']:
                    if not PREDICATES: self.assertEqual(unit['relation_overflow'],'')
                    errors=[f for f in row['findings'] if f['unit']==unit['where'] and f['phase']=='lowering' and f['reason']=='a dynamic import()']
                    self.assertTrue(errors,(mode,unit,row['findings']))
                    self.assertTrue(all(f['kind']=='NotYet' and '/b-dynamic.ts:' in f['where'] for f in errors))
                if mode=='full':
                    self.assertEqual(sum(f['kind']=='Boundary' and f['reason']=='module ordering failed; unit not lowered' for f in row['findings']),2)
                if mode=='speculative':
                    self.assertEqual(row['speculative_coverage']['attempted_nodes'],0)
                    self.assertTrue(row['failed_boundaries'])

    @unittest.skipIf(PREDICATES,'historical predicates driver has no full-mode selector')
    def test_order_error_is_emitted_for_file_with_only_deferred_units(self):
        row=self.run_mode('full','a-deferred.ts')['a-deferred.ts']
        self.assertEqual([u['status'] for u in row['units']],['deferred_generic_instantiation'])
        self.assertTrue(any(f['phase']=='registration' and f['reason']=='a dynamic import()' for f in row['findings']),row)

if __name__=='__main__':
    unittest.main()
