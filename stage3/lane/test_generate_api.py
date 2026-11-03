#!/usr/bin/env python3
"""Exercise adaptation sanctions against pinned upstream and the lane API guard."""
import difflib
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import unittest

import check

LANE = Path(__file__).resolve().parent


class ApiSanctionGeneratorTests(unittest.TestCase):
    def setUp(self):
        upstream = os.environ.get('ADAMIC_TYPESCRIPT_SOURCE')
        if not upstream:
            self.fail('ADAMIC_TYPESCRIPT_SOURCE names no checkout of the pinned TypeScript, '
                      'the gate input the adapted tree is made from')
        self.upstream = Path(upstream)
        self.scratch = tempfile.TemporaryDirectory(prefix='lane-generator-')
        self.addCleanup(self.scratch.cleanup)
        self.root = Path(self.scratch.name)
        self.stage = self.root / 'stage3'
        self.lane = self.stage / 'lane'
        self.lane.mkdir(parents=True)
        for name in ('generate-api.cjs', 'generate-api.py', 'normalize-api.cjs'):
            shutil.copyfile(LANE / name, self.lane / name)
        shutil.copyfile(LANE.parent / 'source.json', self.stage / 'source.json')
        # Stage exactly the proof inputs bound by the checked-in manifest.
        for proof in json.loads((LANE / 'sanctioned-api.json').read_text())['proofs']:
            relative = Path(proof['file']).relative_to('stage3')
            target = self.stage / relative
            target.parent.mkdir(parents=True, exist_ok=True)
            target.symlink_to(LANE.parent / relative)
        self.file = self.stage / 'adapt/zz-sanction-fixture/api-sanctions.json'
        self.file.parent.mkdir()
        self.row = dict(path=['TypeReferenceNode', 'typeName'], kind='PropertySignature',
                        before='EntityName', after='EntityName | undefined',
                        why='Fixture public optional widening')
        self.environment = dict(os.environ, NODE_PATH=str(LANE.parent / 'api/node_modules'),
                                TSC_ADAPT_TYPESCRIPT=str(LANE.parent / 'api/node_modules/typescript'))

    def generate(self, rows=None):
        self.file.write_text(json.dumps([self.row] if rows is None else rows))
        result = subprocess.run(['node', str(self.lane / 'generate-api.cjs'), str(self.upstream)],
                                env=self.environment, capture_output=True, text=True)
        return result

    def generated(self):
        result = self.generate()
        self.assertEqual(result.returncode, 0, result.stderr)
        proof = json.loads(result.stdout)
        self.assertIn(dict(file='stage3/adapt/zz-sanction-fixture/api-sanctions.json',
                           sha256=hashlib.sha256(self.file.read_bytes()).hexdigest()), proof['proofs'])
        return proof

    def guard(self, proof, local):
        results = self.root / 'results'
        tree = results / 'adapted-tree'
        tree.mkdir(parents=True)
        subprocess.run(['git', 'init', '-q', str(tree)], check=True, capture_output=True)
        objects = subprocess.check_output(['git', '-C', str(self.upstream), 'rev-parse',
                                           '--path-format=absolute', '--git-path', 'objects'], text=True).strip()
        (tree / '.git/objects/info/alternates').write_text(objects + '\n')
        subprocess.run(['git', '-C', str(tree), 'update-ref', 'HEAD', proof['source_commit']],
                       check=True, capture_output=True)
        for folder, text in [('reference', proof['original']), ('local', local)]:
            file = tree / f'tests/baselines/{folder}/api/typescript.d.ts'
            file.parent.mkdir(parents=True, exist_ok=True)
            file.write_text(text)
        oracle = results / 'oracle'
        oracle.mkdir()
        (oracle / 'baseline.diff').write_text(''.join(difflib.unified_diff(
            proof['original'].splitlines(True), local.splitlines(True),
            fromfile='reference/api/typescript.d.ts', tofile='local/api/typescript.d.ts')))
        return check.api_check(results, proof)[0]

    def test_api_sanction_applied_and_guard_accepts(self):
        proof = self.generated()
        self.assertIn('typeName: EntityName | undefined;', proof['expected'])
        self.assertEqual(self.guard(proof, proof['expected']), [])

    def test_api_sanction_before_mismatch_names_row(self):
        self.row['before'] = 'wrong upstream type'
        result = self.generate()
        self.assertNotEqual(result.returncode, 0)
        for text in ('api-sanctions.json row 1', 'wrong upstream type', 'EntityName'):
            self.assertIn(text, result.stderr)

    def test_api_sanction_missing_owner(self):
        self.row['path'] = ['NoSuchPublicOwner', 'missing']
        result = self.generate()
        self.assertNotEqual(result.returncode, 0)
        self.assertIn('api-sanctions.json row 1', result.stderr)
        self.assertIn('missing proof owner', result.stderr)

    def test_api_sanction_ambiguous_owner(self):
        self.row.update(path=['NodeFactory', 'createToken'], kind='MethodSignature')
        result = self.generate()
        self.assertNotEqual(result.returncode, 0)
        self.assertIn('api-sanctions.json row 1', result.stderr)
        self.assertIn('ambiguous', result.stderr)

    def test_api_sanction_duplicate_claim(self):
        result = self.generate([self.row, self.row])
        self.assertNotEqual(result.returncode, 0)
        self.assertIn('api-sanctions.json row 2', result.stderr)
        self.assertIn('already claimed', result.stderr)

    def test_api_sanction_hard_coded_claim(self):
        self.row.update(path=['System', 'exit'], kind='MethodSignature', before='void')
        result = self.generate()
        self.assertNotEqual(result.returncode, 0)
        self.assertIn('already claimed', result.stderr)

    def test_api_sanction_undo_is_rejected(self):
        proof = self.generated()
        local = proof['expected'].replace('typeName: EntityName | undefined;',
                                          'typeName: EntityName;')
        errors = self.guard(proof, local)
        self.assertTrue(any('missing sanctioned API declaration' in error and
                            'TypeReferenceNode' in error for error in errors), errors)

    def member_row(self):
        self.row = dict(path=['ModuleResolutionHost', 'getCurrentDirectory'], kind='MethodSignature',
                        replace='member', before='getCurrentDirectory?(): string;',
                        after='getCurrentDirectory?: (() => string) | undefined;',
                        why='Fixture optional method conversion')

    def test_api_sanction_member_applied_and_guard_accepts(self):
        self.member_row()
        proof = self.generated()
        self.assertTrue(self.row['after'] in proof['expected'], 'member conversion was not applied')
        self.assertEqual(self.guard(proof, proof['expected']), [])

    def test_api_sanction_member_before_requires_full_text(self):
        self.member_row()
        self.row['before'] = 'string'
        result = self.generate()
        self.assertNotEqual(result.returncode, 0)
        for text in ('api-sanctions.json row 1', 'getCurrentDirectory?(): string;', 'before "string"'):
            self.assertIn(text, result.stderr)

    def test_api_sanction_member_undo_is_rejected(self):
        self.member_row()
        proof = self.generated()
        local = proof['expected'].replace(self.row['after'], self.row['before'])
        self.assertNotEqual(local, proof['expected'])
        errors = self.guard(proof, local)
        self.assertTrue(any('missing sanctioned API declaration' in error and
                            'getCurrentDirectory' in error for error in errors), errors)

    def test_api_sanction_invalid_replace_is_rejected(self):
        self.row['replace'] = None
        result = self.generate()
        self.assertNotEqual(result.returncode, 0)
        self.assertIn('api-sanctions.json row 1', result.stderr)
        self.assertIn('invalid sanction', result.stderr)

    def insert_row(self):
        self.row = dict(insert='after', path=['TextRange'], kind='InterfaceDeclaration',
                        text='type LaneInsertedValue = string;\ninterface LaneInsertedRecord { value: LaneInsertedValue; }',
                        why='Fixture public declarations')

    def test_api_sanction_insert_applied_and_guard_accepts(self):
        self.insert_row()
        second = dict(self.row, text='type LaneInsertedSecond = LaneInsertedRecord;')
        result = self.generate([self.row, second])
        self.assertEqual(result.returncode, 0, result.stderr)
        proof = json.loads(result.stdout)
        for line in self.row['text'].splitlines() + [second['text']]:
            self.assertTrue('\n    ' + line in proof['expected'], 'insert missing or incorrectly indented')
        self.assertLess(proof['expected'].index('type LaneInsertedValue'),
                        proof['expected'].index('type LaneInsertedSecond'))
        self.assertEqual(self.guard(proof, proof['expected']), [])

    def test_api_sanction_insert_undo_is_rejected(self):
        self.insert_row()
        proof = self.generated()
        inserted = '\n' + '\n'.join('    ' + line for line in self.row['text'].splitlines())
        local = proof['expected'].replace(inserted, '')
        self.assertNotEqual(local, proof['expected'])
        errors = self.guard(proof, local)
        self.assertTrue(any('missing sanctioned API declaration' in error and
                            'LaneInserted' in error for error in errors), errors)

    def test_api_sanction_insert_non_declaration_is_rejected(self):
        self.insert_row()
        for text in ('const laneInserted = 1;', 'interface {', ''):
            with self.subTest(text=text):
                self.row['text'] = text
                result = self.generate()
                self.assertNotEqual(result.returncode, 0)
                self.assertIn('api-sanctions.json row 1', result.stderr)
                self.assertIn('only interface or type alias declarations', result.stderr)

    def test_api_sanction_insert_existing_name_is_rejected(self):
        self.insert_row()
        self.row['text'] = 'type TextRange = string;'
        result = self.generate()
        self.assertNotEqual(result.returncode, 0)
        self.assertIn('insert name already exists: TextRange', result.stderr)
        self.insert_row()
        result = self.generate([self.row, self.row])
        self.assertNotEqual(result.returncode, 0)
        self.assertIn('api-sanctions.json row 2', result.stderr)
        self.assertIn('insert name already exists: LaneInsertedValue', result.stderr)

    def test_api_sanction_insert_invalid_fields_are_rejected(self):
        for field in ('before', 'replace'):
            with self.subTest(field=field):
                self.insert_row()
                self.row[field] = 'member'
                result = self.generate()
                self.assertNotEqual(result.returncode, 0)
                self.assertIn('api-sanctions.json row 1', result.stderr)
                self.assertIn('invalid insert sanction', result.stderr)

    def test_api_sanction_function_return_applied_and_guard_accepts(self):
        self.row = dict(path=['getDefaultLibFileName'], kind='FunctionDeclaration', before='string',
                        after='string | undefined', why='Fixture function return')
        proof = self.generated()
        self.assertTrue('function getDefaultLibFileName(options: CompilerOptions): string | undefined;'
                        in proof['expected'], 'sanctioned function return was not emitted')
        self.assertEqual(self.guard(proof, proof['expected']), [])

    def test_api_sanction_overloaded_function_is_rejected(self):
        self.row = dict(path=['getOriginalNode'], kind='FunctionDeclaration', before='Node',
                        after='Node | undefined', why='Fixture overloaded function')
        result = self.generate()
        self.assertNotEqual(result.returncode, 0)
        self.assertIn('api-sanctions.json row 1', result.stderr)
        self.assertIn('ambiguous or missing proof owner', result.stderr)

    def line_row(self):
        # Two members share this path; 98-e sanctions server.ModuleImportResult.error, so the fixture
        # uses a pair no adaptation claims.
        self.row = dict(path=['TypeOnlyExportDeclaration', 'isTypeOnly'], kind='PropertySignature',
                        line=5651, before='true', after='true | 1',
                        why='Fixture disambiguated failure member')

    def test_api_sanction_line_applied_and_guard_accepts(self):
        self.line_row()
        proof = self.generated()
        self.assertTrue(self.row['after'] in proof['expected'], 'selected member was not widened')
        self.assertEqual(self.guard(proof, proof['expected']), [])

    def test_api_sanction_ambiguous_path_without_line_is_rejected(self):
        self.line_row()
        del self.row['line']
        result = self.generate()
        self.assertNotEqual(result.returncode, 0)
        self.assertIn('ambiguous or missing proof owner', result.stderr)
        self.assertIn('TypeOnlyExportDeclaration,isTypeOnly', result.stderr)

    def test_api_sanction_wrong_line_names_path_and_line(self):
        self.line_row()
        self.row['line'] = 5650
        result = self.generate()
        self.assertNotEqual(result.returncode, 0)
        self.assertIn('TypeOnlyExportDeclaration,isTypeOnly line 5650', result.stderr)
        self.assertIn('ambiguous or missing proof owner', result.stderr)
        # A unique path must also refuse a drifted line, as the hard-coded proofs do.
        self.row = dict(path=['TypeReferenceNode', 'typeName'], kind='PropertySignature',
                        line=2706, before='EntityName', after='EntityName | undefined',
                        why='Fixture unique owner drift')
        result = self.generate()
        self.assertNotEqual(result.returncode, 0)
        self.assertIn('TypeReferenceNode,typeName line 2706', result.stderr)

    def test_api_sanction_non_integer_line_is_rejected(self):
        self.line_row()
        self.row['line'] = 5651.5
        result = self.generate()
        self.assertNotEqual(result.returncode, 0)
        self.assertIn('expected a positive integer', result.stderr)
        self.assertIn('TypeOnlyExportDeclaration,isTypeOnly line 5651.5', result.stderr)


if __name__ == '__main__':
    unittest.main()
