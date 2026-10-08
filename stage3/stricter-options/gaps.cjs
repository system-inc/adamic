// Hold each runtime representation gap to source Node and the current compiler.
const fs = require('fs');
const path = require('path');
const os = require('os');
const {spawnSync} = require('child_process');
const ts = require('typescript');
if (process.argv.length !== 3) throw Error('usage: NODE_PATH=<stage3/api/node_modules> node gaps.cjs <adamic>');
const compiler = path.resolve(process.argv[2]);
const scratch = fs.mkdtempSync(path.join(os.tmpdir(), 'stricter-options-gaps-'));
for (const [name, expected] of Object.entries({
    'array-hole': 'undefined\n',
    'typed-array': 'undefined\n',
    'record': 'undefined\n',
    'catch-value': 'not Error\n',
})) {
    const file = path.join(__dirname, 'gaps', name + '.a');
    const javascript = ts.transpileModule(fs.readFileSync(file, 'utf8'), {
        compilerOptions: {target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext},
    }).outputText;
    const node = spawnSync(process.execPath, ['--input-type=module', '-e', javascript], {encoding: 'utf8'});
    if (node.status !== 0 || node.stdout !== expected || node.stderr !== '') throw Error(`${name}: source Node disagrees: ${JSON.stringify(node)}`);
    const native = spawnSync(compiler, ['build', file, '-o', path.join(scratch, name)], {encoding: 'utf8'});
    console.log(`${name}: Node exit=${node.status} stdout=${JSON.stringify(node.stdout)}; native build exit=${native.status}`);
    console.log(native.stderr.trim());
    if (native.status !== 1 || !/stage 0 can.t lower|refus/.test(native.stderr)) throw Error(`${name}: gap changed; inspect ${scratch}`);
}
console.log('4 representation gaps observed; no exit-70 native runtime checks claimed');
