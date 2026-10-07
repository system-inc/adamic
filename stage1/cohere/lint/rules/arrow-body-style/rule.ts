export { syntaxKinds } from './listeners.a';
import { panic } from 'adamic';
import { Parser } from '../../../../typescript/parser/parser.ts';
import type { RuleContext } from '../../context.ts';
import { Fix, requireDetailedReporting } from '../typescript-no-non-null-assertion/suggestions.a';
import { FixedFinding } from './edits.a';
import { message } from './messages.a';
import { comments, type Comment } from '../eslint-comments-require-description/comments.a';
const name = 'arrow-body-style';
export class Rule {
    readonly context: RuleContext;
    readonly trivia: readonly Comment[];
    constructor(context: RuleContext) { this.context = context; this.trivia = context.enabled(name) ? comments(context) : []; }
    report(body: number, id: string, fixes: Fix[]): void {
        this.context.findings.push(new FixedFinding(name, id, message(id), this.context.start(body), this.context.node(body).end, fixes));
    }
    hasIn(index: number): boolean {
        const node = this.context.node(index);
        const token = node.children[1] ?? -1;
        if(node.kind === 'BinaryExpression' && token >= 0 && this.context.node(token).kind === 'InKeyword') return true;
        return node.children.some((child) => this.hasIn(child));
    }
    forInitializer(index: number): boolean {
        let cursor = index;
        let parent = this.context.parents[cursor] ?? -1;
        while(parent >= 0) {
            if(this.context.node(parent).kind === 'ForStatement') {
                const node = this.context.node(parent);
                const opening = this.context.source.indexOf('(', this.context.start(parent));
                const scanner = this.context.scanner;
                scanner.pos = opening + 1;
                scanner.scan();
                return node.children[0] === cursor && scanner.kind !== 'SemicolonToken';
            }
            cursor = parent; parent = this.context.parents[cursor] ?? -1;
        }
        return false;
    }
    hazard(body: number): boolean {
        const text = this.context.source;
        for(let position = this.context.node(body).end; position < text.length; position++) {
            const character = text.slice(position, position + 1);
            if(' \t\n\r'.includes(character)) continue;
            if(character === '/' && ['/', '*'].includes(text.slice(position + 1, position + 2))) {
                if(text.slice(position + 1, position + 2) === '/') {
                    while(position + 1 < text.length && !'\n\r'.includes(text.slice(position + 1, position + 2))) position++;
                }
                else {
                    const close = text.indexOf('*/', position + 2);
                    position = close < 0 ? text.length - 1 : close + 1;
                }
                continue;
            }
            return '([/`+-'.includes(character);
        }
        return false;
    }
    between(from: number, to: number): boolean {
        return this.trivia.some((comment) => comment.pos >= from && comment.end <= to);
    }
    semicolon(statement: number): number {
        const node = this.context.node(statement);
        for(let position = node.end - 1; position >= node.pos; position--) {
            const character = this.context.source.slice(position, position + 1);
            if(' \t\n\r'.includes(character)) continue;
            return character === ';' ? position : -1;
        }
        return -1;
    }
    remove(index: number, body: number, statements: readonly number[]): Fix[] {
        const context = this.context;
        const statement = statements[0] ?? -1;
        if(statements.length !== 1 || statement < 0 || context.node(statement).kind !== 'ReturnStatement') return [];
        const value = context.node(statement).children[0] ?? -1;
        if(value < 0 || this.hazard(body)) return [];
        const block = context.node(body);
        const opening = context.source.indexOf('{', block.pos);
        const closing = block.end - 1;
        const start = context.start(value);
        const end = context.node(value).end;
        const semicolon = this.semicolon(statement);
        const unwrapped = context.unwrap(value);
        const operator = context.node(unwrapped).children[1] ?? -1;
        let parens = context.source.slice(start, start + 1) === '{' ||
            (context.node(unwrapped).kind === 'BinaryExpression' && operator >= 0 && context.node(operator).kind === 'CommaToken') ||
            (this.hasIn(index) && this.forInitializer(index));
        if(context.node(value).kind === 'ParenthesizedExpression') parens = false;
        const open = parens ? '(' : '';
        const close = parens ? ')' : '';
        const trailing = semicolon >= 0 ? semicolon + 1 : end;
        if(this.between(opening + 1, start) || this.between(trailing, closing)) {
            const keyword = context.start(statement);
            const fixes = [new Fix(opening, opening + 1, ''), new Fix(keyword, keyword + 6, open)];
            if(semicolon >= 0) {
                fixes.push(new Fix(semicolon, semicolon + 1, close));
                fixes.push(new Fix(closing, closing + 1, ''));
            }
            else fixes.push(new Fix(closing, closing + 1, close));
            return fixes;
        }
        return [new Fix(opening, start, open), new Fix(semicolon >= 0 ? semicolon : end, block.end, close)];
    }
    leftmost(index: number): number {
        let cursor = this.context.unwrap(index);
        while(['PropertyAccessExpression', 'ElementAccessExpression', 'CallExpression'].includes(this.context.node(cursor).kind)) cursor = this.context.unwrap(this.context.node(cursor).children[0] ?? panic('missing receiver'));
        return cursor;
    }
    forcedObject(body: number): number {
        const context = this.context;
        if(context.source.slice(context.start(body), context.start(body) + 1) !== '(') return -1;
        let cursor = body;
        while(true) {
            const node = context.node(cursor);
            if(node.kind === 'ParenthesizedExpression') {
                const first = this.leftmost(node.children[0] ?? panic('missing parenthesized value'));
                return context.node(first).kind === 'ObjectLiteralExpression' ? first : -1;
            }
            if(!['PropertyAccessExpression', 'ElementAccessExpression', 'CallExpression'].includes(node.kind)) return -1;
            cursor = node.children[0] ?? panic('missing leftmost operand');
        }
    }
    matching(from: number): number {
        const text = this.context.source;
        let depth = 0;
        for(let position = from; position < text.length; position++) {
            const character = text.slice(position, position + 1);
            if("'\"`".includes(character)) {
                const quote = character;
                position++;
                while(position < text.length) {
                    const next = text.slice(position, position + 1);
                    if(next === '\\') position++;
                    else if(next === quote) break;
                    position++;
                }
            }
            else if(character === '(') depth++;
            else if(character === ')') { depth--; if(depth === 0) return position; }
        }
        return -1;
    }
    braces(index: number, body: number): Fix[] {
        const context = this.context;
        const start = context.start(body);
        const last = context.node(index).end;
        const object = this.forcedObject(body);
        if(object < 0) return [new Fix(start, start, '{return '), new Fix(last, last, '}')];
        const objectStart = context.start(object);
        const closing = this.matching(start);
        const sameLine = !/[\n\r]/u.test(context.source.slice(start, objectStart));
        const fixes = [new Fix(start, start + 1, sameLine ? '{return ' : '{')];
        if(!sameLine) fixes.push(new Fix(objectStart, objectStart, 'return '));
        if(closing < 0) fixes.push(new Fix(last, last, '}'));
        else if(closing + 1 === last) fixes.push(new Fix(closing, last, '}'));
        else { fixes.push(new Fix(closing, closing + 1, '')); fixes.push(new Fix(last, last, '}')); }
        return fixes;
    }
    visit(index: number): void {
        const context = this.context;
        const node = context.node(index);
        const body = node.children[node.children.length - 1] ?? -1;
        if(body < 0) return;
        let configured = context.settings.read('mode', 'AsNeeded');
        let requireObject = context.settings.read('requirereturnforobjectliteral', 'false') === 'true';
        if(context.settings.text.startsWith('[')) {
            const parser = new Parser(`(${context.settings.text});`, 'arrow body settings');
            const root = parser.file();
            const statement = parser.node(root).children[0] ?? panic('missing option statement');
            const expression = parser.node(statement).children[0] ?? panic('missing option expression');
            const array = parser.node(expression).children[0] ?? panic('missing option list');
            const entries = parser.node(array).children;
            const first = entries[0] ?? -1;
            const rawMode = first < 0 ? 'as-needed' : parser.node(first).text;
            configured = rawMode === 'as-needed' ? 'AsNeeded' : rawMode === 'always' ? 'Always' : rawMode === 'never' ? 'Never' : rawMode;
            const second = entries[1] ?? -1;
            if(second >= 0) for(const property of parser.node(second).children) {
                const key = parser.node(property).children[0] ?? -1;
                const value = parser.node(property).children[1] ?? -1;
                if(key >= 0 && value >= 0 && parser.node(key).text === 'requireReturnForObjectLiteral') requireObject = parser.node(value).kind === 'TrueKeyword';
            }
        }
        const mode = configured === '' ? 'AsNeeded' : configured;
        if(!['AsNeeded', 'Always', 'Never'].includes(mode)) panic('arrow-body-style: unsupported decoded mode');
        const block = context.node(body);
        if(block.kind !== 'Block') {
            if(mode === 'Always' || (mode === 'AsNeeded' && requireObject && context.node(context.unwrap(body)).kind === 'ObjectLiteralExpression')) this.report(body, 'expectedBlock', this.braces(index, body));
            return;
        }
        const statements = block.children;
        const statement = statements[0] ?? -1;
        if(statements.length !== 1 && mode !== 'Never') return;
        const returned = statement >= 0 && context.node(statement).kind === 'ReturnStatement' ? context.node(statement).children[0] ?? -1 : -1;
        if(mode === 'AsNeeded' && requireObject && statements.length === 1 && returned >= 0 && context.node(context.unwrap(returned)).kind === 'ObjectLiteralExpression') return;
        if(mode !== 'Never' && !(mode === 'AsNeeded' && statements.length === 1 && statement >= 0 && context.node(statement).kind === 'ReturnStatement')) return;
        const id = statements.length === 0 ? 'unexpectedEmptyBlock' : statements.length > 1 || statement < 0 || context.node(statement).kind !== 'ReturnStatement' ? 'unexpectedOtherBlock' : returned >= 0 && context.source.slice(context.start(returned), context.start(returned) + 1) === '{' ? 'unexpectedObjectBlock' : 'unexpectedSingleBlock';
        this.report(body, id, this.remove(index, body, statements));
    }
}
export function create(context: RuleContext): Rule { requireDetailedReporting(context, name); return new Rule(context); }
