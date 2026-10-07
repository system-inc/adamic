import { panic } from 'adamic';
import type { RuleContext } from '../../context.ts';
import { Finding } from '../../finding.ts';
import { description } from './messages.a';

export class Frame {
    readonly index: number;
    count = 1;
    constructor(index: number) { this.index = index; }
}
export class Rule {
    readonly context: RuleContext;
    readonly frames: Frame[] = [];
    readonly maximum: number;
    readonly modified: boolean;
    constructor(context: RuleContext) {
        this.context = context;
        this.maximum = Number(context.settings.read('maximum', '20'));
        this.modified = context.settings.read('variant', 'Classic') === 'Modified';
    }
    initializer(index: number): number {
        const children = this.context.node(index).children;
        if(children.length < 2) { return -1; }
        const last = children[children.length - 1] ?? panic('missing child');
        const previous = children[children.length - 2] ?? panic('missing previous child');
        this.context.scanner.pos = this.context.node(previous).end;
        this.context.scanner.scan();
        return this.context.scanner.kind === 'EqualsToken' ? last : -1;
    }
    nameNode(index: number): number {
        const node = this.context.node(index);
        if(['ArrowFunction', 'Constructor', 'ClassStaticBlockDeclaration'].includes(node.kind)) { return -1; }
        for(const child of node.children) {
            const kind = this.context.node(child).kind;
            if(kind.endsWith('Keyword') || ['AsteriskToken', 'Decorator'].includes(kind)) { continue; }
            return ['Identifier', 'PrivateIdentifier', 'StringLiteral', 'NumericLiteral', 'ComputedPropertyName', 'NoSubstitutionTemplateLiteral'].includes(kind) ? child : -1;
        }
        return -1;
    }
    owner(index: number): number {
        if(!['ArrowFunction', 'FunctionExpression'].includes(this.context.node(index).kind)) { return -1; }
        const parent = this.context.parents[index] ?? -1;
        if(parent < 0) { return -1; }
        const node = this.context.node(parent);
        if(node.kind === 'PropertyAssignment' && node.children[node.children.length - 1] === index) { return parent; }
        if(node.kind === 'PropertyDeclaration' && this.initializer(parent) === index) { return parent; }
        return -1;
    }
    readable(index: number): string {
        if(index < 0) { return ''; }
        const node = this.context.node(index);
        if(node.kind === 'ComputedPropertyName') {
            const expression = node.children[0] ?? -1;
            if(expression < 0 || !['StringLiteral', 'NumericLiteral', 'NoSubstitutionTemplateLiteral'].includes(this.context.node(expression).kind)) { return ''; }
            return this.context.node(expression).text;
        }
        return node.text;
    }
    name(index: number): string {
        const node = this.context.node(index);
        if(node.kind === 'ClassStaticBlockDeclaration') { return 'Class static block'; }
        if(node.kind === 'PropertyDeclaration') { return 'Class field initializer'; }
        if(node.kind === 'Constructor') { return 'Constructor'; }
        const kinds = node.children.map(child => this.context.node(child).kind);
        const own = this.nameNode(index);
        const tokens: string[] = [];
        if(kinds.includes('StaticKeyword')) { tokens.push('static'); }
        const private_ = own >= 0 && this.context.node(own).kind === 'PrivateIdentifier';
        if(private_) { tokens.push('private'); }
        if(kinds.includes('AsyncKeyword')) { tokens.push('async'); }
        if(kinds.includes('AsteriskToken') && ['FunctionDeclaration', 'FunctionExpression', 'MethodDeclaration'].includes(node.kind)) { tokens.push('generator'); }
        const owner = this.owner(index);
        const kind = owner >= 0 ? 'method' : node.kind === 'GetAccessor' ? 'getter' : node.kind === 'SetAccessor' ? 'setter' : node.kind === 'MethodDeclaration' ? 'method' : node.kind === 'ArrowFunction' ? 'arrow function' : 'function';
        tokens.push(kind);
        let name = owner >= 0 ? this.readable(this.nameNode(owner)) : '';
        if(name === '') { name = this.readable(own); }
        if(private_) { tokens.push(this.context.node(own).text); }
        else if(name !== '') { tokens.push("'" + name + "'"); }
        const rendered = tokens.join(' ');
        return rendered.slice(0, 1).toUpperCase() + rendered.slice(1);
    }
    report(frame: Frame): void {
        if(frame.count <= this.maximum) { return; }
        const node = this.context.node(frame.index);
        let start = this.context.start(frame.index);
        let end = node.end;
        const owner = this.owner(frame.index);
        if(node.kind === 'PropertyDeclaration') {
            const initializer = this.initializer(frame.index);
            if(initializer >= 0) { start = this.context.start(initializer); end = this.context.node(initializer).end; }
        }
        else if(node.kind === 'ClassStaticBlockDeclaration') { this.context.scanner.pos = node.pos; this.context.scanner.scan(); start = this.context.scanner.start; end = this.context.scanner.pos; }
        else if(node.kind === 'ArrowFunction' && owner < 0) {
            const arrow = node.children.find(child => this.context.node(child).kind === 'EqualsGreaterThanToken');
            if(arrow !== undefined) { start = this.context.start(arrow); end = this.context.node(arrow).end; }
        }
        else {
            const anchor = owner >= 0 ? owner : frame.index;
            const anchorStart = this.context.start(anchor);
            const last = node.children[node.children.length - 1] ?? -1;
            const limit = last >= 0 && (this.context.node(last).kind === 'Block' || node.kind === 'ArrowFunction') ? this.context.node(last).pos : node.end;
            let depth = 0;
            let opening = -1;
            for(let cursor = node.pos; cursor < limit; cursor++) {
                const character = this.context.source[cursor] ?? '';
                if(character === '<') { depth++; }
                if(character === '>' && depth > 0) { depth--; }
                if(character === '(' && depth === 0) { opening = cursor; break; }
            }
            if(opening > anchorStart) { start = anchorStart; end = opening; }
        }
        this.context.findings.push(new Finding('complexity', 'complex', description(this.name(frame.index), frame.count, this.maximum), start, end, '', '', ''));
    }
    walk(index: number): void {
        const node = this.context.node(index);
        const initializer = node.kind === 'PropertyDeclaration' ? this.initializer(index) : -1;
        if(initializer >= 0) {
            for(const child of node.children) { if(child !== initializer) { this.walk(child); } }
            this.frames.push(new Frame(index)); this.walk(initializer);
            this.report(this.frames.pop() ?? panic('missing field frame')); return;
        }
        const starts = ['FunctionDeclaration', 'FunctionExpression', 'ArrowFunction', 'MethodDeclaration', 'GetAccessor', 'SetAccessor', 'Constructor', 'ClassStaticBlockDeclaration'].includes(node.kind);
        if(starts) { this.frames.push(new Frame(index)); }
        else if(this.frames.length > 0) {
            let counts = ['CatchClause', 'ConditionalExpression', 'ForStatement', 'ForInStatement', 'ForOfStatement', 'IfStatement', 'WhileStatement', 'DoStatement'].includes(node.kind);
            if(node.kind === 'BinaryExpression') { counts = ['AmpersandAmpersandToken', 'BarBarToken', 'QuestionQuestionToken', 'AmpersandAmpersandEqualsToken', 'BarBarEqualsToken', 'QuestionQuestionEqualsToken'].includes(this.context.node(node.children[1] ?? panic('missing binary operator')).kind); }
            if(['Parameter', 'BindingElement'].includes(node.kind)) { counts = this.initializer(index) >= 0; }
            if(['PropertyAccessExpression', 'ElementAccessExpression', 'CallExpression'].includes(node.kind)) { counts = node.children.some(child => this.context.node(child).kind === 'QuestionDotToken'); }
            if(node.kind === 'CaseClause') { counts = !this.modified; }
            if(node.kind === 'SwitchStatement') { counts = this.modified; }
            if(counts) { (this.frames[this.frames.length - 1] ?? panic('missing frame')).count++; }
        }
        for(const child of node.children) { this.walk(child); }
        if(starts) { this.report(this.frames.pop() ?? panic('missing function frame')); }
    }
    visit(index: number, parent: number): void { this.walk(index); }
}
export function create(context: RuleContext): Rule { return new Rule(context); }
