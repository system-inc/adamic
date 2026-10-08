// Package-local planning API for the unchanged production pilot. No type judgment.
import { panic } from 'adamic';
import type { RuleContext } from '../context.ts';
import type { Checker } from '../checker.a';
import { Frames } from '../../typeaware/frames.ts';
export class PlannedQuestion {
    readonly index: number;
    readonly question: string;
    constructor(index: number, question: string) { this.index = index; this.question = question; }
}
export class QuestionPlan {
    readonly questions: readonly PlannedQuestion[];
    constructor(questions: readonly PlannedQuestion[]) { this.questions = questions; }
}
export class PilotPlanner {
    readonly context: RuleContext;
    constructor(context: RuleContext) { this.context = context; }
    plan(root: number): QuestionPlan {
        const questions: PlannedQuestion[] = [new PlannedQuestion(root, 'options')];
        this.walk(root, questions);
        return new QuestionPlan(questions);
    }
    walk(index: number, questions: PlannedQuestion[]): void {
        const parser = this.context.parser;
        const node = parser.node(index);
        if(node.kind === 'BinaryExpression') {
            const operator = parser.node(node.children[1] ?? panic('missing operator')).kind;
            if(operator === 'EqualsEqualsToken' || operator === 'EqualsEqualsEqualsToken' || operator === 'ExclamationEqualsToken' || operator === 'ExclamationEqualsEqualsToken') {
                const left = this.context.unwrap(node.children[0] ?? panic('missing left'));
                const right = this.context.unwrap(node.children[2] ?? panic('missing right'));
                const rightKind = parser.node(right).kind;
                const leftKind = parser.node(left).kind;
                if(rightKind === 'TrueKeyword' || rightKind === 'FalseKeyword') { questions.push(new PlannedQuestion(left, 'type-shape')); }
                else if(leftKind === 'TrueKeyword' || leftKind === 'FalseKeyword') { questions.push(new PlannedQuestion(right, 'type-shape')); }
            }
        }
        for(const child of node.children) { this.walk(child, questions); }
    }
}
function frame(value: string): string { return `${value.length}\n${value}`; }
export function validatePlan(plan: QuestionPlan, checker: Checker): void {
    if(plan.questions.length !== checker.recorded.length) { panic('declared question count differs'); }
    for(let index = 0; index < plan.questions.length; index++) {
        const question = plan.questions[index] ?? panic('missing planned question');
        const node = checker.parser.node(question.index);
        const start = checker.offsets[node.pos] ?? panic('planned start outside source');
        const end = checker.offsets[node.end] ?? panic('planned end outside source');
        const expected = frame(checker.path) + frame(`${start}`) + frame(`${end}`) + frame(node.kind) + frame(question.question);
        const frames = new Frames(checker.recorded[index] ?? panic('missing recorded question'));
        if(frames.field() !== expected) { panic('declared question order or selector differs'); }
        frames.field(); frames.field(); frames.end();
    }
}
// Failure controls for the planning contract; never used by the production rule.
export function planMutant(plan: QuestionPlan, kind: string): QuestionPlan {
    const questions: PlannedQuestion[] = [];
    for(let index = 0; index < plan.questions.length; index++) {
        const question = plan.questions[index] ?? panic('missing question');
        if(index === 0 && kind === 'omit') { continue; }
        questions.push(index === 0 && kind === 'wrong' ? new PlannedQuestion(question.index, 'scout-plan-mutant') : question);
    }
    return new QuestionPlan(questions);
}
