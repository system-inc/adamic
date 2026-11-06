'use strict';
const e = require('./engine.cjs');
const {fs,path,assert,ts} = e;
const pin = require('./inventory.json').sites[0];
const file = path.join(path.resolve(process.argv[2]),'src/compiler',pin.file);
const raw = fs.readFileSync(file,'utf8');
let text = e.normalize(raw);
const owner = e.ownerMap(e.parse(file,text)).get(pin.owner);
assert(owner, 'adapted owner absent');
const start = text.indexOf(pin.after,owner.start);
assert(start >= owner.start && start + pin.after.length <= owner.end, 'adapted site absent');
text = text.slice(0,start) + pin.before + text.slice(start + pin.after.length);
const source = e.parse(file,text);
const imports = source.statements.filter(ts.isImportDeclaration);
const declaration = imports.find(node => node.moduleSpecifier.text === './_namespaces/ts.js');
assert(declaration && declaration.importClause?.namedBindings && ts.isNamedImports(declaration.importClause.namedBindings));
const elements = declaration.importClause.namedBindings.elements;
if (!elements.some(node => node.name.text === 'maybeBind')) {
    const anchor = elements.find(node => node.name.text === 'noop');
    assert(anchor, 'reviewed import anchor absent');
    const line = text.lastIndexOf('\n',anchor.getStart(source)) + 1;
    text = text.slice(0,line) + '    maybeBind,\n' + text.slice(line);
}
if(raw.includes('\r\n')) text = text.replace(/\n/g,'\r\n');
fs.writeFileSync(file,text);
console.log(JSON.stringify({file:pin.file,owner:pin.owner,restored:pin.before,importRestored:true}));
