package rule_testing

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

var inventoryMutex sync.Mutex

func inventoryRecord(name string) {
	path := os.Getenv("ADAMIC_INVENTORY_CAPTURE")
	if path == "" {
		return
	}
	location := "unknown caller"
	for depth := 1; depth < 40; depth++ {
		_, file, line, ok := runtime.Caller(depth)
		if !ok {
			break
		}
		if strings.Contains(filepath.ToSlash(file), "/internal/lint/rules/") && strings.HasSuffix(file, "_test.go") {
			index := strings.Index(filepath.ToSlash(file), "/internal/lint/rules/")
			location = "cohere" + filepath.ToSlash(file)[index:] + ":" + itoaInventory(line)
			break
		}
	}
	data, err := json.Marshal(map[string]string{"rule": name, "location": location})
	if err != nil {
		panic(err)
	}
	inventoryMutex.Lock()
	defer inventoryMutex.Unlock()
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	_, err = file.Write(append(data, '\n'))
	closeErr := file.Close()
	if err != nil {
		panic(err)
	}
	if closeErr != nil {
		panic(closeErr)
	}
}
func itoaInventory(value int) string {
	if value == 0 {
		return "0"
	}
	var bytes [20]byte
	i := len(bytes)
	for value > 0 {
		i--
		bytes[i] = byte('0' + value%10)
		value /= 10
	}
	return string(bytes[i:])
}
