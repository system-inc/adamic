"""Private upstream measurement views of immutable prepared artifacts."""
from pathlib import Path
import shutil


def make_view(tree, output, private_case=None):
    """Share immutable inputs; give each measurement its own baseline scratch."""
    output.mkdir(parents=True, exist_ok=False)
    for file in tree.iterdir():
        if file.name == 'Herebyfile.mjs':
            # Hereby requires an ordinary file; its tests still use the built harness.
            shutil.copyfile(file, output / file.name)
        elif file.name not in ('tests', 'built', 'test.config', 'mytest.config', '.failed-tests'):
            (output / file.name).symlink_to(file, target_is_directory=file.is_dir())
    for directory, except_names in [('tests', {'baselines', 'cases'}), ('built', {'local'})]:
        target = output / directory
        target.mkdir()
        for file in (tree / directory).iterdir():
            if file.name not in except_names:
                (target / file.name).symlink_to(file, target_is_directory=file.is_dir())
    baselines = output / 'tests/baselines'
    baselines.mkdir()
    # Only reference baselines are immutable; all output paths remain private.
    (baselines / 'reference').symlink_to(tree / 'tests/baselines/reference', target_is_directory=True)
    (baselines / 'local').mkdir()
    local = output / 'built/local'
    local.mkdir()
    for file in (tree / 'built/local').iterdir():
        if file.name in ('run.js', 'run.js.map'):
            shutil.copyfile(file, local / file.name)
        else:
            (local / file.name).symlink_to(file, target_is_directory=file.is_dir())
    def cases(source, target, relative):
        if private_case is None or not private_case.is_relative_to(relative):
            target.symlink_to(source, target_is_directory=source.is_dir())
        elif source.is_dir():
            target.mkdir()
            for file in source.iterdir():
                cases(file, target / file.name, relative / file.name)
        else:
            shutil.copyfile(source, target)
    cases(tree / 'tests/cases', output / 'tests/cases', Path('tests/cases'))
