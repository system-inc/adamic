package rule_testing

import (
	"encoding/json"
	"os"
	"strings"
	"sync"
)

var slot02Batch8Mutex sync.Mutex

func slot02Batch8Capture(name, file, source string, options any, directory string) {
	if !strings.Contains("\n"+os.Getenv("ADAMIC_SLOT02_BATCH8_RULES")+"\n", "\n"+name+"\n") {
		return
	}
	optionText, err := json.Marshal(options)
	if err != nil {
		panic(err)
	}
	normalized := string(optionText)
	if directory != "" {
		normalized = strings.ReplaceAll(normalized, directory, "<fixture>")
	}
	data, err := json.Marshal(map[string]string{"rule": name, "file": file, "source": source, "options": normalized})
	if err != nil {
		panic(err)
	}
	slot02Batch8Mutex.Lock()
	defer slot02Batch8Mutex.Unlock()
	f, err := os.OpenFile(os.Getenv("ADAMIC_SLOT02_BATCH8_CAPTURE"), os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	_, err = f.Write(append(data, '\n'))
	if err != nil {
		panic(err)
	}
	if err = f.Close(); err != nil {
		panic(err)
	}
}
