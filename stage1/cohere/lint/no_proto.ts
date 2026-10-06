// Port of pinned Go cohere's no-proto.
import type { RuleContext } from './rule_context.ts';

const messageNoProto =
    "This reads or writes `__proto__`, which reaches an object's prototype through a deprecated accessor. It is normative only for web compatibility, it is slow because writing it deoptimizes the object's shape, and on an object parsed from untrusted input it is the prototype pollution vector. Use `Object.getPrototypeOf` and `Object.setPrototypeOf`, or `Object.create(null)` for a map that has no prototype to reach.";

export function visit(ctx: RuleContext, index: number): void {
    if(!ctx.enabled('no-proto')) return;
    if(ctx.accessName(index) === '__proto__') ctx.report(index, 'no-proto', 'unexpectedProto', messageNoProto);
}
