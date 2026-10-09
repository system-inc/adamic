interface Result<T> { readonly value:T; }
function make(value:string|number|undefined) {
 function evaluate(input:string):Result<string|undefined>;
 function evaluate(input:string|number):Result<string|number|undefined>;
 function evaluate(input:string|number):Result<string|number|undefined> { return {value}; }
 return evaluate;
}
const evaluate=make('word');
const narrow:(input:string)=>Result<string|undefined>=evaluate;
console.log(`${narrow===evaluate}`);
console.log(`${narrow('text').value}`);
function invoke(visitor:(input:string)=>Result<string|undefined>) {
 return visitor('text').value;
}
console.log(`${invoke(narrow)}`);
