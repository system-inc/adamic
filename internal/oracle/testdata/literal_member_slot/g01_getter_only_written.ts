interface W { t: number }
const w: W = { t: 2 };
function make(): { p: W } {
    return { get p(): W { return w; } };
}
const x = make();
try { x.p = { t: 5 }; console.log('wrote'); } catch (e) { console.log(e instanceof TypeError ? 'TypeError' : 'other'); }
console.log(String(x.p.t));
