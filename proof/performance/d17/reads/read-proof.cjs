const path = require('node:path');
const ts = require('typescript');
const root = path.resolve(process.argv[2], 'src');
const parser = path.join(root, 'compiler/parser.ts');
const files = ts.sys.readDirectory(root, ['.ts']);
const program = ts.createProgram(files, {
    strict: true, skipLibCheck: true, module: ts.ModuleKind.NodeNext,
});
const c = program.getTypeChecker();
const source = program.getSourceFile(parser);
const declaration = source.statements.flatMap(s => s.declarationList?.declarations ?? [])
    .find(d => d.name.getText() === 'forEachChildTable');
const table = c.getSymbolAtLocation(declaration.name);
const place = n => {
    const f = n.getSourceFile();
    const p = f.getLineAndCharacterOfPosition(n.getStart());
    return path.relative(root, f.fileName) + ':' + (p.line + 1) + ':' + (p.character + 1);
};
let references = 0;
for (const f of program.getSourceFiles()) {
    function visit(n) {
        if (ts.isIdentifier(n) && c.getSymbolAtLocation(n) === table) {
            references++;
            console.log('Reference:', place(n), n === declaration.name ? 'declaration with initializer' : n.parent.getText());
        }
        ts.forEachChild(n, visit);
    }
    ts.forEachChild(f, visit);
}
console.log('Source files walked:', program.getSourceFiles().length);
console.log('References:', references);
console.log('Module-local const syntax:', !!(declaration.parent.flags & ts.NodeFlags.Const) &&
    declaration.parent.parent.parent === source);
console.log('Exported:', c.getExportsOfModule(c.getSymbolAtLocation(source)).some(s =>
    (s.flags & ts.SymbolFlags.Alias ? c.getAliasedSymbol(s) : s) === table));
const fn = source.statements.find(s => s.name?.text === 'ForEachChildFunction');
console.log('Function type alias:', ts.isTypeAliasDeclaration(fn) && ts.isFunctionTypeNode(fn.type));
const entries = declaration.initializer.properties;
console.log('Initializer entries:', entries.length);
const dataFunction = e => ts.isPropertyAssignment(e) &&
    (ts.isFunctionExpression(e.initializer) || ts.isArrowFunction(e.initializer) ||
        (ts.isIdentifier(e.initializer) &&
            c.getSymbolAtLocation(e.initializer)?.declarations?.some(ts.isFunctionDeclaration)));
console.log('Entries other than data properties holding functions:', entries.filter(e => !dataFunction(e)).length);
const tableType = c.getTypeAtLocation(declaration.type);
const mapped = tableType.declaration;
console.log('Mapped type:', mapped.getText());
const groups = new Map();
let nonNumeric = 0;
for (const m of c.getTypeAtLocation(mapped.typeParameter.constraint).types) {
    const k = c.getTypeOfPropertyOfType(m, 'kind');
    if (!(k?.flags & ts.TypeFlags.NumberLiteral)) nonNumeric++;
    if (!groups.has(k?.value)) groups.set(k?.value, []);
    groups.get(k?.value).push(m);
}
console.log('Mapped union members:', [...groups.values()].flat().length);
console.log('Distinct numeric keys:', groups.size, 'Non-numeric kinds:', nonNumeric);
for (const [key, members] of groups) {
    if (members.length < 2) continue;
    const labels = members.map(m => c.typeToString(m));
    const property = tableType.getProperty(String(key));
    const signatures = c.getSignaturesOfType(c.getTypeOfSymbol(property), ts.SignatureKind.Call);
    console.log('Duplicate key:', key, labels.join(' / '));
    console.log('Property type:', c.typeToString(c.getTypeOfSymbol(property)));
    console.log('Call signatures:', signatures.map(s => c.signatureToString(s)).join(' / '));
    console.log('Distinct type objects:', new Set(members).size,
        'Mutually assignable:', c.isTypeAssignableTo(members[0], members[1]) &&
        c.isTypeAssignableTo(members[1], members[0]));
    for (const m of members) console.log('Member declaration:', place(m.symbol.declarations[0]));
    for (const e of entries) {
        if (ts.isComputedPropertyName(e.name) && c.getTypeAtLocation(e.name.expression).value === key)
            console.log('Initializer:', place(e), e.initializer.parameters?.[0]?.getText());
    }
}
const dispatch = source.statements.find(s => s.name?.text === 'forEachChild');
const node = c.getTypeAtLocation(dispatch.parameters[0]);
const kind = c.getTypeOfPropertyOfType(node, 'kind');
console.log('Dispatch parameter:', c.typeToString(node), 'kind:', c.typeToString(kind));
console.log('Dispatch kind literals:', kind.types.length,
    'Non-numeric:', kind.types.filter(k => !(k.flags & ts.TypeFlags.NumberLiteral)).length);
let singleton = 0;
let multiple = 0;
let mismatches = 0;
for (const [key, members] of [...groups].sort(([a], [b]) => a - b)) {
    const property = c.getPropertyOfType(tableType, String(key));
    const sigs = c.getSignaturesOfType(c.getTypeOfSymbol(property), ts.SignatureKind.Call);
    const parameter = sigs.length === 1 && sigs[0].parameters.length > 0
        ? c.getTypeOfSymbol(sigs[0].parameters[0]) : undefined;
    const constituents = parameter?.isUnion() ? parameter.types : parameter ? [parameter] : [];
    const exact = constituents.length === members.length &&
        members.every(member => constituents.includes(member));
    if (members.length === 1) singleton++;
    else multiple++;
    if (!exact) mismatches++;
    console.log(`Key ${key}: members ${members.map(displayType).join(' | ')}; parameter ${displayType(parameter)}; exactly same constituents: ${exact}`);
}
console.log(`Key contracts: ${singleton} singleton, ${multiple} multiple, ${mismatches} mismatches`);
if (mismatches || singleton !== 173 || multiple !== 2) process.exitCode = 1;

function displayType(type) {
    return type ? c.typeToString(type) : '<none>';
}

let signatureFailures = 0;
for (const [key] of [...groups].sort(([a], [b]) => a - b)) {
    const property = c.getPropertyOfType(tableType, String(key));
    const type = c.getTypeOfSymbol(property);
    const signatures = c.getSignaturesOfType(type, ts.SignatureKind.Call);
    if (type.isUnion() || signatures.length !== 1) signatureFailures++;
    console.log(`Signature key ${key}: union=${type.isUnion()}, calls=${signatures.length}`);
}
const forms = new Map();
const identifiers = new Set();
let identifierFailures = 0;
let anyEntries = 0;
for (const entry of entries) {
    const init = entry.initializer;
    const form = ts.SyntaxKind[init.kind];
    forms.set(form, (forms.get(form) || 0) + 1);
    const symbol = ts.isIdentifier(init) ? c.getSymbolAtLocation(init) : undefined;
    const decls = symbol?.declarations ?? [];
    const type = c.getTypeAtLocation(init);
    const sigs = c.getSignaturesOfType(type, ts.SignatureKind.Call);
    const implementation = symbol ? decls.find(d => ts.isFunctionDeclaration(d) && d.body) : init;
    const any = !!(type.flags & ts.TypeFlags.Any) ||
        [...(implementation?.parameters ?? [])].some(parameter => !!(c.getTypeAtLocation(parameter).flags & ts.TypeFlags.Any));
    if (any) anyEntries++;
    if (ts.isIdentifier(init)) {
        identifiers.add(symbol);
        if (decls.length !== 1 || implementation === undefined || sigs.length !== 1) identifierFailures++;
    }
    console.log(`Entry ${place(entry)}: ${form}, any=${any}` +
        (symbol ? `, symbol=${symbol.getName()}, declarations=${decls.length}, overloads=${decls.filter(d => ts.isFunctionDeclaration(d) && !d.body).length}, ` +
            decls.map(d => `${ts.SyntaxKind[d.kind]} at ${place(d)}, body=${!!d.body}`).join(' / ') : ''));
}
console.log('Entry forms:', JSON.stringify(Object.fromEntries(forms)));
console.log('Distinct identifier symbols:', identifiers.size);
console.log('Failures:', signatureFailures, 'key signatures,', identifierFailures, 'identifier entries,', anyEntries, 'any entries');
if (signatureFailures || identifierFailures || anyEntries) process.exitCode = 1;

// Round 14 (Performance, #d1chbm7), item 1 (e): Adamic's count, from the checker. For every key, its mapped
// property's one call signature's parameters; for every entry, its declared parameters (an arrow's or a function
// expression's own, an identifier entry's through the declaration its symbol binds); then forEachChild's call.
const parameterShape = parameters => parameters.map(p =>
    `${p.name.getText()}${p.questionToken || p.initializer ? '?' : ''}${p.dotDotDotToken ? '...' : ''}`);
const signatureShape = signature => {
    const declared = signature.parameters.map(s => s.valueDeclaration);
    return { count: declared.length, required: signature.minArgumentCount, rest: !!signature.hasRestParameter || declared.some(d => d?.dotDotDotToken), names: parameterShape(declared.filter(Boolean)) };
};
let overSlot = 0;
let restCount = 0;
const slotByKey = new Map();
for (const [key] of [...groups].sort(([a], [b]) => a - b)) {
    const property = c.getPropertyOfType(tableType, String(key));
    const [signature] = c.getSignaturesOfType(c.getTypeOfSymbol(property), ts.SignatureKind.Call);
    const shape = signatureShape(signature);
    slotByKey.set(key, { signature, shape });
    if (shape.rest) restCount++;
    console.log(`Slot ${key}: ${shape.count} parameters (${shape.names.join(', ')}), required ${shape.required}, rest ${shape.rest}`);
}
for (const entry of entries) {
    const key = c.getTypeAtLocation(entry.name.expression).value;
    const init = entry.initializer;
    const declaration = ts.isIdentifier(init) ? c.getSymbolAtLocation(init).declarations.find(d => ts.isFunctionDeclaration(d) && d.body) : init;
    const parameters = [...declaration.parameters];
    const rest = parameters.some(p => p.dotDotDotToken);
    if (rest) restCount++;
    const slot = slotByKey.get(key).shape;
    const over = parameters.length > slot.count;
    if (over) overSlot++;
    console.log(`Entry ${place(entry)} key ${key}: slot ${slot.count}, entry ${parameters.length} (${parameterShape(parameters).join(', ')})${rest ? ', rest' : ''}${over ? ', more than its slot' : ''}`);
}
console.log(`Entries with more parameters than their slot: ${overSlot}; rest parameters in slots or entries: ${restCount}`);
// The call fn(node, cbNode, cbNodes): its count against each slot, and each argument after node against the slot's
// own parameter at that position. The slot is generic in its own T; the checker resolves the call through the view
// (ForEachChildFunction<any>), inferring T from the arguments, and the view and every slot are instantiations of the
// one alias, so their parameters after the first are the alias's own declarations: shown by declaration identity.
let call;
(function find(n) { if (!call && ts.isCallExpression(n) && n.expression.getText() === 'fn') call = n; ts.forEachChild(n, find); })(dispatch);
const resolved = c.getResolvedSignature(call);
console.log(`Call ${place(call)}: ${call.arguments.length} arguments; resolved ${c.signatureToString(resolved)}`);
let callFailures = 0;
for (const [key, { signature, shape }] of slotByKey) {
    if (call.arguments.length > shape.count || call.arguments.length < shape.required) callFailures++;
    for (let i = 1; i < call.arguments.length; i++) {
        const same = signature.parameters[i]?.valueDeclaration === resolved.parameters[i]?.valueDeclaration;
        const argument = c.getTypeAtLocation(call.arguments[i]);
        const parameter = c.getTypeOfSymbol(resolved.parameters[i]);
        const assignable = c.isTypeAssignableTo(argument, parameter);
        if (!same || !assignable) {
            callFailures++;
            console.log(`Call key ${key} argument ${i}: same declaration ${same}, ${c.typeToString(argument)} assignable to ${c.typeToString(parameter)}: ${assignable}`);
        }
    }
}
console.log(`Call failures across ${slotByKey.size} slots: ${callFailures}`);
if (overSlot || restCount || callFailures) process.exitCode = 1;
