import { programArguments } from 'adamic';
const count = programArguments().length + 1;
try {
    console.log('a'.repeat(count));
}
catch(error) {
    if(error instanceof Error) {
        console.log(error.message);
    }
}
