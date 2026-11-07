function copy(xs:readonly unknown[]):unknown[]{return xs.slice();}
console.log(String(copy([7]).length));
