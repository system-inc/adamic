package nextjs

import (
	"encoding/json"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
)

var adamicCaptureLock sync.Mutex

func adamicBytes(s string) string {
	a := []string{}
	for _, b := range []byte(s) {
		a = append(a, strconv.Itoa(int(b)))
	}
	return strings.Join(a, ",")
}
func adamicList(a []string) string {
	b := []string{}
	for _, s := range a {
		b = append(b, adamicBytes(s))
	}
	return strings.Join(b, ":")
}
func adamicNames(m map[string]bool) string {
	if m == nil {
		return "nil"
	}
	a := []string{}
	for k, v := range m {
		if v {
			a = append(a, k)
		}
	}
	sort.Strings(a)
	return strings.Join(a, ",")
}
func adamicContracts(m map[string]map[string]bool) string {
	keys := []string{}
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	a := []string{}
	for _, k := range keys {
		a = append(a, k+"="+adamicNames(m[k]))
	}
	return strings.Join(a, ";")
}
func adamicRecord(kind, path, want string) {
	target := os.Getenv("ADAMIC_NEXTJS_CAPTURE")
	if target == "" {
		return
	}
	adamicCaptureLock.Lock()
	defer adamicCaptureLock.Unlock()
	f, err := os.OpenFile(target, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err = json.NewEncoder(f).Encode(struct{ Kind, Path, Want string }{kind, path, want}); err != nil {
		panic(err)
	}
}
func IsDocumentFile(path string) bool {
	v := adamicRawIsDocumentFile(path)
	adamicRecord("IsDocumentFile", path, strconv.FormatBool(v))
	return v
}
func lastSeparator(path string) int {
	v := adamicRawlastSeparator(path)
	adamicRecord("lastSeparator", path, strconv.Itoa(v))
	return v
}
func splitPath(path string) (string, string) {
	a, b := adamicRawsplitPath(path)
	adamicRecord("splitPath", path, adamicBytes(a)+"/"+adamicBytes(b))
	return a, b
}
func IsDocumentPage(path string) bool {
	v := adamicRawIsDocumentPage(path)
	adamicRecord("IsDocumentPage", path, strconv.FormatBool(v))
	return v
}
func IsInApplicationDirectory(path string) bool {
	v := adamicRawIsInApplicationDirectory(path)
	adamicRecord("IsInApplicationDirectory", path, strconv.FormatBool(v))
	return v
}
func IsInPagesDirectory(path string) bool {
	v := adamicRawIsInPagesDirectory(path)
	adamicRecord("IsInPagesDirectory", path, strconv.FormatBool(v))
	return v
}
func splitSegments(path string) []string {
	v := adamicRawsplitSegments(path)
	adamicRecord("splitSegments", path, adamicList(v))
	return v
}
func RouteContractExports(path string) map[string]bool {
	v := adamicRawRouteContractExports(path)
	adamicRecord("RouteContractExports", path, adamicNames(v))
	return v
}
func buildRouteFileContracts() map[string]map[string]bool {
	v := adamicRawbuildRouteFileContracts()
	adamicRecord("buildRouteFileContracts", "", adamicContracts(v))
	return v
}
