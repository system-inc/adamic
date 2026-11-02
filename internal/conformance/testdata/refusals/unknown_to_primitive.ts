// Spec: A cast the runtime can't check is a type the compiler would have to trust
// Expect: refused a cast the runtime can't check
const values: unknown[] = ['disk failed', 42];
const first: unknown = values[1];
console.log((first as string).toUpperCase());
