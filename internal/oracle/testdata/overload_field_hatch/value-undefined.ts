interface EvaluatorResult<T> {readonly value:T; syntacticallyString:boolean;}
function run() {
 const shared:EvaluatorResult<string|number|undefined>={value:undefined,syntacticallyString:false};
 let calls=0;
 function evaluate(input:string):EvaluatorResult<string|undefined>;
 function evaluate(input:number):EvaluatorResult<string|number|undefined>;
 function evaluate(input:string|number):EvaluatorResult<string|number|undefined> {calls++; return shared;}
 console.log(`${evaluate('node').value ?? 'missing'}`);
 console.log(`${evaluate('node').value ?? 'missing'}`);
 console.log(`${evaluate(1).value ?? 'missing'}`);
 console.log(`${calls}`);
}
run();
