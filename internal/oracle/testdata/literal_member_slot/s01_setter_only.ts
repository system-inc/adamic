interface W { t: number }
let seen = 0;
function make(): { p: W } {
    return { set p(v: W) { seen = v.t; } };
}
const o = make();
o.p = { t: 3 };
console.log(String(seen));
