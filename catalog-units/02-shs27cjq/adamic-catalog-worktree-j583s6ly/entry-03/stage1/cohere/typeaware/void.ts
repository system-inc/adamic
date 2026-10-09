import type { Rules } from './rules.ts';
import type { Volume } from './volume.ts';
import { Repair } from './repair.ts';

export class VoidRule {
    readonly rules: Rules;
    readonly volume: Volume;
    readonly repaired: number[] = [];
    constructor(rules: Rules, volume: Volume) {
        this.rules = rules;
        this.volume = volume;
    }
    ancestor(index: number): number {
        let current = index;
        while((this.rules.parents[current] ?? -1) >= 0) {
            const parentIndex = this.rules.parents[current] ?? -1;
            const parent = this.rules.parser.node(parentIndex);
            if(parent.kind === 'ExpressionStatement') {
                return -1;
            }
            if(
                parent.kind === 'ParenthesizedExpression' ||
                (parent.kind === 'ConditionalExpression' && parent.children[0] !== current)
            ) {
                current = parentIndex;
                continue;
            }
            if(parent.kind === 'BinaryExpression') {
                const operator = this.rules.parser.node(parent.children[1] ?? -1).kind;
                if(operator === 'CommaToken') {
                    let child = current;
                    let comma = parentIndex;
                    while(
                        comma >= 0 &&
                        this.rules.parser.node(comma).kind === 'BinaryExpression' &&
                        this.rules.parser.node(this.rules.parser.node(comma).children[1] ?? -1).kind === 'CommaToken'
                    ) {
                        if(this.rules.parser.node(comma).children[0] === child) {
                            return -1;
                        }
                        child = comma;
                        comma = this.rules.parents[comma] ?? -1;
                    }
                    return parentIndex;
                }
                if(
                    ['AmpersandAmpersandToken', 'BarBarToken', 'QuestionQuestionToken'].includes(operator) &&
                    parent.children[2] === current
                ) {
                    current = parentIndex;
                    continue;
                }
            }
            return parentIndex;
        }
        return -1;
    }
    gap(start: number, end: number, semicolon = false): boolean {
        for(let at = start; at < end; at++) {
            const code = this.rules.scanner.text.charCodeAt(at);
            if(![32, 9, 10, 13, 11, 12].includes(code) && !(semicolon && code === 59)) {
                return true;
            }
        }
        return false;
    }
    visit(index: number): void {
        const node = this.rules.parser.node(index);
        if(['AwaitExpression', 'CallExpression', 'TaggedTemplateExpression'].includes(node.kind)) {
            const invalid = this.ancestor(index);
            if(invalid >= 0 && (this.volume.facts(index, 'type-shape').root().flags & 16) !== 0) {
                const parent = this.rules.parser.node(invalid);
                if(parent.kind === 'ArrowFunction') {
                    const finding = this.rules.add(
                        'no-confusing-void-expression',
                        'invalidVoidExprArrow',
                        'Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function. The concise body makes the arrow return undefined, which reads as deliberate when it is usually an accident of the syntax.',
                        index,
                    );
                    const wrappedBody = parent.children[parent.children.length - 1] ?? -1;
                    const body = this.rules.skip(wrappedBody);
                    const arrow =
                        parent.children.find(
                            (child) => this.rules.parser.node(child).kind === 'EqualsGreaterThanToken',
                        ) ?? -1;
                    if(
                        arrow >= 0 &&
                        !this.repaired.includes(invalid) &&
                        (this.volume.facts(body, 'type-shape').root().flags & 20) !== 0 &&
                        !this.gap(
                            this.rules.parser.node(arrow).end,
                            this.rules.start(this.rules.parser.node(wrappedBody)),
                        ) &&
                        !this.gap(this.rules.parser.node(wrappedBody).end, parent.end)
                    ) {
                        finding.repairs.push(
                            new Repair(
                                this.rules.byte(this.rules.parser.node(arrow).end),
                                this.rules.byte(this.rules.start(this.rules.parser.node(body))),
                                ' { ',
                            ),
                        );
                        finding.repairs.push(
                            new Repair(
                                this.rules.byte(this.rules.parser.node(body).end),
                                this.rules.byte(parent.end),
                                '; }',
                            ),
                        );
                        this.repaired.push(invalid);
                    }
                }
                else if(parent.kind === 'ReturnStatement') {
                    const blockIndex = this.rules.parents[invalid] ?? -1;
                    const block = this.rules.parser.node(blockIndex);
                    const functionIndex = this.rules.parents[blockIndex] ?? -1;
                    const final =
                        block.kind === 'Block' &&
                        functionIndex >= 0 &&
                        ['ArrowFunction', 'FunctionDeclaration', 'FunctionExpression'].includes(
                            this.rules.parser.node(functionIndex).kind,
                        ) &&
                        block.children[block.children.length - 1] === invalid;
                    const finding = this.rules.add(
                        'no-confusing-void-expression',
                        final ? 'invalidVoidExprReturnLast' : 'invalidVoidExprReturn',
                        final
                            ? 'Returning a void expression from a function is forbidden. Please remove the `return` statement. It is the last statement, so the return adds nothing but the suggestion that a value comes back.'
                            : 'Returning a void expression from a function is forbidden. Please move it before the `return` statement. Returning it says the value matters when it is undefined.',
                        index,
                    );
                    const argument = parent.children[0] ?? -1;
                    const value = this.rules.parser.node(argument);
                    const start = this.rules.start(value);
                    const statementStart = this.rules.start(parent);
                    if(
                        (this.volume.facts(argument, 'type-shape').root().flags & 20) !== 0 &&
                        !this.gap(statementStart + 6, start) &&
                        !this.gap(value.end, parent.end, true)
                    ) {
                        let text = `${this.rules.text(value)};${final ? '' : ' return;'}`;
                        if(['(', '[', '`'].includes(this.rules.scanner.text.slice(start, start + 1))) {
                            text = `;${text}`;
                        }
                        if(!final && block.kind !== 'Block') {
                            text = `{ ${text} }`;
                        }
                        finding.repairs.push(
                            new Repair(this.rules.byte(statementStart), this.rules.byte(parent.end), text),
                        );
                    }
                }
                else {
                    this.rules.add(
                        'no-confusing-void-expression',
                        'invalidVoidExpr',
                        'Placing a void expression inside another expression is forbidden. Move it to its own statement instead. A void expression evaluates to undefined, so whatever this outer expression does with the value is working on undefined rather than on a result.',
                        index,
                    );
                }
            }
        }
        for(const child of node.children) {
            this.visit(child);
        }
    }
    run(): void {
        this.visit(this.rules.parser.nodes.length - 1);
    }
}
