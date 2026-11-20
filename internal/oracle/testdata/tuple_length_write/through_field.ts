const holder: { xs: [number, number] } = { xs: [1, 2] };
holder.xs.length = 2;
console.log(String(holder.xs[0]));
