package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/system-inc/adamic/internal/worker"
)

func workerCommand(arguments []string) int {
	entry, output := arguments[0], ""
	var names []string
	for index := 1; index < len(arguments); index++ {
		flag := arguments[index]
		index++
		if index >= len(arguments) {
			fmt.Fprintln(os.Stderr, usage)
			return 2
		}
		switch flag {
		case "--out":
			if output != "" {
				fmt.Fprintln(os.Stderr, usage)
				return 2
			}
			output = arguments[index]
		case "--wasm":
			for _, name := range strings.Split(arguments[index], ",") {
				if name == "" {
					fmt.Fprintln(os.Stderr, "adamic: --wasm requires function names; fix: use --wasm first,second")
					return 2
				}
				names = append(names, name)
			}
		default:
			fmt.Fprintln(os.Stderr, usage)
			return 2
		}
	}
	if output == "" {
		fmt.Fprintln(os.Stderr, usage)
		return 2
	}
	options := worker.Options{Wasm: names, BuildWasm: func(entry, module, abi string, names []string) error {
		arguments := []string{"--target", "wasm32-wasi", "--abi-json", abi}
		for _, name := range names {
			arguments = append(arguments, "--export", name)
		}
		if code := buildWithExports(entry, module, arguments, true); code != 0 {
			return fmt.Errorf("worker: Wasm build refused; fix: use ABI version 1 signatures and a configured WASI toolchain")
		}
		return nil
	}}
	if err := worker.BuildWith(entry, output, options); err != nil {
		fmt.Fprintln(os.Stderr, relative("adamic: "+err.Error()))
		return 1
	}
	return 0
}
