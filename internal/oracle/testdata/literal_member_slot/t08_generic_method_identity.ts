interface V { t: number }
interface W { t: number }
const v: V = { t: 2 };
function make(): { get(x: V): W } {
    return { get<T>(x: T): T { return x; } };
}
make().get(v).t = 1;
console.log(String(v.t));
