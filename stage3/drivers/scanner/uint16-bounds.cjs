// Check writes in the actual sliced parsePseudoBigInt, including a bounds mutant.
const fs = require('node:fs');
const vm = require('node:vm');
const path = require('node:path');
const ts = require(process.env.TSC_ADAPT_TYPESCRIPT);
const source = fs.readFileSync(process.argv[2], 'utf8');
const ast = ts.createSourceFile('utilities.ts', source, ts.ScriptTarget.Latest, true);
const declaration = ast.statements.find(n => ts.isFunctionDeclaration(n) && n.name?.text === 'parsePseudoBigInt');
if (!declaration) throw Error('parsePseudoBigInt not found');
const functionText = declaration.getText(ast);
const codes = {};
for (const [name, char] of Object.entries({ b:'b', B:'B', o:'o', O:'O', x:'x', X:'X', _0:'0', _9:'9', F:'F', A:'A', a:'a' })) codes[name] = char.charCodeAt(0);
let writes = 0;
function CheckedArray(length) {
    const target = new Uint16Array(length);
    return new Proxy(target, {
        get(target, key) { return Reflect.get(target, key, target); },
        set(target, key, value) {
            if (/^\d+$/.test(String(key))) {
                if (Number(key) >= target.length) throw Error('out-of-range write');
                writes++;
            }
            return Reflect.set(target, key, value, target);
        },
    });
}
function compile(text) {
    const context = { exports:{}, Uint16Array:CheckedArray, CharacterCodes:codes, Debug:{ fail(message) { throw Error(message); } } };
    vm.runInNewContext(ts.transpileModule(text, { compilerOptions:{ target:ts.ScriptTarget.ES2022, module:ts.ModuleKind.CommonJS } }).outputText, context);
    return context.exports.parsePseudoBigInt;
}
const parse = compile(functionText);
const inputs = new Set(['0n', '0000n', '000123n']);
for (const [base, digit] of [['b','1'], ['o','7'], ['x','f']]) {
    for (let count = 1; count <= 512; count++) {
        inputs.add('0'+base+digit.repeat(count)+'n');
        inputs.add('0'+base+'0'.repeat(count)+'n');
        inputs.add('0'+base+'1'+'0'.repeat(count-1)+'n');
    }
}
let corpusLiterals = 0;
for (const file of fs.readdirSync(process.argv[3])) {
    const full = path.join(process.argv[3], file);
    if (!fs.statSync(full).isFile()) continue;
    for (const match of fs.readFileSync(full,'utf8').matchAll(/\b(?:0[xX][0-9a-fA-F_]+|0[bB][01_]+|0[oO][0-7_]+|[0-9][0-9_]*)n\b/g)) {
        inputs.add(match[0].replaceAll('_','')); corpusLiterals++;
    }
}
for (const input of inputs) {
    const actual = parse(input);
    const expected = BigInt(input.slice(0,-1)).toString();
    if (actual !== expected) throw Error('value mismatch: '+input);
}
const mutant = compile(functionText.replace('// Add the digits, one at a time', 'segments[segments.length] = 1;\n// Add the digits, one at a time'));
let caught = false;
try { mutant('0xfn'); } catch (error) { if (error.message === 'out-of-range write') caught = true; else throw error; }
if (!caught) throw Error('bounds mutant survived');
console.log(JSON.stringify({ inputs:inputs.size, corpus_literal_occurrences:corpusLiterals, writes_checked:writes, values_match_BigInt:true, out_of_range_mutant_caught:caught }));
