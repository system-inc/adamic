// Find and remove the actual emitted cast calls and their transitive read stops.
const ts = require(require('node:path').resolve(__dirname, '../../api/node_modules/typescript'));
function checks(code) {
 const file = ts.createSourceFile('emitted.mjs', code, ts.ScriptTarget.Latest, true, ts.ScriptKind.JS), edits = [];
 function visit(node) {
  if (ts.isCallExpression(node)) {
   const name = node.expression.getText(file);
   let failure = false;
   function message(part) {if (ts.isStringLiteral(part) && /^(cast failed:|field read failed:|element read failed:)/.test(part.text)) failure = true; ts.forEachChild(part, message);}
   for (const argument of node.arguments) message(argument);
   if (failure && ['adamicCast', 'adamicCheckedViewCast', 'panic'].includes(name)) edits.push({start: node.getStart(file), end: node.end, value: name === 'panic' ? 'void 0' : node.arguments[0].getText(file)});
  }
  ts.forEachChild(node, visit);
 }
 visit(file);
 return edits;
}
function remove(code, keep = -1) {
 const edits = checks(code);
 if (!edits.length) throw Error('missing emitted check mutant');
 let result = code;
 for (const edit of edits.filter((_, index) => index !== keep).sort((a,b) => b.start - a.start)) result = result.slice(0, edit.start) + edit.value + result.slice(edit.end);
 return {code: result, removed: edits.length - (keep >= 0 ? 1 : 0)};
}
function requireComplete(code) {
 const remaining = checks(code);
 if (remaining.length) throw Error(`observer left ${remaining.length} emitted cast/view check(s) in place`);
}
module.exports = {checks, remove, requireComplete};
