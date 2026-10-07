package tailwind

import (
	"encoding/json"
	"os"
	"sort"
	"sync"
)

func AdamicSortedKeys(set map[string]bool) []string { return sortedKeys(set) }

var adamicSortLock sync.Mutex

func AdamicRecordSortedKeys(set map[string]bool) {
	path := os.Getenv("ADAMIC_SLOT04_SORTED_KEYS")
	if path == "" {
		return
	}
	adamicSortLock.Lock()
	defer adamicSortLock.Unlock()
	type entry struct {
		Key   string `json:"key"`
		Value bool   `json:"value"`
	}
	keys := []string{}
	for key := range set {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	values := []entry{}
	for _, key := range keys {
		values = append(values, entry{key, set[key]})
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	if err = json.NewEncoder(file).Encode(values); err != nil {
		panic(err)
	}
	if err = file.Close(); err != nil {
		panic(err)
	}
}
