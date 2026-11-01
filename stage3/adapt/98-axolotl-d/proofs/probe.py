# Reproduces the row 2 probe. It is a scratch probe, not an adapter: it edits
# copies of an applied tree by exact text, so it does not meet the adapter
# contract (stock compiler API, parsed text). Usage:
#   python3 probe.py <applied-tree> <scratch-directory> <adamic-binary>
# It writes <scratch>/ruled-four and <scratch>/six, runs `adamic types` rooted at
# moduleNameResolver.ts on each and on the applied tree, and prints the
# sorted primary diagnostic lines of each run.
import re
import shutil
import subprocess
import sys
from pathlib import Path

tree, scratch, adamic = Path(sys.argv[1]), Path(sys.argv[2]), sys.argv[3]


def replace_after(text, anchor, old, new, window):
    start = text.index(anchor)
    at = text.index(old, start)
    assert at - start < window, (anchor, old)
    return text[:at] + new + text[at + len(old):]


def ruled_four(root):
    path = root / 'src/compiler/types.ts'
    text = path.read_text()
    full = 'export interface ResolvedModuleFull extends'
    directive = 'export interface ResolvedTypeReferenceDirective {'
    text = replace_after(text, full, '    readonly originalPath?: string;\n', '    readonly originalPath?: (string) | undefined;\n', 300)
    text = replace_after(text, full, '    packageId?: PackageId;\n', '    packageId?: (PackageId) | undefined;\n', 800)
    text = replace_after(text, directive, '    originalPath?: string;\n', '    originalPath?: (string) | undefined;\n', 600)
    text = replace_after(text, directive, '    packageId?: PackageId;\n', '    packageId?: (PackageId) | undefined;\n', 800)
    path.write_text(text)


def six(root):
    ruled_four(root)
    path = root / 'src/compiler/types.ts'
    text = path.read_text()
    text = replace_after(text, 'export interface ResolvedModule {', '    isExternalLibraryImport?: boolean;\n', '    isExternalLibraryImport?: (boolean) | undefined;\n', 300)
    path.write_text(text)
    path = root / 'src/compiler/program.ts'
    text = path.read_text()
    old = 'export interface ResolutionWithResolvedFileName {\n    resolvedFileName: string | undefined;\n    packageId?: PackageId;\n}'
    assert text.count(old) == 1
    path.write_text(text.replace(old, old.replace('packageId?: PackageId;', 'packageId?: (PackageId) | undefined;')))


def diagnostics(root):
    run = subprocess.run([adamic, 'types', str(root / 'src/compiler/moduleNameResolver.ts')], capture_output=True, text=True)
    prefix = str(root) + '/'
    lines = [line.replace(prefix, '') for line in (run.stdout + run.stderr).splitlines() if re.match(r'^\S+:[0-9]+:[0-9]+: error TS', line)]
    return sorted(lines)


for name, edit in (('ruled-four', ruled_four), ('six', six)):
    copy = scratch / name
    if copy.exists():
        raise SystemExit(f'refusing existing scratch copy: {copy}')
    shutil.copytree(tree, copy, symlinks=True)
    edit(copy)
for name, root in (('before', tree), ('ruled-four', scratch / 'ruled-four'), ('six', scratch / 'six')):
    print(f'== {name}')
    print('\n'.join(diagnostics(root)))
