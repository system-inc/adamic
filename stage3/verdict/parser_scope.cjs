// Eligibility follows the stock parser, never the tested binary's output.
const fs = require('node:fs');
const ts = require(process.argv[2]);
if (ts.version !== '6.0.3') throw new Error('selection requires TypeScript 6.0.3');
const inputs = JSON.parse(fs.readFileSync(0, 'utf8'));
const results = {};
for (const input of inputs) {
    const file = ts.createSourceFile(input.name, input.content, ts.ScriptTarget.Latest, true);
    results[input.name] = file.parseDiagnostics.map(diagnostic => diagnostic.code);
}
process.stdout.write(JSON.stringify(results));
