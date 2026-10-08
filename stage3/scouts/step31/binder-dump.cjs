'use strict';
// This observer calls only the parser and binder. It never creates a checker.
const fs = require('node:fs');

function dump(api, request) {
    if (api.version !== '6.0.3') throw new Error('expected TypeScript 6.0.3');
    const lines = [];
    const projectNames = new Set();
    for (const project of request.projects) {
        if (projectNames.has(project.id)) throw new Error('duplicate project id');
        projectNames.add(project.id);
        const converted = api.convertCompilerOptionsFromJson(project.options || {}, '/');
        if (converted.errors.length) throw new Error(api.flattenDiagnosticMessageText(converted.errors[0].messageText, '\n'));
        const files = [...project.files].sort((a, b) => a.path < b.path ? -1 : a.path > b.path ? 1 : 0);
        if (new Set(files.map(file => file.path)).size !== files.length) throw new Error('duplicate file');
        lines.push({record: 'project', format: 'adamic-binder-v1', typescript: api.version, id: project.id,
            options: Object.fromEntries(Object.entries(project.options || {}).sort(([a], [b]) => a < b ? -1 : a > b ? 1 : 0))});
        for (const input of files) {
            if (!input.path || input.path.startsWith('/') || input.path.includes('\\') || input.path.split('/').includes('..')) throw new Error('expected relative virtual file path');
            const file = api.createSourceFile(input.path.replace(/\.a$/, '.ts'), input.text,
                converted.options.target ?? api.ScriptTarget.Latest, true,
                input.scriptKind ?? (input.path.endsWith('.tsx') ? api.ScriptKind.TSX : input.path.endsWith('.js') ? api.ScriptKind.JS : api.ScriptKind.TS));
            api.bindSourceFile(file, converted.options);
            const nodes = [];
            const nodeIds = new Map();
            function visit(node) {
                nodeIds.set(node, nodes.length);
                nodes.push(node);
                api.forEachChild(node, child => { visit(child); }, children => { for (const child of children) visit(child); });
                // forEachChild intentionally omits attached JSDoc; the binder visits it.
                if (node.jsDoc) for (const doc of node.jsDoc) if (!nodeIds.has(doc)) visit(doc);
            }
            visit(file);
            const symbols = [];
            const symbolIds = new Map();
            function symbolId(symbol) {
                if (!symbol) return null;
                if (!symbolIds.has(symbol)) { symbolIds.set(symbol, symbols.length); symbols.push(symbol); }
                return symbolIds.get(symbol);
            }
            function table(value) {
                if (!value) return null;
                return [...value.entries()].sort(([a], [b]) => a < b ? -1 : a > b ? 1 : 0)
                    .map(([name, symbol]) => [name, symbolId(symbol)]);
            }
            function location(node) {
                if (!node) return null;
                if (!nodeIds.has(node)) throw new Error('binder declaration outside observed AST');
                return nodeIds.get(node);
            }
            const nodeRows = nodes.map((node, index) => ({id: index, kind: node.kind, pos: node.pos, end: node.end,
                flags: node.flags, symbol: symbolId(node.symbol), localSymbol: symbolId(node.localSymbol),
                locals: table(node.locals), flowFlags: node.flowNode?.flags ?? null}));
            const symbolRows = [];
            for (let index = 0; index < symbols.length; index++) {
                const symbol = symbols[index];
                symbolRows.push({id: index, name: symbol.escapedName, flags: symbol.flags,
                    declarations: (symbol.declarations || []).map(location), valueDeclaration: location(symbol.valueDeclaration),
                    parent: symbolId(symbol.parent), exportSymbol: symbolId(symbol.exportSymbol),
                    exports: table(symbol.exports), members: table(symbol.members), constEnumOnlyModule: symbol.constEnumOnlyModule ?? null});
            }
            function diagnostic(value) {
                return {start: value.start ?? null, length: value.length ?? null, code: value.code,
                    category: value.category, message: typeof value.messageText === 'string' ? value.messageText : chain(value.messageText),
                    related: (value.relatedInformation || []).map(diagnostic)};
            }
            function chain(value) {
                return {message: value.messageText, code: value.code, category: value.category, next: (value.next || []).map(chain)};
            }
            lines.push({record: 'file', path: input.path, nodes: nodeRows, symbols: symbolRows,
                parseDiagnostics: file.parseDiagnostics.map(diagnostic), bindDiagnostics: file.bindDiagnostics.map(diagnostic)});
        }
    }
    return lines.map(line => JSON.stringify(line)).join('\n') + '\n';
}
module.exports = {dump};
if (require.main === module) {
    const api = require(process.env.STEP31_TYPESCRIPT);
    process.stdout.write(dump(api, JSON.parse(fs.readFileSync(process.argv[2], 'utf8'))));
}
