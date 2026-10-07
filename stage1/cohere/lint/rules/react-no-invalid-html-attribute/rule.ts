import { panic } from 'adamic';
import type { RuleContext } from '../../context.ts';
import type { Suggestion } from './repairs.ts';
const html = ["a", "abbr", "acronym", "address", "applet", "area", "article", "aside", "audio", "b", "base", "basefont", "bdi", "bdo", "bgsound", "big", "blink", "blockquote", "body", "br", "button", "canvas", "caption", "center", "cite", "code", "col", "colgroup", "content", "data", "datalist", "dd", "del", "details", "dfn", "dialog", "dir", "div", "dl", "dt", "em", "embed", "fieldset", "figcaption", "figure", "font", "footer", "form", "frame", "frameset", "h1", "h2", "h3", "h4", "h5", "h6", "head", "header", "hgroup", "hr", "html", "i", "iframe", "image", "img", "input", "ins", "kbd", "keygen", "label", "legend", "li", "link", "main", "map", "mark", "marquee", "math", "menu", "menuitem", "meta", "meter", "nav", "nobr", "noembed", "noframes", "noscript", "object", "ol", "optgroup", "option", "output", "p", "param", "picture", "plaintext", "portal", "pre", "progress", "q", "rb", "rp", "rt", "rtc", "ruby", "s", "samp", "script", "section", "select", "shadow", "slot", "small", "source", "spacer", "span", "strike", "strong", "style", "sub", "summary", "sup", "svg", "table", "tbody", "td", "template", "textarea", "tfoot", "th", "thead", "time", "title", "tr", "track", "tt", "u", "ul", "var", "video", "wbr", "xmp"];
const relations = new Map<string, string[]>([
    ["alternate", ["link", "area", "a"]],
    ["apple-touch-icon", ["link"]],
    ["apple-touch-startup-image", ["link"]],
    ["author", ["link", "area", "a"]],
    ["bookmark", ["area", "a"]],
    ["canonical", ["link"]],
    ["dns-prefetch", ["link"]],
    ["external", ["area", "a", "form"]],
    ["help", ["link", "area", "a", "form"]],
    ["icon", ["link"]],
    ["license", ["link", "area", "a", "form"]],
    ["manifest", ["link"]],
    ["mask-icon", ["link"]],
    ["modulepreload", ["link"]],
    ["next", ["link", "area", "a", "form"]],
    ["nofollow", ["area", "a", "form"]],
    ["noopener", ["area", "a", "form"]],
    ["noreferrer", ["area", "a", "form"]],
    ["opener", ["area", "a", "form"]],
    ["pingback", ["link"]],
    ["preconnect", ["link"]],
    ["prefetch", ["link"]],
    ["preload", ["link"]],
    ["prerender", ["link"]],
    ["prev", ["link", "area", "a", "form"]],
    ["search", ["link", "area", "a", "form"]],
    ["shortcut", ["link"]],
    ["shortcut icon", ["link"]],
    ["stylesheet", ["link"]],
    ["tag", ["area", "a"]]
]);
const messages = new Map<string, string>([
    ["neverValid", "`%s` is never a valid `%s` attribute value on any element. The browser does not warn about a rel token it does not recognise, it simply ignores it, so a typo here fails in the direction that looks fine: the relationship the author meant to declare is absent and nothing says so. Check it against the registered link relations."],
    ["notValidFor", "`%s` is a real `%s` value but not one that means anything on `<%s>`. Link relations are scoped to the elements they can appear on, so a token borrowed from a `<link>` and placed on an `<a>` is silently dropped rather than applied. Either move it to an element it is defined for, or use the token that carries the same meaning here."],
    ["onlyMeaningfulFor", "The `%s` attribute only has meaning on the tags: %s. On any other element it is not a recognised attribute at all, so React passes it through to the DOM where it sits as inert markup, and a reader who sees it reasonably assumes it is doing something. Remove it, or move it to the element that actually takes it."],
    ["emptyIsMeaningless", "An empty `%s` attribute is meaningless. A bare attribute with no value is a boolean attribute, and this one is not boolean: it names a relationship, so with nothing to name it declares nothing. This is almost always a half-finished edit. Give it a value or delete it."],
    ["noEmpty", "An empty `%s` attribute is meaningless. The attribute is present, so a reader takes it as deliberate, and the value is blank, so the browser takes it as absent. Those two readings disagree and the markup is the only place that would say which was meant. Give it a value or delete it."],
    ["onlyStrings", "The `%s` attribute only supports strings. A number, a boolean, `null` or an object here is coerced on its way to the DOM, so `rel={true}` becomes the literal text `true` and `rel={null}` removes the attribute entirely. Neither is what the expression looks like it does. Pass the token as a string."],
    ["noMethod", "The `%s` attribute cannot be a method. A method shorthand in a props object produces a function value, and this attribute takes a string, so the function is coerced to its own source text on the way to the DOM. That is never intended and it is invisible until someone reads the rendered markup."],
    ["notAlone", "`%s` must be directly followed by `%s`. On its own it is the legacy half of a two-token spelling and no browser treats it as complete, so the icon it was meant to declare is simply not declared. Write both tokens."],
    ["notPaired", "`%s` can not be directly followed by `%s` without `%s`. The first token is only meaningful as the opening half of that pair, so following it with anything else leaves both tokens doing nothing rather than one of them working."],
    ["spaceDelimited", "The `%s` attribute's values should be space delimited, with exactly one space and no leading or trailing whitespace. A tab, a newline, or a run of spaces parses correctly in every browser, so this never breaks anything; it is flagged because the irregular spacing usually means the value was assembled by concatenation and nobody looked at the result."]
]);
const goWhitespace = /[\t\n\v\f\r \u0085\u00a0\u1680\u2000-\u200a\u2028\u2029\u202f\u205f\u3000]+/;
export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) { this.context = context; }
    suggestions(index: number): Suggestion[] { return []; }
    report(index: number, id: string, values: string[]): void {
        let message = messages.get(id) ?? panic('missing message');
        for(const value of values) { message = message.replace('%s', value); }
        this.context.report(index, 'react/no-invalid-html-attribute', id, message, '', '', '');
    }
    meaningful(tag: string): boolean { return ['link','a','area','form'].includes(tag); }
    name(index: number): string {
        const node = this.context.node(index);
        return ['Identifier','StringLiteral'].includes(node.kind) ? node.text : '';
    }
    literal(index: number): boolean { return ['StringLiteral','NumericLiteral','BigIntLiteral','TrueKeyword','FalseKeyword','NullKeyword','RegularExpressionLiteral'].includes(this.context.node(index).kind); }
    literalText(index: number): string {
        const node = this.context.node(index);
        if(node.kind === 'TrueKeyword') { return 'true'; }
        if(node.kind === 'FalseKeyword') { return 'false'; }
        if(node.kind === 'NullKeyword') { return 'null'; }
        return node.text;
    }
    checkToken(index: number, value: string, tag: string): void {
        const tags = relations.get(value);
        if(tags === undefined) { this.report(index, 'neverValid', [value, 'rel']); }
        else if(!tags.includes(tag)) { this.report(index, 'notValidFor', [value, 'rel', tag]); }
    }
    value(index: number, tag: string): void {
        const node = this.context.node(index);
        const tokens: string[] = [];
        let start = -1;
        for(let cursor = 0; cursor <= node.text.length; cursor++) {
            if(cursor === node.text.length || goWhitespace.test(node.text.slice(cursor, cursor + 1))) {
                if(start >= 0) { tokens.push(node.text.slice(start, cursor)); start = -1; }
            }
            else if(start < 0) { start = cursor; }
        }
        if(tokens.length === 0) { this.report(index, 'noEmpty', ['rel']); return; }
        for(const token of tokens) { this.checkToken(index, token, tag); }
        for(let cursor = 0; cursor < tokens.length; cursor++) {
            if(tokens[cursor] !== 'shortcut') { continue; }
            if(cursor + 1 >= tokens.length) { this.report(index, 'notAlone', ['shortcut','icon']); }
            else if(tokens[cursor + 1] !== 'icon') { this.report(index, 'notPaired', ['shortcut',tokens[cursor + 1] ?? panic('missing token'),'icon']); }
        }
        let raw = this.context.source.slice(node.pos, node.end).replace(/^[ \t\r\n]+/, '');
        if(raw.length >= 2 && (raw[0] === '"' || raw[0] === "'") && raw[raw.length - 1] === raw[0]) { raw = raw.slice(1, -1); }
        else { raw = node.text; }
        let cursor = 0;
        while(cursor < raw.length) {
            if(!/\s/.test(raw[cursor] ?? '')) { cursor++; continue; }
            const start = cursor;
            while(cursor < raw.length && /\s/.test(raw[cursor] ?? '')) { cursor++; }
            if(start === 0 || cursor === raw.length || raw.slice(start, cursor) !== ' ') { this.report(index, 'spaceDelimited', ['rel']); }
        }
    }
    jsx(index: number): void {
        const node = this.context.node(index);
        const name = node.children[0] ?? -1;
        if(name < 0 || this.name(name) !== 'rel') { return; }
        const attributes = this.context.parents[index] ?? -1;
        const opening = attributes < 0 ? -1 : this.context.parents[attributes] ?? -1;
        if(opening < 0 || !['JsxOpeningElement','JsxSelfClosingElement'].includes(this.context.node(opening).kind)) { return; }
        const tagIndex = this.context.node(opening).children[0] ?? -1;
        if(tagIndex < 0 || this.context.node(tagIndex).kind !== 'Identifier') { return; }
        const tag = this.context.node(tagIndex).text;
        if(!html.includes(tag)) { return; }
        if(!this.meaningful(tag)) { this.report(name, 'onlyMeaningfulFor', ['rel','<link>, <a>, <area>, <form>']); return; }
        const initializer = node.children[1] ?? -1;
        if(initializer < 0) { this.report(name, 'emptyIsMeaningless', ['rel']); return; }
        const value = this.context.node(initializer);
        if(value.kind === 'StringLiteral') { this.value(initializer, tag); return; }
        if(value.kind !== 'JsxExpression' || value.children.length === 0) { return; }
        const inner = value.children[value.children.length - 1] ?? panic('missing JSX expression');
        const expression = this.context.node(inner);
        if(expression.kind === 'StringLiteral') { this.value(inner, tag); }
        else if(this.literal(inner)) { this.report(inner, 'onlyStrings', ['rel']); }
        else if(expression.kind === 'ObjectLiteralExpression' || (expression.kind === 'Identifier' && expression.text === 'undefined')) { this.report(initializer, 'onlyStrings', ['rel']); }
    }
    prop(index: number, tag: string): void {
        if(!this.literal(index)) { return; }
        if(this.context.node(index).kind === 'StringLiteral') { this.checkToken(index, this.context.node(index).text, tag); }
        else { this.report(index, 'neverValid', [this.literalText(index), 'rel']); }
    }
    call(index: number): void {
        const node = this.context.node(index);
        const callee = this.context.node(node.children[0] ?? panic('missing callee'));
        if(callee.kind !== 'PropertyAccessExpression') { return; }
        const object = this.context.node(callee.children[0] ?? panic('missing object'));
        const method = this.context.node(callee.children[callee.children.length - 1] ?? panic('missing method'));
        if(object.kind !== 'Identifier' || object.text !== 'React' || method.kind !== 'Identifier' || method.text !== 'createElement') { return; }
        if(node.list < 2) { return; }
        const args = node.children.slice(node.children.length - node.list);
        const element = args[0] ?? panic('missing tag');
        if(!this.literal(element)) { return; }
        const tag = this.literalText(element);
        if(this.context.node(element).kind === 'StringLiteral' && !html.includes(tag)) { return; }
        const props = this.context.node(args[1] ?? panic('missing props'));
        if(props.kind !== 'ObjectLiteralExpression') { return; }
        for(const property of props.children) {
            const item = this.context.node(property);
            const name = item.children[0] ?? -1;
            if(name < 0 || this.name(name) !== 'rel') { continue; }
            if(!this.meaningful(tag)) { this.report(name, 'onlyMeaningfulFor', ['rel','<link>, <a>, <area>, <form>']); continue; }
            if(item.kind === 'MethodDeclaration') { this.report(property, 'noMethod', ['rel']); continue; }
            if(item.kind !== 'PropertyAssignment') { continue; }
            const value = item.children[item.children.length - 1] ?? panic('missing initializer');
            if(this.context.node(value).kind === 'ArrayLiteralExpression') {
                for(const child of this.context.node(value).children) { this.prop(child, tag); }
            }
            else { this.prop(value, tag); }
        }
    }
    visit(index: number, parent: number): void {
        const node = this.context.node(index);
        if(node.kind !== 'JsxAttribute' && node.kind !== 'CallExpression') { return; }
        if(this.context.settings.list('attributes', ['rel']).length === 0 || this.context.settings.read('option', 'default') === '[]') { return; }
        if(node.kind === 'JsxAttribute') { this.jsx(index); }
        else if(node.kind === 'CallExpression') { this.call(index); }
    }
}
export function create(context: RuleContext): Rule { return new Rule(context); }
