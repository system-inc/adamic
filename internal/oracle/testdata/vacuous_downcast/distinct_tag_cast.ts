// kept_check_downcast_refused/distinct_tag_cast.a as .ts, where the cast stays a checked cast: Viewed
// declares kind 'Viewed', which a Base's kind can fail, the test passes, and the program prints what Node
// prints.
interface Base { readonly kind: 'Holder' | 'Viewed'; }
interface Viewed extends Base { readonly kind: 'Viewed'; count: number; }
const raw = { kind: 'Viewed' as const, count: 4 };
const base: Base = raw;
const viewed = base as Viewed;
console.log(`viewed ${viewed.count + 1}`);
