#!/usr/bin/env python3
"""diff.py reads a day back out of its shards' job logs whole, refuses logs that don't make one day, and holds a day
against the one before: every section, a void that never moves, and the exit a task opens on."""
import base64
import gzip
import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest

DIFF = Path(os.environ.get('AGREEMENT_DIFF_UNDER_TEST', Path(__file__).resolve().parent / 'diff.py'))
HEAD = 'a' * 40


def diff(*arguments):
    # -I, as job.sh runs it.
    return subprocess.run([sys.executable, '-I', str(DIFF), *arguments], capture_output=True, text=True, timeout=30)


def report(shard, programs, head=HEAD, **counts):
    """A tool's --json report for one shard, with fields compact drops beside the ones it keeps."""
    fields = {'head': head, 'manifest_blob': 'b' * 40, 'shard': shard, 'complete_corpus': False, 'verdict': 'partial',
              'void': False, 'disagreements': [], 'compiler_failures': [], 'stale_runtime_stops': [], 'timed_out_runs': 0,
              'programs': programs}
    fields.update({name: counts.get(name, 0) for name in
                   ('compared', 'agreeing', 'ruled', 'checked_stops', 'pending', 'disagreeing', 'memory_unsafe', 'refused')})
    return fields


class Scratch(unittest.TestCase):
    def setUp(self):
        self.scratch = tempfile.TemporaryDirectory()
        self.addCleanup(self.scratch.cleanup)
        self.directory = Path(self.scratch.name)

    def write(self, name, text):
        path = self.directory / name
        path.write_text(text)
        return str(path)

    def log(self, name, full_report, stamp='', width=0):
        """A job log as job.sh prints it: the tool's lines, the AGREEMENT_DAILY line, then the compact report, made by
        diff.py compact itself, gzipped and base64'd on one line between the markers. stamp goes before every line, as a
        log that stamps its lines would, and a width splits the report's line into lines of that width, as Loom's log
        does at 49,152 bytes."""
        source = self.write(name + '.report.json', json.dumps(full_report))
        made = diff('compact', source)
        self.assertEqual(made.returncode, 0, made.stderr)
        blob = base64.b64encode(gzip.compress(made.stdout.encode())).decode()
        lines = ['sandboxed-job: npm installed offline in 1 projects', 'pending stage3/p.ts: awaits #abc1234',
                 'agreement: compared 2, agreeing 1, ruled 0, checked stops 0, pending 1, disagreeing 0, memory-unsafe 0, refused 0',
                 f"AGREEMENT_DAILY shard={full_report['shard']} head={full_report['head']} exit=0",
                 'AGREEMENT_REPORT_BEGIN', *([blob[at:at + width] for at in range(0, len(blob), width)] if width else [blob]),
                 'AGREEMENT_REPORT_END']
        return self.write(name, ''.join(stamp + line + '\n' for line in lines))

    def extract(self, *logs):
        return diff('extract', *logs)


class Compact(Scratch):
    def test_keeps_the_named_fields_and_leaves_absent_ones_absent(self):
        source = self.write('report.json', json.dumps(report('1/4', [
            {'path': 'a.ts', 'corpus': 'stage3', 'judgment': 'agreeing', 'outcome_hash': 'h1'},
            {'path': 'b.ts', 'corpus': 'stage3', 'judgment': 'refused'},
            {'path': 'c.ts', 'corpus': 'gaps', 'judgment': 'pending', 'outcome_hash': 'h3', 'task': 'abc1234'},
        ], compared=2, agreeing=1, pending=1, refused=1)))
        made = diff('compact', source)
        self.assertEqual(made.returncode, 0, made.stderr)
        self.assertEqual(json.loads(made.stdout), {
            'head': HEAD, 'shard': '1/4', 'verdict': 'partial', 'void': False, 'void_reason': '',
            'counts': {'compared': 2, 'agreeing': 1, 'ruled': 0, 'checked_stops': 0, 'pending': 1, 'disagreeing': 0,
                       'memory_unsafe': 0, 'refused': 1},
            'programs': [{'path': 'a.ts', 'judgment': 'agreeing', 'outcome_hash': 'h1'},
                         {'path': 'b.ts', 'judgment': 'refused'},
                         {'path': 'c.ts', 'judgment': 'pending', 'outcome_hash': 'h3', 'task': 'abc1234'}]})
        self.assertEqual(len(made.stdout.splitlines()), 1)

    def test_a_report_that_doesnt_parse_is_refused(self):
        for text in ('', '{"programs": ', '[]', '{"head": "x"}'):
            with self.subTest(text=text):
                made = diff('compact', self.write('report.json', text))
                self.assertEqual(made.returncode, 2)
                self.assertIn('report.json', made.stderr)


class Extract(Scratch):
    def test_round_trip_merges_the_shards(self):
        first = self.log('one.log', report('2/2', [
            {'path': 'b.ts', 'judgment': 'refused'},
            {'path': 'd.ts', 'judgment': 'disagreeing', 'outcome_hash': 'h4'},
        ], compared=1, disagreeing=1, refused=1))
        second = self.log('two.log', report('1/2', [
            {'path': 'a.ts', 'judgment': 'agreeing', 'outcome_hash': 'h1'},
            {'path': 'c.ts', 'judgment': 'pending', 'outcome_hash': 'h3', 'task': 'abc1234'},
        ], compared=2, agreeing=1, pending=1), stamp='2026-11-21T06:00:00Z ', width=16)
        made = self.extract(first, second)
        self.assertEqual(made.returncode, 0, made.stderr)
        self.assertEqual(json.loads(made.stdout), {
            'head': HEAD, 'shards': ['1/2', '2/2'],
            'counts': {'compared': 3, 'agreeing': 1, 'ruled': 0, 'checked_stops': 0, 'pending': 1, 'disagreeing': 1,
                       'memory_unsafe': 0, 'refused': 1},
            'programs': {'a.ts': {'judgment': 'agreeing', 'outcome_hash': 'h1'},
                         'b.ts': {'judgment': 'refused'},
                         'c.ts': {'judgment': 'pending', 'outcome_hash': 'h3', 'task': 'abc1234'},
                         'd.ts': {'judgment': 'disagreeing', 'outcome_hash': 'h4'}}})

    def assertRefused(self, made, name):
        self.assertEqual(made.returncode, 2, made.stdout)
        self.assertIn(name, made.stderr)
        self.assertEqual(made.stdout, '')

    def test_a_report_with_no_end_is_refused(self):
        good = self.log('good.log', report('1/2', []))
        open_ended = self.write('open.log', 'AGREEMENT_REPORT_BEGIN\nbm90IGd6aXA=\n')
        self.assertRefused(self.extract(good, open_ended), 'open.log')

    def test_a_log_with_no_report_is_refused(self):
        good = self.log('good.log', report('1/2', []))
        bare = self.write('bare.log', 'agreement-daily: go build ./cmd/adamic-admission-delta failed, so nothing ran\n')
        self.assertRefused(self.extract(good, bare), 'bare.log')

    def test_a_report_that_doesnt_decode_is_refused(self):
        broken = self.write('broken.log', 'AGREEMENT_REPORT_BEGIN\nbm90IGd6aXA=\nAGREEMENT_REPORT_END\n')
        self.assertRefused(self.extract(broken), 'broken.log')

    def test_a_log_with_two_reports_is_refused(self):
        one = Path(self.log('one.log', report('1/2', [])))
        twice = self.write('twice.log', one.read_text() * 2)
        self.assertRefused(self.extract(twice), 'twice.log')

    def test_two_logs_of_one_shard_are_refused(self):
        first = self.log('first.log', report('1/2', [{'path': 'a.ts', 'judgment': 'refused'}]))
        again = self.log('again.log', report('1/2', [{'path': 'b.ts', 'judgment': 'refused'}]))
        self.assertRefused(self.extract(first, again), 'again.log')

    def test_logs_of_different_heads_are_refused(self):
        first = self.log('first.log', report('1/2', []))
        other = self.log('other.log', report('2/2', [], head='c' * 40))
        self.assertRefused(self.extract(first, other), 'other.log')

    def test_shards_of_different_counts_are_refused(self):
        first = self.log('first.log', report('1/2', []))
        other = self.log('other.log', report('2/4', []))
        self.assertRefused(self.extract(first, other), 'other.log')

    def test_a_path_in_two_shards_is_refused(self):
        first = self.log('first.log', report('1/2', [{'path': 'a.ts', 'judgment': 'refused'}]))
        other = self.log('other.log', report('2/2', [{'path': 'a.ts', 'judgment': 'refused'}]))
        self.assertRefused(self.extract(first, other), 'other.log')


class Compare(Scratch):
    def day(self, name, programs, head=HEAD, **counts):
        counts = {name: counts.get(name, 0) for name in
                  ('compared', 'agreeing', 'ruled', 'checked_stops', 'pending', 'disagreeing', 'memory_unsafe', 'refused')}
        return self.write(name, json.dumps({'head': head, 'shards': ['1/2', '2/2'], 'counts': counts, 'programs': programs}))

    def compare(self, yesterday, today):
        made = diff('compare', self.day('yesterday.json', yesterday, head='b' * 40), self.day('today.json', today, agreeing=7))
        self.assertIn(made.returncode, (0, 1), made.stderr)
        sections, name = {}, None
        for line in made.stdout.splitlines()[2:]:
            if line.startswith('  '):
                sections[name].append(line.strip())
            else:
                name = line.rsplit(' (', 1)[0]
                sections[name] = []
        return made, sections

    def test_a_day_like_the_one_before_holds(self):
        programs = {'a.ts': {'judgment': 'agreeing', 'outcome_hash': 'h1'}, 'b.ts': {'judgment': 'refused'}}
        made, sections = self.compare(programs, programs)
        self.assertEqual(made.returncode, 0)
        self.assertEqual(sections, {name: [] for name in ('moved', 'output changed', 'new', 'gone', 'void', 'pending')})
        self.assertEqual(made.stdout.splitlines()[0], f"head {HEAD}, shards 1/2 2/2; yesterday {'b' * 40}")
        self.assertIn('agreeing 7', made.stdout.splitlines()[1])

    def test_a_judgment_that_changes_moves(self):
        made, sections = self.compare({'a.ts': {'judgment': 'agreeing', 'outcome_hash': 'h1'}},
                                      {'a.ts': {'judgment': 'disagreeing', 'outcome_hash': 'h2'}})
        self.assertEqual(made.returncode, 1)
        self.assertEqual(sections['moved'], ['a.ts: agreeing -> disagreeing'])
        self.assertEqual(sections['output changed'], [])

    def test_an_outcome_that_changes_under_the_same_judgment_is_output_changed(self):
        made, sections = self.compare({'a.ts': {'judgment': 'ruled', 'outcome_hash': 'h1'}},
                                      {'a.ts': {'judgment': 'ruled', 'outcome_hash': 'h2'}})
        self.assertEqual(made.returncode, 1)
        self.assertEqual(sections['output changed'], ['a.ts'])
        self.assertEqual(sections['moved'], [])

    def test_an_outcome_hash_on_one_day_alone_changes_nothing(self):
        made, sections = self.compare({'a.ts': {'judgment': 'agreeing'}}, {'a.ts': {'judgment': 'agreeing', 'outcome_hash': 'h2'}})
        self.assertEqual(made.returncode, 0)
        self.assertEqual(sections['output changed'], [])

    def test_new_and_gone(self):
        for yesterday, today, section, line in (({}, {'a.ts': {'judgment': 'agreeing'}}, 'new', 'a.ts: agreeing'),
                                                ({'a.ts': {'judgment': 'refused'}}, {}, 'gone', 'a.ts: refused')):
            with self.subTest(section=section):
                made, sections = self.compare(yesterday, today)
                self.assertEqual(made.returncode, 1)
                self.assertEqual(sections[section], [line])

    def test_a_void_never_moves_or_changes(self):
        for was, now in (('agreeing', 'void'), ('void', 'disagreeing'), ('void', 'void')):
            with self.subTest(was=was, now=now):
                made, sections = self.compare({'a.ts': {'judgment': was, 'outcome_hash': 'h1'}},
                                              {'a.ts': {'judgment': now, 'outcome_hash': 'h2'}})
                self.assertEqual(made.returncode, 0, made.stdout)
                self.assertEqual(sections['void'], [f'a.ts: {was} -> {now}'])
                self.assertEqual(sections['moved'], [])
                self.assertEqual(sections['output changed'], [])

    def test_pending_names_each_task_and_holds(self):
        pending = {'p.ts': {'judgment': 'pending', 'outcome_hash': 'h1', 'task': 'abc1234'},
                   'q.ts': {'judgment': 'pending', 'outcome_hash': 'h2', 'task': 'def5678'}}
        made, sections = self.compare(pending, pending)
        self.assertEqual(made.returncode, 0)
        self.assertEqual(sections['pending'], ['p.ts: #abc1234', 'q.ts: #def5678'])

    def test_a_day_that_doesnt_parse_is_refused(self):
        made = diff('compare', self.write('yesterday.json', '{'), self.day('today.json', {}))
        self.assertEqual(made.returncode, 2)
        self.assertIn('yesterday.json', made.stderr)


if __name__ == '__main__':
    unittest.main()
