// Message literals from pinned Go cohere and its TypeScript policy terms.
export const texts = new Map<string, string>([
    [
        'default_case_last',
        'This `default` clause is not the last one in its `switch`. A reader scanning the clauses in order expects the fallback at the bottom, and a `default` sitting in the middle also falls through into whatever case follows it when its body has no `break`, which is a control flow almost nobody writes on purpose. Move it to the end.',
    ],
    [
        'default_param_last',
        'A parameter carrying a default or an optional marker sits before a parameter without one, so the default can never be taken: reaching the later parameter obliges every caller to pass something here. Move it after the required parameters, where omitting it is what selects the default.',
    ],
    [
        'for_direction',
        'The update clause of this loop moves the counter away from the condition, so the condition can never become false and the loop runs forever. Nothing about this fails to compile and nothing throws: the program simply stops making progress, which is why it is usually found by a hung process rather than by reading the loop.',
    ],
    [
        'guard_for_in',
        "This `for-in` loop visits inherited properties as well as the object's own. `for-in` walks the whole prototype chain, so anything added to `Object.prototype` by a library, a polyfill, or older code shows up as another iteration, and the body runs on a key the object never declared. The usual symptom is a serializer emitting a phantom field or a counter that is one too high, on a machine where some other package loaded first. Guard the body with `if (Object.prototype.hasOwnProperty.call(o, key))`, or use `Object.keys(o)`, which returns own enumerable keys and needs no guard.",
    ],
    [
        'max_classes_per_file',
        'File has too many classes (%d). Maximum allowed is %d. A file holding several classes makes the reader hold several vocabularies at once, and the import that names the file no longer says which one it wanted. Split them into a file each, named for the class.',
    ],
    [
        'max_depth',
        'Each level of nesting is another condition the reader has to hold in their head to know whether this line runs at all. Lift a branch into a named function, return early to flatten the happy path, or invert a condition so the exceptional case exits first.',
    ],
    [
        'max_lines',
        'File has too many lines (%d). Maximum allowed is %d. A file this long holds more than one concern, and a reader has to hold all of them to change any. Split it along a real seam, a concern, a table, a command group, never by line count, and have every caller import the new home directly.',
    ],
    [
        'max_nested_callbacks',
        'Too many nested callbacks (%d). Maximum allowed is %d. Each layer of callback moves the code that runs later further from the code that decided to run it, so by this depth the order of execution can no longer be read off the page. Name the inner functions and call them by name, or move to promises.',
    ],
    [
        'grouped_accessor_pairs',
        "Accessor pair is in the wrong order for this project's convention. The pair is grouped, which is the substance; this is the ordering the configuration asks for on top of that.",
    ],
    [
        'ts_default_param_last',
        'A parameter with a default sits before a parameter without one, so the default can never be taken without passing `undefined` explicitly at every call site. Move it after the required parameters, where omitting it is what selects the default.',
    ],
    [
        'ts_ban_tslint_comment',
        'This file carries the tslint directive %s. tslint has been deprecated since 2019 and nothing in this project reads its directives, so the comment suppresses nothing and instead reads as a live suppression to anyone who finds it. Delete it, and if the code it guarded still warrants a suppression, write the equivalent eslint-disable comment naming the rule.',
    ],
    ['ts_init_declarations', "Variable '"],
    ['base_boundary_no_global_container:boundaryNoGlobalContainer', "Usage of 'getGlobalContainer' is not allowed."],
    [
        'base_consistency_no_hand_built_declared_error:consistencyNoHandBuiltDeclaredError',
        "Do not build a BaseError that names a declared failure. Call the tier that declares it: 'AccountModule.error(identifier, data, cause)', 'ApiWorker.error(...)', or 'Base.error(...)'. The status comes from the declaration, so writing it here is how one failure ends up answering two.",
    ],
    [
        'base_consistency_require_pagination_argument_name:invalidName',
        'A GraphQlArgument of PaginationInput type must be named "pagination", or end with "Pagination" when a query takes more than one paginator and the names have to be told apart. A caller reads the argument name to know which paginator it is driving, and a name outside that convention makes them guess.',
    ],
    [
        'base_correctness_require_orm_column_declare:missingDeclare',
        "Property '{{propertyName}}' decorated with @{{decoratorName}}() must use 'declare' (e.g. `declare {{propertyName}}: ...`)",
    ],
    [
        'nexus_consistency_no_for_in:forIn',
        'Iterate an object with `for (const [key, value] of Object.entries(object))`, or `Object.keys(object)` when only the key matters, rather than `for...in`. `for...in` walks the prototype chain, so it needs a guard to be correct, and a guard is a line someone forgets; it also hands back every key as a bare `string`, which is why these loops grow `key as keyof typeof object` casts to read the value. `Object.entries` returns own keys only, with the value already in hand, so there is nothing to guard and nothing to cast, and every loop over an object has one shape.',
    ],
    [
        'nexus_consistency_no_utils_folder:noUnderscoreUtils',
        'Folder name "_utils" is not allowed. Use "_utilities" instead.',
    ],
    ['nexus_consistency_no_utils_folder:noUtils', 'Folder name "utils" is not allowed. Use "utilities" instead.'],
    [
        'nexus_consistency_no_screaming_snake_case:noScreamingSnakeCaseExported',
        'Constant "{{name}}" is exported and should be PascalCase ("{{suggestion}}"). `const` already tells the reader and the compiler the value is immutable, so shouting adds no information. The casing is free to signal scope instead, the way Go does it: PascalCase means the value came from somewhere else, camelCase means it lives in this file. Reserve the shouting form for values that mirror an external system\'s grammar, where matching the upstream spelling keeps the value greppable across a boundary you do not control.',
    ],
    [
        'nexus_consistency_no_screaming_snake_case:noScreamingSnakeCaseLocal',
        'Constant "{{name}}" is file-local and should be camelCase ("{{suggestion}}"). `const` already tells the reader and the compiler the value is immutable, so shouting adds no information. The casing is free to signal scope instead, the way Go does it: PascalCase means the value came from somewhere else, camelCase means it lives in this file. Reserve the shouting form for values that mirror an external system\'s grammar, where matching the upstream spelling keeps the value greppable across a boundary you do not control.',
    ],
    [
        'nexus_consistency_no_stuttering_name:stutteringName',
        '"{{name}}.{{name}}" stutters, which means the name is carrying nothing: it repeats the field instead of saying which {{name}} this is. Rename the value for what it holds, the type it came back as or whatever distinguishes it from another {{name}} in this scope, so a reader forty lines down does not have to find the declaration.',
    ],
]);
