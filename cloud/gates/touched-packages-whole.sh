source /opt/adamic-block/environment.sh && source cloud/sandboxed-job.sh && source cloud/gates/steps.sh && gate_begin touched-packages-whole || exit 1
# ask: --size F16
# touched-packages-whole: go test of every package holding a changed file, whole (cloud/gates/diff-to-gates.json). A
# changed file's package is the nearest directory at or above it that holds .go files, above any testdata, so a new
# oracle fixture runs internal/oracle whole; a path under cohere/ is the submodule's, and one with no package above it
# (a doc at the root) is named and needs none.
# This file is the whole command of a Loom sealed job at the head's own sha, from its first line, which opens sealed so a
# rented machine takes it. cloud/submit-gates.py matches it by sha256 against the job's commandSha256, so any change here
# is a new recipe, and jobs of the old one stop counting.
gate_base
python3 -I - "$gateBase" > "$gateScratch/packages" << 'PYTHON' || gate_abort packages "reading the head's packages"
import os, subprocess, sys
changed = subprocess.run(['git', 'diff', '--no-renames', '--name-only', '-z', sys.argv[1], 'HEAD'],
                         capture_output=True, check=True).stdout.decode().split('\0')
packages = set()
for path in filter(None, changed):
    if path == 'cohere' or path.startswith('cohere/'):
        print(f'# {path}: the submodule, not read here', file=sys.stderr)
        continue
    parts = path.split('/')[:-1]
    if 'testdata' in parts:
        parts = parts[:parts.index('testdata')]
    while parts:
        directory = '/'.join(parts)
        if os.path.isdir(directory) and any(name.endswith('.go') for name in os.listdir(directory)):
            packages.add('./' + directory)
            break
        parts.pop()
    else:
        print(f'# {path}: no package holds it', file=sys.stderr)
for package in sorted(packages):
    print(package)
PYTHON
mapfile -t gatePackages < "$gateScratch/packages"
printf 'touched packages: %s\n' "${gatePackages[*]:-none}"
if [ ${#gatePackages[@]} -eq 0 ]; then
	gate_pass "no package holds a changed file"
	gate_end
fi
gate_preamble npm sandboxed_npm
gate_products "${gatePackages[@]}"
gate_go_test touched-packages-whole "${gatePackages[@]/%//}"
gate_end
