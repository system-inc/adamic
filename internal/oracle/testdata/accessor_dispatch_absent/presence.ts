// #4gpbn7w: a presence test of an optional field, in a program where a class implements it with a
// getter, reads through finishAccessors' dispatcher. On a plain literal without the field, the
// dispatcher's fallback must read it as absent, as the read it replaced did, not load a slot the
// literal doesn't have (ASan heap-buffer-overflow on main 2f327160).
interface H {
    readonly m?: () => number;
}
class G implements H {
    get m(): () => number {
        return (): number => 7;
    }
}
function probe(h: H): string {
    return h.m ? 'has' : 'none';
}
console.log(probe(new G()));
console.log(probe({}));
console.log(probe({ m: (): number => 1 }));
