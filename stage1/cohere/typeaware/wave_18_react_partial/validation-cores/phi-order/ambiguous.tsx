declare function createA():any;declare function createB():any;function Component(props:any){let Inner;if(props.c){Inner=createA();}else{Inner=createB();}return <Inner/>;}
