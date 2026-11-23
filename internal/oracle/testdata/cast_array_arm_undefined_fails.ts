// A lying undefined: the cast's target holds no undefined, and undefined is no array, so the cast fails with its
// own message. Node runs on without a check and prints undefined.
interface Program { readonly rootNames: readonly string[] }
interface BuilderProgram { readonly program: Program }
function cast(source: readonly string[] | Program | BuilderProgram | undefined): readonly string[] { return source as readonly string[]; }
console.log('casting');
console.log(typeof cast(undefined));
