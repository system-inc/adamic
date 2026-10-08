// Research-only call tracing. Original parser methods execute unchanged.
import { Parser } from '../../../parser/parser.ts';
import { readFileSync } from 'node:fs';
const trace = [];
for(const name of ['arrowCandidate', 'parameters', 'parameter', 'recoverList', 'expect']) {
    const original = Parser.prototype[name];
    Parser.prototype[name] = function(...args) {
        const before = {token: this.kind(), position: this.scanner.fullStart, contexts: [...this.listContexts], diagnostics: this.diagnostics.length};
        const result = original.apply(this, args);
        const after = {token: this.kind(), position: this.scanner.fullStart, contexts: [...this.listContexts], diagnostics: this.diagnostics.length};
        if(name !== 'expect' || result === false) trace.push({method: name, args, before, after, result});
        return result;
    };
}
const source = readFileSync(process.argv[2], 'utf8');
const parser = new Parser(source);
const root = process.argv[3] === 'expression' ? parser.expression() : parser.file();
console.log(JSON.stringify({source, mode: process.argv[3] ?? 'file', root_kind: parser.node(root).kind, diagnostics: parser.diagnostics, trace}, null, 2));
