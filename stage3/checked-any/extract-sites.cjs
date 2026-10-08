// Extract byte-for-byte declarations. Other stock modules become declarations only.
const fs = require('fs');
const path = require('path');
const crypto = require('crypto');
const ts = require('../api/node_modules/typescript');
const [stock, output, inventory] = process.argv.slice(2);
const sites = JSON.parse(fs.readFileSync(inventory)).records;
const files = [...new Set(sites.map(s => path.join(stock, s.file)))];
const options = { target: ts.ScriptTarget.ES2023, module: ts.ModuleKind.NodeNext,
    moduleResolution: ts.ModuleResolutionKind.NodeNext, declaration: true,
    emitDeclarationOnly: true, noEmitOnError: false, skipLibCheck: true,
    strict: true, exactOptionalPropertyTypes: true, stripInternal: false };
const program = ts.createProgram(files, options);
program.emit(undefined, (name, text) => {
    if (!name.endsWith('.d.ts')) return;
    const relative = path.relative(stock, name);
    const destination = path.join(output, 'declarations', relative);
    fs.mkdirSync(path.dirname(destination), {recursive: true});
    fs.writeFileSync(destination, text);
});
fs.mkdirSync(output, {recursive:true});
const manifest = [];
const checker = program.getTypeChecker();
for (let index = 0; index < sites.length; index++) {
    const site = sites[index];
    const file = program.getSourceFile(path.join(stock, site.file));
    const start = file.getPositionOfLineAndCharacter(site.line - 1, site.column - 1);
    let leaf = file;
    function find(node) {
        if (node.getStart(file) <= start && start < node.end) {
            leaf = node; ts.forEachChild(node, find);
        }
    }
    find(file);
    let declaration = leaf;
    while (declaration.parent && !ts.isFunctionDeclaration(declaration) &&
        !ts.isVariableStatement(declaration) && !ts.isInterfaceDeclaration(declaration) &&
        !ts.isTypeAliasDeclaration(declaration) && !ts.isClassDeclaration(declaration)) declaration = declaration.parent;
    const begin = declaration.getStart(file), end = declaration.end;
    let owner=leaf;
    while(owner.parent && !ts.isParameter(owner) && !ts.isVariableDeclaration(owner) && !ts.isAsExpression(owner) && !ts.isTypeAssertionExpression(owner)) owner=owner.parent;
    const binding=owner.name && ts.isIdentifier(owner.name) ? owner.name.text : ts.isAsExpression(owner)||ts.isTypeAssertionExpression(owner) ? owner.getText(file) : null;
    const text = file.text.slice(begin, end);
    const directory = path.join(output, String(index).padStart(3, '0'));
    fs.mkdirSync(directory, {recursive:true});
    // Dependency contracts retain their stock relative module paths.
    const destination = path.join(directory, path.dirname(site.file), '__any_probe.ts');
    fs.mkdirSync(path.dirname(destination), {recursive:true});
    fs.cpSync(path.join(output, 'declarations'), directory, {recursive:true});
    const importedNames = new Set();
    for (const n of file.statements.filter(ts.isImportDeclaration)) { for (const e of n.importClause?.namedBindings?.elements ?? []) importedNames.add(e.name.text); }
    const declarationFile = path.join(output, 'declarations', site.file.replace(/\.ts$/, '.d.ts'));
    const own = fs.existsSync(declarationFile) ? ts.createSourceFile(declarationFile, fs.readFileSync(declarationFile,'utf8'), ts.ScriptTarget.Latest,true) : undefined;
    const declaredNames = new Set();
    function names(n) { if (n.name && ts.isIdentifier(n.name)) declaredNames.add(n.name.text); ts.forEachChild(n,names); }
    names(declaration);
    const localImports = (own?.statements ?? []).filter(n=>n.name && (n.modifiers ?? []).some(m=>m.kind===ts.SyntaxKind.ExportKeyword)).map(n=>n.name.text).filter(n=>!importedNames.has(n)&&!declaredNames.has(n));
    const captured = new Map();
    function captures(node) {
        if(ts.isIdentifier(node) && !declaredNames.has(node.text)) {
            const symbol=checker.getSymbolAtLocation(node);
            if(symbol && symbol.flags & ts.SymbolFlags.Value && (symbol.declarations ?? []).some(d=>d.getSourceFile()===file && (d.end<=begin || d.getStart(file)>=end))) {
                captured.set(node.text,checker.typeToString(checker.getTypeOfSymbolAtLocation(symbol,node),node,ts.TypeFormatFlags.NoTruncation));
            }
        }
        ts.forEachChild(node,captures);
    }
    captures(declaration);
    
    // Keep source offsets and line numbers by blanking all other source text, including bodies.
    const mask = file.text.slice(0, begin).replace(/[^\r\n]/g, ' ');
    // Stub imported bindings by declarations referring to the exact stock types.
    // Import-type queries load checker contracts without adding runtime modules.
    let prefix = mask;
    const bindings=[];
    for(const n of file.statements.filter(ts.isImportDeclaration)) {
        const module=JSON.stringify(n.moduleSpecifier.text);
        for(const e of n.importClause?.namedBindings?.elements ?? []) {
            const name=e.name.text, exported=e.propertyName?.text ?? name;
            const symbol=checker.getSymbolAtLocation(e.name);
            const target=symbol && (symbol.flags&ts.SymbolFlags.Alias) ? checker.getAliasedSymbol(symbol):symbol;
            if(target?.flags&ts.SymbolFlags.Type) { const params=target.declarations?.find(d=>d.typeParameters)?.typeParameters ?? []; const generic=params.length?'<'+params.map(p=>p.getText(p.getSourceFile())).join(', ')+'>':''; const args=params.length?'<'+params.map(p=>p.name.text).join(', ')+'>':''; bindings.push('type '+name+generic+' = import('+module+').'+exported+args+';'); }
            if(target?.flags&ts.SymbolFlags.Value) bindings.push('declare const '+name+': typeof import('+module+').'+exported+';');
        }
        const namespace=n.importClause?.namedBindings;
        if(namespace&&ts.isNamespaceImport(namespace)) bindings.push('declare const '+namespace.name.text+': typeof import('+module+');');
    }
    // Same-file dependencies and lexical captures are declarations too.
    for(const name of localImports) {
        const module=JSON.stringify('./'+path.basename(site.file).replace(/\.ts$/,'.js'));
        const symbol=checker.getExportsOfModule(checker.getSymbolAtLocation(file)).find(s=>s.name===name);
        if(symbol?.flags&ts.SymbolFlags.Type) { const params=symbol.declarations?.find(d=>d.typeParameters)?.typeParameters ?? []; const generic=params.length?'<'+params.map(p=>p.getText(p.getSourceFile())).join(', ')+'>':''; const args=params.length?'<'+params.map(p=>p.name.text).join(', ')+'>':''; bindings.push('type '+name+generic+' = import('+module+').'+name+args+';'); }
        if(symbol?.flags&ts.SymbolFlags.Value) bindings.push('declare const '+name+': typeof import('+module+').'+name+';');
    }
    const valueNames=new Set(bindings.filter(b=>b.startsWith('declare const ')).map(b=>b.split(' ')[2].replace(':','')));
    const captureText=[...captured].filter(([name])=>!valueNames.has(name)).map(([name,type])=>'declare const '+name+': '+type+';').join('\n');
    fs.writeFileSync(destination, prefix + text + '\n' + [...new Set(bindings)].join('\n') + '\n'+captureText+'\n');
    manifest.push({...site, index, source:destination, declaration_start:begin, declaration_end:end,
        binding, first_line:file.getLineAndCharacterOfPosition(begin).line+1, last_line:file.getLineAndCharacterOfPosition(end).line+1, declaration_kind:ts.SyntaxKind[declaration.kind], declaration_sha256:crypto.createHash('sha256').update(text).digest('hex'),
        captures:[...captured].map(([name,type])=>({name,type})), imports_stubbed_as:'local ambient bindings with exact stock import-type queries; no runtime dependency modules', extraction:'original declaration text unchanged'});
}
fs.writeFileSync(path.join(output, 'manifest.json'), JSON.stringify(manifest,null,2)+'\n');
