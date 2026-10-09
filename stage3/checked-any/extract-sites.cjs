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
    strict: true, exactOptionalPropertyTypes: true, stripInternal: false, types: ["node"],
    typeRoots: [path.resolve(__dirname, "../api/node_modules/@types")] };
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
    let begin = declaration.getStart(file), end = declaration.end;
    let selectedNodes = [declaration];
    if (ts.isFunctionDeclaration(declaration) && !declaration.body) {
        const symbol = checker.getSymbolAtLocation(declaration.name);
        const family = (symbol?.declarations ?? []).filter(d=>ts.isFunctionDeclaration(d)&&d.getSourceFile()===file);
        if (family.some(d=>d.body)) {
            begin = Math.min(...family.map(d=>d.getStart(file)));
            end = Math.max(...family.map(d=>d.end));
            selectedNodes = family;
        }
    }
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
    function names(n) {
        if (n.name && ts.isIdentifier(n.name)) {
            const symbol = checker.getSymbolAtLocation(n.name);
            if (symbol && symbol.declarations?.some(d => d.getStart(file) >= begin && d.end <= end)) declaredNames.add(n.name.text);
        }
        ts.forEachChild(n,names);
    }
    for (const selected of selectedNodes) names(selected);
    const neededNames = new Set();
    function referenced(n) {
        if (ts.isIdentifier(n)) {
            const symbol = checker.getSymbolAtLocation(n);
            if (symbol && !symbol.declarations?.some(d => d.getSourceFile() === file && d.getStart(file) >= begin && d.end <= end)) neededNames.add(n.text);
        }
        ts.forEachChild(n,referenced);
    }
    for (const selected of selectedNodes) referenced(selected);
    const localImports = (own?.statements ?? []).filter(n=>n.name && (n.modifiers ?? []).some(m=>m.kind===ts.SyntaxKind.ExportKeyword)).map(n=>n.name.text).filter(n=>neededNames.has(n)&&!importedNames.has(n)&&!declaredNames.has(n));
    const captured = new Map();
    const captureKinds = new Map();
    function captures(node) {
        if(ts.isIdentifier(node)) {
            const symbol=ts.isShorthandPropertyAssignment(node.parent) ? checker.getShorthandAssignmentValueSymbol(node.parent) : checker.getSymbolAtLocation(node);
            if(symbol && symbol.flags & (ts.SymbolFlags.FunctionScopedVariable | ts.SymbolFlags.BlockScopedVariable | ts.SymbolFlags.Function | ts.SymbolFlags.Class | ts.SymbolFlags.RegularEnum | ts.SymbolFlags.ConstEnum | ts.SymbolFlags.ValueModule) && !(symbol.declarations ?? []).some(d=>d.getSourceFile()===file && d.getStart(file)>=begin && d.end<=end) && (symbol.declarations ?? []).some(d=>d.getSourceFile()===file && (d.end<=begin || d.getStart(file)>=end))) {
                captured.set(node.text,checker.typeToString(checker.getTypeOfSymbolAtLocation(symbol,node),node,ts.TypeFormatFlags.NoTruncation));
                const origin = symbol.valueDeclaration ?? symbol.declarations[0];
                captureKinds.set(node.text, ts.isVariableDeclaration(origin) && origin.parent.flags & ts.NodeFlags.Const ? 'const' : 'let');
            }
        }
        ts.forEachChild(node,captures);
    }
    for (const selected of selectedNodes) captures(selected);
    // A value's readonly enum members do not require a namespace. Keep
    // namespaces only for qualified type references present in stock text or
    // generated capture contracts, so unused ambient modules cannot block lowering.
    const qualifiedTypes = new Map();
    function typeQualifiers(n) {
        if (ts.isTypeReferenceNode(n) && ts.isQualifiedName(n.typeName) && ts.isIdentifier(n.typeName.left)) {
            const name = n.typeName.left.text;
            if (!qualifiedTypes.has(name)) qualifiedTypes.set(name, new Set());
            qualifiedTypes.get(name).add(n.typeName.right.text);
        }
        ts.forEachChild(n, typeQualifiers);
    }
    for (const selected of selectedNodes) typeQualifiers(selected);
    for (const type of captured.values()) typeQualifiers(ts.createSourceFile('capture.d.ts', 'type Capture = '+type+';', ts.ScriptTarget.Latest, true));
    // Follow only contracts referenced by the declaration or its captures. Adding
    // every export creates unrelated strict-checker failures before the site.
    const contracts = new Map();
    const unboundTypeParameters = new Set();
    function requireType(name) {
        if (declaredNames.has(name) || contracts.has(name)) return;
        const symbol = checker.resolveName(name, declaration, ts.SymbolFlags.Type, false);
        if (!symbol) return;
        const decl = symbol.declarations?.find(d => d.getSourceFile() === file);
        if (!decl || (decl.getStart(file) >= begin && decl.end <= end)) return;
        if (ts.isTypeParameterDeclaration(decl)) {unboundTypeParameters.add(name); return;}
        if (!ts.isInterfaceDeclaration(decl) && !ts.isTypeAliasDeclaration(decl) && !ts.isEnumDeclaration(decl)) return;
        contracts.set(name, decl.getText(file));
        typeQualifiers(decl);
        function dependencies(n) {if(ts.isIdentifier(n)) {neededNames.add(n.text); requireType(n.text);} ts.forEachChild(n,dependencies);}
        ts.forEachChild(decl,dependencies);
    }
    for (const name of [...neededNames]) requireType(name);
    for (const type of captured.values()) {
        const parsed = ts.createSourceFile('capture.d.ts', 'type Capture = '+type+';', ts.ScriptTarget.Latest, true);
        const bound = new Set();
        function parameters(n) {if(ts.isTypeParameterDeclaration(n)) bound.add(n.name.text);ts.forEachChild(n,parameters);}
        parameters(parsed);
        function dependencies(n) {if(ts.isIdentifier(n) && !bound.has(n.text)) {neededNames.add(n.text);requireType(n.text);}ts.forEachChild(n,dependencies);}
        dependencies(parsed);
    }
    fs.symlinkSync(path.resolve(__dirname, '../api/node_modules'), path.join(directory, 'node_modules'), 'dir');
    
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
            if (!neededNames.has(name) || contracts.has(name)) continue;
            const symbol=checker.getSymbolAtLocation(e.name);
            const target=symbol && (symbol.flags&ts.SymbolFlags.Alias) ? checker.getAliasedSymbol(symbol):symbol;
            const enumDeclaration = target?.declarations?.find(ts.isEnumDeclaration);
            if (enumDeclaration && ts.isEnumDeclaration(enumDeclaration) && enumDeclaration.modifiers?.some(m=>m.kind===ts.SyntaxKind.ConstKeyword) && enumDeclaration.name.text===name) {
                bindings.push('type '+name+' = import('+module+').'+exported+';');
                const members=enumDeclaration.members.map(m=>{
                    const key=m.name.text;
                    return 'readonly '+JSON.stringify(key)+': import('+module+').'+exported+'.'+key+';';
                });
                bindings.push('export declare const '+name+': {'+members.join(' ')+'};');
                const typeMembers = enumDeclaration.members.filter(m=>qualifiedTypes.get(name)?.has(m.name.text));
                if (typeMembers.length) bindings.push('export declare namespace '+name+' {'+typeMembers.map(m=>'export type '+m.name.text+' = import('+module+').'+exported+'.'+m.name.text+';').join(' ')+'}');
                continue;
            }
            if(target?.flags&ts.SymbolFlags.Type) { const params=target.declarations?.find(d=>d.typeParameters)?.typeParameters ?? []; const generic=params.length?'<'+params.map(p=>p.getText(p.getSourceFile())).join(', ')+'>':''; const args=params.length?'<'+params.map(p=>p.name.text).join(', ')+'>':''; bindings.push('type '+name+generic+' = import('+module+').'+exported+args+';'); }
            if(target?.flags&ts.SymbolFlags.Value) bindings.push('export declare const '+name+': typeof import('+module+').'+exported+';');
        }
        const namespace=n.importClause?.namedBindings;
        if(namespace&&ts.isNamespaceImport(namespace)) {
            const name=namespace.name.text;
            bindings.push('export declare const '+name+': typeof import('+module+');');
            const alias=checker.getSymbolAtLocation(namespace.name);
            const target=alias&&checker.getAliasedSymbol(alias);
            const types=(target?checker.getExportsOfModule(target):[]).filter(s=>neededNames.has(s.name)&&(s.flags&ts.SymbolFlags.Type));
            if (types.length) bindings.push('export declare namespace '+name+' {'+types.map(symbol=>{
                const params=symbol.declarations?.find(d=>d.typeParameters)?.typeParameters ?? [];
                const generic=params.length?'<'+params.map(p=>p.getText(p.getSourceFile())).join(', ')+'>':'';
                const args=params.length?'<'+params.map(p=>p.name.text).join(', ')+'>':'';
                return 'export type '+symbol.name+generic+' = import('+module+').'+symbol.name+args+';';
            }).join(' ')+'}');
        }
    }
    // Same-file dependencies and lexical captures are declarations too.
    for(const name of localImports.filter(n=>!contracts.has(n))) {
        const module=JSON.stringify('./'+path.basename(site.file).replace(/\.ts$/,'.js'));
        const symbol=checker.getExportsOfModule(checker.getSymbolAtLocation(file)).find(s=>s.name===name);
        if(symbol?.flags&ts.SymbolFlags.Type) { const params=symbol.declarations?.find(d=>d.typeParameters)?.typeParameters ?? []; const generic=params.length?'<'+params.map(p=>p.getText(p.getSourceFile())).join(', ')+'>':''; const args=params.length?'<'+params.map(p=>p.name.text).join(', ')+'>':''; bindings.push('type '+name+generic+' = import('+module+').'+name+args+';'); }
        if(symbol?.flags&ts.SymbolFlags.Value) bindings.push('export declare const '+name+': typeof import('+module+').'+name+';');
    }
    if ([...neededNames].some(n=>['process','require','Buffer','module','global'].includes(n)) || [...captured.values()].some(t=>t.includes('node:'))) bindings.push('import type {} from "node:module";');
    if (neededNames.has('require')) bindings.push('export declare const require: ReturnType<typeof import("node:module").createRequire>;');
    if (neededNames.has('process')) bindings.push('export declare const process: typeof import("node:process");');
        const valueNames=new Set(bindings.filter(b=>b.startsWith('export declare const ')).map(b=>b.split(' ')[3].replace(':','')));
    const captureText=[...captured].filter(([name])=>!valueNames.has(name)&&!(contracts.has(name)&&checker.resolveName(name,declaration,ts.SymbolFlags.Type,false)?.declarations?.some(ts.isEnumDeclaration))).map(([name,type])=>'export declare '+captureKinds.get(name)+' '+name+': '+type+';').join('\n');
    fs.writeFileSync(destination, prefix + text + '\n' + [...contracts.values()].join('\n')+'\n'+[...new Set(bindings)].join('\n') + '\n'+captureText+'\nexport {};\n');
    const referenceSites = [];
    const bindingSymbol = owner.name && ts.isIdentifier(owner.name) ? checker.getSymbolAtLocation(owner.name) : undefined;
    function uses(n) {
        if (ts.isIdentifier(n) && bindingSymbol && checker.getSymbolAtLocation(n) === bindingSymbol && n !== owner.name) {
            const pos=file.getLineAndCharacterOfPosition(n.getStart(file));
            referenceSites.push({line:pos.line+1,column:pos.character+1,kind:ts.SyntaxKind[n.parent.kind],within_declaration:n.getStart(file)>=begin&&n.end<=end});
        }
        ts.forEachChild(n,uses);
    }
    uses(file);
    manifest.push({...site, index, declaration_name:declaration.name?.getText(file) ?? binding ?? ts.SyntaxKind[declaration.kind],
        generic_declaration:!!declaration.typeParameters?.length, binding_kind:ts.SyntaxKind[owner.kind], references:referenceSites,
        source_stock_diagnostics:program.getSemanticDiagnostics(file).filter(d=>d.start>=begin&&d.start<end).map(d=>({code:d.code,message:ts.flattenDiagnosticMessageText(d.messageText,' ')})), source:destination, declaration_start:begin, declaration_end:end,
        binding, first_line:file.getLineAndCharacterOfPosition(begin).line+1, last_line:file.getLineAndCharacterOfPosition(end).line+1, declaration_kind:ts.SyntaxKind[declaration.kind], declaration_sha256:crypto.createHash('sha256').update(text).digest('hex'),
        unbound_type_parameters:[...unboundTypeParameters], ambient_contracts:[...contracts.keys()],
        captures:[...captured].map(([name,type])=>({name,type})), imports_stubbed_as:'local ambient bindings with exact stock import-type queries; no runtime dependency modules', extraction:'original declaration text unchanged'});
}
fs.writeFileSync(path.join(output, 'manifest.json'), JSON.stringify(manifest,null,2)+'\n');
