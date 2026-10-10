#!/usr/bin/env python3
"""Build the moved-object fixture and measure whole processes against source Node."""
import argparse
import json
import os
from pathlib import Path
import platform
import subprocess
import tempfile
import time

ROOT = Path(__file__).resolve().parents[4]
FIXTURE = ROOT / 'internal/oracle/testdata/moves/accepted/objects.a'

def run(arguments, *, env=None):
    start = time.perf_counter()
    result = subprocess.run(arguments, cwd=ROOT, env=env, capture_output=True,
                            text=True, timeout=600)
    elapsed = time.perf_counter() - start
    if result.returncode != 0:
        raise RuntimeError(f'{arguments}: exit {result.returncode}\n{result.stderr}')
    return elapsed, result.stdout, result.stderr

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--threads', default='1,2,3,4,5')
    parser.add_argument('--rounds', type=int, default=5)
    args = parser.parse_args()
    scratch = Path(tempfile.mkdtemp(prefix='adamic-move-measure-'))
    compiler, binary, counted = (scratch/name for name in ('adamic','objects','counted'))
    for command in ([ 'go','build','-o',str(compiler),'./cmd/adamic'],
                    [str(compiler),'build',str(FIXTURE),'-o',str(binary)],
                    [str(compiler),'build',str(FIXTURE),'-o',str(counted),'--count']):
        _, out, err = run(command)
        (scratch/('build-'+str(len(list(scratch.glob('build-*'))))+'.log')).write_text(out+err)
    threads = [int(value) for value in args.threads.split(',')]
    node = ['node','--disable-warning=ExperimentalWarning','oracle/node.mjs',str(FIXTURE)]
    evidence = {'platform':platform.platform(), 'online_cpus':os.cpu_count(), 'nproc':int(run(['nproc'])[1]),
                'affinity':len(os.sched_getaffinity(0)),
                'cpu_max':Path('/sys/fs/cgroup/cpu.max').read_text().strip(),
                'load_before':os.getloadavg(), 'threads':threads, 'rounds':[],
                'versions':{tool:run([tool,'--version'])[1].splitlines()[0]
                            for tool in ['clang','node']}}
    witness = run(node)[1]
    for number in range(args.rounds):
        sample = {'round':number+1}
        for count in threads:
            env = os.environ.copy()
            env['ADAMIC_THREADS'] = str(count)
            elapsed, out, err = run([str(binary)],env=env)
            if out != witness or err:
                raise RuntimeError(f'threads {count}: disagreed with source Node')
            sample[str(count)] = elapsed
        elapsed,out,err = run(node)
        if out != witness or err:
            raise RuntimeError('Node witness changed')
        sample['node'] = elapsed
        evidence['rounds'].append(sample)
        print(sample,flush=True)
    evidence['load_after'] = os.getloadavg()
    evidence['stdout'] = witness
    evidence['counts'] = {}
    for count in threads:
        env = os.environ.copy()
        env['ADAMIC_THREADS'] = str(count)
        _,out,err = run([str(counted)],env=env)
        if out != witness:
            raise RuntimeError('counted build differs')
        evidence['counts'][str(count)] = err.strip()
    evidence['best_seconds'] = {name:min(row[name] for row in evidence['rounds'])
                                for name in [*map(str,threads),'node']}
    target = ROOT / 'internal/oracle/testdata/moves/measurements.json'
    target.write_text(json.dumps(evidence,indent=2)+'\n')
    print(json.dumps(evidence,indent=2))
    print('raw data:',target)
    print('build logs:',scratch)

if __name__ == '__main__':
    main()
