import { panic } from 'adamic';
import type { RuleContext } from '../../context.ts';
import { checked } from './options.ts';
import { missingProtector } from './messages.ts';
const modifierKinds = ['PublicKeyword', 'PrivateKeyword', 'ProtectedKeyword', 'ReadonlyKeyword', 'StaticKeyword', 'AbstractKeyword', 'AsyncKeyword', 'DeclareKeyword', 'ExportKeyword', 'DefaultKeyword', 'OverrideKeyword', 'AccessorKeyword'];
export class Rule {
    readonly context: RuleContext;
    readonly required: Map<string, string[]>;
    constructor(context: RuleContext, rawOptions: string = context.settings.text) { this.context = context; if(context.selected === 'base/security-require-context-access' && rawOptions === '' && context.settings.values.size > 0) { panic('NotYet: shared factory must supply raw context-access options'); } this.required = context.enabled('base/security-require-context-access') ? checked(context.selected === 'all' ? '' : rawOptions).requirements : new Map<string, string[]>(); }
    first(index: number): number { return this.context.node(index).children[0] ?? -1; }
    name(index: number): number {
        for(const child of this.context.node(index).children) {
            const kind = this.context.node(child).kind;
            if(kind !== 'Decorator' && !modifierKinds.includes(kind) && kind !== 'DotDotDotToken') { return child; }
        }
        return -1;
    }
    decoratorName(index: number, callsOnly: boolean): string {
        let expression = this.first(index);
        if(expression < 0) { return ''; }
        if(this.context.node(expression).kind === 'CallExpression') { expression = this.first(expression); }
        else if(callsOnly) { return ''; }
        return expression >= 0 && this.context.node(expression).kind === 'Identifier' ? this.context.node(expression).text : '';
    }
    names(index: number): string[] {
        const names: string[] = [];
        for(const child of this.context.node(index).children) { if(this.context.node(child).kind === 'Decorator') { names.push(this.decoratorName(child, false)); } }
        return names;
    }
    injectedKey(index: number): string {
        const expression = this.first(index);
        if(expression < 0 || this.context.node(expression).kind !== 'CallExpression' || this.decoratorName(index, true) !== 'InjectRequestContext') { return ''; }
        const children = this.context.node(expression).children;
        const callee = children[0] ?? -1;
        if(callee < 0) { return ''; }
        // Type arguments precede the argument list in ForEachChild. Find its opening token.
        this.context.scanner.pos = this.context.node(callee).end;
        while(this.context.scanner.scan() !== 'EndOfFile' && this.context.scanner.kind !== 'OpenParenToken') {}
        const opening = this.context.scanner.pos;
        for(let i = 1; i < children.length; i++) {
            const argument = children[i] ?? panic('missing call child');
            if(this.context.start(argument) < opening) { continue; }
            if(this.context.node(argument).kind === 'Identifier') { return this.context.node(argument).text; }
            if(this.context.node(argument).kind === 'PropertyAccessExpression') {
                const parts = this.context.node(argument).children;
                const property = parts[parts.length - 1] ?? -1;
                if(property >= 0 && this.context.node(property).kind === 'Identifier') { return this.context.node(property).text; }
            }
            return '';
        }
        return '';
    }
    admitsMissing(index: number): boolean {
        const node = this.context.node(index);
        if(['UndefinedKeyword', 'NullKeyword', 'AnyKeyword', 'UnknownKeyword', 'VoidKeyword'].includes(node.kind)) { return true; }
        if(node.kind === 'LiteralType' || node.kind === 'ParenthesizedType' || node.kind === 'UnionType') { return node.children.some(child => this.admitsMissing(child)); }
        return false;
    }
    soft(index: number): boolean {
        const name = this.name(index);
        if(name < 0) { return false; }
        this.context.scanner.pos = this.context.node(name).end;
        const next = this.context.scanner.scan();
        if(next === 'QuestionToken') { return true; }
        if(next !== 'ColonToken') { return false; }
        const children = this.context.node(index).children;
        const position = children.indexOf(name);
        const annotation = children[position + 1] ?? -1;
        return annotation >= 0 && this.admitsMissing(annotation);
    }
    aliases(root: number): Map<string, string> {
        const aliases = new Map<string, string>();
        for(const statement of this.context.node(root).children) {
            if(this.context.node(statement).kind !== 'ImportDeclaration') { continue; }
            const pending = [statement];
            while(pending.length > 0) {
                const index = pending.pop() ?? panic('missing import child');
                const node = this.context.node(index);
                if(node.kind === 'ImportSpecifier' && node.children.length >= 2) {
                    const exported = node.children[node.children.length - 2] ?? -1;
                    const local = node.children[node.children.length - 1] ?? -1;
                    if(exported >= 0 && local >= 0 && this.context.node(exported).kind === 'Identifier') { aliases.set(this.context.node(local).text, this.context.node(exported).text); }
                }
                for(const child of node.children) { pending.push(child); }
            }
        }
        return aliases;
    }
    check(index: number, aliases: Map<string, string>): void {
        const node = this.context.node(index);
        for(const child of node.children) { if(this.context.node(child).kind === 'Decorator' && this.decoratorName(child, true) === 'GraphQlFieldResolver') { return; } }
        const present = this.names(index);
        let parent = this.context.parents[index] ?? -1;
        while(parent >= 0) {
            const kind = this.context.node(parent).kind;
            if(kind === 'ClassDeclaration' || kind === 'ClassExpression') { for(const name of this.names(parent)) { present.push(name); } break; }
            if(kind === 'SourceFile') { break; }
            parent = this.context.parents[parent] ?? -1;
        }
        const injected: string[] = [];
        for(const parameter of node.children) {
            if(this.context.node(parameter).kind !== 'Parameter' || this.soft(parameter)) { continue; }
            for(const decorator of this.context.node(parameter).children) {
                if(this.context.node(decorator).kind !== 'Decorator') { continue; }
                const surface = this.injectedKey(decorator);
                if(surface === '') { continue; }
                const key = aliases.get(surface) ?? surface;
                if(this.required.has(key) && !injected.includes(key)) { injected.push(key); }
            }
        }
        for(const key of injected) {
            const requires = this.required.get(key) ?? panic('missing configured requirement');
            if(requires.some(name => present.includes(name))) { continue; }
            const target = node.kind === 'Constructor' ? index : this.name(index);
            this.context.report(target < 0 ? index : target, 'base/security-require-context-access', 'missingProtector', missingProtector(key, requires), '', '', '');
        }
    }
    visit(root: number): void {
        if(this.required.size === 0) { return; }
        const aliases = this.aliases(root);
        const pending = [root];
        while(pending.length > 0) {
            const index = pending.pop() ?? panic('missing member');
            const node = this.context.node(index);
            if(['MethodDeclaration', 'Constructor', 'GetAccessor', 'SetAccessor'].includes(node.kind)) { this.check(index, aliases); }
            for(let i = node.children.length - 1; i >= 0; i--) { pending.push(node.children[i] ?? panic('missing member child')); }
        }
    }
}
export function create(context: RuleContext, rawOptions: string = context.settings.text): Rule { return new Rule(context, rawOptions); }
