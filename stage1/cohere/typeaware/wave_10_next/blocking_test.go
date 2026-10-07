package wave10next_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// These are TypeScript oracle inputs, not Adamic source modules. Keeping each
// entry independent exercises program-wide import classification in one program.
func TestBlockingFixturePreparation(t *testing.T) {
	directory := os.Getenv("ADAMIC_WAVE10_BLOCKING_FIXTURES")
	if directory == "" {
		t.Skip("set ADAMIC_WAVE10_BLOCKING_FIXTURES to prepare source comparison inputs")
	}
	repository, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	constants, _ := processSources(t, repository)
	h := &harness{t: t, repository: repository, directory: directory}
	prelude := constants["correctnessNoProcessExitAfterOutputPrelude"]
	stream := filepath.Join(directory, "libraries/nexus/source/system")
	if err := os.MkdirAll(stream, 0755); err != nil {
		t.Fatal(err)
	}
	h.write("libraries/nexus/source/system/StandardStreams.ts", "export function blockStandardStreams():void{}\n")
	h.write("libraries/nexus/source/system/Reexport.ts", "export {blockStandardStreams as blocking} from './StandardStreams';\n")
	sources := []string{
		"console.log('a');process.exit(0);",
		"blockStandardStreams();console.log('a');process.exit(0);",
		"console.log('a');process.exit(0);blockStandardStreams();",
		"if(flag)blockStandardStreams();console.log('a');process.exit(0);",
		"console.log('a');if(flag)blockStandardStreams();process.exit(0);",
		"console.log('a');await run();process.exit(0);",
		"console.log('a');process.exit(0);await run();",
		"function main(){console.log('a');process.exit(0)}main();",
		"function main(){blockStandardStreams();console.log('a');process.exit(0)}main();",
		"function main(){console.log('a');process.exit(0)}main(blockStandardStreams());",
		"function unused(){console.log('a');process.exit(0)}",
		"const main=()=>{console.log('a');process.exit(0)};main();",
		"let main=()=>{console.log('a');process.exit(0)};main();",
		"function* main(){console.log('a');process.exit(0)}main();",
		"function main(){console.log('a');process.exit(0)}onEvent(main);",
		"function main(){console.log('a');process.exit(0)}unknown(main);",
		"console.log('a');try{process.exit(0)}catch{process.exit(1)}",
		"console.log('a');try{use(1)}catch{process.exit(1)}process.exit(2);",
		"function main(){console.log('a');try{return;}finally{process.exit(0)}}main();",
		"console.log('a');process.exit(0);process.exit(1);",
		"function main(){console.log('a');if(flag)process.exit(0);else process.exit(1)}main();",
		"function main(){return main()}main();console.log('a');process.exit(0);",
		"class A{static{console.log('a');process.exit(0)}}",
		"const main=()=>{console.log('a');process.exit(0)};(main)();",
		"for(;;){console.log('a');break;}process.exit(0);",
		"console.log('a');blocking();process.exit(0);",
		"blockStandardStreams();function main(){console.log('a');process.exit(0)}main();",
		"console.log('a');require('./unknown');process.exit(0);",
	}
	var paths []string
	for i, source := range sources {
		imports := ""
		if i%2 == 1 || strings.Contains(source, "blockStandardStreams") {
			imports = "import {blockStandardStreams} from './libraries/nexus/source/system/StandardStreams';\n"
		}
		if strings.Contains(source, "blocking()") {
			imports += "import {blocking} from './libraries/nexus/source/system/Reexport';\n"
		}
		header := ""
		if i%3 == 0 {
			header = "#!/usr/bin/env tsx\n"
		}
		paths = append(paths, h.write(fmt.Sprintf("blocking-%03d.ts", i), header+prelude+imports+"declare function onEvent(callback:()=>void):void;declare function unknown(...args:any[]):any;declare function require(path:string):unknown;\n"+source+"\nexport {};\n"))
	}
	base, err := os.ReadFile(filepath.Join(directory, "controls.manifest"))
	if err != nil {
		t.Fatal(err)
	}
	h.write("all-controls.manifest", string(base)+strings.Join(paths, "\n")+"\n")
	t.Logf("prepared %d additional blocking controls", len(paths))
}
