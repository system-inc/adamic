// Generated from pinned Cohere policy and Go rule descriptions.
import { panic } from 'adamic';

export function policyMessage(rule: string, id: string, fields: readonly string[]): string {
    let text: string;
    switch(`${rule}/${id}`) {
        case 'nexus/consistency-no-abbreviated-identifier/abbreviatedIdentifier':
            text =
                'Identifier "{{name}}" should not be abbreviated. {{advice}} A name is written once and read everywhere, so the letters saved at the declaration are paid back at every call site by a reader who has to expand the abbreviation themselves and hope they expanded it the way the author meant.';
            break;
        case 'nexus/consistency-no-abbreviated-identifier/abbreviatedSuffix':
            text =
                'Identifier "{{name}}" should not end with "{{suffix}}". {{advice}} A name is written once and read everywhere, so the letters saved at the declaration are paid back at every call site by a reader who has to expand the abbreviation themselves and hope they expanded it the way the author meant.';
            break;
        case 'nexus/consistency-no-abbreviated-identifier/millisecondSuffix':
            text =
                'Identifier "{{name}}" should not abbreviate milliseconds as "Ms". Use "{{suggestion}}", which is how the rest of the tree spells a millisecond value: "durationInMilliseconds" outnumbers "durationMs" more than two to one for the identical value, so the rename follows what the codebase already decided rather than introducing a third spelling.';
            break;
        case 'nexus/consistency-no-abbreviated-identifier/abbreviatedWordSegment':
            text =
                'Identifier "{{name}}" abbreviates "{{word}}". Use "{{suggestion}}". A name is written once and read everywhere, so the letters saved at the declaration are paid back at every call site by a reader who has to expand the abbreviation themselves and hope they expanded it the way the author meant.';
            break;
        case 'nexus/consistency-no-ambiguous-identifier/noAmbiguousE':
            text =
                'Variable named "e" is too ambiguous<<context>>. It is the one name that could be an error or an event, and a reader has to find the declaration to learn which. Use "{{suggestedName}}" or a more descriptive name.';
            break;
        case 'nexus/consistency-no-ambiguous-identifier/noSingleLetter':
            text =
                'Single-letter identifier "{{name}}" is not descriptive enough. The name is read everywhere it is used and declared only once, so the saving is at the declaration and the cost is at every call site.';
            break;
        case 'nexus/consistency-no-ambiguous-identifier/noUnderscore':
            text =
                'Identifier "_" is not descriptive enough. Name it explicitly, or prefix an underscore to a real name such as "_event" when the point is that the value is deliberately unused.';
            break;
        case 'base/consistency-no-console/consistencyNoConsole':
            text =
                "Do not call 'console.{{methodName}}'. Reach the tier that owns this failure and call '.log.error(identifier, data, error)' or '.log.warning(...)' for a row, or '.log.debug(message)' for a line that never becomes one.";
            break;
        case 'nexus/consistency-no-enum/noEnum':
            text =
                "TypeScript enum is banned. Use `as const` with the Kind and KindType pattern instead: `export const FooKind = { A: 'A' } as const; export type FooKindType = (typeof FooKind)[keyof typeof FooKind];`";
            break;
        case 'nexus/consistency-no-long-line-comment/longLineComment':
            text =
                'A run of five or more double-slash lines should be a block comment. Past four lines a stack of slashes stops reading as one thought and starts reading as a wall, with no cue where the passage ends. This run is {{lineCount}} lines.';
            break;
        case 'nexus/consistency-no-multiline-arrow-function/reactHookArrow':
            text =
                'Use a regular function instead of an arrow function with React.{{hookName}}. A named function shows up in a stack trace and in the React devtools as itself.';
            break;
        case 'nexus/consistency-no-multiline-arrow-function/addEventListenerArrow':
            text =
                'Use a regular function instead of an arrow function with addEventListener. A listener you cannot name is a listener you cannot remove, since removeEventListener needs the same reference back.';
            break;
        case 'nexus/consistency-no-multiline-arrow-function/multilineArrow':
            text =
                'Use a regular function instead of a multi-line arrow function. An arrow earns its terseness on a single-line implicit return, and past that a named function declaration reads better and hoists. Arrows that use `this` are exempt, since a function would rebind it.';
            break;
        case 'nexus/consistency-no-shouting/shoutingInComment':
            text =
                '{{tokens}} in a comment reads as shouting, and the register spreads: the next reader mirrors it, so emphasis on everything becomes emphasis on nothing. If it is code, put it in backticks. If it is a point, make it in a normal voice, and if the sentence needed the volume to land, it probably needed the reason instead.';
            break;
        case 'nexus/consistency-no-single-line-jsdoc/useSimpleComment':
            text =
                "A JSDoc comment carrying one line of description should be a line comment. JSDoc's delimiters exist to hold structure, so spending three lines of them on a sentence tells a reader to look for tags that are not there.";
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
        case '@typescript-eslint/no-non-null-assertion/noNonNull':
            text =
                "A non-null assertion tells the compiler to stop checking, and it is erased before the code runs, so when the value is missing anyway the failure lands somewhere else, as a property access on `undefined` in code the types promised was safe. Fix it in this order. If the value can really be missing on a reachable path (input, disk, the network, an external API), guard it: return early, `continue`, or throw an error that says what was missing, because that is the case the type is warning about. If it cannot, fix the type that says it can (a type predicate in `.filter`, a tuple type, a field made required), so there is no claim left to check. If an index is what makes it optional, restructure so the index goes away (`for...of`, `.entries()`, a destructured head checked once), because a value you never index is never possibly missing. Where none of those fit, state the proof with `assert(value !== undefined, 'why it holds')` or, in an expression, `required(value, 'why it holds')` from `@nexus/source/errors/Assert`, which narrows the same way and throws that sentence if the proof is ever wrong. `?.`, `?? ''` and `as T` are not repairs on their own: each makes the missing value disappear instead of handling it.";
            break;
        case '@typescript-eslint/no-non-null-assertion/suggestOptionalChain':
            text =
                "Use the optional chain operator `?.`, but only where a missing value is really possible and the code after it handles `undefined`. It reads `undefined` where `!` would throw, so it changes the expression's type rather than repairing the value.";
            break;
        case 'adamic/no-type-predicate/typePredicate':
            text =
                "A written type predicate is trusted and never checked: tsc narrows the argument on this function's word, and nothing makes its body agree, so a guard that returns true for the wrong value turns every narrowed use into a TypeError. Narrow where you use the value (`if (pet.kind === 'Cat')`), or drop the annotation and let TypeScript infer the predicate, which it does only when the body proves it (`(x) => x !== undefined`).";
            break;
        default:
            return panic(`unknown policy message ${rule}/${id}`);
    }
    for(let index = 0; index < fields.length; index += 2) {
        const name = fields[index] ?? panic('field name');
        const value = fields[index + 1] ?? panic('field value');
        text = text.split(`{{${name}}}`).join(value).split(`<<${name}>>`).join(value);
    }
    return text;
}

export function cohereMessage(name: string): string {
    switch(name) {
        case 'messageOneVarCombineUninitialized':
            return 'This declares uninitialized variables in a second statement where the scope already has one. Reading the declarations of a scope means finding every statement that declares, and one list per scope makes that a single place to look.';
        case 'messageOneVarCombineInitialized':
            return 'This declares initialized variables in a second statement where the scope already has one. Reading the declarations of a scope means finding every statement that declares, and one list per scope makes that a single place to look.';
        case 'messageOneVarSplitUninitialized':
            return 'This declares several uninitialized variables in one statement. One binding per statement makes each declaration its own line to move, comment or delete, and keeps a diff that touches one variable from touching the others.';
        case 'messageOneVarSplitInitialized':
            return 'This declares several initialized variables in one statement. One binding per statement makes each declaration its own line to move, comment or delete, and keeps a diff that touches one variable from touching the others.';
        case 'messageOneVarSplitRequires':
            return "This mixes require calls with other initializers in one declaration. Separating the requires keeps the module's dependencies readable as a block rather than interleaved with ordinary locals.";
        case 'messageOneVarCombine':
            return 'This declares variables in a second statement where the scope already has one. Reading the declarations of a scope means finding every statement that declares, and one list per scope makes that a single place to look.';
        case 'messageOneVarSplit':
            return 'This declares several variables in one statement. One binding per statement makes each declaration its own line to move, comment or delete, and keeps a diff that touches one variable from touching the others.';
        case 'messageMethodSignatureStyleErrorMethod':
            return 'This is written as a shorthand method signature, which TypeScript checks bivariantly: a parameter can be narrowed by an implementer and the compiler will not object, so a call that type-checks can still be wrong at runtime. A function property is checked the strict way and catches that.';
        case 'messageMethodSignatureStyleErrorProperty':
            return 'This project writes interface members as shorthand methods, and this one is a property holding a function type. The two differ in how the compiler checks parameters, so mixing them in one codebase means a reader cannot tell which rule applies without looking at the punctuation.';
        case 'messageMethodSignatureStyleConvertToMethod':
            return 'Convert to a method signature. There is no syntax for a readonly method, so this drops the `readonly` modifier and the member becomes reassignable.';
        case 'preferLiteralEnumMemberNotLiteralMessage':
            return 'Explicit enum value must only be a literal value (string or number).';
        case 'preferLiteralEnumMemberNotLiteralOrBitwiseMessage':
            return 'Explicit enum value must only be a literal value (string or number) or a bitwise expression.';
        case 'messageBannedWrapperObjectType':
            return "This names the boxed object rather than the primitive: `String` is what `new String('x')` produces, and `string` is what `'x'` is. The wrapper accepts the object and behaves differently from the primitive under comparison and truthiness, which is the reverse of what the annotation almost always means. Use the lowercase primitive.";
        case 'messagePreferDestructuring':
            return 'Pulling one name out of an object or array by hand repeats the name on both sides, so a rename touches two places and the two can drift. Destructuring writes it once.';
        case 'messageNoNegatedCondition':
            return 'This condition is negated and has an else branch, so the reader has to invert it mentally to work out which branch runs when. Swapping the two branches and dropping the negation says the same thing without the double take.';
        case 'messageReturnAssignment':
            return 'This return statement assigns rather than returns a value it already has, so it does two things at once and the assignment is the one the reader is likeliest to miss. It is usually a `==` typed as `=`, and where it is deliberate it reads as a value being returned rather than as a write happening on the way out. Assign on its own line and return the variable, or wrap the assignment in its own parentheses to say the write was meant.';
        case 'messageArrowAssignment':
            return "This arrow function's body is an assignment, so calling it writes to something outside itself and returns what it wrote. A concise arrow body reads as a value, which is what makes the write easy to miss at the call site. Give the arrow a block body and assign inside it, or wrap the assignment in its own parentheses to say the write was meant.";
        default:
            return panic(`unknown Go message ${name}`);
    }
}
