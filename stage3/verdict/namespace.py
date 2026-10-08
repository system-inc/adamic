"""Run an unchanged CLI over temporary absolute paths in a Linux mount namespace."""
from pathlib import Path
import shutil


def command(argv, folder, virtual_cwd, tree):
    executable = shutil.which('bwrap')
    if not executable:
        raise RuntimeError('absolute-path cases require bubblewrap; no API or source-rewrite fallback')
    filesystem = folder / 'filesystem'
    result = [executable, '--die-with-parent', '--unshare-user', '--unshare-pid', '--tmpfs', '/']
    # Keep the executable, dynamic loader and its installed libraries available.
    # Test absolute roots overlay these read-only runtime mounts inside the child.
    for name in ('usr', 'lib', 'lib64', 'bin', 'sbin', 'etc', 'opt', 'home', 'workspace', 'tmp'):
        host = Path('/') / name
        if host.exists():
            result += ['--ro-bind', str(host), str(host)]
    result += ['--proc', '/proc', '--dev', '/dev']
    result += ['--ro-bind', str(tree / 'tests/lib'), '/.lib']
    result += ['--ro-bind', str(folder.parent / '.namespace-types'), '/.ts']
    for path in sorted(filesystem.iterdir()):
        result += ['--bind', str(path), '/' + path.name]
        if path.name in ('lib', 'lib64') and path.is_dir():
            # Runtime loader subtrees remain read-only; test emit destinations
            # at this root are writable and remain under the temporary directory.
            for runtime in sorted((Path('/') / path.name).iterdir()):
                if not (path / runtime.name).exists():
                    result += ['--ro-bind', str(runtime), '/' + path.name + '/' + runtime.name]
    return result + ['--chdir', virtual_cwd, '--', *argv]


def virtual_arguments(argv, folder):
    prefix = str(folder / 'filesystem')
    return [value.replace(prefix + '/', '/') for value in argv]
