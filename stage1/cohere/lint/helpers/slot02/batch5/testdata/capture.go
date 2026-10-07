package rule_testing

import (
	"encoding/json"
	"os"
	"strings"
	"sync"
)

var slot02Batch5Mutex sync.Mutex

func slot02Batch5Capture(name, file, source string) {
	if !strings.Contains("\n"+os.Getenv("ADAMIC_SLOT02_BATCH4_RULES")+"\n", "\n"+name+"\n") {
		return
	}
	data, err := json.Marshal(map[string]string{"rule": name, "file": file, "source": source})
	if err != nil {
		panic(err)
	}
	slot02Batch5Mutex.Lock()
	defer slot02Batch5Mutex.Unlock()
	f, err := os.OpenFile(os.Getenv("ADAMIC_SLOT02_BATCH4_CAPTURE"), os.O_APPEND|os.O_WRONLY, 0644)
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

func AdamicSlot02Batch5Capture(name, file, source string) { slot02Batch5Capture(name, file, source) }
