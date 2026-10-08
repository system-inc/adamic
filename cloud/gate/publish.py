#!/usr/bin/env python3
"""Publish one shard attempt without touching source or overwriting newer attempts."""
import argparse
import json
from pathlib import Path
import re
import shutil
import subprocess
import tempfile


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--origin', required=True)
    parser.add_argument('--branch', required=True)
    parser.add_argument('--attempt', type=int, required=True)
    parser.add_argument('--commit', required=True)
    payload = parser.add_mutually_exclusive_group(required=True)
    payload.add_argument('--failure')
    payload.add_argument('--archive', type=Path)
    parser.add_argument('--setup-log', type=Path)
    args = parser.parse_args()
    match = re.fullmatch(r'gate-logs/([0-9a-f]{12})/([0-9A-Za-z_-]+)/(?P<shard>shard-[0-9]+)', args.branch)
    if not match or args.attempt < 1 or args.attempt > 3:
        parser.error('only a run-specific shard log branch and attempts 1..3 are allowed')
    if args.failure is not None and (not args.failure or args.setup_log is None):
        parser.error('a failure needs a reason and --setup-log')
    if args.archive is not None and args.archive.name != match['shard']+'.tgz':
        parser.error('archive name must match the shard')
    # A failed checkout can publish a marker for the requested commit. Successful
    # logs must actually come from that checkout; the runner verifies the frozen inputs.
    if not re.fullmatch(r'[0-9a-f]{40}', args.commit) or args.commit[:12] != match[1]:
        parser.error('requested commit differs from the log branch')
    if args.archive is not None:
        commit = subprocess.check_output(['git', 'rev-parse', 'HEAD'], text=True).strip()
        if commit != args.commit:
            parser.error('checkout commit differs from requested commit')
    record = dict(Commit=args.commit, Shard=match['shard'], Attempt=args.attempt)
    with tempfile.TemporaryDirectory(prefix='gate-publish-') as directory:
        root = Path(directory)
        def git(*arguments, check=True):
            return subprocess.run(['git', '-C', directory, *arguments], check=check,
                                  stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True, timeout=60)
        git('init', '-q')
        git('remote', 'add', 'origin', args.origin)
        for retry in range(3):
            # ls-remote errors must not look like an absent branch.
            existing = git('ls-remote', '--heads', 'origin', 'refs/heads/'+args.branch).stdout.strip()
            if existing:
                git('fetch', '-q', 'origin', 'refs/heads/'+args.branch)
                git('checkout', '-q', '-B', 'publish', 'FETCH_HEAD')
                previous = root/'attempt.json'
                if not previous.exists():
                    previous = root/'failure.json'
                if previous.exists():
                    old = json.loads(previous.read_text())
                    if old['Attempt'] > args.attempt:
                        raise SystemExit('refusing to overwrite newer shard attempt')
                    if old['Attempt'] == args.attempt and (root/(match['shard']+'.tgz')).exists() and args.failure is not None:
                        raise SystemExit('refusing failure after this attempt published logs')
                git('rm', '-q', '-r', '--ignore-unmatch', '.')
            else:
                git('checkout', '-q', '--orphan', 'publish-'+str(retry))
            if args.failure is not None:
                (root/'failure.json').write_text(json.dumps(dict(record, Reason=args.failure), indent=2)+'\n')
                shutil.copyfile(args.setup_log, root/'setup.log')
            else:
                shutil.copyfile(args.archive, root/args.archive.name)
                (root/'attempt.json').write_text(json.dumps(record, indent=2)+'\n')
            git('add', '.')
            git('-c', 'user.name=Adamic gate', '-c', 'user.email=gate@system.inc',
                'commit', '-q', '--allow-empty', '-m', f"Gate {match['shard']} attempt {args.attempt} {'failure' if args.failure is not None else 'logs'}")
            pushed = git('push', '-q', 'origin', 'HEAD:refs/heads/'+args.branch, check=False)
            if pushed.returncode == 0:
                print(f'published {args.branch} attempt {args.attempt}')
                return
            if retry == 2:
                raise SystemExit(pushed.stderr)


if __name__ == '__main__':
    main()
