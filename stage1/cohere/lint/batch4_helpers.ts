import { panic } from 'adamic';
import type { RuleContext } from './rule_context.ts';
import { Finding } from './finding.ts';
import { texts } from './batch4_messages.ts';
import { space } from './comments.ts';
export function text(messageName: string): string {
    return texts.get(messageName) ?? panic('missing message');
}
export function child(ctx: RuleContext, index: number, offset: number): number {
    return ctx.node(index).children[offset] ?? -1;
}
export function last(ctx: RuleContext, index: number): number {
    return child(ctx, index, ctx.node(index).children.length - 1);
}
export function kind(ctx: RuleContext, index: number): string {
    return index < 0 ? '' : ctx.node(index).kind;
}
export function named(ctx: RuleContext, index: number, name: string): boolean {
    return kind(ctx, index) === 'Identifier' && ctx.node(index).text === name;
}
export function key(ctx: RuleContext, index: number): number {
    for(const part of ctx.node(index).children)
        if(
            ![
                'Decorator',
                'PublicKeyword',
                'PrivateKeyword',
                'ProtectedKeyword',
                'ReadonlyKeyword',
                'OverrideKeyword',
                'DeclareKeyword',
                'StaticKeyword',
                'AbstractKeyword',
                'ExportKeyword',
                'DefaultKeyword',
                'AsyncKeyword',
                'AccessorKeyword',
                'DotDotDotToken',
            ].includes(kind(ctx, part))
        )
            return part;
    return -1;
}
export function range(
    ctx: RuleContext,
    rule: string,
    id: string,
    message: string,
    start: number,
    end: number,
    repair = '',
    replacement = '',
    editStart = start,
    editEnd = end,
): void {
    const finding = new Finding(rule, id, message, start, end, repair, replacement, '');
    finding.editStart = editStart;
    finding.editEnd = editEnd;
    ctx.findings.push(finding);
}
export function blank(value: string): boolean {
    for(const character of value) if(!space(character) && character !== '\ufeff') return false;
    return true;
}
export function trim(value: string): string {
    let start = 0;
    let end = value.length;
    while(start < end && space(value[start] ?? '')) start++;
    while(end > start && space(value[end - 1] ?? '')) end--;
    return value.slice(start, end);
}
export class Comment {
    readonly start: number;
    readonly end: number;
    readonly block: boolean;
    constructor(start: number, end: number, block: boolean) {
        this.start = start;
        this.end = end;
        this.block = block;
    }
}
function literals(ctx: RuleContext, index: number, ends: number[]): void {
    if(
        [
            'StringLiteral',
            'RegularExpressionLiteral',
            'NoSubstitutionTemplateLiteral',
            'TemplateHead',
            'TemplateMiddle',
            'TemplateTail',
            'JsxText',
        ].includes(kind(ctx, index))
    )
        ends[ctx.start(index)] = ctx.node(index).end;
    for(const part of ctx.node(index).children) literals(ctx, part, ends);
}
export function comments(ctx: RuleContext, root: number): Comment[] {
    const ends = new Array<number>(ctx.source.length + 1).fill(-1);
    literals(ctx, root, ends);
    const result: Comment[] = [];
    let cursor = 0;
    while(cursor < ctx.source.length) {
        const end = ends[cursor] ?? -1;
        if(end >= 0) {
            cursor = end;
            continue;
        }
        if(ctx.source.slice(cursor, cursor + 2) === '//') {
            const start = cursor;
            cursor += 2;
            while(cursor < ctx.source.length && !['\n', '\r', '\u2028', '\u2029'].includes(ctx.source[cursor] ?? ''))
                cursor++;
            result.push(new Comment(start, cursor, false));
        }
        else if(ctx.source.slice(cursor, cursor + 2) === '/*') {
            const start = cursor;
            const close = ctx.source.indexOf('*/', cursor + 2);
            cursor = close < 0 ? ctx.source.length : close + 2;
            result.push(new Comment(start, cursor, true));
        }
        else cursor++;
    }
    return result;
}
export function defaultParameters(ctx: RuleContext, index: number, rule: string, message: string, rest: boolean): void {
    if(!ctx.functionLike(index)) return;
    const body = last(ctx, index);
    if(body < 0 || (kind(ctx, index) !== 'ArrowFunction' && kind(ctx, body) !== 'Block')) return;
    const parameters = ctx.node(index).children.filter((part) => kind(ctx, part) === 'Parameter');
    let boundary = -1;
    for(let i = 0; i < parameters.length; i++) {
        const parameter = parameters[i] ?? panic('parameter');
        if(
            ctx.initializer(parameter) < 0 &&
            !ctx.hasKind(parameter, 'QuestionToken') &&
            !ctx.hasKind(parameter, 'DotDotDotToken')
        )
            boundary = i;
    }
    for(let i = 0; i < boundary; i++) {
        const parameter = parameters[i] ?? panic('parameter');
        if(
            ctx.initializer(parameter) >= 0 ||
            ctx.hasKind(parameter, 'QuestionToken') ||
            (rest && ctx.hasKind(parameter, 'DotDotDotToken'))
        )
            ctx.report(parameter, rule, 'shouldBeLast', message);
    }
}

export function integer(value: string): number {
    let result = 0;
    let sign = 1;
    let position = 0;
    if(value[0] === '-') {
        sign = -1;
        position++;
    }
    for(; position < value.length; position++) {
        const digit = '0123456789'.indexOf(value[position] ?? '');
        if(digit < 0) panic('invalid integer option');
        result = result * 10 + digit;
    }
    return result * sign;
}

export function args(ctx: RuleContext, index: number): number[] {
    const node = ctx.node(index);
    return node.list > 0 ? node.children.slice(node.children.length - node.list) : [];
}
