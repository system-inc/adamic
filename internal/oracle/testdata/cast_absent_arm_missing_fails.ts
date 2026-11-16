// A missing optional slot under exactOptionalPropertyTypes, beside an explicit undefined (Language's v17): the
// slot that's missing reads undefined at run time as the checker says, so the cast fails its absent check with
// the cast's message. Node runs on without a check and prints undefined.
enum Kind { A = 1, B = 2 }
type A = { readonly kind: Kind.A; readonly n: number };
type B = { readonly kind: Kind.B; readonly n: number };
interface Holder { readonly d?: A | B; readonly e: A | B | undefined }
function cast(holder: Holder): A { return holder.d as A; }
const holder: Holder = { e: undefined };
console.log('casting');
console.log(typeof cast(holder));
