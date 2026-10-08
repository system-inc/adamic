"""Filesystem units and compiler options, following the pinned compiler runner."""
import itertools
import json
from pathlib import Path, PurePosixPath
import posixpath
import re

SCHEMA_DATA = json.loads((Path(__file__).parent / 'options.json').read_text())
if SCHEMA_DATA['version'] != '6.0.3':
    raise ValueError('compiler option schema is not pinned to TypeScript 6.0.3')
SCHEMA = SCHEMA_DATA['options']
DECLARATION_ERRORS = set(SCHEMA_DATA['declarationErrors'])
HEADER = re.compile(r'^//\s*@(\w+)\s*:\s*([^\r\n]*)')
IGNORED = {'filename', 'notypesandsymbols', 'fullemitpaths', 'symlink', 'typescriptversion'}
HOST = {'baselinefile',
        'capturesuggestions', 'suppressoutputpathcheck'}


def safe_name(name):
    path = PurePosixPath(name)
    if not name or path.is_absolute() or '..' in path.parts or '\\' in name or ':' in name:
        raise ValueError('rooted or escaping virtual paths require an isolated virtual filesystem')
    return str(path)


def unit_name(name):
    if '\\' in name or ':' in name or not name:
        raise ValueError('Windows virtual paths require drive and separator semantics unavailable on Linux')
    if name.startswith('/') or '..' in PurePosixPath(name).parts:
        return posixpath.normpath(name)
    return safe_name(name)


def source_text(raw):
    if raw.startswith(b'\xff\xfe'):
        return raw[2:].decode('utf-16-le', errors='replace')
    if raw.startswith(b'\xfe\xff'):
        return raw[2:].decode('utf-16-be', errors='replace')
    return raw.decode('utf-8-sig', errors='replace')


def virtual_directory(settings):
    return posixpath.normpath(posixpath.join('/.src', settings.get('currentdirectory') or '/.src'))


def virtual_name(name, settings):
    absolute = posixpath.normpath(posixpath.join(virtual_directory(settings), name))
    return absolute.removeprefix('/.src/') if absolute.startswith('/.src/') else absolute


def parse(raw, name):
    text = source_text(raw)
    settings, units, links = {}, [], []
    current_name, content, current_symlinks = None, '', []
    for line in re.split(r'\r\n|\n|\r', text):
        match = HEADER.match(line)
        if match:
            key, value = match[1].lower(), match[2].strip()
            settings[key] = value
            if key == 'symlink':
                current_symlinks = [unit_name(path.strip()) for path in value.split(',')]
            if key == 'link':
                target, destination = value.split('->', 1)
                links.append((unit_name(target.strip()), unit_name(destination.strip())))
            if key == 'filename':
                if current_name:
                    units.append({'name': unit_name(current_name), 'content': content})
                    links.extend((unit_name(current_name), destination) for destination in current_symlinks)
                    current_symlinks = []
                current_name, content = value, ''
        else:
            content += ('\n' if content else '') + line
    units.append({'name': unit_name(current_name or name), 'content': content})
    links.extend((unit_name(current_name or name), destination) for destination in current_symlinks)
    project = next((u['name'] for u in units if Path(u['name']).name.lower()
                    in ('tsconfig.json', 'jsconfig.json')), None)
    for unit in units:
        if re.search(r'(?:\b(?:from|require|import)\s*(?:\(\s*)?|<reference\s+[^>]*(?:path|types)\s*=\s*)[\'\"]/', unit['content']):
            raise ValueError('absolute source references require mounted virtual roots; source text is preserved')
        if re.search(r'@jsxImportSource\s+/', unit['content']):
            raise ValueError('absolute JSX pragmas require mounted virtual roots; source text is preserved')
        if unit['name'].endswith('.json') and re.search(r'[\'\"](?:/|[A-Za-z]:)', unit['content']):
            raise ValueError('absolute package/config paths require mounted virtual roots; source text is preserved')
    for key, value in settings.items():
        if key in HOST:
            raise ValueError(f'harness directive @{key}: {value}')
        if key == 'typescriptversion' and value not in ('6.0', '6.0.3'):
            raise ValueError(f'API compiler version override @{key}: {value}')
        if key == 'usecasesensitivefilenames' and value.lower() != 'true':
            raise ValueError('case-insensitive virtual filesystem differs from Linux')
        if key not in SCHEMA and key not in IGNORED | {'noimplicitreferences', 'usecasesensitivefilenames', 'link', 'currentdirectory'}:
            raise ValueError(f'unknown harness directive @{key}: {value}')
    if settings.get('currentdirectory'):
        unit_name(settings['currentdirectory'])
    last = units[-1]
    # Upstream tests truthiness of the setting string, even when it says false.
    roots = [last['name']] if (settings.get('noimplicitreferences')
             or re.search(r'require\(|reference\spath', last['content'])) else [u['name'] for u in units]
    roots = [name for name in roots if not name.lower().endswith('.json')]
    settings['__project'] = project
    settings['__links'] = links
    return units, settings, roots


def variations(value, option):
    includes, excludes = [], []
    star = False
    for part in value.lower().split(','):
        part = part.strip()
        if part == '*':
            star = True
        elif part.startswith(('-', '!')):
            excludes.append(part[1:])
        elif part:
            includes.append(part)
    if len(includes) <= 1 and not star and not excludes:
        return None
    values = option.get('values') or {'true': 1, 'false': 0}
    result = []
    for key in includes + (list(values) if star else []):
        if not any(key == existing or (key in values and values[key] == values.get(existing)) for existing in result):
            result.append(key)
    result = [key for key in result if not any(key == excluded or
              (excluded in values and values.get(key) == values[excluded]) for excluded in excludes)]
    if not result:
        raise ValueError('empty option variation set')
    return result


def configurations(settings):
    varying = [(key, variations(settings[key], option)) for key, option in SCHEMA.items()
               if key in settings and option['vary']]
    varying = [(key, values) for key, values in varying if values is not None]
    combinations = list(itertools.product(*(values for _, values in varying)))
    if len(combinations) > 25:
        raise ValueError('option variants exceed upstream limit of 25')
    result = []
    for values in combinations:
        override = dict(zip((key for key, _ in varying), values))
        options = {}
        for key, value in {**settings, **override}.items():
            if key not in SCHEMA:
                continue
            option = SCHEMA[key]
            if option['type'] == 'boolean':
                parsed = value.lower() == 'true'
            elif option['type'] in ('list', 'listOrElement'):
                parsed = [v.strip() for v in value.split(',') if v.strip()]
            elif option['type'] == 'number':
                parsed = int(value)
            elif option['type'] == 'enum':
                parsed = value.lower()
            else:
                parsed = value
            if option['filePath'] or option['elementPath']:
                for path in parsed if isinstance(parsed, list) else [parsed]:
                    unit_name(path)
            options[option['name']] = parsed
        suffix = ','.join(f'{key}={override[key]}' for key in sorted(override))
        result.append((suffix, options))
    return result


def arguments(options, roots, ambient_roots=True, defaults=None):
    # Command-line options avoid invented tsconfig locations on global diagnostics.
    merged = {'skipDefaultLibCheck': (defaults or {}).get('skipDefaultLibCheck', True),
              'noErrorTruncation': True}
    if ambient_roots:
        merged['typeRoots'] = ['node_modules/@types']
    merged.update(options)
    # Resolution traces are a separate upstream baseline, not diagnostics.
    merged['traceResolution'] = False
    for name in ('listFiles', 'listFilesOnly', 'listEmittedFiles', 'diagnostics', 'extendedDiagnostics', 'explainFiles'):
        merged[name] = False
    if 'locale' in merged:
        merged['locale'] = 'en'
    merged['pretty'] = options.get('pretty', False)
    argv = []
    for name, value in merged.items():
        if value == []:
            # The CLI cannot spell an empty list; the isolated empty type root
            # already implements an empty ambient-types set.
            if name == 'types':
                # Explicit @types: empty disables even the materialized packages.
                if '--typeRoots' in argv:
                    position = argv.index('--typeRoots') + 1
                    argv[position] = '.verdict-empty-types'
                else:
                    argv += ['--typeRoots', '.verdict-empty-types']
                continue
            raise ValueError(f'empty @{name} list requires config-file option parsing')
        argv += ['--' + name, ','.join(value) if isinstance(value, list) else
                 str(value).lower() if isinstance(value, bool) else str(value)]
    return argv + roots


def layout(units, options, folder, settings=None):
    settings = settings or {}
    paths = [value for key, value in options.items() if (SCHEMA[key.lower()]['filePath'] or key == 'jsxImportSource')
             and isinstance(value, str)]
    virtual = bool(settings.get('currentdirectory') or settings.get('__links')) or any(name.startswith('/') or '..' in PurePosixPath(name).parts
                  for name in [u['name'] for u in units] + paths)
    virtual_cwd = virtual_directory(settings)
    cwd = folder / 'filesystem' / virtual_cwd.lstrip('/') if virtual else folder
    cwd.mkdir(parents=True, exist_ok=True)
    def physical(name):
        if not virtual:
            return cwd / safe_name(name)
        absolute = posixpath.normpath(posixpath.join(virtual_cwd, unit_name(name)))
        return folder / 'filesystem' / absolute.lstrip('/')
    mapped = dict(options)
    if virtual:
        for key, value in options.items():
            option = SCHEMA[key.lower()]
            if (option['filePath'] or key == 'jsxImportSource' and value.startswith('/')) and isinstance(value, str):
                mapped[key] = str(physical(value))
            elif option['elementPath'] and isinstance(value, list):
                mapped[key] = [str(physical(path)) for path in value]
        if 'typeRoots' not in mapped:
            mapped['typeRoots'] = [str(cwd / 'node_modules/@types'),
                                   str(folder / 'filesystem/node_modules/@types')]
        # The compiler runner resolves a relative header baseUrl against /.src,
        # independently of currentDirectory.
        if options.get('baseUrl') and not options['baseUrl'].startswith('/'):
            mapped['baseUrl'] = str(folder / 'filesystem/.src' / options['baseUrl'])
    return cwd, physical, mapped


def diagnostics(stdout, folder, cwd, units, settings=None):
    settings = settings or {}
    virtual_cwd = virtual_directory(settings)
    if cwd == folder or virtual_cwd == '/.src':
        stdout = stdout.replace((str(cwd) + '/').encode(), b'')
    if cwd != folder:
        root = folder / 'filesystem'
        stdout = stdout.replace((str(root) + '/').encode(), b'/')
        for unit in units:
            name = posixpath.normpath(posixpath.join(virtual_cwd, unit['name']))
            if unit['name'].startswith('/') or virtual_cwd != '/.src':
                relative = posixpath.relpath(name, virtual_cwd)
                virtual = name.removeprefix('/.src/') if name.startswith('/.src/') else name
                stdout = re.sub(rb'^' + re.escape(relative.encode()) + rb'(?=\(\d+,\d+\):)',
                                lambda match: virtual.encode(), stdout, flags=re.M)
                stdout = stdout.replace(b'\x1b[96m' + relative.encode() + b'\x1b[0m',
                                        b'\x1b[96m' + virtual.encode() + b'\x1b[0m')
    return stdout


def expected_exit(options, expected):
    if not expected:
        return 0
    # Pinned handleNoEmitOptions and emitter.ts define emitSkipped. noEmit
    # short-circuits noEmitOnError; declaration transform errors block d.ts.
    if options.get('noEmit'):
        return 2
    plain = re.sub(rb'\x1b\[[0-9;]*m', b'', expected)
    codes = set(map(int, re.findall(rb'error TS(\d+):', plain)))
    declarations = options.get('declaration') or options.get('composite')
    if (options.get('noEmitOnError')
            or options.get('emitDeclarationOnly') and not declarations
            or codes & {5055, 5056}
            or declarations and codes & DECLARATION_ERRORS):
        return 1
    return 2


def config_options(options):
    # "null" has an unsetting meaning in parseCommandLine; JSON preserves a
    # literal string. Some compiler options are only accepted in config files.
    return {key: value for key, value in options.items()
            if value == 'null' or value == [] or SCHEMA[key.lower()].get('configOnly') and value is not False}


def exhaustive_emit(options, expected):
    plain = re.sub(rb'\x1b\[[0-9;]*m', b'', expected)
    codes = set(map(int, re.findall(rb'error TS(\d+):', plain)))
    return bool(options.get('noEmitOnError') and not options.get('noEmit')
                and not codes & DECLARATION_ERRORS)


def has_type_packages(units, settings):
    names = [unit['name'] for unit in units] + [destination for _, destination in settings['__links']]
    return any('/node_modules/@types/' in '/' + name for name in names)


def pretty_diagnostics(stdout):
    if not stdout:
        return stdout
    # createDiagnosticReporter formats one diagnostic at a time and appends
    # one wrapper newline. The API formatter joins the same blocks directly.
    content = stdout.split(b'\nFound ', 1)[0]
    header = rb'(?=^(?:\x1b\[96m[^\n]*? - )?\x1b\[91m[^\n]*?\x1b\[0m\x1b\[90m TS\d+:)'
    blocks = re.split(header, content, flags=re.M)
    return b''.join(block[:-1] if block.endswith(b'\n') else block for block in blocks)
