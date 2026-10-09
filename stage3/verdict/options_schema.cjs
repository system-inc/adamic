// Freeze the pinned compiler's declarations; the verdict needs no Node runtime.
const fs = require('node:fs');
const ts = require(process.argv[2]);
if (ts.version !== '6.0.3') throw new Error('option metadata requires TypeScript 6.0.3');
const schema = {version: ts.version, options: {}};
for (const option of ts.optionDeclarations) {
    schema.options[option.name.toLowerCase()] = {
        name: option.name,
        type: typeof option.type === 'object' ? 'enum' : option.type,
        values: typeof option.type === 'object' ? Object.fromEntries(option.type) : undefined,
        filePath: !!option.isFilePath,
        elementPath: !!option.element?.isFilePath,
        configOnly: !!option.isTSConfigOnly,
        vary: !option.isCommandLineOnly && (option.type === 'boolean' || typeof option.type === 'object') &&
            (option.affectsProgramStructure || option.affectsEmit || option.affectsModuleResolution ||
                option.affectsBindDiagnostics || option.affectsSemanticDiagnostics || option.affectsSourceFile ||
                option.affectsDeclarationPath || option.affectsBuildInfo) || ['noEmit', 'isolatedModules'].includes(option.name),
    };
}
const path = require('node:path');
const crypto = require('node:crypto');
const declarationFiles = ['transformers/declarations.ts', 'transformers/declarations/diagnostics.ts'];
const codes = new Set();
schema.declarationSources = {};
for (const file of declarationFiles) {
    const content = fs.readFileSync(path.join(process.argv[3], 'src/compiler', file));
    schema.declarationSources[file] = crypto.createHash('sha256').update(content).digest('hex');
    for (const match of content.toString('utf8').matchAll(/Diagnostics\.(\w+)/g)) {
        if (ts.Diagnostics[match[1]]) codes.add(ts.Diagnostics[match[1]].code);
    }
}
schema.declarationErrors = [...codes].sort((a, b) => a - b);
process.stdout.write(JSON.stringify(schema, null, 2) + '\n');
