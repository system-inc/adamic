import {blockStandardStreams} from './nexus/source/system/StandardStreams';async function main(){await run();console.log('x');process.exit(0);}main();blockStandardStreams();
export {};
