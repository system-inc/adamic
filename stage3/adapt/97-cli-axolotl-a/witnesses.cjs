'use strict';
const fs = require('node:fs');
const assert = require('node:assert/strict');
const ts = require('typescript');

// Runs verifyCompilerOptions' substitution loop, lifted out of each program.ts and transpiled, on the same entries (README.md).
// usage: node witnesses.cjs <program.ts before> <program.ts after>
function loop(fileName) {
    const source = ts.createSourceFile(fileName, fs.readFileSync(fileName, 'utf8'), ts.ScriptTarget.Latest, true);
    const loops = [];
    (function visit(node) {
        if (ts.isForStatement(node) && ts.isBlock(node.statement) && node.statement.statements.some(s => ts.isVariableStatement(s) && s.declarationList.declarations.some(d => ts.isIdentifier(d.name) && d.name.text === 'subst'))) loops.push(node);
        ts.forEachChild(node, visit);
    })(source);
    assert.equal(loops.length, 1, `${fileName}: ${loops.length} substitution loops, not one`);
    const javascript = ts.transpileModule(loops[0].getText(source), {compilerOptions: {target: ts.ScriptTarget.ES2020}}).outputText;
    const condition = loops[0].statement.statements.find(ts.isIfStatement).expression.getText(source);
    const run = new Function('options', 'key', 'len', 'Diagnostics', 'hasZeroOrOneAsteriskCharacter', 'pathIsRelative', 'pathIsAbsolute', 'createDiagnosticForOptionPathKeyValue', javascript);
    return {condition, run};
}

// One entry per case: a string with two asterisks (and no leading ./ while baseUrl is unset), a number, and a sparse hole.
const cases = [
    {name: 'two asterisks', paths: {'@app/*': ['*/*']}},
    {name: 'a number', paths: {'@app/*': [42]}},
    // eslint-disable-next-line no-sparse-arrays
    {name: 'a hole', paths: {'@app/*': [, ]}},
];

function diagnostics(run, paths) {
    const reported = [];
    const options = {paths};
    for (const key in paths) {
        // A loop that throws, as a wrong condition does when a number reaches pathIsAbsolute, is an observation too.
        try {
            run(options, key, paths[key].length, ts.Diagnostics, ts.hasZeroOrOneAsteriskCharacter, ts.pathIsRelative, ts.pathIsAbsolute, (reportedKey, index, message, ...args) => {
                // formatStringFromArgs is what tsc's diagnostic creation calls; it refuses an undefined argument with a Debug Failure.
                let text;
                try {
                    text = ts.formatStringFromArgs(message.message, args);
                } catch (error) {
                    text = `throws: ${error.message}`;
                }
                reported.push({key: reportedKey, index, code: message.code, message: text, args: args.map(a => a === undefined ? '<undefined>' : a)});
            });
        } catch (error) {
            reported.push({throws: `${error.name}: ${error.message}`});
        }
    }
    return reported;
}

assert.equal(ts.version, '6.0.3');
assert.equal(process.argv.length, 4, 'usage: node witnesses.cjs <program.ts before> <program.ts after>');
const before = loop(process.argv[2]);
const after = loop(process.argv[3]);
const results = cases.map(c => ({name: c.name, before: diagnostics(before.run, c.paths), after: diagnostics(after.run, c.paths)}));
console.log(JSON.stringify({conditions: {before: before.condition, after: after.condition}, results}, null, 2));
for (const result of results) assert.deepEqual(result.after, result.before, `${result.name}: the adapted loop reports differently`);
console.log('same diagnostics, by message and argument, in all', results.length, 'cases');
