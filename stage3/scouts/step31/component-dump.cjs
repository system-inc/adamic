"use strict";
// Hermetic compiler host shared by the checker and emitter observations.
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const hash = bytes => crypto.createHash('sha256').update(bytes).digest('hex');
const order = (a, b) => a < b ? -1 : a > b ? 1 : 0;
function canonical(value) {
    if (Array.isArray(value)) return value.map(canonical);
    if (value && typeof value === 'object') return Object.fromEntries(Object.keys(value).sort(order).map(key => [key, canonical(value[key])]));
    return value;
}
function inputHash(project) {
    return hash(JSON.stringify(canonical({...project, files: [...project.files].sort((a, b) => order(a.path, b.path))})));
}
function environment(apiPath, mode) {
    const directory = path.dirname(apiPath);
    const libraries = fs.readdirSync(directory).filter(name => /^lib\..*\.d\.ts$/.test(name)).sort(order);
    return {format: `adamic-${mode}-v1`, typescript: '6.0.3', node: process.version,
        api: hash(fs.readFileSync(apiPath)),
        libraries: hash(JSON.stringify(libraries.map(name => [name, hash(fs.readFileSync(path.join(directory, name)))]))),
        driver: hash(fs.readFileSync(__filename))};
}
function observer(ts, libraryDirectory, reuseLibraries = true) {
    if (ts.version !== '6.0.3') throw new Error('expected TypeScript 6.0.3');
    const libraryText = new Map();
    const libraryAst = new Map();
    function virtual(name) {
        const normalized = path.posix.normalize(name);
        if (normalized.startsWith('/project/')) return normalized.slice('/project/'.length);
        if (normalized.startsWith('/lib/')) return '@lib/' + normalized.slice('/lib/'.length);
        return '@host' + normalized;
    }
    function chain(value) {
        return typeof value === 'string' ? value : {message: value.messageText, code: value.code,
            category: value.category, next: (value.next || []).map(chain)};
    }
    function diagnostic(value) {
        return {file: value.file ? virtual(value.file.fileName) : null, code: value.code,
            start: value.start ?? null, length: value.length ?? null, category: value.category,
            message: chain(value.messageText), related: (value.relatedInformation || []).map(diagnostic)};
    }
    function diagnostics(values) {
        return values.map(diagnostic).sort((a, b) => order(a.file ?? '', b.file ?? '') ||
            (a.start ?? -1) - (b.start ?? -1) || (a.length ?? -1) - (b.length ?? -1) ||
            a.code - b.code || order(JSON.stringify(a), JSON.stringify(b)));
    }
    function observe(project, mode) {
        if (!['checker', 'emitter'].includes(mode)) throw new Error('unknown component');
        const inputs = [...project.files].sort((a, b) => order(a.path, b.path));
        const files = new Map();
        const original = new Map();
        for (const input of inputs) {
            if (!input.path || input.path.startsWith('/') || input.path.includes('\\') || input.path.split('/').some(part => !part || part === '.' || part === '..')) throw new Error('expected relative virtual file path');
            const name = '/project/' + input.path.replace(/\.a$/, '.ts');
            if (files.has(name)) throw new Error('duplicate or extension-colliding file');
            files.set(name, input);
            original.set(virtual(name), input.path);
        }
        const converted = ts.convertCompilerOptionsFromJson(project.options || {}, '/project');
        if (converted.errors.length) throw new Error(ts.flattenDiagnosticMessageText(converted.errors[0].messageText, '\n'));
        const options = converted.options;
        function isLibrary(name) { return /^\/lib\/lib\.[^/]+\.d\.ts$/.test(name); }
        function readFile(name) {
            name = path.posix.normalize(name);
            if (files.has(name)) return files.get(name).text;
            if (isLibrary(name)) {
                if (!libraryText.has(name)) {
                    const disk = path.join(libraryDirectory, path.posix.basename(name));
                    libraryText.set(name, fs.existsSync(disk) ? fs.readFileSync(disk, 'utf8') : undefined);
                }
                return libraryText.get(name);
            }
            return undefined;
        }
        const emitted = [];
        const host = {
            getSourceFile(name, languageVersion) {
                name = path.posix.normalize(name);
                const text = readFile(name);
                if (text === undefined) return undefined;
                // TypeScript itself reuses SourceFiles between Programs. Only
                // standard-library ASTs are shared here; project ASTs are always fresh.
                const key = name + ':' + JSON.stringify(languageVersion);
                if (reuseLibraries && isLibrary(name) && libraryAst.has(key)) return libraryAst.get(key);
                const source = ts.createSourceFile(name, text, languageVersion, true, files.get(name)?.scriptKind);
                if (reuseLibraries && isLibrary(name)) libraryAst.set(key, source);
                return source;
            },
            getDefaultLibFileName: () => '/lib/' + ts.getDefaultLibFileName(options),
            getDefaultLibLocation: () => '/lib',
            writeFile(name, text, bom, onError, sourceFiles) {
                emitted.push({path: virtual(name), text, bom,
                    sources: (sourceFiles || []).map(file => original.get(virtual(file.fileName)) ?? virtual(file.fileName)).sort(order)});
            },
            getCurrentDirectory: () => '/project',
            getCanonicalFileName: name => name,
            useCaseSensitiveFileNames: () => true,
            getNewLine: () => '\n',
            fileExists: name => readFile(name) !== undefined,
            readFile,
            directoryExists: name => name === '/' || name === '/lib' || name === '/project' ||
                [...files.keys()].some(file => file.startsWith(name.replace(/\/$/, '') + '/')),
            getDirectories: () => [],
            realpath: name => name,
        };
        const program = ts.createProgram([...files.keys()], options, host);
        const header = {record: 'project', format: `adamic-${mode}-v1`, typescript: ts.version,
            id: project.id, options: canonical(project.options || {})};
        function restoredDiagnostics(values) {
            const all = diagnostics(values);
            function restore(values) {
                for (const value of values) {
                    if (original.has(value.file)) value.file = original.get(value.file);
                    restore(value.related);
                }
            }
            restore(all);
            return all;
        }
        if (mode === 'checker') {
            const all = restoredDiagnostics(ts.getPreEmitDiagnostics(program));
            const names = new Set(inputs.map(file => file.path));
            for (const item of all) if (item.file !== null) names.add(item.file);
            return [header, {record: 'global', diagnostics: all.filter(item => item.file === null)},
                ...[...names].sort(order).map(name => ({record: 'file', path: name,
                    diagnostics: all.filter(item => item.file === name)}))];
        }
        const result = program.emit();
        emitted.sort((a, b) => order(a.path, b.path));
        if (new Set(emitted.map(file => file.path)).size !== emitted.length) throw new Error('duplicate emitted path');
        return [header, {record: 'result', emitSkipped: result.emitSkipped, diagnostics: restoredDiagnostics(result.diagnostics)},
            ...emitted.map(file => ({record: 'output', ...file}))];
    }
    return {observe};
}
function run(mode) {
    const args = process.argv.slice(2);
    const requestPath = args.shift();
    const output = args.shift();
    if (!requestPath || !output) throw new Error('usage: driver REQUEST NEW_RESULTS [--previous RESULTS] [--changed-only] [--ids JSON_IDS] [--fresh-libs]');
    let previous, ids, changedOnly = false, fresh = false;
    while (args.length) {
        const arg = args.shift();
        if (arg === '--previous') previous = args.shift();
        else if (arg === '--ids') ids = new Set(JSON.parse(fs.readFileSync(args.shift(), 'utf8')));
        else if (arg === '--changed-only') changedOnly = true;
        else if (arg === '--fresh-libs') fresh = true;
        else throw new Error('unknown argument: ' + arg);
    }
    if (changedOnly && !previous) throw new Error('--changed-only requires --previous');
    if (fs.existsSync(output)) throw new Error('results must be new');
    const started = performance.now();
    const apiPath = process.env.STEP31_TYPESCRIPT;
    const ts = require(apiPath);
    const env = environment(apiPath, mode);
    const old = previous ? JSON.parse(fs.readFileSync(path.join(previous, 'manifest.json'), 'utf8')) : undefined;
    const compatible = old && JSON.stringify(old.environment) === JSON.stringify(env);
    const oldRows = new Map(compatible ? old.projects.map(row => [row.id, row]) : []);
    const request = JSON.parse(fs.readFileSync(requestPath, 'utf8'));
    const seen = new Set();
    for (const project of request.projects) {
        if (typeof project.id !== 'string' || seen.has(project.id)) throw new Error('duplicate or invalid project id');
        seen.add(project.id);
    }
    if (ids && [...ids].some(id => !seen.has(id))) throw new Error('requested id not found');
    fs.mkdirSync(path.join(output, 'projects'), {recursive: true});
    const compiler = observer(ts, path.dirname(apiPath), !fresh);
    const manifest = {environment: env, projects: [], selected: [], changed: 0, reused: 0};
    const selected = [];
    const stdout = fs.openSync(path.join(output, 'golden.stdout'), 'w');
    for (const project of request.projects) {
        if (ids && !ids.has(project.id)) continue;
        const input = inputHash(project);
        const before = oldRows.get(project.id);
        const shard = 'projects/' + hash(project.id) + '.jsonl';
        let bytes, elapsed = 0;
        const reused = before?.input === input;
        if (reused) {
            bytes = fs.readFileSync(path.join(previous, before.shard));
            if (hash(bytes) !== before.sha256) throw new Error('cached output hash mismatch');
            manifest.reused++;
        } else {
            const tick = performance.now();
            bytes = Buffer.from(compiler.observe(project, mode).map(row => JSON.stringify(row)).join('\n') + '\n');
            elapsed = performance.now() - tick;
            manifest.changed++;
        }
        fs.writeFileSync(path.join(output, shard), bytes);
        manifest.projects.push({id: project.id, input, shard, sha256: hash(bytes), bytes: bytes.length,
            milliseconds: elapsed, reused});
        if (!changedOnly || !reused) {
            fs.writeSync(stdout, bytes);
            selected.push(project);
            manifest.selected.push(project.id);
        }
    }
    fs.closeSync(stdout);
    const selectedRequest = JSON.stringify({projects: selected}) + '\n';
    fs.writeFileSync(path.join(output, 'request.json'), selectedRequest);
    manifest.requestSha256 = hash(selectedRequest);
    manifest.milliseconds = performance.now() - started;
    manifest.sha256 = hash(fs.readFileSync(path.join(output, 'golden.stdout')));
    fs.writeFileSync(path.join(output, 'manifest.json'), JSON.stringify(manifest, null, 2) + '\n');
    console.log(JSON.stringify({mode, projects: manifest.projects.length, selected: selected.length,
        changed: manifest.changed, reused: manifest.reused, milliseconds: manifest.milliseconds,
        sha256: manifest.sha256}));
}
module.exports = {observer, inputHash, environment, run};
