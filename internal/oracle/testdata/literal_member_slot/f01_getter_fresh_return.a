interface Fixed { readonly t: number }
interface W { t: number }
function make(): { p: W } {
    return { get p(): Fixed { return { t: 1 }; } };
}
const x = make();
x.p.t = 5;
console.log(String(x.p.t));
