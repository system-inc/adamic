function copy(xs:readonly unknown[]):unknown[]{return [...xs];}
console.log(String(copy([7]).length));
