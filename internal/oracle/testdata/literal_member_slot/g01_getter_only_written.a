interface W { t: number }
const w: W = { t: 2 };
function make(): { p: W } {
    return { get p(): W { return w; } };
}
const x = make();
try { x.p = { t: 5 }; console.log('wrote'); } catch (error) { console.log(error instanceof Error ? error.name : 'other'); }
console.log(String(x.p.t));
