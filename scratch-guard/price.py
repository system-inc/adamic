"""Scratch, never landing (#8nj5pn4, Emit's guard): the guard's price. Arguments: samples, then the directory holding
each program's <name>.base and <name>.head release binaries, then the names. Each sample runs base, head, head, base,
the order mirrored on odd samples, and takes each side's mean user plus system time; a program's row is the median of
the head-over-base ratios and the sign test's 95% interval over them (the 4th and 12th of 15), as bench's -compare
reads it. Every run's stdout must be the first base run's."""
import os
import statistics
import subprocess
import sys

samples = int(sys.argv[1])
directory = sys.argv[2]
names = sys.argv[3:]


def timed(binary):
    process = subprocess.Popen([binary], stdout=subprocess.PIPE, stderr=subprocess.DEVNULL)
    output = process.stdout.read()
    _, status, usage = os.wait4(process.pid, 0)
    if status != 0:
        sys.exit(f'FAIL price: {binary} exited {status}')
    return usage.ru_utime + usage.ru_stime, output


failed = False
for name in names:
    base, head = f'{directory}/{name}.base', f'{directory}/{name}.head'
    _, want = timed(base)
    ratios, base_times, head_times = [], [], []
    for sample in range(samples):
        order = [base, head, head, base] if sample % 2 == 0 else [head, base, base, head]
        times = {base: [], head: []}
        for binary in order:
            seconds, output = timed(binary)
            if output != want:
                print(f'FAIL price: {name}: {binary} printed other than the base')
                failed = True
            times[binary].append(seconds)
        base_mean, head_mean = statistics.mean(times[base]), statistics.mean(times[head])
        base_times.append(base_mean)
        head_times.append(head_mean)
        ratios.append(head_mean / base_mean)
    ratios.sort()
    k = 4 if samples == 15 else max(1, samples // 4)
    low, high = ratios[k - 1], ratios[-k]
    verdict = 'faster' if high < 1 else 'slower' if low > 1 else 'no change at this confidence'
    print(f'{name}: base median {statistics.median(base_times):.3f} s, head median {statistics.median(head_times):.3f} s, ratio median {statistics.median(ratios):.4f}, interval [{low:.4f}, {high:.4f}], {verdict}')
print('FAIL price' if failed else 'PASS price')
sys.exit(1 if failed else 0)
