// Reduced stock-TypeScript and Node checks for the survey's mechanical claims.
const assert = require('node:assert/strict');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const vm = require('node:vm');
const ts = require(process.env.TSC_SURVEY_TYPESCRIPT);
assert.equal(ts.version, '6.0.3');
const directory = fs.mkdtempSync(path.join(os.tmpdir(), 'adamic-layer-probes-'));
function codes(source) {
    const name = path.join(directory, 'probe.ts');
    fs.writeFileSync(name, source);
    const program = ts.createProgram([name], {strict: true, noUncheckedIndexedAccess: true, exactOptionalPropertyTypes: true, noImplicitReturns: true, noEmit: true, target: ts.ScriptTarget.ES2024});
    return ts.getPreEmitDiagnostics(program).map(d => d.code);
}
function output(source) {
    let result = '';
    vm.runInNewContext(ts.transpileModule(source, {compilerOptions: {target: ts.ScriptTarget.ES2024}}).outputText, {console: {log: (...values) => {result += values.join(' ') + '\n';}}});
    return result;
}
try {
    const alias = `function f(subst: string | undefined) { const typeOfSubst = typeof subst; if (typeOfSubst === "string") console.log(subst.toUpperCase()); } f("a"); f(undefined);`;
    const narrowed = alias.replace('typeOfSubst === "string"', 'typeof subst === "string"');
    assert(codes(alias).includes(18048));
    assert.deepEqual(codes(narrowed), []);
    assert.equal(output(alias), output(narrowed));
    console.log('PASS: direct typeof of const local preserves output and restores narrowing');
    const replace = `const stringReplace = String.prototype.replace; function f(s: string, replacement: string): string { return stringReplace.call(s, "*", replacement); } console.log(f("a*b*c", "$&x"));`;
    const annotated = replace.replace('const stringReplace =', 'const stringReplace: (this: string, search: string, replacement: string) => string =');
    assert(codes(replace).includes(2345));
    assert.deepEqual(codes(annotated), []);
    assert.equal(output(replace), output(annotated));
    console.log('PASS: checked overload annotation preserves captured builtin and replacement expansion');
    const index = `const digits = "ABCD"; for (const input of ["A", "ABCD"]) console.log(digits.indexOf(input[3]));`;
    const converted = index.replace('input[3]', 'String(input[3])');
    assert(codes(index).includes(2345));
    assert.deepEqual(codes(converted), []);
    assert.equal(output(index), output(converted));
    assert.notEqual(output(index), output(index.replace('input[3]', 'input.charAt(3)')));
    console.log('PASS: explicit String preserves indexOf coercion; charAt default changes output');
    const views = `interface Base { isReadonly?: boolean; } interface Source { isReadonly?: boolean | undefined; } const source: Source = {isReadonly: undefined}; function f(x: Base) { console.log(Object.hasOwn(x,"isReadonly"), x.isReadonly); } f(source);`;
    const aligned = views.replace('interface Base { isReadonly?: boolean;', 'interface Base { isReadonly?: boolean | undefined;');
    assert(codes(views).some(code => code === 2345 || code === 2379));
    assert.deepEqual(codes(aligned), []);
    assert.equal(output(views), output(aligned));
    console.log('PASS: optional view alignment preserves present undefined');
    const holes = `function forEach<T>(array: readonly T[], callback: (value: T) => void): void { for (let i=0; i<array.length; i++) callback(array[i]); } const holes: number[] = new Array<number>(1); forEach(holes, value => console.log(typeof value)); const values = [1,2]; forEach(values, value => { console.log(typeof value); delete values[1]; });`;
    assert(codes(holes).includes(2345));
    assert.equal(output(holes), 'undefined\nnumber\nundefined\n');
    console.log('PASS: admitted sparse arrays and callback deletion supply undefined to T callback');
} finally {
    fs.rmSync(path.join(directory, 'probe.ts'));
    fs.rmdirSync(directory);
}
