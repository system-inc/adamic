import { panic } from 'adamic';
import type { RuleContext } from '../../context.ts';
import { Parser } from '../../../../typescript/parser/parser.ts';
import { graphemes } from './graphemes.a';
const name = 'id-length';
const modifiers: readonly string[] = ['StaticKeyword', 'PublicKeyword', 'PrivateKeyword', 'ProtectedKeyword', 'ReadonlyKeyword', 'AbstractKeyword', 'DeclareKeyword', 'OverrideKeyword', 'AsyncKeyword', 'AccessorKeyword', 'Decorator', 'AsteriskToken', 'DotDotDotToken', 'ExportKeyword', 'DefaultKeyword'];
function exceptionMap(text: string): string[] {
    const parser = new Parser(`(${text});`, 'identifier exception map');
    const root = parser.file();
    const statement = parser.node(root).children[0] ?? panic('missing exception settings');
    const expression = parser.node(statement).children[0] ?? panic('missing exception settings');
    const object = parser.node(expression).children[0] ?? panic('missing exception map');
    const result: string[] = [];
    for(const property of parser.node(object).children) {
        const key = parser.node(property).children[0] ?? panic('missing exception key');
        const value = parser.node(property).children[1] ?? panic('missing exception value');
        if(parser.node(value).kind === 'TrueKeyword') result.push(parser.node(key).text);
    }
    return result;
}
export class Rule {
    readonly context: RuleContext;
    readonly minimum: number;
    readonly maximum: number;
    readonly hasMaximum: boolean;
    readonly properties: boolean;
    readonly exceptions: readonly string[];
    constructor(context: RuleContext) {
        this.context = context;
        this.minimum = Number(context.settings.read('minimum', context.settings.read('min', '2')));
        this.maximum = Number(context.settings.read('maximum', context.settings.read('max', '0')));
        this.hasMaximum = context.settings.read('hasmaximum', context.settings.read('max', '') === '' ? 'false' : 'true') === 'true';
        this.properties = context.settings.read('checkproperties', context.settings.read('properties', 'always') === 'never' ? 'false' : 'true') !== 'false';
        // Captured Go settings encode Exceptions as a map; raw options use an array.
        const exceptions = context.settings.list('exceptions', []);
        if(exceptions.length === 1 && (exceptions[0] ?? '').startsWith('{')) {
            const text = exceptions[0] ?? '';
            this.exceptions = exceptionMap(text);
        }
        else this.exceptions = exceptions;
        if(context.enabled(name) && context.settings.list('exceptionpatterns', []).length > 0) panic('id-length: runtime regex option requires shared matcher support');
    }
    parent(index: number): number { return this.context.parents[index] ?? -1; }
    ownName(index: number): number {
        const node = this.context.node(index);
        const children = node.children.filter((child) => !modifiers.includes(this.context.node(child).kind));
        const first = children[0] ?? -1;
        if(first < 0 || node.kind === 'ArrowFunction') return -1;
        if(node.kind === 'BindingElement') {
            const second = children[1] ?? -1;
            if(second >= 0 && this.context.source.slice(this.context.node(first).end, this.context.start(second)).includes(':')) return second;
        }
        if(['FunctionDeclaration', 'FunctionExpression', 'ClassDeclaration', 'ClassExpression'].includes(node.kind) && this.context.node(first).kind !== 'Identifier') return -1;
        return first;
    }
    target(index: number): boolean {
        let current = index;
        while(true) {
            const parent = this.parent(current);
            if(parent < 0) return false;
            const node = this.context.node(parent);
            if(node.kind === 'BinaryExpression') {
                const operator = node.children[1] ?? -1;
                return operator >= 0 && node.children[0] === current && this.context.node(operator).kind === 'EqualsToken';
            }
            if(['ForOfStatement', 'ForInStatement', 'Parameter'].includes(node.kind)) return true;
            if(!['PropertyAssignment', 'ShorthandPropertyAssignment', 'ObjectLiteralExpression', 'ArrayLiteralExpression', 'SpreadAssignment', 'SpreadElement', 'ParenthesizedExpression'].includes(node.kind)) return false;
            current = parent;
        }
    }
    importOptions(index: number): boolean {
        const parent = this.parent(index);
        if(parent < 0) return false;
        const node = this.context.node(parent);
        if(node.kind === 'CallExpression') {
            const callee = node.children[0] ?? -1;
            const args = node.list >= 0 ? node.children.slice(node.children.length - node.list) : node.children.slice(1);
            return callee >= 0 && this.context.node(callee).kind === 'ImportKeyword' && args[1] === index;
        }
        if(node.kind !== 'PropertyAssignment' || node.children[1] !== index) return false;
        const outer = this.parent(parent);
        return outer >= 0 && this.context.node(outer).kind === 'ObjectLiteralExpression' && this.importOptions(outer);
    }
    written(access: number): boolean {
        const parent = this.parent(access);
        if(parent < 0) return false;
        const node = this.context.node(parent);
        if(node.kind === 'BinaryExpression') {
            const operator = node.children[1] ?? -1;
            return operator >= 0 && node.children[0] === access && this.context.node(operator).kind === 'EqualsToken';
        }
        const object = this.parent(parent);
        return node.kind === 'PropertyAssignment' && node.children[1] === access && object >= 0 && this.context.node(object).kind === 'ObjectLiteralExpression' && this.target(object);
    }
    naming(index: number, parent: number): boolean {
        const node = this.context.node(parent);
        if(['VariableDeclaration', 'FunctionDeclaration', 'FunctionExpression', 'ArrowFunction', 'ClassDeclaration', 'ClassExpression', 'MethodDeclaration', 'GetAccessor', 'SetAccessor', 'PropertyDeclaration', 'Parameter', 'ImportClause', 'NamespaceImport'].includes(node.kind)) return this.ownName(parent) === index;
        if(node.kind === 'BindingElement') {
            if(this.ownName(parent) !== index) return false;
            const outer = this.parent(parent);
            if(outer >= 0 && this.context.node(outer).kind === 'ArrayBindingPattern') return true;
            if(node.children.some((child) => this.context.node(child).kind === 'DotDotDotToken')) return true;
            return this.properties || this.ownName(parent) !== node.children[0];
        }
        if(['PropertyAssignment', 'ShorthandPropertyAssignment'].includes(node.kind)) {
            const object = this.parent(parent);
            if(object < 0 || this.context.node(object).kind !== 'ObjectLiteralExpression') return false;
            if(this.target(object)) return node.kind === 'PropertyAssignment' ? node.children[1] === index : node.children[0] === index && this.properties;
            return this.properties && node.children[0] === index && !this.importOptions(object);
        }
        if(node.kind === 'PropertyAccessExpression') return this.properties && node.children[node.children.length - 1] === index && this.written(parent);
        if(node.kind === 'ImportSpecifier') {
            const local = node.children[1] ?? -1;
            const exported = node.children[0] ?? -1;
            return local === index && exported >= 0 && this.context.node(exported).text !== this.context.node(index).text;
        }
        return false;
    }
    visit(index: number): void {
        const context = this.context;
        const parent = this.parent(index);
        if(parent < 0) return;
        const node = context.node(index);
        const privateName = node.kind === 'PrivateIdentifier';
        const text = privateName ? node.text.slice(1) : node.text;
        const length = graphemes(text);
        const tooLong = this.hasMaximum && length > this.maximum;
        if((length >= this.minimum && !tooLong) || this.exceptions.includes(text) || !this.naming(index, parent)) return;
        const id = `${tooLong ? 'tooLong' : 'tooShort'}${privateName ? 'Private' : ''}`;
        const rendered = privateName ? tooLong ? `#'${text}'` : `'#${text}'` : `'${text}'`;
        const policy = tooLong
            ? `A ${privateName ? 'private field ' : ''}name this long is usually carrying context that belongs in its type or its scope rather than in its spelling.`
            : `A ${privateName ? 'private field ' : ''}name this short carries no information about what it holds, so every reader has to reconstruct that from the surrounding code. Name it for what it is.`;
        const message = `Identifier name ${rendered} is too ${tooLong ? 'long' : 'short'} (${tooLong ? '>' : '<'} ${tooLong ? this.maximum : this.minimum}). ${policy}`;
        context.report(index, name, id, message, '', '', '');
    }
}
export function create(context: RuleContext): Rule { return new Rule(context); }
