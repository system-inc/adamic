"use strict";
const assert = require("node:assert/strict");
function normalizeDecoderJS(text) {
    const pattern = /while \(nextCode !== undefined && \(nextCode & 0[bB]11000000\) === 0[bB]10000000\)/g;
    assert([...text.matchAll(pattern)].length <= 1, "duplicate decoder adaptation");
    return text.replace(pattern, match => match.replace("nextCode !== undefined && ", ""));
}
function compareDecoders(ts, before, after) {
    function compile(text) {
        const source = ts.createSourceFile("utilities.ts",text,ts.ScriptTarget.Latest,true);
        const fn = source.statements.find(node => ts.isFunctionDeclaration(node) && node.name.text === "getStringFromExpandedCharCodes");
        assert(fn);
        const js = ts.transpileModule(fn.getText(source),{compilerOptions:{target:ts.ScriptTarget.ES2024}}).outputText;
        return new Function(js + ";return getStringFromExpandedCharCodes;")();
    }
    const original = compile(before), adapted = compile(after);
    let cases = 0;
    function observe(fn, values, sequence) {
        const trace = []; let reads = 0;
        const codes = new Proxy([...values], {get(target,key,receiver) {
            trace.push(String(key));
            if (sequence && String(key)==="1") return sequence[Math.min(reads++,sequence.length-1)];
            return Reflect.get(target,key,receiver);
        }});
        return { result:fn(codes), trace };
    }
    function check(values, sequence) {assert.deepEqual(observe(adapted,values,sequence),observe(original,values,sequence),"decoder output or reads changed: "+JSON.stringify(values));cases++;}
    check([]);
    const tails = [0,1,65,127,128,129,191,192,193,224,240,255,NaN,Infinity,-Infinity];
    for(let first=0;first<256;first++) {check([first]);for(const second of tails) check([first,second]);}
    for(const first of [192,193,224,240,255]) for(const second of tails) for(const third of tails) check([first,second,third]);
    for(const values of [[65,66,67],[192,128,128,128],[240,159,152,128],[255,191,191,191,191],[NaN,Infinity,-Infinity],[-1,65535,2048]])check(values);
    for(const sequence of [[65,128],[128,65],[undefined,128],[128,undefined]])check([192,128],sequence);
    return { decoder_node_cases:cases, output_and_indexed_read_traces:"identical" };
}
module.exports={normalizeDecoderJS,compareDecoders};
