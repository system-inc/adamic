import {blockStandardStreams} from './nexus/source/system/StandardStreams';function main(){if(flag){console.log('x');process.exit(0);}blockStandardStreams();console.log('x');process.exit(1);}main();
export {};
