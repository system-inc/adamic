#!/usr/bin/env python3
"""The review lane's census (@system_adamic, Oct 8, from the witness): every program attached to a review task is a
fixture in the repository's review lane, internal/oracle/testdata/review/{agree,refused}/, so a reviewed program
can't stay outside the gate. Matched by content (git's blob hash), so a renamed fixture still counts. Each pass lists
the programs not on origin/main yet and, when that set changes, pages integration (the reader who commits them).

  cloud/review-census.py            # the loop (launchd: com.adamic.review-census), every ten minutes
  cloud/review-census.py --once     # one pass: prints the missing programs, pages if the set changed
"""
import hashlib
import os
from pathlib import Path
import subprocess
import sys
import time

repository = Path(__file__).resolve().parents[1]
attachments = Path(os.environ.get('ADAMIC_REVIEW_ATTACHMENTS', '/Users/kirkouimet/Projects/ahra/modules/tasks/data/attachments'))
tasks = Path(os.environ.get('ADAMIC_REVIEW_TASKS', str(repository / 'cloud' / 'review-tasks.txt')))
state = Path(os.environ.get('ADAMIC_REVIEW_STATE', os.path.expanduser('~/.adamic-review-census')))
ahraDirectory = os.environ.get('ADAMIC_REVIEW_AHRA_DIR', '/Users/kirkouimet/Projects/ahra')
lane = 'internal/oracle/testdata/review/'


def blob(content):
    """git's blob hash of these bytes, as git hash-object computes it."""
    return hashlib.sha1(b'blob %d\0' % len(content) + content).hexdigest()


def reviewTasks():
    return [line.split()[0] for line in tasks.read_text().splitlines() if line.strip() and not line.startswith('#')]


def laneBlobs():
    """The blob hashes of every program in the lane on origin/main."""
    subprocess.run(['git', '-C', str(repository), 'fetch', '-q', 'origin', '+refs/heads/main:refs/remotes/origin/main'], capture_output=True)
    listing = subprocess.run(['git', '-C', str(repository), 'ls-tree', '-r', 'refs/remotes/origin/main', '--', lane],
                             capture_output=True, text=True).stdout
    return {line.split()[2] for line in listing.splitlines() if line.split()[3].startswith((lane + 'agree/', lane + 'refused/'))}


def missing():
    inLane = laneBlobs()
    outside = []
    for task in reviewTasks():
        for path in sorted((attachments / task).glob('*.a')):
            if blob(path.read_bytes()) not in inLane:
                outside.append('%s/%s' % (task, path.name))
    return outside


def once():
    outside = missing()
    stamp = time.strftime('%H:%M:%S', time.gmtime())
    print('%s %d review programs outside the lane' % (stamp, len(outside)), flush=True)
    state.mkdir(parents=True, exist_ok=True)
    # Pages only for a program it hasn't paged about: a set shrinking as integration commits is progress, not news.
    previous = set((state / 'paged').read_text().split()) if (state / 'paged').exists() else set()
    if set(outside) - previous:
        shown = ', '.join(outside[:12]) + (' and %d more' % (len(outside) - 12) if len(outside) > 12 else '')
        text = ('Review lane census: %d programs attached to review tasks aren\'t fixtures on main yet (internal/oracle/testdata/'
                'review/agree/ or refused/, matched by content): %s. Commit each, with a pending skip if main still gets it wrong.'
                % (len(outside), shown))
        try:
            subprocess.run(['ahra', 'os', 'send', 'system_adamic_integration', text, '--from', 'system_adamic_developer_tools'],
                           cwd=ahraDirectory, capture_output=True, check=True)
            (state / 'paged').write_text('\n'.join(outside) + '\n')
            print('%s paged integration' % stamp, flush=True)
        except (subprocess.CalledProcessError, OSError) as error:
            print('%s could not page integration: %s' % (stamp, error), flush=True)
    else:
        (state / 'paged').write_text('\n'.join(outside) + '\n' if outside else '')


def main():
    while True:
        once()
        if '--once' in sys.argv:
            return
        time.sleep(600)


if __name__ == '__main__':
    main()
