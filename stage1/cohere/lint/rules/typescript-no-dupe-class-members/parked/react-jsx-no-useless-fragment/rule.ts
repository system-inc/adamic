import { panic } from 'adamic';
import type { RuleContext } from '../../context.ts';
import { Edit, type Suggestion } from './repairs.ts';
const goBlank = /^[\t\n\v\f\r \u0085\u00a0\u1680\u2000-\u200a\u2028\u2029\u202f\u205f\u3000]*$/;
const why = ' A fragment exists to give several siblings a single parent without emitting an element';
const childMessage = 'This fragment is the child of an HTML element, which already accepts any number of children. The fragment groups nothing that the element was not going to accept anyway, so it only adds a layer for a reader to see through.';
export class Rule {
    readonly context: RuleContext;
    readonly allowExpressions: boolean = false;
    constructor(context: RuleContext) { this.context = context; this.allowExpressions = context.settings.read('allowexpressions', 'false') === 'true'; }
    suggestions(index: number): Suggestion[] { return []; }
    opening(index: number): number {
        const node = this.context.node(index);
        return node.kind === 'JsxElement' ? node.children[0] ?? panic('missing opener') : index;
    }
    tag(index: number): number { return this.context.node(this.opening(index)).children[0] ?? -1; }
    fragmentName(index: number): boolean {
        if(index < 0) { return false; }
        const node = this.context.node(index);
        if(node.kind === 'Identifier') { return node.text === 'Fragment'; }
        if(node.kind !== 'PropertyAccessExpression') { return false; }
        const object = this.context.node(node.children[0] ?? panic('missing object'));
        const name = this.context.node(node.children[node.children.length - 1] ?? panic('missing property'));
        return object.kind === 'Identifier' && object.text === 'React' && name.kind === 'Identifier' && name.text === 'Fragment';
    }
    attributes(index: number): number[] {
        const node = this.context.node(this.opening(index));
        if(!['JsxOpeningElement', 'JsxSelfClosingElement'].includes(node.kind)) { return []; }
        const attributes = node.children[node.children.length - 1] ?? -1;
        return attributes >= 0 && this.context.node(attributes).kind === 'JsxAttributes' ? this.context.node(attributes).children : [];
    }
    children(index: number): number[] {
        const node = this.context.node(index);
        return node.kind === 'JsxSelfClosingElement' ? [] : node.children.slice(1, -1);
    }
    kept(index: number): number[] {
        return this.children(index).filter(child => {
            const node = this.context.node(child);
            return !(node.kind === 'JsxText' && goBlank.test(node.text) && node.text.includes('\n'));
        });
    }
    html(parent: number): boolean {
        if(parent < 0 || this.context.node(parent).kind !== 'JsxElement') { return false; }
        const tag = this.tag(parent);
        return tag >= 0 && this.context.node(tag).kind === 'Identifier' && /^[a-z]*$/.test(this.context.node(tag).text);
    }
    ordinaryCall(index: number): boolean {
        const child = this.context.node(index);
        if(child.kind !== 'JsxExpression' || child.children.length === 0) { return false; }
        let cursor = child.children[child.children.length - 1] ?? -1;
        if(cursor < 0 || this.context.node(cursor).kind !== 'CallExpression') { return false; }
        while(cursor >= 0) {
            const node = this.context.node(cursor);
            if(['CallExpression','PropertyAccessExpression','ElementAccessExpression'].includes(node.kind)) {
                if(node.children.some(part => this.context.node(part).kind === 'QuestionDotToken')) { return false; }
                cursor = node.children[0] ?? -1;
            }
            else if(['NonNullExpression','ParenthesizedExpression'].includes(node.kind)) { cursor = node.children[0] ?? -1; }
            else { return true; }
        }
        return true;
    }
    canFix(index: number): boolean {
        for(const attribute of this.attributes(index)) {
            const node = this.context.node(attribute);
            if(node.kind === 'JsxSpreadAttribute') { return false; }
            if(node.kind === 'JsxAttribute') {
                const name = this.context.node(node.children[0] ?? panic('missing attribute name'));
                if(name.kind !== 'Identifier' || name.text !== 'key') { return false; }
            }
        }
        const parent = this.context.parents[index] ?? -1;
        const isJsx = parent >= 0 && ['JsxElement','JsxFragment'].includes(this.context.node(parent).kind);
        const children = this.children(index);
        if(!isJsx && (children.length === 0 || children.some(child => {
            const node = this.context.node(child);
            return node.kind === 'JsxExpression' || (node.kind === 'JsxText' && !goBlank.test(node.text));
        }))) { return false; }
        return !(parent >= 0 && this.context.node(parent).kind === 'JsxElement' && !this.html(parent) && !this.fragmentName(this.tag(parent)));
    }
    fixes(index: number): Edit[] {
        if(!this.canFix(index)) { return []; }
        const node = this.context.node(index);
        let text = '';
        if(node.kind !== 'JsxSelfClosingElement') {
            const opening = this.context.node(node.children[0] ?? panic('missing opening'));
            const closing = this.context.node(node.children[node.children.length - 1] ?? panic('missing closing'));
            text = this.context.source.slice(opening.end, closing.pos);
            let leading = 0; let trailing = text.length;
            while(leading < text.length && /[ \t\n\r\v\f]/.test(text[leading] ?? '')) { leading++; }
            while(trailing > leading && /[ \t\n\r\v\f]/.test(text[trailing - 1] ?? '')) { trailing--; }
            const start = text.slice(0, leading).includes('\n') ? leading : 0;
            const end = text.slice(trailing).includes('\n') ? trailing : text.length;
            text = start > end ? '' : text.slice(start, end);
        }
        return [new Edit(this.context.start(index), node.end, text)];
    }
    report(index: number, id: string, message: string): void {
        const edits = this.fixes(index);
        this.context.report(index, 'react/jsx-no-useless-fragment', id, message, edits.length > 0 ? 'fix' : '', edits[0]?.text ?? '', '');
    }
    visit(index: number, parent: number): void {
        const node = this.context.node(index);
        if(node.kind !== 'JsxFragment' && (!['JsxElement','JsxSelfClosingElement'].includes(node.kind) || !this.fragmentName(this.tag(index)))) { return; }
        for(const attribute of this.attributes(index)) {
            const item = this.context.node(attribute);
            const name = item.children[0] ?? -1;
            if(item.kind === 'JsxAttribute' && name >= 0 && this.context.node(name).kind === 'Identifier' && this.context.node(name).text === 'key') { return; }
        }
        const children = this.children(index); const kept = this.kept(index);
        const textOnlyOutside = children.length === 1 && this.context.node(children[0] ?? panic('missing child')).kind === 'JsxText' && !(parent >= 0 && ['JsxElement','JsxFragment'].includes(this.context.node(parent).kind));
        const expressionAllowed = this.allowExpressions && kept.length === 1 && this.context.node(kept[0] ?? panic('missing child')).kind === 'JsxExpression';
        if(kept.length < 2 && !(kept.length === 1 && this.ordinaryCall(kept[0] ?? panic('missing child'))) && !textOnlyOutside && !expressionAllowed) {
            let message = 'This fragment wraps one thing, so it does nothing.' + why + '; around a single child it is pure noise in the tree and in the source. Write the child on its own.';
            const only = kept[0] ?? -1;
            if(kept.length === 0 || (only >= 0 && this.context.node(only).kind === 'JsxExpression' && this.context.node(only).children.length === 0)) {
                message = 'This fragment is empty, so it renders nothing.' + why + ', and here there is nothing to group. Delete it; where a value is still required, write `null`.';
                if(only >= 0) {
                    const item = this.context.node(only); const raw = this.context.source.slice(item.pos, item.end);
                    if(raw.includes('/*') || raw.includes('//')) { message = 'This fragment holds only a comment, which renders nothing, so the fragment renders nothing either.' + why + ', and here there is nothing to group. Delete it, and the comment with it if that is dead code; where a value is still required, write `null`.'; }
                }
            }
            this.report(index, 'NeedsMoreChildren', message);
        }
        if(this.html(parent)) { this.report(index, 'ChildOfHtmlElement', childMessage); }
    }
}
export function create(context: RuleContext): Rule { return new Rule(context); }
