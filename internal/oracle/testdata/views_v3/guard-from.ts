function copy(xs:readonly unknown[]):unknown[]{return Array.from(xs);}
console.log(String(copy([7]).length));
