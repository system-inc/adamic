// Count syntax in the gathered closure and map each site back to pinned source.
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.SCOUT_TYPESCRIPT);
const [tree, slice] = process.argv.slice(2);
const manifest = JSON.parse(fs.readFileSync(path.join(slice, 'slice.json'), 'utf8'));
const sites = [];
for (const file of [...new Set(manifest.declarations.map(d => d.file))]) {
 const text = fs.readFileSync(path.join(slice, file), 'utf8');
 const original = ts.createSourceFile(file, fs.readFileSync(path.join(tree, file), 'utf8'), ts.ScriptTarget.Latest, true);
 const parsed = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
 function visit(node) {
  let category;
  if (ts.isModuleDeclaration(node)) category = 'namespace';
  if (ts.isVariableDeclaration(node) && node.name.getText(parsed) === 'isDebugging') category = 'mutable-namespace';
  if (ts.isAsExpression(node) || ts.isTypeAssertionExpression(node)) category = 'assertion-syntax';
  if (ts.isNewExpression(node) && node.expression.getText(parsed) === 'Array') category = 'array-length';
  if (ts.isNewExpression(node) && node.expression.getText(parsed) === 'Uint16Array') category = 'uint16';
  if (ts.isCallExpression(node) && node.expression.getText(parsed) === 'Object.entries') category = 'entries';
  if (ts.isComputedPropertyName(node)) category = 'computed-name';
  if (ts.isFunctionLike(node) && node.typeParameters?.length) category = 'generic-signature';
  if (category) {
   const start = node.getStart(parsed);
   const spans = manifest.declarations.filter(d => d.file === file && start >= d.output_start && start < d.output_end && d.kind !== 'NamespaceClosing');
   const span = spans.sort((a,b) => (a.output_end-a.output_start)-(b.output_end-b.output_start))[0];
   if (span) {
    const position = span.start + start - span.output_start;
    const lc = original.getLineAndCharacterOfPosition(position);
    sites.push({category,file,line:lc.line+1,column:lc.character+1,closure_line:parsed.getLineAndCharacterOfPosition(start).line+1,text:node.getText(parsed).split(/\r?\n/)[0].slice(0,180)});
   }
  }
  ts.forEachChild(node, visit);
 }
 visit(parsed);
}
const counts = {};
for (const s of sites) counts[s.category] = (counts[s.category] || 0)+1;
console.log(JSON.stringify({typescript:ts.version,summary:manifest.summary,counts,sites},null,2));
