"""Extend the pinned admission generator with both step 16 fixture directories."""
import json
import subprocess
import sys

repository, sha = sys.argv[1:]
revision = '543925aa5e950830507f6907f70e32d125beef11'
generator = subprocess.check_output(['git', '-C', repository, 'show', revision+':cloud/admission-corpus/manifest.py'], text=True)
result = subprocess.run([sys.executable, '-', '--repository', repository, '--sha', sha], input=generator, capture_output=True, text=True, check=True)
manifest = json.loads(result.stdout)
listing = subprocess.check_output(['git', '-C', repository, 'ls-tree', '-r', '--full-tree', manifest['sha'], 'stage3/fixtures/generic-values', 'stage3/fixtures/generics'], text=True)
programs = []
for line in listing.splitlines():
    metadata, path = line.split('\t', 1)
    if path.endswith('.a'):
        programs.append({'path': path, 'blob': metadata.split()[2]})
manifest['corpora'].append({'name': 'generic-values-step16', 'programs': programs})
json.dump(manifest, sys.stdout, indent=2)
sys.stdout.write('\n')
