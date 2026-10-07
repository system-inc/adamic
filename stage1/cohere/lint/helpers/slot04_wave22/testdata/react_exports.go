package react

import (
	"encoding/json"
	"os"
	"sync"
)

var adamicOptionLock sync.Mutex

func adamicRecordOptions(raw []byte) {
	p := os.Getenv("ADAMIC_SLOT04_OPTIONS")
	if p == "" {
		return
	}
	adamicOptionLock.Lock()
	defer adamicOptionLock.Unlock()
	f, e := os.OpenFile(p, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if e != nil {
		panic(e)
	}
	defer f.Close()
	if e = json.NewEncoder(f).Encode(map[string]string{"OptionRaw": string(raw)}); e != nil {
		panic(e)
	}
}
