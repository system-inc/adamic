// #k8ed1f2: stock tsc's emitted JavaScript and declarations for src/compiler in one adapted tree, written under an
// output directory, so two trees' outputs can be compared byte for byte. Never merged.
// Usage: node emit.mjs <adapted tree> <out directory>
import fs from 'node:fs';
import path from 'node:path';
import {createRequire} from 'node:module';
import {fileURLToPath} from 'node:url';
const treeRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..', '..', '..');
const ts = createRequire(import.meta.url)(path.join(treeRoot, 'stage3', 'api', 'node_modules', 'typescript'));
const source = path.join(fs.realpathSync(process.argv[2]), 'src');
const out = path.resolve(process.argv[3]);
const roots = [];
for (const entry of fs.readdirSync(path.join(source, 'compiler'), {recursive: true})) if (String(entry).endsWith('.ts')) roots.push(path.join(source, 'compiler', String(entry)));
const config = ts.readConfigFile(path.join(source, 'tsconfig-base.json'), ts.sys.readFile);
const parsed = ts.parseJsonConfigFileContent(config.config, ts.sys, source);
const options = {...parsed.options, noEmit: false, outDir: out, rootDir: source, composite: false, declaration: true, declarationMap: false,
    sourceMap: false, emitDeclarationOnly: false, isolatedDeclarations: false, incremental: false, stripInternal: false,
    types: ['node'], typeRoots: [path.join(treeRoot, 'stage3', 'api', 'node_modules', '@types')]};
const built = ts.createProgram({rootNames: roots, options});
const result = built.emit();
console.log(`emitted ${result.emittedFiles?.length ?? 'some'} files, skipped ${result.emitSkipped}, ${result.diagnostics.length} emit diagnostics`);
