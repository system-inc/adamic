const fs = require('fs');
const path = require('path');
const ts = require(path.join(process.argv[2], 'lib/typescript.js'));
const evidence = JSON.parse(fs.readFileSync(path.join(__dirname, process.argv[3] || 'direct-optional-original-witnesses.json'), 'utf8'));
const directories = new Map([[17, 'variable-declaration'], [18, 'variable-list'], [21, 'variable-statement'], [32, 'parameter-declaration']]);
const normalized = text => text.replace(/\s+/g, '');
let count = 0;
for (const member of evidence.members) {
    const directory = path.join(__dirname, 'optional-aggregates', directories.get(member.rank));
    for (const file of fs.readdirSync(directory).filter(file => (file.endsWith('.a') || file.endsWith('.ts')))) {
        const source = ts.createSourceFile(file, fs.readFileSync(path.join(directory, file), 'utf8'), ts.ScriptTarget.Latest, true);
        let declaration, binding;
        function visit(node) {
            if (ts.isMethodSignature(node) && node.parent.name?.text === 'Target' && node.name.getText(source) === member.field) declaration = node;
            if (member.readKind === 'binding' && ts.isBindingElement(node) && node.propertyName?.getText(source) === member.field) binding = node;
            if (member.readKind === 'property' && ts.isPropertyAccessExpression(node) && node.name.text === member.field) binding = node;
            ts.forEachChild(node, visit);
        }
        visit(source);
        if (!declaration || normalized(declaration.getText(source)) !== normalized(member.declaration)) throw Error(file + ': original declaration changed');
        if (!binding || normalized(binding.getText(source)) !== normalized(member.read)) throw Error(file + ': original member read changed');
        count++;
    }
}
console.log('Verified ' + count + ' fixtures retain both complete original declarations and member reads.');
