#!/usr/bin/env python3
"""The automatic A/B for a red that may be load (@system_adamic from the witness, Oct 9 00:34Z: nobody calls a red
load by hand again). A red whose first failure is a go test timeout, a stall guard or a kill gets one job for
Loom's side pool, which times that test on the red's main and on the candidate on one Codex instance and writes
'regression <ratio>x' or 'load <ratio>x' beside the job. The verdict goes to whoever heard the red.

    cloud/ab-trigger.py write --candidate <sha> --main <sha> --red <gate-logs ref> --first-failure <file> [--notify a,b]
    cloud/ab-trigger.py collect

write reads the first failure's text (the gate's first-failure.txt, or a fast gate's FIRST FAILURE block): its
first line names the package and the test, and the rest is searched for the kind. A red of any other kind writes
nothing. The job is <jobs>/<candidate sha12>-<hash of package and test>.json, once per red and test.

collect sends each finished verdict once, to the job's notify list, then marks it sent.

The test is compared leaf to leaf, never by a parent's Elapsed (Oct 9: a parent's Elapsed excludes its parallel
subtests, and a 37x 'regression' was 19 s against 720 s of the same work), so the -run pattern names the leaf.
"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import subprocess
import sys
import time

jobs = Path(os.environ.get('ADAMIC_AB_JOBS', os.path.expanduser('~/.loom/jobs/ab')))
ahra = Path(os.environ.get('ADAMIC_FAST_GATE_AHRA_DIR', '/Users/kirkouimet/Projects/ahra'))
# What a red that may be load looks like, in its first failure's text.
kinds = (
    ('timeout', re.compile(r'panic: test timed out after')),
    ('stall', re.compile(r'stalled: no output for|stall guard')),
    ('kill', re.compile(r'signal: killed|killed by (?:the )?(?:guard|watchdog)')),
    # A test's own budget kill (Oct 9: cohere's split units stop a step at 90 s and print this), which may be load.
    ('deadline', re.compile(r'unit deadline exceeded name=')),
)
testLine = re.compile(r'^(\S+/\S+) (Test[^\s]*)$')
# A go test timeout panics naming the package alone; the test is the first one it lists as running.
packageLine = re.compile(r'^(\S+/\S+)$')
runningTest = re.compile(r'running tests:\s*\n\s*(Test\S*) \(')


def kindOf(text):
    for kind, pattern in kinds:
        if pattern.search(text):
            return kind
    return None


def runPattern(test):
    """go test -run for exactly this leaf: each level anchored and quoted."""
    return '/'.join('^%s$' % re.escape(part) for part in test.split('/'))


def write(candidate, main, red, text, notify):
    lines = [line.strip() for line in text.splitlines() if line.strip()]
    named = testLine.match(lines[0]) if lines else None
    if named is None and lines and packageLine.match(lines[0]) and runningTest.search(text):
        named = re.match(r'(.*) (.*)', '%s %s' % (lines[0], runningTest.search(text).group(1)))
    kind = kindOf(text)
    if named is None or kind is None:
        return None
    package, test = named.groups()
    identifier = '%s-%s' % (candidate[:12], hashlib.sha1(('%s %s' % (package, test)).encode()).hexdigest()[:8])
    jobs.mkdir(parents=True, exist_ok=True)
    path = jobs / (identifier + '.json')
    job = {'candidate': candidate, 'main': main, 'package': package, 'test': runPattern(test), 'leaf': test,
           'red': red, 'kind': kind, 'notify': notify, 'written': time.strftime('%Y-%m-%dT%H:%M:%SZ', time.gmtime())}
    try:
        with open(path, 'x') as handle:
            json.dump(job, handle, indent=2)
            handle.write('\n')
    except FileExistsError:
        return None
    return path


def collect():
    sent = []
    for verdictPath in sorted(jobs.glob('*.verdict')):
        jobPath = verdictPath.with_suffix('.json')
        marker = verdictPath.with_suffix('.sent')
        if marker.exists() or not jobPath.exists():
            continue
        job = json.loads(jobPath.read_text())
        verdict = verdictPath.read_text().strip()
        if not verdict:
            continue
        first = verdict.splitlines()[0]
        reading = ('a regression in the candidate, not load' if first.startswith('regression') else
                   'load, not the change' if first.startswith('load') else 'unread')
        text = ('A/B for the %s red %s (%s %s, %s at %s against main %s, same instance): %s. So it reads as %s.' % (
            job['kind'], job['red'], job['package'].split('/adamic/', 1)[-1], job['leaf'], job['candidate'][:12],
            job['kind'], job['main'][:12], ' '.join(verdict.split()), reading))
        text = re.sub(r'[A-Z]{3,}', lambda match: match.group(0).lower(), text)
        for name in job.get('notify') or ['system_adamic_integration']:
            subprocess.run(['ahra', 'os', 'send', name, text, '--from', 'system_adamic_developer_tools'], cwd=str(ahra), capture_output=True)
        marker.write_text(time.strftime('%Y-%m-%dT%H:%M:%SZ', time.gmtime()) + '\n')
        sent.append(text)
    return sent


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('action', choices=('write', 'collect'))
    parser.add_argument('--candidate')
    parser.add_argument('--main')
    parser.add_argument('--red')
    parser.add_argument('--first-failure')
    parser.add_argument('--notify', default='')
    arguments = parser.parse_args()
    if arguments.action == 'collect':
        for text in collect():
            print('%s sent: %s' % (time.strftime('%H:%M:%S', time.gmtime()), text))
        return
    if not all(re.fullmatch(r'[0-9a-f]{40}', value or '') for value in (arguments.candidate, arguments.main)):
        sys.exit('ab-trigger: --candidate and --main must be 40 hex digits')
    path = write(arguments.candidate, arguments.main, arguments.red or '', Path(arguments.first_failure).read_text(errors='replace'),
                 [name for name in arguments.notify.split(',') if name])
    if path is not None:
        print('%s wrote %s' % (time.strftime('%H:%M:%S', time.gmtime()), path))


if __name__ == '__main__':
    main()
