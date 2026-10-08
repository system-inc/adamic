// The source AST decides the cast location independently of Adamic's IR.
const fs = require('node:fs'), path = require('node:path');
const ts = require(path.resolve(__dirname, '../../api/node_modules/typescript'));
function castSite(row) {
 const source = ts.createSourceFile(row.file, fs.readFileSync(path.join(__dirname, row.file), 'utf8'), ts.ScriptTarget.Latest, true, ts.ScriptKind.TS);
 const casts = [];
 function visit(node) {if (ts.isAsExpression(node)) casts.push(node); ts.forEachChild(node, visit);}
 visit(source);
 if (casts.length !== 1) throw Error(`expected one reduced cast: ${row.file}`);
 const position = source.getLineAndCharacterOfPosition(casts[0].getStart(source));
 return `${row.file}:${position.line + 1}:${position.character + 1}`;
}
function contract(row, result) {
 if (result.exit !== row.runtime_exit || result.stdout !== row.runtime_stdout) return false;
 if (!row.failing) return result.stderr === '';
 if (row.category === 'tagged') return result.stderr.startsWith('adamic: panic: cast failed:');
 if (row.category === 'untagged') return result.stderr.startsWith('adamic: panic: field read failed:') && result.stderr.includes(castSite(row));
 if (row.category === 'generic') return /^adamic: panic: (field|element) read failed:/.test(result.stderr) && result.stderr.includes(castSite(row));
 return result.stderr.startsWith('adamic: panic: cast failed:');
}
module.exports = {contract, castSite};
