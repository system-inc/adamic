#!/usr/bin/env python3
"""Exercise the real fleet controller and publisher with disposable local Git origins."""
import io
import json
import os
from pathlib import Path
import re
import subprocess
import tarfile
import tempfile
import time
import unittest

DIRECTORY = Path(__file__).resolve().parent
FLEET = Path(os.environ.get('FLEET_SCRIPT', DIRECTORY/'gate/fleet.sh')).resolve()
PUBLISH = Path(os.environ.get('FLEET_PUBLISH_SCRIPT', DIRECTORY/'gate/publish.py')).resolve()

# This launcher uses the publisher embedded in the actual shard brief. It does not
# pretend that a branch exists: every marker and archive goes through a normal Git push.
LAUNCHER = r'''#!/usr/bin/env python3
import io,json,os,re,subprocess,sys,tarfile,time
from pathlib import Path
root=Path(os.environ['FLEET_TEST_ROOT'])
brief=Path(sys.argv[2]).read_text()
sha=os.environ['FLEET_TEST_SHA']
prefix=re.search(r'gate-logs/[a-f0-9]{12}/[0-9A-Za-z_-]+',brief)[0]
name=Path(sys.argv[2]).stem.split('-attempt-')[0]
record=dict(name=name,label=sys.argv[1],time=time.monotonic(),brief=sys.argv[2])
if name.startswith('shard-'):
    record['attempt']=int(re.search(r'Attempt: (\d+)',brief)[1]) if 'Attempt: ' in brief else 1+sum(json.loads(line)['name']==name for line in (root/'launches.jsonl').read_text().splitlines())
with (root/'launches.jsonl').open('a') as stream:stream.write(json.dumps(record)+'\n')
def git(*args):return subprocess.check_output(['git',*args],text=True).strip()
def publish_other(payload,message):
    scratch=root/('publish-'+name)
    scratch.mkdir()
    git('init','-q',str(scratch))
    branch=prefix+'/'+name
    if git('ls-remote','--heads',str(root/'origin.git'),'refs/heads/'+branch):
        git('-C',str(scratch),'fetch','-q',str(root/'origin.git'),'refs/heads/'+branch)
        git('-C',str(scratch),'checkout','-q','-B','logs','FETCH_HEAD')
        git('-C',str(scratch),'rm','-q','-r','--ignore-unmatch','.')
    else:git('-C',str(scratch),'checkout','-q','--orphan','logs')
    for filename,data in payload.items():(scratch/filename).write_bytes(data)
    git('-C',str(scratch),'add','.')
    git('-C',str(scratch),'-c','user.name=Test','-c','user.email=test@example.invalid','commit','-q','-m',message)
    git('-C',str(scratch),'push','-q',str(root/'origin.git'),'HEAD:refs/heads/'+prefix+'/'+name)
def archive(files):
    output=io.BytesIO()
    with tarfile.open(fileobj=output,mode='w:gz') as tar:
        for filename,data in files.items():
            entry=tarfile.TarInfo(filename);entry.size=len(data);tar.addfile(entry,io.BytesIO(data))
    return output.getvalue()
if name=='plan':
    plan=json.dumps(dict(Commit=sha,Count=2,Units=[])).encode()
    publish_other({'plan.tgz':archive({'plan.json':plan})},'Gate plan')
elif name=='merge':
    observed={}
    for i in range(2):
        shard='shard-'+str(i)
        data=subprocess.check_output(['git','--git-dir='+str(root/'origin.git'),'show',prefix+'/'+shard+':'+shard+'.tgz'])
        with tarfile.open(fileobj=io.BytesIO(data),mode='r:gz') as tar:
            observed[shard]=json.load(tar.extractfile('gate-out/summary.json'))
    (root/'merged.json').write_text(json.dumps(observed))
    publish_other({'merge.tgz':archive({'merged/merged.json':json.dumps(observed).encode()})},'GATE GREEN '+sha)
else:
    attempt=record['attempt']
    if os.environ.get('FLEET_TEST_MODE')=='launcher' and name=='shard-0' and attempt==1:
        raise SystemExit(7)
    publisher=root/(name+'-publisher.py')
    if '```python\n' in brief:
        publisher.write_text(brief.split('```python\n',1)[1].split('\n```',1)[0])
    else:publisher.write_text(Path(os.environ['FLEET_TEST_PUBLISH']).read_text())
    args=['python3',str(publisher),'--origin',str(root/'origin.git'),'--branch',prefix+'/'+name,'--attempt',str(attempt),'--commit',sha]
    fail=name=='shard-0' and (attempt==1 or os.environ.get('FLEET_TEST_MODE')=='cap')
    if fail:
        if os.environ.get('FLEET_TEST_MODE')=='notes-unwritable':
            (Path(sys.argv[2]).parent/'notes.jsonl').mkdir()
        if os.environ.get('FLEET_TEST_MODE')=='late-marker':
            if os.fork(): raise SystemExit(0)
            time.sleep(0.5)
        log=root/(name+'-setup.log');log.write_text('golang.org: Forbidden\n')
        subprocess.run(args+['--failure','setup: golang.org Forbidden','--setup-log',str(log)],check=True)
        with (root/'markers.jsonl').open('a') as stream:stream.write(json.dumps(dict(attempt=attempt,time=time.monotonic()))+'\n')
        if os.environ.get('FLEET_TEST_MODE')=='invalid':
            marker=dict(Commit='a'*40,Shard=name,Attempt=attempt,Reason='wrong commit marker')
            publish_other({'failure.json':json.dumps(marker).encode(),'setup.log':b'Forbidden'},'Wrong commit marker')
    else:
        if name=='shard-0' and attempt==2 and os.environ.get('FLEET_TEST_MODE')=='delayed':
            if os.fork(): raise SystemExit(0)
            time.sleep(0.25)
        data=dict(shard=name,attempt=attempt,pass_count=1,fail=0,skip=0)
        output=root/(name+'.tgz');output.write_bytes(archive({'gate-out/summary.json':json.dumps(data).encode()}))
        subprocess.run(args+['--archive',str(output)],check=True)
'''


class FleetTests(unittest.TestCase):
    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory(prefix='fleet-test-')
        self.addCleanup(self.temporary.cleanup)
        self.root = Path(self.temporary.name)
        self.repo = self.root/'checkout'
        self.git('init', '-q', '--bare', str(self.root/'origin.git'))
        self.git('init', '-q', str(self.repo))
        self.git('-C', str(self.repo), '-c', 'user.name=Test', '-c', 'user.email=test@example.invalid',
                 'commit', '-q', '--allow-empty', '-m', 'Source')
        self.sha = self.git('-C', str(self.repo), 'rev-parse', 'HEAD')
        self.git('-C', str(self.repo), 'remote', 'add', 'origin', str(self.root/'origin.git'))
        self.git('-C', str(self.repo), 'push', '-q', 'origin', 'HEAD:refs/heads/source')
        self.launcher = self.root/'launcher.py'
        self.launcher.write_text(LAUNCHER)
        self.launcher.chmod(0o755)
        self.env = dict(os.environ, ADAMIC_GATE_LAUNCH=str(self.launcher),
                        ADAMIC_GATE_FLEET_DIRECTORY=str(self.root/'fleet'),
                        ADAMIC_GATE_FLEET_POLL='0.02', ADAMIC_GATE_PLAN_WAIT='5',
                        ADAMIC_GATE_FLEET_MERGE_WAIT='5', ADAMIC_GATE_FLEET_SHARD_WAIT='10',
                        FLEET_TEST_ROOT=str(self.root), FLEET_TEST_SHA=self.sha,
                        FLEET_TEST_PUBLISH=str(PUBLISH))
        self.command_number = 0

    def git(self, *arguments):
        return subprocess.check_output(['git', *arguments], stderr=subprocess.STDOUT, text=True).strip()

    def run_fleet(self, *arguments, timeout=65):
        self.command_number += 1
        path = self.root/f'controller-{self.command_number}.log'
        with path.open('wb') as stream:
            result = subprocess.run(['bash', str(FLEET), *arguments], cwd=self.repo, env=self.env,
                                    stdout=stream, stderr=subprocess.STDOUT, timeout=timeout)
        return result.returncode, path.read_text()

    def records(self, filename):
        path = self.root/filename
        return [json.loads(line) for line in path.read_text().splitlines()] if path.exists() else []

    def publish(self, branch, attempt, reason=None):
        log = self.root/'setup.log'
        log.write_text('golang.org: Forbidden\n')
        args = ['python3', str(PUBLISH), '--origin', str(self.root/'origin.git'),
                '--branch', branch, '--attempt', str(attempt), '--commit', self.sha]
        if reason is None:
            archive = self.root/'shard-0.tgz'
            with tarfile.open(archive, 'w:gz') as tar:
                data = b'{"replacement":true}'
                entry = tarfile.TarInfo('gate-out/summary.json')
                entry.size = len(data)
                tar.addfile(entry, io.BytesIO(data))
            args += ['--archive', str(archive)]
        else:
            args += ['--failure', reason, '--setup-log', str(log)]
        with (self.root/'publisher.log').open('ab') as stream:
            return subprocess.run(args, cwd=self.repo, stdout=stream, stderr=subprocess.STDOUT).returncode

    def test_first_setup_failure_replaced_within_minute_and_merge_uses_replacement(self):
        # Exercise the production default, not a shortened polling interval.
        self.env.pop('ADAMIC_GATE_FLEET_POLL')
        self.env['ADAMIC_GATE_FLEET_SHARD_WAIT'] = '70'
        started = time.monotonic()
        code, output = self.run_fleet('run', self.sha, '2')
        launches = self.records('launches.jsonl')
        failed = [r for r in launches if r['name']=='shard-0']
        healthy = [r for r in launches if r['name']=='shard-1']
        marker_time = self.records('markers.jsonl')[0]['time']
        elapsed = failed[1]['time']-marker_time if len(failed)>1 else None
        measurement = dict(exit=code, wall=time.monotonic()-started, replacement_seconds=elapsed,
                           launches=[dict(name=r['name'],attempt=r.get('attempt')) for r in launches])
        if os.environ.get('FLEET_BENCHMARK_OUTPUT'):
            Path(os.environ['FLEET_BENCHMARK_OUTPUT']).write_text(json.dumps(measurement,indent=2)+'\n')
        self.assertEqual(code, 0, output)
        self.assertEqual([r['attempt'] for r in failed], [1, 2])
        self.assertEqual(len(healthy), 1)
        self.assertLess(elapsed, 60)
        self.assertIn('Attempt: 2.', Path(failed[1]['brief']).read_text())
        merged = json.loads((self.root/'merged.json').read_text())
        self.assertEqual(merged['shard-0']['attempt'], 2)
        notes = self.records_from_work('notes.jsonl')
        self.assertEqual([(n['shard'],n['attempt'],n['reason']) for n in notes],
                         [('shard-0',2,'setup: golang.org Forbidden')])
        merge_brief = Path(next(r['brief'] for r in launches if r['name']=='merge')).read_text()
        self.assertIn('"reason": "setup: golang.org Forbidden"', merge_brief)
        print(f'replacement-start-after-marker={elapsed:.3f}s default-poll=30s', flush=True)

    def records_from_work(self, filename):
        paths = list((self.root/'fleet'/self.sha[:12]).glob('*/'+filename))
        self.assertEqual(len(paths), 1)
        return [json.loads(line) for line in paths[0].read_text().splitlines()]

    def test_three_failed_attempts_fail_run_by_shard_name(self):
        self.env['FLEET_TEST_MODE'] = 'cap'
        code, output = self.run_fleet('run', self.sha, '2')
        self.assertNotEqual(code, 0, output)
        self.assertIn('shard-0 failed after 3 attempts: setup: golang.org Forbidden', output)
        launches = self.records('launches.jsonl')
        self.assertEqual([r['attempt'] for r in launches if r['name']=='shard-0'], [1,2,3])
        self.assertFalse(any(r['name']=='merge' for r in launches))
        self.assertEqual([r['attempt'] for r in self.records_from_work('notes.jsonl')], [2,3])

    def test_failure_arriving_after_first_poll_replaced_on_next_default_poll(self):
        self.env.pop('ADAMIC_GATE_FLEET_POLL')
        self.env['FLEET_TEST_MODE'] = 'late-marker'
        self.env['ADAMIC_GATE_FLEET_SHARD_WAIT'] = '70'
        code, output = self.run_fleet('run', self.sha, '2')
        self.assertEqual(code,0,output)
        replacement = next(r for r in self.records('launches.jsonl') if r['name']=='shard-0' and r['attempt']==2)
        elapsed = replacement['time']-self.records('markers.jsonl')[0]['time']
        self.assertLess(elapsed,60)
        self.assertEqual(json.loads((self.root/'merged.json').read_text())['shard-0']['attempt'],2)
        print(f'next-poll-replacement-after-marker={elapsed:.3f}s default-poll=30s',flush=True)

    def test_old_marker_does_not_relaunch_replacement_while_it_starts(self):
        self.env['FLEET_TEST_MODE'] = 'delayed'
        code, output = self.run_fleet('run', self.sha, '2')
        self.assertEqual(code, 0, output)
        self.assertEqual([r['attempt'] for r in self.records('launches.jsonl') if r['name']=='shard-0'],[1,2])
        self.assertEqual(len(self.records_from_work('notes.jsonl')),1)

    def test_merge_refuses_branch_with_no_archive(self):
        # An arbitrary branch is not a completed shard.
        branch = f'gate-logs/{self.sha[:12]}/empty/shard-0'
        self.git('-C',str(self.repo),'push','-q','origin','HEAD:refs/heads/'+branch)
        code, output = self.run_fleet('merge', self.sha, 'empty', '2')
        self.assertNotEqual(code,0)
        self.assertIn('merge refuses shard-0: missing shard log archive',output)
        self.assertFalse(self.records('launches.jsonl'))

    def test_launcher_failure_replaced_without_deadline(self):
        self.env['FLEET_TEST_MODE'] = 'launcher'
        code, output = self.run_fleet('run', self.sha, '2')
        self.assertEqual(code, 0, output)
        self.assertIn('launcher failed with exit 7', output)
        self.assertEqual(json.loads((self.root/'merged.json').read_text())['shard-0']['attempt'],2)

    def test_merge_refuses_marker_only_shard(self):
        branch = f'gate-logs/{self.sha[:12]}/marker-only/shard-0'
        self.assertEqual(self.publish(branch, 1, 'setup: Forbidden'), 0)
        code, output = self.run_fleet('merge', self.sha, 'marker-only', '2')
        self.assertNotEqual(code, 0)
        self.assertIn('merge refuses shard-0: failure marker, not shard logs', output)
        self.assertFalse(self.records('launches.jsonl'))
        setup = self.git('--git-dir='+str(self.root/'origin.git'), 'show', branch+':setup.log')
        self.assertIn('Forbidden', setup)

    def test_relaunch_refuses_to_start_if_run_notes_cannot_be_written(self):
        self.env['FLEET_TEST_MODE'] = 'notes-unwritable'
        code,output = self.run_fleet('run',self.sha,'2')
        self.assertNotEqual(code,0)
        self.assertIn('notes.jsonl',output)
        self.assertEqual([r['attempt'] for r in self.records('launches.jsonl') if r['name']=='shard-0'],[1])

    def test_invalid_commit_marker_refused_by_shard_name(self):
        self.env['FLEET_TEST_MODE'] = 'invalid'
        code, output = self.run_fleet('run',self.sha,'2')
        self.assertNotEqual(code,0)
        self.assertIn('shard-0 has an invalid failure marker',output)
        self.assertFalse(any(r['name']=='merge' for r in self.records('launches.jsonl')))

    def test_merge_refuses_marker_even_with_an_old_archive(self):
        branch = f'gate-logs/{self.sha[:12]}/old-archive/shard-0'
        self.assertEqual(self.publish(branch,1,'setup: Forbidden'),0)
        marker = self.git('--git-dir='+str(self.root/'origin.git'),'show',branch+':failure.json')
        self.assertEqual(self.publish(branch,2),0)
        scratch = self.root/'old-archive'
        self.git('init','-q',str(scratch))
        self.git('-C',str(scratch),'fetch','-q',str(self.root/'origin.git'),'refs/heads/'+branch)
        self.git('-C',str(scratch),'checkout','-q','-B','logs','FETCH_HEAD')
        (scratch/'failure.json').write_text(marker)
        self.git('-C',str(scratch),'add','failure.json')
        self.git('-C',str(scratch),'-c','user.name=Test','-c','user.email=test@example.invalid','commit','-q','-m','Marker with stale archive')
        self.git('-C',str(scratch),'push','-q',str(self.root/'origin.git'),'HEAD:refs/heads/'+branch)
        code,output = self.run_fleet('merge',self.sha,'old-archive','2')
        self.assertNotEqual(code,0)
        self.assertIn('merge refuses shard-0: failure marker, not shard logs',output)
        self.assertFalse(self.records('launches.jsonl'))

    def test_failed_checkout_can_publish_marker_for_requested_commit(self):
        requested = 'b'*40
        branch = f'gate-logs/{requested[:12]}/checkout-failed/shard-0'
        setup = self.root/'checkout-failed.log'
        setup.write_text('git fetch: network unavailable\n')
        args = ['python3',str(PUBLISH),'--origin',str(self.root/'origin.git'),'--branch',branch,
                '--attempt','1','--commit',requested,'--failure','checkout: network unavailable','--setup-log',str(setup)]
        with (self.root/'publisher.log').open('ab') as stream:
            result = subprocess.run(args,cwd=self.repo,stdout=stream,stderr=subprocess.STDOUT)
        self.assertEqual(result.returncode,0)
        marker = json.loads(self.git('--git-dir='+str(self.root/'origin.git'),'show',branch+':failure.json'))
        self.assertEqual(marker['Commit'],requested)
        # A successful archive cannot claim to come from that other commit.
        archive = self.root/'shard-0.tgz'
        archive.write_bytes(b'not-used')
        args = args[:args.index('--failure')]+['--archive',str(archive)]
        with (self.root/'publisher.log').open('ab') as stream:
            result = subprocess.run(args,cwd=self.repo,stdout=stream,stderr=subprocess.STDOUT)
        self.assertNotEqual(result.returncode,0)
        self.assertIn('checkout commit differs from requested commit',(self.root/'publisher.log').read_text())

    def test_publisher_replaces_marker_preserves_history_and_refuses_stale_attempt(self):
        branch = f'gate-logs/{self.sha[:12]}/publisher/shard-0'
        self.assertEqual(self.publish(branch, 1, 'setup: Forbidden'), 0)
        old = self.git('--git-dir='+str(self.root/'origin.git'), 'rev-parse', branch)
        self.assertEqual(self.publish(branch, 2), 0)
        parent = self.git('--git-dir='+str(self.root/'origin.git'), 'rev-parse', branch+'^')
        self.assertEqual(parent, old)
        tree = self.git('--git-dir='+str(self.root/'origin.git'), 'ls-tree', '--name-only', branch)
        self.assertIn('shard-0.tgz', tree)
        self.assertNotIn('failure.json', tree)
        self.assertNotEqual(self.publish(branch, 1, 'late stale failure'), 0)
        self.assertNotEqual(self.publish(branch, 2, 'failure after logs'), 0)
        self.assertEqual(self.git('--git-dir='+str(self.root/'origin.git'), 'rev-parse', branch+'^'),old)


if __name__ == '__main__':
    unittest.main(verbosity=2)
