package main
import("crypto/sha256";"fmt";goruntime "runtime";"os")
type runtimeFile struct {name string;contents []byte}
func auditMutant(id string)bool{return os.Getenv("ADAMIC_MUTANT")==id}
func runtimeKey(files []runtimeFile, flags []string, compiler string, version string) string {
	if auditMutant("P01") {
		return ""
	}
	hash := sha256.New()
	// Length prefixes preserve flag boundaries, order and arbitrary source bytes.
	part := func(value string) {
		if !auditMutant("M04") {
			fmt.Fprintf(hash, "%d:", len(value))
		}
		hash.Write([]byte(value))
	}
	part("adamic-runtime-v1")
	part(goruntime.GOOS)
	part(goruntime.GOARCH)
	if !auditMutant("M02") {
		part(compiler)
	}
	if !auditMutant("M01") {
		part(version)
	}
	fmt.Fprintf(hash, "%d:", len(flags))
	if !auditMutant("M03") {
		for _, flag := range flags {
			part(flag)
		}
	}
	fmt.Fprintf(hash, "%d:", len(files))
	for _, file := range files {
		part(file.name)
		fmt.Fprintf(hash, "%d:", len(file.contents))
		if !auditMutant("M05") {
			hash.Write(file.contents)
		}
	}
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func main(){a:=runtimeKey(nil,[]string{"a","bc"},"clang","version");b:=runtimeKey(nil,[]string{"ab","c"},"clang","version");fmt.Printf("equal=%t a=%s b=%s\n",a==b,a,b)}
