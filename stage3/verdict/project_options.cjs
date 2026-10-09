// Config eligibility and effective options come from the pinned external API.
const fs = require('node:fs');
const path = require('node:path').posix;
const ts = require(process.argv[2]);
if (ts.version !== '6.0.3') throw new Error('selection requires TypeScript 6.0.3');
const inputs = JSON.parse(fs.readFileSync(0, 'utf8'));
const results = {};
for (const input of inputs) {
    const files = new Map(input.units.map(unit => [path.resolve('/.src', unit.name), unit.content]));
    const configName = path.resolve('/.src', input.project);
    const config = ts.parseJsonText(configName, files.get(configName));
    const host = {
        useCaseSensitiveFileNames: false,
        fileExists: name => files.has(name),
        readFile: name => files.get(name),
        readDirectory: (directory, extensions, excludes, includes, depth) => ts.matchFiles(
            directory, extensions, excludes, includes, false, '', depth,
            dir => {
                const entries = {files: [], directories: []};
                for (const name of files.keys()) {
                    if (!name.startsWith(dir.replace(/\/$/, '') + '/')) continue;
                    const relative = name.slice(dir.replace(/\/$/, '').length + 1);
                    if (relative.includes('/')) entries.directories.push(relative.split('/')[0]);
                    else entries.files.push(relative);
                }
                entries.directories = [...new Set(entries.directories)];
                return entries;
            }, name => name),
    };
    const parsed = ts.parseJsonSourceFileConfigFileContent(config, host, path.dirname(configName), undefined, configName);
    const options = {};
    for (const declaration of ts.optionDeclarations) {
        let value = parsed.options[declaration.name];
        if (value === undefined) continue;
        if (declaration.type instanceof Map) {
            value = [...declaration.type].find(([, entry]) => entry === value)?.[0];
        }
        options[declaration.name] = value;
    }
    results[input.source] = {options, files: parsed.fileNames, errors: parsed.errors.map(error => error.code)};
}
process.stdout.write(JSON.stringify(results));
