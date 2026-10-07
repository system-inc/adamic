// The stage-1 arena does not retain NodeList Pos/End. Reconstruct only
// delimiters owned by cohere's enumerated list owners, skipping child spans.
import type { Parser } from '../../../../typescript/parser/parser.ts';
import { Scanner } from '../../../../typescript/scanner/scanner.ts';
import { isLineBreak } from '../../../../typescript/scanner/characters.ts';
function tokenStart(parser: Parser, index: number): number {
    const scanner = new Scanner(parser.scanner.text);
    scanner.pos = parser.node(index).pos;
    scanner.scan();
    return scanner.start;
}
function isParameterListKind(kind: string): boolean {
    switch(kind) {
        case 'FunctionDeclaration':
        case 'FunctionExpression':
        case 'ArrowFunction':
        case 'MethodDeclaration':
        case 'Constructor':
        case 'GetAccessor':
        case 'SetAccessor':
        case 'FunctionType':
        case 'ConstructorType':
        case 'CallSignature':
        case 'ConstructSignature':
        case 'MethodSignature':
        case 'IndexSignature':
            return true;
        default:
            return false;
    }
}
function isBracedListKind(kind: string): boolean {
    switch(kind) {
        case 'ClassDeclaration':
        case 'ClassExpression':
        case 'Block':
        case 'CaseBlock':
        case 'ModuleBlock':
        case 'InterfaceDeclaration':
        case 'EnumDeclaration':
        case 'TypeLiteral':
        case 'ObjectLiteralExpression':
        case 'ObjectBindingPattern':
        case 'NamedImports':
        case 'NamedExports':
            return true;
        default:
            return false;
    }
}

export function collectListInteriors(parser: Parser, index: number): number[] {
    const anchors: number[] = [];
    const node = parser.node(index);
    const parameterList = isParameterListKind(node.kind);
    const argumentsList = node.kind === 'CallExpression' || node.kind === 'NewExpression';
    const bracedList = isBracedListKind(node.kind);
    const arrayList = node.kind === 'ArrayLiteralExpression' || node.kind === 'ArrayBindingPattern';
    const caseList = node.kind === 'CaseClause' || node.kind === 'DefaultClause';
    if(parameterList || argumentsList || bracedList || arrayList || caseList) {
        const children = new Map<number, number>();
        const start = tokenStart(parser, index);
        for(const child of node.children) {
            const childStart = Math.max(start, parser.node(child).pos);
            children.set(childStart, Math.max(children.get(childStart) ?? 0, parser.node(child).end));
        }
        let depth = 0;
        for(let position = start; position < node.end; position++) {
            const childEnd = children.get(position);
            if(childEnd !== undefined && childEnd > position) {
                position = childEnd - 1;
                continue;
            }
            const code = parser.scanner.text.charCodeAt(position);
            if(code === 47) {
                const next = parser.scanner.text.charCodeAt(position + 1);
                if(next === 47) {
                    while(position < node.end && !isLineBreak(parser.scanner.text.charCodeAt(position))) {
                        position++;
                    }
                    continue;
                }
                if(next === 42) {
                    const close = parser.scanner.text.indexOf('*/', position + 2);
                    position = close < 0 ? node.end : close + 1;
                    continue;
                }
            }
            const opening =
                parameterList || argumentsList ? (node.kind === 'IndexSignature' ? 91 : 40) : arrayList ? 91 : 123;
            const closing = opening === 40 ? 41 : opening === 91 ? 93 : 125;
            if(code === opening) {
                depth++;
                if(depth === 1) {
                    anchors.push(position + 1);
                }
            }
            if(code === closing) {
                depth--;
            }
            if(code === 44 && depth === 1) {
                anchors.push(position + 1);
            }
            if(caseList && code === 58) {
                anchors.push(position + 1);
            }
        }
    }
    if(node.kind === 'JsxExpression' && node.children.length === 0) {
        anchors.push(node.pos + 1);
    }
    return anchors;
}
