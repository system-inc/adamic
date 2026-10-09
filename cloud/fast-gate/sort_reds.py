#!/usr/bin/env python3
"""Sort gate reds using finished main evidence, content identity and failure kind.

Never turns a red gate green. Integration must also settle infra reruns and stale
units. Missing hashes, missing evidence and lookup errors cannot excuse a bug.
"""
import argparse
import json
import os
import re
import subprocess
import time

INFRA = re.compile(r'killed at \d+(?:\.\d+)? s|(?:exit(?:ed)?|exit status)\s*[:=]?\s*2\b|'
                   r'no space left|(?:clone|fetch) (?:failed|failure)|failed to (?:clone|fetch)|'
                   r'Traceback \(most recent call last\)|ssh.*(?:hang|timed out)|'
                   r'planned stages without a recorded exit|void|never reported', re.I)
TIMING = re.compile(r'budget|deadline|killed|setup exceeded|timed out', re.I)
WRONG = re.compile(r'compile error|build.failed|undefined:|syntax error|assertion|'
                   r'\b(?:got|want|expected)\b|assert.*diff', re.I)
COLD = re.compile(r'cold (?:build|miss)|cache miss|product.*\bmiss\b', re.I)


def identity(row):
    return row.get('unit') or (row.get('package', '') + ' ' + row.get('test', '')).strip()


def ledger(record):
    """Keep leaf units; planned top-level rows fill holes, not duplicate subtests."""
    rows = {}
    for field in ('units', 'product_units', 'cache_drain_units', 'phase_units'):
        for row in record.get(field, []):
            # A Go failure already has its test unit; its enclosing phase isn't a second bug.
            if field == 'phase_units' and any(name in row.get('detail', '') for name in record.get('failed_tests', [])) and row.get('status') == 'failed':
                continue
            rows[identity(row)] = dict(rows.get(identity(row), {}), **row)
    for row in record.get('test_outcomes', []):
        key = identity(row)
        if not any(name == key or name.startswith(key + '/') or name == key + ' (setup)' for name in rows):
            rows[key] = row.copy()
    for key in record.get('failed_tests', []):
        if not any(name == key or name.startswith(key + '/') or name == key + ' (setup)' for name in rows):
            rows[key] = {'unit': key, 'action': 'fail'}
    for name, row in list(rows.items()):
        if name.endswith(' (setup)') and row.get('action') == 'fail':
            parent = name.removesuffix(' (setup)')
            # testUnits inherits the parent's fail action when a child failed. Only
            # an actual setup diagnostic makes that a separate red unit.
            detail = row.get('detail', '')
            if not WRONG.search(detail) and not TIMING.search(detail) and any(
                    key.startswith(parent + '/') and red(child) for key, child in rows.items()):
                del rows[name]
    return rows


def red(row):
    return row.get('action') == 'fail' or row.get('status') in ('failed', 'not run') or (
        'exit' in row and row['exit'] != 0 and row.get('required', True))


def sort_reds(candidate, main=None, context=None):
    context = context or {}
    matching = bool(main and candidate.get('tools_fingerprint') and
                    main.get('tools_fingerprint') == candidate['tools_fingerprint'] and
                    main.get('finished') is True and not main.get('void') and not main.get('stopped'))
    main_rows = ledger(main) if matching else {}
    result = {label: [] for label in ('infra', 'mains', 'candidate', 'stale')}
    result.update(main_record=context.get('main_ref') if matching else None,
                  main_sha=main.get('sha') if matching else context.get('main_sha'),
                  tools_fingerprint=candidate.get('tools_fingerprint'),
                  no_main_record=not matching, version=1)
    rows = ledger(candidate)
    failure = candidate.get('failure') or {}
    step = failure.get('step', '')
    # A stage failure not represented by a ledger red still needs its own label.
    if failure and not any(red(row) and (name == step or name.startswith(step + ' ') or
                           name.startswith(step + '/') or (step in ('tests', 'products', 'budget') and row.get('package')))
                           for name, row in rows.items()):
        rows[step or 'runner'] = {'unit': step or 'runner', 'status': 'failed',
                                 'detail': failure.get('detail', ''), 'tool_crash': failure.get('tool_crash')}
    if candidate.get('void') or candidate.get('stopped'):
        rows['gate void'] = {'unit': 'gate void', 'status': 'not run', 'detail': 'void'}
    for name, row in sorted(rows.items()):
        if not red(row):
            continue
        detail = '\n'.join(str(row.get(field, '')) for field in ('detail', 'error', 'output'))
        if name == step or detail == '\n\n' and len([r for r in rows.values() if red(r)]) == 1:
            detail += '\n' + failure.get('detail', '')
        wrong = bool(WRONG.search(detail))
        cold = row.get('cold_miss') is True or bool(COLD.search(detail))
        paths = set(row.get('input_paths', []))
        touched = bool(paths & set(context.get('candidate_changed', [])))
        known_inputs = bool(row.get('input_hash') and paths and 'candidate_changed' in context)
        other = main_rows.get(name)
        same = bool(other and row.get('input_hash') and row['input_hash'] == other.get('input_hash'))
        if row.get('status') == 'not run' or row.get('tool_crash') or row.get('exit') == 2:
            label, reason = 'infra', 'void, unreported unit or tool failure; rerun'
        elif not wrong and TIMING.search(detail) and cold and known_inputs and not touched and matching:
            label, reason = 'mains', 'cold timing red; candidate does not touch its inputs'
        elif 'Traceback (most recent call last)' in detail or INFRA.search(detail) and not wrong:
            label, reason = 'infra', 'tool limit, storage, transport or runner failure; rerun'
        elif same and red(other) and other.get('status') != 'not run':
            label, reason = 'mains', 'same unit and input hash red in newest main record on these tools'
        elif (other and not red(other) and other.get('input_hash') and not same and known_inputs and
              not touched and paths & set(context.get('main_changed', []))):
            label, reason = 'stale', "main has since fixed this, recut, don't land"
        else:
            label, reason = 'candidate', 'no matching red main evidence for these inputs'
        result[label].append({'unit': name, 'input_hash': row.get('input_hash'), 'reason': reason})
    result['candidate_reds'] = len(result['candidate'])
    result['status'] = ('candidate reds: %d' % result['candidate_reds']) if matching else 'no main record on these tools'
    if result['stale']:
        result['status'] += "; stale: main has since fixed this, recut, don't land"
    return result


class MainRecords:
    """Read newest finished record at origin/main, with a 45 s total lookup budget."""
    def __init__(self, tree):
        self.tree = tree
        self.deadline = time.monotonic() + 45

    def git(self, *args):
        remaining = self.deadline - time.monotonic()
        if remaining <= 0:
            raise TimeoutError('main record lookup exceeded 45 s')
        return subprocess.run(['git', '-C', self.tree] + list(args), capture_output=True,
                              text=True, check=True, timeout=min(10, remaining)).stdout.strip()

    def newest(self, fingerprint):
        main_sha = self.git('ls-remote', 'origin', 'refs/heads/main').split()[0]
        listing = self.git('ls-remote', 'origin', 'refs/heads/gate-logs/' + main_sha[:12] + '/*')
        refs = [line.split()[1] for line in listing.splitlines()
                if line.split()[1].endswith(('/fast', '/full-main'))]
        # Timestamps order records; prefer whole at equal timestamps.
        refs.sort(key=lambda ref: (ref.split('/')[-2], ref.endswith('/full-main')), reverse=True)
        for ref in refs:
            self.git('fetch', '--no-tags', 'origin', ref)
            kind = 'full' if ref.endswith('/full-main') else 'fast'
            try:
                record = json.loads(self.git('show', 'FETCH_HEAD:' + kind + '.json'))
            except (ValueError, subprocess.CalledProcessError):
                continue
            if (record.get('sha') == main_sha and record.get('finished') is True and
                record.get('tools_fingerprint') == fingerprint and not record.get('void') and not record.get('stopped')):
                return record, {'main_sha': main_sha, 'main_ref': ref.removeprefix('refs/heads/')}
        return None, {'main_sha': main_sha}

    def context(self, candidate, info):
        main_sha = info.get('main_sha')
        if not main_sha:
            return info
        self.git('fetch', '--no-tags', 'origin', main_sha)
        fork = self.git('merge-base', candidate['sha'], main_sha)
        info.update(fork=fork,
                    candidate_changed=self.git('diff', '--name-only', fork, candidate['sha']).splitlines(),
                    main_changed=self.git('diff', '--name-only', fork, main_sha).splitlines())
        return info


def sort_record(record, tree):
    reader = MainRecords(tree)
    try:
        main, info = reader.newest(record.get('tools_fingerprint'))
        if main:
            info = reader.context(record, info)
        return sort_reds(record, main, info)
    except (OSError, ValueError, IndexError, subprocess.SubprocessError, TimeoutError) as error:
        result = sort_reds(record)
        result['lookup_error'] = str(error)
        return result


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('record')
    parser.add_argument('--tree', required=True)
    args = parser.parse_args()
    with open(args.record) as handle:
        record = json.load(handle)
    print(json.dumps(sort_record(record, args.tree), indent=2))


if __name__ == '__main__':
    main()
