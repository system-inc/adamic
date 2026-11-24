const view: { next(): unknown } = 'a1'.matchAll(/[0-9]/g);
console.log(view.toString());
