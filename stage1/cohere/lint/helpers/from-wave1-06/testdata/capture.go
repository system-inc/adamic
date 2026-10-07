package rule_testing

import (
	"encoding/json"
	"os"
	"strings"
	"sync"
)

var wave06Mutex sync.Mutex

func wave06Capture(name, file, source string) {
	if !strings.Contains("\n"+os.Getenv("ADAMIC_WAVE06_RULES")+"\n", "\n"+name+"\n") {
		return
	}
	data, err := json.Marshal(map[string]string{"rule": name, "file": file, "source": source})
	if err != nil {
		panic(err)
	}
	wave06Mutex.Lock()
	defer wave06Mutex.Unlock()
	f, err := os.OpenFile(os.Getenv("ADAMIC_WAVE06_CAPTURE"), os.O_APPEND|os.O_WRONLY, 0644)
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

func AdamicWave06Capture(name, file, source string) { wave06Capture(name, file, source) }
