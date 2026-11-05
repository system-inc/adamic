const xs: (number | undefined)[] = [1, 2];
delete xs[2];
xs.push(9);
console.log(String(xs[2]));
