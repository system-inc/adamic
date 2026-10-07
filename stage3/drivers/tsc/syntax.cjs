// Eligibility only: reject parser errors before selecting a CLI corpus.
const fs = require('node:fs');
const ts = require(process.argv[2]);
if (ts.version !== '6.0.3') throw new Error('selection requires stock TypeScript 6.0.3');
const cases = JSON.parse(fs.readFileSync(0, 'utf8'));
const invalid = cases.filter(row => ts.createSourceFile(row.name, row.content, ts.ScriptTarget.Latest, true).parseDiagnostics.length > 0);
process.stdout.write(JSON.stringify(invalid.map(row => row.name)));
