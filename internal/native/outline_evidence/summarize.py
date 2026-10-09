#!/usr/bin/env python3
"""Summarize a reproduce.py scratch directory; exclude separately warmed runtime C."""
import json, pathlib, statistics, sys

scratch = pathlib.Path(sys.argv[1])
builds = json.loads((scratch / 'builds.json').read_text())
summary = []
for mode in ('sanitized', 'release', 'line-tables'):
    for source in ('before', 'after'):
        totals, mains, slowest, slow_names, counts = [], [], [], [], []
        for build in builds:
            if build['Mode'] != mode or build['Source'] != source:
                continue
            label = f"{build['Round']}-{source}-{mode}"
            rows = [json.loads(line) for line in (scratch / (label + '.jsonl')).read_text().splitlines()]
            assert all(row['exit'] == 0 for row in rows)
            compiles = {}
            for row in rows:
                args = row['args']
                if '-c' in args:
                    name = args[args.index('-c') + 1]
                    if pathlib.Path(name).name == name:
                        compiles[name] = row
            slow = max(compiles, key=lambda name: compiles[name]['seconds'])
            totals.append(build['Seconds'])
            mains.append(compiles['main.c']['seconds'])
            slowest.append(compiles[slow]['seconds'])
            slow_names.append(slow)
            counts.append(len(compiles))
        summary.append(dict(mode=mode, source=source, whole_build_median=statistics.median(totals),
                            main_compile_median=statistics.median(mains), slowest_compile_median=statistics.median(slowest),
                            slowest_units=slow_names, whole_build_runs=totals, main_compile_runs=mains,
                            slowest_compile_runs=slowest, units=counts))
(scratch / 'summary.json').write_text(json.dumps(summary, indent=2) + '\n')
for row in summary:
    print(row['mode'], row['source'], 'main', f"{row['main_compile_median']:.3f}",
          'slowest', f"{row['slowest_compile_median']:.3f}", 'build', f"{row['whole_build_median']:.3f}", 'units', row['units'])
