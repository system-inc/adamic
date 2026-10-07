import { syntaxKinds as listenerKinds } from './listeners.a';
export const syntaxKinds: readonly number[] = listenerKinds;
import { panic } from 'adamic';
import { Parser } from '../../../../typescript/parser/parser.ts';
import type { RuleContext } from '../../context.ts';
import { Finding } from '../../finding.ts';
const name = 'grouped-accessor-pairs';
const modifiers: readonly string[] = ['StaticKeyword', 'PublicKeyword', 'PrivateKeyword', 'ProtectedKeyword', 'ReadonlyKeyword', 'AbstractKeyword', 'DeclareKeyword', 'OverrideKeyword', 'AsyncKeyword', 'AccessorKeyword', 'Decorator'];
class Entry {
    readonly key: string;
    getter = -1;
    setter = -1;
    getterPosition = -1;
    setterPosition = -1;
    getters = 0;
    setters = 0;
    constructor(key: string) { this.key = key; }
}
export class Rule {
    readonly context: RuleContext;
    order = 'anyOrder';
    types = false;
    constructor(context: RuleContext) {
        this.context = context;
        if(context.enabled(name)) {
        const text = context.settings.text;
        if(text.startsWith('[')) {
            const parser = new Parser(`(${text});`, 'accessor options');
            const root = parser.file();
            const statement = parser.node(root).children[0] ?? panic('missing options');
            const wrapped = parser.node(statement).children[0] ?? panic('missing options');
            const array = parser.node(wrapped).children[0] ?? panic('missing options');
            const values = parser.node(array).children;
            if(values.length > 2) panic('grouped-accessor-pairs takes at most two options');
            const first = values[0] ?? -1;
            if(first >= 0) {
                if(parser.node(first).kind !== 'StringLiteral') panic('grouped-accessor-pairs takes a string order');
                this.order = parser.node(first).text;
            }
            const second = values[1] ?? -1;
            if(second >= 0) {
                if(parser.node(second).kind !== 'ObjectLiteralExpression') panic('grouped-accessor-pairs takes an option object');
                for(const property of parser.node(second).children) {
                    const key = parser.node(property).children[0] ?? -1;
                    const value = parser.node(property).children[1] ?? -1;
                    if(key < 0 || value < 0 || parser.node(key).text !== 'enforceForTSTypes') panic('unknown accessor option');
                    if(!['TrueKeyword', 'FalseKeyword'].includes(parser.node(value).kind)) panic('accessor option must be boolean');
                    this.types = parser.node(value).kind === 'TrueKeyword';
                }
            }
        }
        else {
            this.order = context.settings.read('order', 'anyOrder');
            this.types = context.settings.read('enforcefortstypes', 'false') === 'true';
        }
        if(!['anyOrder', 'getBeforeSet', 'setBeforeGet'].includes(this.order)) panic('invalid accessor order');
        }
    }
    key(member: number): number {
        for(const child of this.context.node(member).children) {
            if(!modifiers.includes(this.context.node(child).kind)) return child;
        }
        return -1;
    }
    staticMember(member: number): boolean { return this.context.node(member).children.some((child) => this.context.node(child).kind === 'StaticKeyword'); }
    settled(key: number): string {
        const context = this.context;
        let node = context.node(key);
        if(node.kind === 'ComputedPropertyName') {
            const value = node.children[0] ?? -1;
            if(value < 0) return '';
            key = context.unwrap(value); node = context.node(key);
            if(['Identifier', 'PrivateIdentifier'].includes(node.kind)) return '';
        }
        if(['Identifier', 'PrivateIdentifier', 'StringLiteral', 'NumericLiteral', 'NoSubstitutionTemplateLiteral'].includes(node.kind)) return `:${node.text}`;
        return '';
    }
    identity(member: number): string {
        const key = this.key(member);
        if(key < 0) return '';
        const settled = this.settled(key);
        if(settled !== '') return `${this.context.node(key).kind === 'PrivateIdentifier' ? 'private' : 'static'}${settled}`;
        return `computed:${this.context.source.slice(this.context.start(key), this.context.node(key).end)}`;
    }
    label(member: number): string {
        const context = this.context;
        const kind = context.node(member).kind === 'GetAccessor' ? 'getter' : 'setter';
        const prefix = this.staticMember(member) ? 'static ' : '';
        const key = this.key(member);
        if(key < 0) return kind;
        if(context.node(key).kind === 'PrivateIdentifier') return `${prefix}private ${kind} ${context.node(key).text}`;
        const settled = this.settled(key);
        return settled === '' ? `${prefix}${kind}` : `${prefix}${kind} '${settled.slice(1)}'`;
    }
    scan(members: readonly number[], filter: number): void {
        const context = this.context;
        const entries: Entry[] = [];
        for(let position = 0; position < members.length; position++) {
            const member = members[position] ?? panic('missing member');
            const kind = context.node(member).kind;
            if(!['GetAccessor', 'SetAccessor'].includes(kind)) continue;
            if(filter >= 0 && this.staticMember(member) !== (filter === 1)) continue;
            const key = this.identity(member);
            if(key === '') continue;
            let entry = entries.find((value) => value.key === key);
            if(entry === undefined) { entry = new Entry(key); entries.push(entry); }
            if(kind === 'GetAccessor') { entry.getters++; entry.getter = member; entry.getterPosition = position; }
            else { entry.setters++; entry.setter = member; entry.setterPosition = position; }
        }
        for(const entry of entries) {
            if(entry.getters !== 1 || entry.setters !== 1) continue;
            const getterFirst = entry.getterPosition < entry.setterPosition;
            const former = getterFirst ? entry.getter : entry.setter;
            const latter = getterFirst ? entry.setter : entry.getter;
            const distance = Math.abs(entry.getterPosition - entry.setterPosition);
            let id = '';
            let message = '';
            if(distance > 1) { id = 'notGrouped'; message = `Accessor pair ${this.label(former)} and ${this.label(latter)} should be grouped.`; }
            else if((this.order === 'getBeforeSet' && !getterFirst) || (this.order === 'setBeforeGet' && getterFirst)) {
                id = 'invalidOrder'; message = `Expected ${this.label(latter)} to be before ${this.label(former)}.`;
            }
            if(id === '') continue;
            const key = this.key(latter);
            context.findings.push(new Finding(name, id, message, context.start(latter), context.node(key >= 0 ? key : latter).end, '', '', ''));
        }
    }
    visit(index: number): void {
        const node = this.context.node(index);
        if(['InterfaceDeclaration', 'TypeLiteral'].includes(node.kind) && !this.types) return;
        const members = node.children.filter((child) => ['GetAccessor', 'SetAccessor', 'MethodDeclaration', 'Constructor', 'PropertyDeclaration', 'PropertyAssignment', 'ShorthandPropertyAssignment', 'SpreadAssignment', 'SemicolonClassElement', 'ClassStaticBlockDeclaration', 'IndexSignature', 'PropertySignature', 'MethodSignature', 'CallSignature', 'ConstructSignature'].includes(this.context.node(child).kind));
        if(['ClassDeclaration', 'ClassExpression'].includes(node.kind)) { this.scan(members, 0); this.scan(members, 1); }
        else this.scan(members, -1);
    }
}
export function create(context: RuleContext): Rule { return new Rule(context); }
