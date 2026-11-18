// #k8ed1f2: stock tsc's diagnostics over tsc's whole program in one adapted tree, one per line. Never merged.
// Usage: node tscheck.mjs <adapted tree>
import fs from 'node:fs';
import path from 'node:path';
import {createRequire} from 'node:module';
import {fileURLToPath} from 'node:url';
const treeRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..', '..', '..');
const ts = createRequire(import.meta.url)(path.join(treeRoot, 'stage3', 'api', 'node_modules', 'typescript'));
const source = path.join(fs.realpathSync(process.argv[2]), 'src');
const roots = [];
for (const directory of ['compiler', 'services', 'server', 'jsTyping', 'deprecatedCompat', 'tsc', 'tsserver', 'typescript', 'typingsInstaller', 'typingsInstallerCore', 'watchGuard']) {
    const base = path.join(source, directory);
    if (!fs.existsSync(base)) continue;
    for (const entry of fs.readdirSync(base, {recursive: true})) if (String(entry).endsWith('.ts')) roots.push(path.join(base, String(entry)));
}
const config = ts.readConfigFile(path.join(source, 'tsconfig-base.json'), ts.sys.readFile);
const parsed = ts.parseJsonConfigFileContent(config.config, ts.sys, source);
const options = {...parsed.options, noEmit: true, composite: false, declaration: false, declarationMap: false, emitDeclarationOnly: false,
    isolatedDeclarations: false, incremental: false, types: ['node'], typeRoots: [path.join(treeRoot, 'stage3', 'api', 'node_modules', '@types')]};
const built = ts.createProgram({rootNames: roots, options});
for (const d of [...built.getGlobalDiagnostics(), ...built.getSyntacticDiagnostics(), ...built.getSemanticDiagnostics()]) {
    const where = d.file ? `${path.relative(path.dirname(source), d.file.fileName)}:${d.file.getLineAndCharacterOfPosition(d.start).line + 1}` : 'global';
    console.log(`${where} TS${d.code} ${ts.flattenDiagnosticMessageText(d.messageText, ' ').slice(0, 220)}`);
}
