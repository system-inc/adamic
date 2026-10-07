package tailwind

import (
	"encoding/json"
	"os"
	"sync"
)

func AdamicBreakpointBucket(s string) string { return breakpointBucket(s) }

var adamicStringLock sync.Mutex

func AdamicRecordString(helper, s string) {
	path := os.Getenv("ADAMIC_SLOT04_STRINGS")
	if path == "" {
		return
	}
	adamicStringLock.Lock()
	defer adamicStringLock.Unlock()
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	values := []int{}
	for _, b := range []byte(s) {
		values = append(values, int(b))
	}
	if err := json.NewEncoder(f).Encode(struct {
		Name  string
		Bytes []int
	}{"control:" + helper, values}); err != nil {
		panic(err)
	}
}
