import { panic } from 'adamic';
import type { RuleContext } from '../../context.ts';
import { isIdentifierStart, isIdentifierPart } from '../../../../typescript/scanner/characters.ts';
import { record } from '../typescript-no-non-null-asserted-optional-chain/diagnostic.ts';
import { message } from './messages.ts';
export function assignment(kind: string): boolean {
    return ['EqualsToken','PlusEqualsToken','MinusEqualsToken','AsteriskEqualsToken','AsteriskAsteriskEqualsToken','SlashEqualsToken','PercentEqualsToken','LessThanLessThanEqualsToken','GreaterThanGreaterThanEqualsToken','GreaterThanGreaterThanGreaterThanEqualsToken','AmpersandEqualsToken','BarEqualsToken','CaretEqualsToken','AmpersandAmpersandEqualsToken','BarBarEqualsToken','QuestionQuestionEqualsToken'].includes(kind);
}
function identifier(text: string): boolean {
    if(text === '') { return false; }
    for(let pos = 0; pos < text.length; pos++) {
        const code = text.codePointAt(pos) ?? -1;
        if(!(pos === 0 ? isIdentifierStart(code) : isIdentifierPart(code))) { return false; }
        if(code > 65535) { pos++; }
    }
    return true;
}
export class Rule {
    readonly context: RuleContext;
    readonly direction: string;
    readonly descriptors: boolean;
    readonly exports: boolean;
    constructor(context: RuleContext) {
        this.context = context;
        this.direction = context.settings.read('direction', 'always');
        this.descriptors = context.settings.read('considerpropertydescriptor','false') === 'true';
        this.exports = context.settings.read('includecommonjsmoduleexports','false') === 'true';
    }
    functionName(index: number): string {
        const node = this.context.node(index);
        if(node.kind !== 'FunctionExpression') { return ''; }
        for(const child of node.children) { const part = this.context.node(child); if(part.kind === 'Identifier') { return part.text; } }
        return '';
    }
    property(index: number): string {
        const node = this.context.node(index);
        if(!['PropertyAccessExpression','ElementAccessExpression'].includes(node.kind)) { return ''; }
        const last = node.children[node.children.length - 1] ?? panic('missing member');
        const key = this.context.node(last);
        if(node.kind === 'ElementAccessExpression' && key.kind !== 'StringLiteral') { return ''; }
        if(node.kind === 'PropertyAccessExpression' && key.kind !== 'Identifier') { return ''; }
        return key.text;
    }
    call(index: number, base: string, method: string): boolean {
        if(index < 0 || this.context.node(index).kind !== 'CallExpression') { return false; }
        const children = this.context.node(index).children;
        const callee = this.context.unwrap(children[0] ?? panic('missing callee'));
        const access = this.context.node(callee);
        if(access.kind !== 'PropertyAccessExpression') { return false; }
        const object = this.context.node(access.children[0] ?? panic('missing object'));
        return object.kind === 'Identifier' && object.text === base && this.property(callee) === method;
    }
    // Sentinel distinguishes an unreadable descriptor from an ordinary property.
    descriptor(index: number): string {
        const parent = this.context.parents[index] ?? -1;
        if(this.call(parent,'Object','defineProperty') || this.call(parent,'Reflect','defineProperty')) {
            const args = this.context.node(parent).children;
            const second = args[2] ?? -1;
            return second >= 0 && this.context.node(second).kind === 'StringLiteral' ? this.context.node(second).text : '\u0000';
        }
        if(parent < 0 || this.context.node(parent).kind !== 'PropertyAssignment') { return '\u0001'; }
        const container = this.context.parents[parent] ?? -1;
        const outer = container < 0 ? -1 : this.context.parents[container] ?? -1;
        if(!this.call(outer,'Object','defineProperties') && !this.call(outer,'Object','create')) { return '\u0001'; }
        const key = this.context.node(this.context.node(parent).children[0] ?? panic('missing key'));
        return key.kind === 'Identifier' ? key.text : '\u0000';
    }
    report(index: number, name: string, own: string, property: boolean): void {
        if(this.direction === 'never' ? name !== own : name === own) { return; }
        const id = this.direction === 'never' ? (property ? 'notMatchProperty' : 'notMatchVariable') : (property ? 'matchProperty' : 'matchVariable');
        record(this.context,index,'func-name-matching',id,message(id,own,name),[]);
    }
    visit(index: number, parent: number): void {
        const node = this.context.node(index);
        if(node.kind === 'VariableDeclaration') {
            const children = node.children;
            const key = this.context.node(children[0] ?? panic('missing declaration name'));
            const own = this.functionName(children[children.length - 1] ?? panic('missing declaration child'));
            if(key.kind === 'Identifier' && own !== '') { this.report(index,key.text,own,false); }
            return;
        }
        if(node.kind === 'BinaryExpression') {
            const children = node.children;
            if(!assignment(this.context.node(children[1] ?? panic('missing operator')).kind)) { return; }
            const own = this.functionName(children[2] ?? panic('missing right'));
            if(own === '') { return; }
            const left = children[0] ?? panic('missing left');
            const target = this.context.node(left);
            if(target.kind === 'Identifier') { this.report(index,target.text,own,false); return; }
            const name = this.property(left);
            if(!this.exports && name === 'exports') {
                const base = this.context.node(target.children[0] ?? panic('missing base'));
                if(base.kind === 'Identifier' && base.text === 'module') { return; }
            }
            if(identifier(name)) { this.report(index,name,own,true); }
            return;
        }
        const children = node.children;
        const own = this.functionName(children[children.length - 1] ?? panic('missing property value'));
        if(own === '') { return; }
        let keyIndex = children[0] ?? panic('missing property name');
        // Class modifiers precede the property name in the arena.
        for(const child of children) {
            const kind = this.context.node(child).kind;
            if(['Identifier','StringLiteral','NumericLiteral','ComputedPropertyName','PrivateIdentifier'].includes(kind)) { keyIndex = child; break; }
        }
        const key = this.context.node(keyIndex);
        if(key.kind === 'ComputedPropertyName') {
            const expr = this.context.node(key.children[0] ?? panic('missing computed key'));
            if(expr.kind === 'StringLiteral' && identifier(expr.text)) { this.report(index,expr.text,own,true); }
            return;
        }
        if(key.kind === 'Identifier') {
            let name = key.text;
            if(this.descriptors && name === 'value' && parent >= 0 && this.context.node(parent).kind === 'ObjectLiteralExpression') {
                const target = this.descriptor(parent);
                if(target === '\u0000') { return; }
                if(target !== '\u0001') { name = target; }
            }
            this.report(index,name,own,true);
        }
        else if(key.kind === 'StringLiteral' && identifier(key.text)) { this.report(index,key.text,own,true); }
    }
}
export function create(context: RuleContext): Rule { return new Rule(context); }
