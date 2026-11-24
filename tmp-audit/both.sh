#!/usr/bin/env bash
# Correctness for #e24qsqx: one side of the census pair. The first-stop census (both passes) by the in-tree run.sh,
# then full-mode censuses of moduleNameResolver.ts and commandLineParser.ts with the tool and tree run.sh built,
# every finding printed as a FULLROW line so the two sides diff by site and reason.
set -u -o pipefail
repository=$(git rev-parse --show-toplevel)
cd "$repository"
bash stage3/census/firststops/run.sh || { echo "VERDICT-ALL FAIL run.sh"; exit 1; }
work="${LOOM_JOB_SCRATCH:-$TMPDIR}/firststops"
for file in moduleNameResolver.ts commandLineParser.ts; do
	start=$SECONDS
	LATENT_ONLY_FILE="$work/tree/src/compiler/$file" LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 LATENT_TAIL=1 \
		CENSUS_MODULE_FORMAT=CommonJSBundle "$work/tools/directory-census" "$work/tree/src/compiler" "$work/streams/full-$file.jsonl" \
		> "$work/streams/full-$file.log" 2>&1 || { tail -n 30 "$work/streams/full-$file.log"; echo "VERDICT-ALL FAIL full $file"; exit 1; }
	echo "FULL-WALL $file $((SECONDS - start)) s"
	python3 -I - "$work/streams/full-$file.jsonl" "$work/tree/" "$file" << 'READ'
import json, sys
path, root, name = sys.argv[1], sys.argv[2], sys.argv[3]
units = findings = 0
seen = set()
for line in open(path, encoding='utf-8'):
    try:
        row = json.loads(line)
    except json.JSONDecodeError:
        break
    units += 1
    for f in row.get('findings') or []:
        findings += 1
        key = ((f.get('where') or '').replace(root, ''), f.get('phase') or '', (f.get('reason') or '').replace(root, '').replace('\n', ' '))
        if key not in seen:
            seen.add(key)
            print('FULLROW', name, *key, sep='\t')
print(f'FULL {name}: {units} units, {findings} findings, {len(seen)} distinct site and reason pairs')
READ
done
echo "VERDICT-ALL PASS"
