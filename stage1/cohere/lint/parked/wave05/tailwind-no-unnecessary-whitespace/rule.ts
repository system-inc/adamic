import { panic } from 'adamic';
import type { RuleContext } from '../../context.ts';
import { Finding } from '../../finding.ts';
import { tidy, edits, space, type Edit } from './whitespace.a';
import { description } from './messages.a';
export class Report {
    readonly finding: Finding;
    readonly edits: Edit[];
    constructor(finding: Finding, edits: Edit[]) { this.finding = finding; this.edits = edits; }
}
export class Segment {
    readonly index: number;
    readonly leading: boolean;
    readonly trailing: boolean;
    constructor(index: number, leading: boolean, trailing: boolean) { this.index = index; this.leading = leading; this.trailing = trailing; }
}
export class Rule {
    readonly context: RuleContext;
    readonly reports: Report[] = [];
    multiline = true;
    attributes = ['class', 'className'];
    callees = ['mergeClassNames', 'createVariantClassNames'];
    patterns = ['.*[Cc]lassName$', '.*[Cc]lassNames$'];
    configured = false;
    constructor(context: RuleContext) { this.context = context; }
    configure(): void {
        if(this.configured) { return; }
        this.configured = true;
        this.multiline = this.context.settings.read('allowmultiline', 'true') !== 'false';
        const attributes = this.context.settings.list('attributes', []);
        const callees = this.context.settings.list('callees', []);
        const patterns = this.context.settings.list('variables', []);
        if(attributes.length > 0) { this.attributes = attributes; }
        if(callees.length > 0) { this.callees = callees; }
        if(patterns.length > 0) { this.patterns = patterns; }
    }
    segment(segment: Segment): void {
        const node = this.context.node(segment.index); const text = node.text;
        if(tidy(text, segment.leading, segment.trailing, this.multiline) === text) { return; }
        const begin = this.context.start(segment.index); let start = begin + 1;
        let end = node.end - (['TemplateHead', 'TemplateMiddle'].includes(node.kind) ? 2 : 1);
        if(start > end) { start = begin; end = node.end; }
        const proposals = edits(this.context.source, start, end, text, segment.leading, segment.trailing, this.multiline);
        const encoded = proposals.map(edit => `${edit.start}:${edit.end}:${edit.text}`).join('|');
        const finding = new Finding('better-tailwindcss/no-unnecessary-whitespace', 'unnecessaryWhitespace', description, start, end, proposals.length > 0 ? 'fix-edits' : '', encoded, '');
        this.context.findings.push(finding); this.reports.push(new Report(finding, proposals));
    }
    values(index: number, leading: boolean, trailing: boolean, plain: Segment[], templates: Segment[]): void {
        const node = this.context.node(index); const children = node.children;
        if(['StringLiteral', 'NoSubstitutionTemplateLiteral'].includes(node.kind)) { plain.push(new Segment(index, leading, trailing)); return; }
        if(['JsxExpression', 'ParenthesizedExpression', 'AsExpression', 'SatisfiesExpression'].includes(node.kind)) { const child = children[0] ?? -1; if(child >= 0) { this.values(child, leading, trailing, plain, templates); } return; }
        if(node.kind === 'ConditionalExpression') {
            for(const child of [children[2] ?? -1, children[4] ?? -1]) { if(child >= 0) { this.values(child, leading, trailing, plain, templates); } } return;
        }
        if(node.kind === 'BinaryExpression') {
            const operator = children[1] ?? -1;
            if(operator < 0) { return; }
            const kind = this.context.node(operator).kind;
            if(['BarBarToken', 'QuestionQuestionToken'].includes(kind)) { this.values(children[0] ?? panic('missing left operand'), leading, trailing, plain, templates); }
            if(['AmpersandAmpersandToken', 'BarBarToken', 'QuestionQuestionToken'].includes(kind)) { this.values(children[2] ?? panic('missing right operand'), leading, trailing, plain, templates); } return;
        }
        if(node.kind === 'ArrayLiteralExpression') { for(const child of children) { this.values(child, leading, trailing, plain, templates); } return; }
        if(node.kind === 'TemplateExpression') {
            templates.push(new Segment(index, leading, trailing));
            let before = this.context.node(children[0] ?? panic('missing template head')).text;
            const spans = children.slice(1);
            for(let offset = 0; offset < spans.length; offset++) {
                const span = this.context.node(spans[offset] ?? panic('missing template span'));
                const expression = span.children[0] ?? panic('missing hole');
                const after = this.context.node(span.children[span.children.length - 1] ?? panic('missing template tail')).text;
                const left = before === '' ? offset === 0 ? leading : true : !space(before[before.length - 1] ?? '');
                const right = after === '' ? offset === spans.length - 1 ? trailing : true : !space(after[0] ?? '');
                this.values(expression, left, right, plain, templates); before = after;
            }
        }
    }
    matchesName(name: string): boolean {
        for(const pattern of this.patterns) {
            if(pattern === '.*[Cc]lassName$') { if(/.*[Cc]lassName$/.test(name)) { return true; } }
            else if(pattern === '.*[Cc]lassNames$') { if(/.*[Cc]lassNames$/.test(name)) { return true; } }
            else { panic('NotYet better-tailwindcss/no-unnecessary-whitespace: stage 0 cannot compile configurable variable regex patterns'); }
        }
        return false;
    }
    visit(index: number, parent: number): void {
        this.configure();
        const node = this.context.node(index); const children = node.children;
        const roots: number[] = [];
        if(node.kind === 'JsxAttribute') {
            const name = children[0] ?? -1;
            if(name >= 0 && this.attributes.includes(this.context.node(name).text) && children.length > 1) { roots.push(children[1] ?? panic('missing JSX initializer')); }
        }
        if(node.kind === 'CallExpression') {
            const callee = children[0] ?? -1;
            if(callee >= 0 && this.context.node(callee).kind === 'Identifier' && this.callees.includes(this.context.node(callee).text)) { for(const argument of children.slice(children.length - node.list)) { roots.push(argument); } }
        }
        if(node.kind === 'VariableDeclaration') {
            const name = children[0] ?? -1;
            if(name >= 0 && this.context.node(name).kind === 'Identifier' && this.matchesName(this.context.node(name).text) && children.length > 1) {
                const last = children[children.length - 1] ?? panic('missing initializer');
                this.context.scanner.pos = this.context.node(children[children.length - 2] ?? panic('missing prior child')).end; this.context.scanner.scan();
                if(this.context.scanner.kind === 'EqualsToken') { roots.push(last); }
            }
        }
        const plain: Segment[] = []; const templates: Segment[] = [];
        for(const root of roots) { this.values(root, false, false, plain, templates); }
        for(const segment of plain) { this.segment(segment); }
        for(const template of templates) {
            const pieces = this.context.node(template.index).children;
            this.segment(new Segment(pieces[0] ?? panic('missing head'), template.leading, true));
            for(let offset = 1; offset < pieces.length; offset++) {
                const span = this.context.node(pieces[offset] ?? panic('missing span'));
                this.segment(new Segment(span.children[span.children.length - 1] ?? panic('missing tail'), true, offset < pieces.length - 1 || template.trailing));
            }
        }
    }
}
export function create(context: RuleContext): Rule { return new Rule(context); }
