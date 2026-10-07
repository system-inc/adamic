const fs = require('node:fs'), path = require('node:path');
const ts = require(process.env.SLICE_TYPESCRIPT);
const [beforeTree, afterTree] = process.argv.slice(2);
function helper(tree) {
    const text = fs.readFileSync(path.join(tree, 'src/compiler/scanner.ts'), 'utf8');
    const source = ts.createSourceFile('scanner.ts', text, ts.ScriptTarget.ESNext, true);
    const nodes = source.statements.filter(node => ts.isFunctionDeclaration(node)
        ? node.name?.text === 'utf16EncodeAsStringFallback'
        : ts.isVariableStatement(node) && node.declarationList.declarations.some(d => d.name.getText(source) === 'utf16EncodeAsStringWorker'));
    if (nodes.length !== 2) throw new Error('unexpected UTF-16 helper declarations');
    return nodes.map(node => node.getText(source)).join('\n');
}
function observe(source, available) {
    let calls = 0;
    const intrinsic = {fromCharCode: String.fromCharCode, fromCodePoint: available
        ? (...codes) => { calls++; return String.fromCodePoint(...codes); } : undefined};
    const Debug = {assert(value) { if (!value) throw new Error('False expression.'); }, fail(message) { throw new Error(message); }};
    const js = ts.transpileModule(source, {compilerOptions: {target: ts.ScriptTarget.ESNext}}).outputText;
    const worker = Function('String', 'Debug', js + '\nreturn utf16EncodeAsStringWorker;')(intrinsic, Debug);
    const values = [0, 65, 0xD800, 0xFFFF, 0x10000, 0x1F600, 0x10FFFF, 1.25].map(code => {
        try { return {code, result: worker(code)}; }
        catch (error) { return {code, error: error.name, message: error.message}; }
    });
    return {available, calls, values};
}
const before = helper(beforeTree), after = helper(afterTree);
const normal = [true, false].map(available => {
    const old = observe(before, available), current = observe(after, available);
    if (JSON.stringify(old) !== JSON.stringify(current)) throw new Error('codepoint availability differs');
    return current;
});
const mutant = after.replace('typeof String.fromCodePoint === "function"', 'typeof String.fromCodePoint === "undefined"');
if (mutant === after) throw new Error('availability mutant did not change source');
if (JSON.stringify(observe(mutant, true)) === JSON.stringify(normal[0])) throw new Error('availability mutant survived');
console.log(JSON.stringify({normal, availabilityMutantCaught: true}, null, 2));
