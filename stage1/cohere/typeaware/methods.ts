import type { Rules } from './rules.ts';
import { Diagnostic } from './diagnostic.ts';
import { Repair } from './repair.ts';

export class Methods {
    readonly rules: Rules;
    constructor(rules: Rules) {
        this.rules = rules;
    }
    name(index: number): number {
        return this.rules.parser.node(index).children.find((child) =>
            ['Identifier', 'StringLiteral', 'NumericLiteral', 'ComputedPropertyName', 'PrivateIdentifier'].includes(
                this.rules.parser.node(child).kind)) ?? -1;
    }
    type(index: number): number {
        const node = this.rules.parser.node(index);
        const last = node.children[node.children.length - 1] ?? -1;
        return last < 0 || last === this.name(index) || ['Parameter', 'TypeParameter', 'QuestionToken',
            'ReadonlyKeyword'].includes(this.rules.parser.node(last).kind) ? -1 : last;
    }
    contains(index: number, kind: string): boolean {
        const node = this.rules.parser.node(index);
        return node.kind === kind || node.children.some((child) => this.contains(child, kind));
    }
    key(index: number): string {
        const name = this.name(index);
        return name < 0 ? '' : this.rules.text(this.rules.parser.node(name)) +
            (this.rules.parser.node(index).children.some((child) =>
                this.rules.parser.node(child).kind === 'QuestionToken') ? '?' : '');
    }
    parameters(index: number): string {
        const node = this.rules.parser.node(index);
        const source = this.rules.scanner.text;
        const floor = this.rules.start(node);
        let result = '()';
        for(const kind of ['Parameter', 'TypeParameter']) {
            const children = node.children.filter((child) => this.rules.parser.node(child).kind === kind);
            if(children.length === 0) {
                continue;
            }
            const first = this.rules.parser.node(children[0] ?? -1);
            const last = this.rules.parser.node(children[children.length - 1] ?? -1);
            const opening = source.slice(0, this.rules.start(first)).lastIndexOf(kind === 'Parameter' ? '(' : '<');
            const closing = source.indexOf(kind === 'Parameter' ? ')' : '>', last.end);
            if(opening < floor || closing < 0 || closing >= node.end) {
                return '';
            }
            const text = source.slice(opening, closing + 1);
            result = kind === 'Parameter' ? text : text + result;
        }
        return result;
    }
    delimiter(index: number): string {
        const last = this.rules.scanner.text.slice(this.rules.parser.node(index).end - 1, this.rules.parser.node(index).end);
        return last === ';' || last === ',' ? last : '';
    }
    visit(index: number): void {
        const parser = this.rules.parser;
        const node = parser.node(index);
        const key = this.key(index);
        if(key === '') {
            return;
        }
        const finding = new Diagnostic('method-signature-style', 'errorMethod',
            'This is written as a shorthand method signature, which TypeScript checks bivariantly: a parameter can be narrowed by an implementer and the compiler will not object, so a call that type-checks can still be wrong at runtime. A function property is checked the strict way and catches that.',
            this.rules.byte(this.rules.start(node)), this.rules.byte(node.end));
        this.rules.findings.push(finding);
        let parent = this.rules.parents[index] ?? -1;
        while(parent >= 0) {
            if(parser.node(parent).kind === 'ModuleDeclaration') {
                return;
            }
            parent = this.rules.parents[parent] ?? -1;
        }
        const type = this.type(index);
        if(type >= 0 && this.contains(type, 'ThisType')) {
            return;
        }
        const owner = this.rules.parents[index] ?? -1;
        const group = owner < 0 ? [index] : parser.node(owner).children.filter((child) =>
            parser.node(child).kind === 'MethodSignature' && this.key(child) === key);
        if(group[0] !== index) {
            return;
        }
        const parts: string[] = [];
        for(const member of group) {
            const parameters = this.parameters(member);
            if(parameters === '') {
                return;
            }
            const annotation = this.type(member);
            const text = `${parameters} => ${annotation < 0 ? 'any' : this.rules.text(parser.node(annotation))}`;
            parts.push(group.length > 1 ? `(${text})` : text);
        }
        const last = group[group.length - 1] ?? index;
        finding.repairs.push(new Repair(finding.start, this.rules.byte(parser.node(last).end),
            `${key}: ${parts.join(' & ')}${this.delimiter(last)}`));
    }
    run(): void {
        for(let index = 0; index < this.rules.parser.nodes.length; index++) {
            if((this.rules.parents[index] ?? -1) >= 0 && this.rules.parser.node(index).kind === 'MethodSignature') {
                this.visit(index);
            }
        }
    }
}
