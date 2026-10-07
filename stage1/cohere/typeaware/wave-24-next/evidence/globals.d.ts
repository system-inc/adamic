declare namespace NodeJS { interface Process { exit(code?:number):never; stdout:{write(text:string,callback?:()=>void):void}; stderr:{write(text:string,callback?:()=>void):void}; } }
declare var process:NodeJS.Process;
