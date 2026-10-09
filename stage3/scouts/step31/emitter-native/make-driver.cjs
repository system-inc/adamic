// Generate a closed emitter replay, using the actual compiler interfaces.
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.STEP31_TYPESCRIPT);
const [recordPath, tree, slice, destination] = process.argv.slice(2);
const recorded = JSON.parse(fs.readFileSync(recordPath, 'utf8'));
const program = ts.createProgram([path.join(tree, 'src/compiler/types.ts')], {strict: true, target: ts.ScriptTarget.ESNext, module: ts.ModuleKind.ESNext});
const checker = program.getTypeChecker();
const source = program.getSourceFile(path.join(tree, 'src/compiler/types.ts'));
function properties(name) {
    const declaration = source.statements.find(n => ts.isInterfaceDeclaration(n) && n.name.text === name);
    return checker.getPropertiesOfType(checker.getTypeAtLocation(declaration));
}
const q = JSON.stringify;
const calls = recorded.calls.filter(c => c.group === 'resolver');
function value(v) {
    if (v?.$undefined) return 'undefined';
    if (v?.$node) {
        if (v.kind !== ts.SyntaxKind.SourceFile) throw new Error('only source-file resolver answers supported by this closed replay');
        return 'source';
    }
    if (v === null || ['string', 'number', 'boolean'].includes(typeof v)) return q(v);
    throw new Error('unsupported resolver result');
}
const methods = properties('EmitResolver').map(symbol => {
    const signature = checker.getSignaturesOfType(checker.getTypeOfSymbolAtLocation(symbol, source), ts.SignatureKind.Call)[0];
    const name = symbol.name;
    const parameters = signature.parameters.map((_,i) => 'arg' + i);
    let annotation = '';
    if (name === 'isLateBound') annotation = ': arg0 is LateBoundDeclaration';
    let body = `throw new Error(${q('unrecorded resolver method: '+name)});`;
    if (calls.some(c => c.name === name)) {
        body = 'switch (resolverCursor) {\n';
        calls.forEach((call,index) => {
            if (call.name !== name) return;
            const tests = call.args.map((a,i) => a?.$node ? `arg${i}.kind !== ${a.kind} || arg${i}.pos !== ${a.pos} || arg${i}.end !== ${a.end}` : `arg${i} !== ${value(a)}`);
            body += `case ${index}:\nif (${tests.join(' || ') || 'false'}) throw new Error("resolver argument mismatch");\nresolverCursor++; return ${value(call.result)};\n`;
        });
        body += `default: throw new Error(${q('resolver order mismatch: '+name)});\n}`;
    }
    return `${name}: (${parameters.join(', ')})${annotation} => { ${body} }`;
});
const hostBodies = {
    getCompilerOptions: 'return options;', getSourceFiles: 'return [source];',
    getSourceFile: 'return source;', getSourceFileByPath: 'return source;',
    getCurrentDirectory: 'return "/project";', getCommonSourceDirectory: 'return "/project/";',
    getCanonicalFileName: 'return arg0;', useCaseSensitiveFileNames: 'return true;',
    isSourceFileFromExternalLibrary: 'return false;', isSourceOfProjectReferenceRedirect: 'return false;',
    getRedirectFromSourceFile: 'return undefined;', isEmitBlocked: 'return false;',
    shouldTransformImportCall: 'return false;', getEmitModuleFormatOfFile: 'return ModuleKind.CommonJS;',
    getDefaultResolutionModeForFile: 'return undefined;', getModeForResolutionAtIndex: 'return undefined;',
    getBuildInfo: 'return undefined;', getSourceFileFromReference: 'return undefined;',
    fileExists: 'return arg0 === source.fileName;', getFileIncludeReasons: 'return createMultiMap<Path, FileIncludeReason>();',
    writeFile: 'outputs.push({name: arg0, text: arg1, bom: arg2, sources: (arg4 ?? []).map(file => file.fileName === source.fileName ? "recorded.a" : file.fileName)});',
};
const hostMethods = properties('EmitHost').filter(symbol => !(symbol.flags & ts.SymbolFlags.Optional)).map(symbol => {
    if (symbol.name === 'redirectTargetsMap') return 'redirectTargetsMap: new Map<Path, readonly string[]>()';
    const signature = checker.getSignaturesOfType(checker.getTypeOfSymbolAtLocation(symbol, source), ts.SignatureKind.Call)[0];
    if (!signature) throw new Error('unexpected nonmethod host field '+symbol.name);
    return `${symbol.name}: (${signature.parameters.map((_,i)=>'arg'+i).join(', ')}) => { ${hostBodies[symbol.name] || `throw new Error(${q('unrecorded host method: '+symbol.name)});`} }`;
});
const imports = ['emitFiles','createSourceFile','bindSourceFile','getTransformers','createMultiMap','ScriptTarget','ModuleKind','type EmitResolver','type EmitHost','type CompilerOptions','type LateBoundDeclaration','type Path','type FileIncludeReason'];
const compilerImport = path.join(slice,'src/compiler/_namespaces/ts.ts');
const entry = `import { programArguments, readTextFile } from "adamic";\nimport type {} from "node:util";\nimport { ${imports.join(', ')} } from ${q(compilerImport)};\n`+
    `const requestPath = programArguments()[0];\nif (requestPath === undefined) throw new Error("request pathname required");\nconst request = readTextFile(requestPath);\nif (request.kind === "Error") throw new Error(request.message);\nif (request.text !== ${q(JSON.stringify(recorded.request)+"\n")}) throw new Error("this closed replay requires its recorded input bytes");\n`+
    `const options: CompilerOptions = ${q(recorded.context.options)};\n`+
    `const source = createSourceFile(${q(recorded.context.sources[0].fileName)}, ${q(recorded.context.sources[0].text)}, ScriptTarget.ES2020, true);\nbindSourceFile(source, options);\n`+
    `let resolverCursor = 0;\nconst resolver: EmitResolver = {\n${methods.join(',\n')}\n};\n`+
    `const outputs: {name: string; text: string; bom: boolean; sources: string[]}[] = [];\nconst host: EmitHost = {\n${hostMethods.join(',\n')}\n};\n`+
    `const result = emitFiles(resolver, host, undefined, getTransformers(options), false, false);\n`+
    `if (resolverCursor !== ${calls.length}) throw new Error("resolver transcript incomplete");\n`+
    `console.log(${q(JSON.stringify(recorded.rows[0]))});\n`+
    `console.log(JSON.stringify({record: "result", emitSkipped: result.emitSkipped, diagnostics: result.diagnostics}));\n`+
    `for (const output of outputs) console.log(JSON.stringify({record: "output", path: output.name.slice(9), text: output.text, bom: output.bom, sources: output.sources}));\n`;
fs.writeFileSync(destination, entry);
