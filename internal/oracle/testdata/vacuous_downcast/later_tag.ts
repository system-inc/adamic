// kept_check_downcast_refused/later_tag.a as .ts, where the cast stays a checked cast: .ts tests the
// first tag, version, which every Root passes, the read of count is checked, and the program prints what
// Node prints.
interface Root { readonly version: 1; readonly kind: string; }
interface View extends Root { readonly version: 1; readonly kind: 'Viewed'; count: number; }
const raw = { version: 1 as const, kind: 'Viewed', count: 4 };
const base: Root = raw;
const viewed = base as View;
console.log(`viewed ${viewed.count + 1}`);
