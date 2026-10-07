"use strict";
const assert = require("node:assert/strict");
function normalizeClampJS(text) {
    const adapted = '        const nextLineStart = lineStarts[line + 1];\n        return nextLineStart !== undefined && res > nextLineStart ? lineStarts[line + 1] :';
    const original = '        return res > lineStarts[line + 1] ? lineStarts[line + 1] :';
    const count = text.split(adapted).length - 1;
    assert(count <= 1, "duplicate clamp adaptation");
    return text.replace(adapted, original);
}
function compareClamps(ts, before, after) {
    function compile(text) {
        const source = ts.createSourceFile("scanner.ts", text, ts.ScriptTarget.Latest, true);
        const fn = source.statements.find(node => ts.isFunctionDeclaration(node) && node.name.text === "computePositionOfLineAndCharacter");
        assert(fn, "missing real scanner function");
        const js = ts.transpileModule(fn.getText(source), { compilerOptions: { target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.CommonJS } }).outputText;
        return new Function("Debug", "arrayIsEqual", "computeLineStarts", "exports", js + "; return exports.computePositionOfLineAndCharacter;");
    }
    const original = compile(before), adapted = compile(after);
    let cases = 0;
    function observe(factory, starts, line, character, debugText, edits, sequence) {
        const trace = [], debug = [];
        let reads = 0;
        const map = new Proxy([...starts], { get(target, key, receiver) {
            trace.push(String(key));
            if (sequence && String(key) === "1") return sequence[Math.min(reads++, sequence.length - 1)];
            return Reflect.get(target, key, receiver);
        } });
        const fn = factory({ fail(message) { debug.push(["fail", message]); throw new Error(message); }, assert(value, message) { debug.push(["assert", value, message]); if (!value) throw new Error("assertion"); } },
            (a,b) => a.length === b.length && a.every((v,i) => v === b[i]), ts.computeLineStarts, {});
        try { const result = fn(map, line, character, debugText, edits); return { result: Object.is(result, -0) ? "-0" : String(result), trace, debug }; }
        catch (error) { return { error: error.message, trace, debug }; }
    }
    function check(...args) { assert.deepEqual(observe(adapted, ...args), observe(original, ...args), "scanner result or read trace changed: " + JSON.stringify(args)); cases++; }
    for (const text of ["", "x", "\n", "x\n", "a\nb", "a\r\nb\n", "\r\n", "a\u2028b\u2029c", "one\ntwo\nthree\n"]) {
        const starts = ts.computeLineStarts(text);
        for (let line = -2; line <= starts.length + 2; line++) {
            for (const character of [-10,-1,0,1,2,5,100,NaN,Infinity,-Infinity]) {
                for (const debugText of [undefined, text]) for (const edits of [undefined, true]) check(starts, line, character, debugText, edits);
            }
        }
    }
    // Volatile accessors establish that the second read was not cached or skipped.
    for (const sequence of [[5,7],[5,undefined],[undefined,5],[NaN,7],[Infinity,9],[-Infinity,3]]) {
        for (const character of [-1,0,5,6,100]) for (const debugText of [undefined,"0123456789"]) check([0,5],0,character,debugText,true,sequence);
    }
    return { scanner_clamp_node_cases: cases, stdout_value_and_indexed_read_trace: "identical" };
}
module.exports = { normalizeClampJS, compareClamps };
