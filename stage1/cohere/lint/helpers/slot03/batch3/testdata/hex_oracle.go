package main

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/system-inc/cohere/internal/lint/ecmascript/text"
	"os"
	"sort"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func main() {
	f, e := os.Open(os.Args[1])
	must(e)
	defer f.Close()
	z, e := gzip.NewReader(f)
	must(e)
	defer z.Close()
	scan := bufio.NewScanner(z)
	scan.Buffer(make([]byte, 4096), 16<<20)
	points := map[int]bool{-2147483648: true, -1: true, 0x110000: true, 2147483647: true}
	for scan.Scan() {
		var row struct{ Source string }
		must(json.Unmarshal(scan.Bytes(), &row))
		for _, point := range row.Source {
			points[int(point)] = true
		}
	}
	must(scan.Err())
	list := []int{}
	for point := range points {
		list = append(list, point)
	}
	sort.Ints(list)
	data, e := json.Marshal(map[string]any{"points": list})
	must(e)
	must(os.WriteFile(os.Args[2], data, 0644))
	for _, point := range list {
		fmt.Println(text.AdamicHexValue(rune(point)))
	}
	for point := 0; point <= 0x10ffff; point++ {
		fmt.Println(text.AdamicHexValue(rune(point)))
	}
}
