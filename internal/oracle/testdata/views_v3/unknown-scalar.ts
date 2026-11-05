function first(xs:readonly unknown[]):unknown {return xs[0];}
const xs=[7];
console.log(typeof first(xs));
