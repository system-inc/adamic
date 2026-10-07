import type { RuleContext } from '../../context.ts';
import { Fix, Suggestion, SuggestedFinding, requireDetailedReporting } from '../typescript-no-non-null-assertion/suggestions.a';
import { graphEnd } from './graph-boundary.a';
import { message } from './messages.a';
const name = 'array-callback-return';
const functions: readonly string[] = ['FunctionDeclaration', 'FunctionExpression', 'ArrowFunction', 'MethodDeclaration', 'Constructor', 'GetAccessor', 'SetAccessor'];
const targets: readonly string[] = ['every', 'filter', 'find', 'findIndex', 'findLast', 'findLastIndex', 'flatMap', 'forEach', 'map', 'reduce', 'reduceRight', 'some', 'sort', 'toSorted'];
export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) { this.context = context; }
    unwrap(index: number): number {
        let cursor = index;
        while(cursor >= 0 && ['ParenthesizedExpression', 'NonNullExpression'].includes(this.context.node(cursor).kind)) cursor = this.context.node(cursor).children[0] ?? -1;
        return cursor;
    }
    member(index: number): string {
        const cursor = this.unwrap(index);
        if(cursor < 0) return '';
        const node = this.context.node(cursor);
        if(!['PropertyAccessExpression', 'ElementAccessExpression'].includes(node.kind)) return '';
        const key = this.unwrap(node.children[node.children.length - 1] ?? -1);
        if(key < 0) return '';
        const property = this.context.node(key);
        return node.kind === 'PropertyAccessExpression' || ['StringLiteral', 'NoSubstitutionTemplateLiteral'].includes(property.kind) ? property.text : '';
    }
    consuming(index: number): number {
        let cursor = index;
        let parent = this.context.parents[cursor] ?? -1;
        while(parent >= 0 && this.context.node(parent).kind === 'ParenthesizedExpression') {
            cursor = parent; parent = this.context.parents[cursor] ?? -1;
        }
        return parent >= 0 && this.context.node(parent).kind === 'CallExpression' && this.context.node(parent).children[0] === cursor ? parent : -1;
    }
    method(index: number): string {
        const context = this.context;
        const node = context.node(index);
        if(node.children.some((child) => context.node(child).kind === 'AsteriskToken')) return '';
        const async = node.children.some((child) => context.node(child).kind === 'AsyncKeyword');
        let cursor = index;
        while(cursor >= 0) {
            const parent = context.parents[cursor] ?? -1;
            if(parent < 0) return '';
            const outer = context.node(parent);
            if(outer.kind === 'BinaryExpression') {
                const operator = outer.children[1] ?? -1;
                if(operator < 0 || !['AmpersandAmpersandToken', 'BarBarToken', 'QuestionQuestionToken'].includes(context.node(operator).kind)) return '';
                cursor = parent; continue;
            }
            if(['ConditionalExpression', 'ParenthesizedExpression'].includes(outer.kind)) { cursor = parent; continue; }
            if(outer.kind === 'ReturnStatement') {
                let enclosing = context.parents[parent] ?? -1;
                while(enclosing >= 0 && !functions.includes(context.node(enclosing).kind)) enclosing = context.parents[enclosing] ?? -1;
                if(enclosing < 0) return '';
                cursor = this.consuming(enclosing); continue;
            }
            if(outer.kind !== 'CallExpression') return '';
            const callee = this.unwrap(outer.children[0] ?? -1);
            if(callee < 0) return '';
            const method = this.member(callee);
            const object = this.unwrap(context.node(callee).children[0] ?? -1);
            const objectName = object >= 0 && context.node(object).kind === 'Identifier' ? context.node(object).text : '';
            const args = outer.children.slice(outer.children.length - outer.list);
            if(!async && method === 'from' && objectName.endsWith('Array') && args[1] === cursor) return method;
            if(!async && targets.includes(method) && args[0] === cursor) return method;
            if(method === 'fromAsync' && objectName === 'Array' && args[1] === cursor) return method;
            return '';
        }
        return '';
    }
    returns(index: number, output: number[]): void {
        const node = this.context.node(index);
        if(functions.includes(node.kind) || node.kind === 'ClassStaticBlockDeclaration') return;
        if(node.kind === 'ReturnStatement') output.push(index);
        for(const child of node.children) this.returns(child, output);
    }
    head(index: number): number[] {
        const context = this.context;
        const node = context.node(index);
        const start = context.start(index);
        if(node.kind === 'ArrowFunction') {
            const token = node.children.find((child) => context.node(child).kind === 'EqualsGreaterThanToken') ?? -1;
            if(token >= 0) return [context.start(token), context.node(token).end];
        }
        const parameters = node.children.filter((child) => context.node(child).kind === 'Parameter');
        const last = parameters[parameters.length - 1] ?? -1;
        const from = last >= 0 ? context.node(last).end : context.source.indexOf('(', start) + 1;
        const close = context.source.indexOf(')', from);
        return [start, close >= from ? close + 1 : node.end];
    }
    sentence(index: number, method: string): string {
        const node = this.context.node(index);
        const label = node.children.find((child) => this.context.node(child).kind === 'Identifier') ?? -1;
        const functionName = node.kind === 'ArrowFunction' ? 'arrow function' : label >= 0 ? "function '" + this.context.node(label).text.trim() + "'" : 'function';
        const qualified = ['from', 'fromAsync', 'of', 'isArray'].includes(method) ? 'Array.' : 'Array.prototype.';
        return ' Here ' + qualified + method + '() is being passed ' + functionName + '.';
    }
    report(index: number, id: string, callback: number, method: string, head: boolean, suggestions: Suggestion[]): void {
        const range = head ? this.head(index) : [this.context.start(index), this.context.node(index).end];
        this.context.findings.push(new SuggestedFinding(name, id, message(id) + this.sentence(callback, method), range[0] ?? 0, range[1] ?? 0, suggestions));
    }
    prepend(expression: number): Suggestion {
        const context = this.context;
        const node = context.node(expression);
        const start = context.start(expression);
        const parent = context.parents[expression] ?? -1;
        const adjacent = parent >= 0 && context.node(parent).kind === 'ReturnStatement' && context.source.slice(start - 6, start) === 'return';
        const prefix = adjacent ? ' void ' : 'void ';
        const parens = ['BinaryExpression', 'ConditionalExpression', 'YieldExpression'].includes(node.kind);
        const fixes = [new Fix(start, start, prefix + (parens ? '(' : ''))];
        if(parens) fixes.push(new Fix(node.end, node.end, ')'));
        return new Suggestion('prependVoid', message('prependVoid'), fixes);
    }
    visit(index: number): void {
        const method = this.method(index);
        if(method === '') return;
        const context = this.context;
        const node = context.node(index);
        const body = node.children[node.children.length - 1] ?? -1;
        if(body < 0) return;
        const returns: number[] = [];
        for(const child of context.node(body).children) this.returns(child, returns);
        const allowVoid = context.settings.read('allowvoid', 'false') === 'true';
        if(method === 'forEach') {
            if(context.settings.read('checkforeach', 'false') !== 'true') return;
            if(node.kind === 'ArrowFunction' && context.node(body).kind !== 'Block') {
                if(allowVoid && context.node(body).kind === 'VoidExpression') return;
                const start = context.start(body);
                const end = context.node(body).end;
                const suggestions = [new Suggestion('wrapBraces', message('wrapBraces'), [new Fix(start, start, '{'), new Fix(end, end, '}')])];
                if(allowVoid) suggestions.push(this.prepend(body));
                this.report(index, 'expectedNoReturnValue', index, method, true, suggestions);
                return;
            }
            for(const statement of returns) {
                const expression = context.node(statement).children[0] ?? -1;
                if(expression < 0 || (allowVoid && context.node(expression).kind === 'VoidExpression')) continue;
                this.report(statement, 'expectedNoReturnValue', index, method, false, allowVoid ? [this.prepend(expression)] : []);
            }
            return;
        }
        if(context.node(body).kind === 'Block' && graphEnd(context.start(index))) this.report(index, returns.length > 0 ? 'expectedAtEnd' : 'expectedInside', index, method, true, []);
        if(context.settings.read('allowimplicit', 'false') !== 'true') {
            for(const statement of returns) if(context.node(statement).children.length === 0) this.report(statement, 'expectedReturnValue', index, method, false, []);
        }
    }
}
export function create(context: RuleContext): Rule { requireDetailedReporting(context, name); return new Rule(context); }
