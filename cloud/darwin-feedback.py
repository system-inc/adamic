#!/usr/bin/env python3
"""The outcomes that moved between two `go test -json` runs, for cloud/darwin-feedback.sh.

    cloud/darwin-feedback.py <base.jsonl> <candidate.jsonl> <out dir>

Writes moved.txt (one "base -> candidate  package test" line per moved outcome, then each newly failing
test's last output lines) and summary.txt (one line), and prints the summary. A test's outcome is its
last pass, fail or skip; a package with no test outcome but a fail (a build failure) is keyed by the
package alone. A test only one side ran counts as moved from or to "absent".
"""
import json
import sys
from pathlib import Path

Outcomes = ('pass', 'fail', 'skip')


def outcomes(path):
    """{(package, test): outcome} and {(package, test): [output lines]} from one go test -json file."""
    final, output = {}, {}
    for line in Path(path).read_text(errors='replace').splitlines():
        try:
            event = json.loads(line)
        except ValueError:
            continue
        key = (event.get('Package', ''), event.get('Test', ''))
        if event.get('Action') in Outcomes:
            final[key] = event['Action']
        elif event.get('Action') == 'output':
            output.setdefault(key, []).append(event.get('Output', '').rstrip('\n'))
    # A package's own pass or fail only says something when none of its tests reported.
    for package in {p for p, t in final if t == ''}:
        if any(p == package and t != '' for p, t in final):
            del final[(package, '')]
    return final, output


def moved(base, candidate):
    """Sorted (key, base outcome, candidate outcome) for every key whose outcome differs."""
    keys = set(base) | set(candidate)
    return sorted((key, base.get(key, 'absent'), candidate.get(key, 'absent'))
                  for key in keys if base.get(key, 'absent') != candidate.get(key, 'absent'))


def summary(changes):
    counts = {}
    for _, before, after in changes:
        counts['%s->%s' % (before, after)] = counts.get('%s->%s' % (before, after), 0) + 1
    if not changes:
        return '0 outcomes moved'
    return '%d outcomes moved (%s)' % (len(changes), ', '.join('%d %s' % (n, kind) for kind, n in sorted(counts.items())))


def main():
    basePath, candidatePath, out = sys.argv[1:4]
    base, _ = outcomes(basePath)
    candidate, candidateOutput = outcomes(candidatePath)
    changes = moved(base, candidate)
    lines = ['%s -> %s  %s %s' % (before, after, package, test) for (package, test), before, after in changes]
    for (package, test), before, after in changes:
        if after == 'fail':
            tail = [line for line in candidateOutput.get((package, test), []) if line.strip() and not line.startswith(('=== ', '--- '))][-8:]
            lines += ['', '%s %s' % (package, test)] + ['    ' + line for line in tail]
    Path(out, 'moved.txt').write_text('\n'.join(lines) + '\n')
    # The deepest newly failing tests, for the message: a parent fails whenever a child does, and a
    # fixture can fail on its own before its children run.
    failing = {test for (package, test), _, after in changes if after == 'fail'}
    leaves = [test or package for (package, test), _, after in changes
              if after == 'fail' and not (test and any(name.startswith(test + '/') for name in failing))]
    Path(out, 'failing.txt').write_text(''.join(leaf + '\n' for leaf in leaves))
    text = summary(changes)
    Path(out, 'summary.txt').write_text(text + '\n')
    print(text)


if __name__ == '__main__':
    main()
