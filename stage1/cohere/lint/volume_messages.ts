// Pinned Cohere policy text, with only TypeScript term substitutions.
import { panic } from 'adamic';

export function policyMessage(rule: string, id: string, fields: readonly string[]): string {
    let text: string;
    switch(`${rule}/${id}`) {
        case 'base/consistency-no-console/consistencyNoConsole':
            text =
                "Do not call 'console.{{methodName}}'. Reach the tier that owns this failure and call '.log.error(identifier, data, error)' or '.log.warning(...)' for a row, or '.log.debug(message)' for a line that never becomes one.";
            break;
        case 'nexus/consistency-require-type-suffix/noTypeAliasSuffix':
            text =
                'Type alias "{{name}}" should end in "Type", "Properties", "Interface", or "Options". Rename to "{{name}}Type", "{{name}}Properties" if it shapes React component props, "{{name}}Options" if it is an options or config bag, or convert it to an interface and use "{{name}}Interface".';
            break;
        case 'nexus/consistency-require-type-suffix/noInterfaceSuffix':
            text =
                'Interface "{{name}}" should end in "Interface", "Properties", or "Options". Rename to "{{name}}Interface", "{{name}}Properties" if it shapes React component props, or "{{name}}Options" if it is an options or config bag.';
            break;
        case 'nexus/consistency-require-type-suffix/noConstEnumSuffix':
            text =
                'Const enum-shaped object "{{name}}" should end in "Kind". Rename to "{{name}}Kind" with a paired "{{name}}KindType" alias, so the runtime value and the type that indexes it are visibly one thing.';
            break;
        case 'adamic/no-type-predicate/typePredicate':
            text =
                "A written type predicate is trusted and never checked: tsc narrows the argument on this function's word, and nothing makes its body agree, so a guard that returns true for the wrong value turns every narrowed use into a TypeError. Narrow where you use the value (`if (pet.kind === 'Cat')`), or drop the annotation and let TypeScript infer the predicate, which it does only when the body proves it (`(x) => x !== undefined`).";
            break;
        default:
            return panic(`unknown policy message ${rule}/${id}`);
    }
    for(let index = 0; index < fields.length; index += 2) {
        text = text
            .split(`{{${fields[index] ?? panic('field name')}}}`)
            .join(fields[index + 1] ?? panic('field value'));
    }
    return text;
}
