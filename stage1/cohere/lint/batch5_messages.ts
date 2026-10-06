// Static messages from pinned Go Cohere; no computed verdicts.
import { panic } from 'adamic';
const messages: string[][] = [
    [
        'useTopLevelQualifier',
        "Every name in this import carries its own inline `type` qualifier, so TypeScript erases the names and leaves the bare `import 'module'` behind, which still runs the module for its side effects. The import reads as type-only and is not. Move the qualifier to the top level, where it removes the whole statement.",
    ],
    [
        'interfaceConstruct',
        'This construct signature returns the interface that declares it, which describes a constructor whose instances are more constructors. An interface describing a static class object should have its `new` signature return the instance type being constructed, not the constructor interface itself.',
    ],
    [
        'interfaceConstructor',
        'An interface cannot be constructed, so a member named `constructor` declares an ordinary method that happens to share a name with the class keyword and constructs nothing. Use a `new` construct signature if a constructor is meant, and rename the method otherwise.',
    ],
    [
        'classNew',
        'This method is named `new` and declares that it returns the class it belongs to, which reads as a constructor and is not one. Calling it does not construct anything, so the name misleads every reader; a constructor is spelled `constructor`.',
    ],
    [
        'thisAssignment',
        'This binding holds `this` under another name. The alias exists to carry the receiver into a nested `function`, which rebinds `this` to something else, and an arrow function closes over `this` directly and needs no carrier. Every later reader now has to establish which of the two names is the real receiver and whether they are still the same object. Use an arrow function and write `this`.',
    ],
    [
        'thisDestructure',
        'This pattern pulls members off `this` into local bindings. Each one is read once, at the moment of destructuring, so a later write to the property is invisible here, and a method taken this way has lost its receiver and throws when called. Read through `this` at the point of use instead.',
    ],
    [
        'preferAsConst',
        'This literal type restates a value the compiler can already read off the literal itself. `as const` infers it, so the value and its type cannot drift apart; writing the type out means every edit to the value has to be made twice, and the compiler accepts the version where only one of them was.',
    ],
    [
        'noExtraNonNullAssertion',
        'This non-null assertion is redundant: the value it asserts on has already been asserted non-null, or is about to be checked at runtime by the `?.` that follows it. Either way the second `!` narrows nothing the first one did not, so it is noise that reads as a claim. Remove it.',
    ],
    [
        'noDuplicateEnumValues',
        'Two members of this enum are initialized to the same value. TypeScript permits it, but a reader expects members of one enum to name distinct things, and a reverse lookup by value can only return one of them, so the later member silently wins. The usual cause is a copied line whose value was never changed. Give each member its own value.',
    ],
    [
        'unexpectedAny',
        'This annotation is `any`, which switches the type system off for every value that flows through it: no property is checked, no argument is checked, and no assignment is refused. The compiler stops answering questions about this value and reports nothing when it is used wrongly. Write the type it actually holds, or `unknown` when the shape is genuinely not known yet, which keeps the value opaque until it is narrowed rather than treating it as anything at all.',
    ],
    ['dynamicDelete', 'Do not delete dynamically computed property keys.'],
    [
        'uselessEmptyExport',
        'An `export {}` exists to make a script into a module, and this file is already a module because something else in it imports or exports. The statement therefore changes nothing, and it reads as though it were load-bearing to anyone deciding whether they may delete it. Remove it.',
    ],
];
export function ruleMessage(id: string): string {
    for(const row of messages) {
        if(row[0] === id) {
            return row[1] ?? panic('missing rule message');
        }
    }
    return panic('unknown batch5 message');
}
