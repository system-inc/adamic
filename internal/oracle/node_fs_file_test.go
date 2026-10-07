package oracle

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

var fsFileFixtures = []string{"read", "open", "write", "close", "write_file", "exists", "stat", "mkdir", "unlink", "utimes", "date", "system", "buffer", "read_sync", "write_buffer", "mkdtemp", "rm"}

func fsFilePrepare(t *testing.T, shared, name string) inputRun {
	t.Helper()
	root := filepath.Join(shared, name)
	if err := os.Mkdir(root, 0777); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(root, 0777); err != nil {
		t.Fatal(err)
	}
	for _, dir := range []string{"directory", "locked", "closed"} {
		if err := os.Mkdir(filepath.Join(root, dir), 0755); err != nil {
			t.Fatal(err)
		}
	}
	for name, data := range map[string][]byte{
		"file": []byte("abc"), "utf8": []byte("héllo 🌍"), "bom": []byte("\xef\xbb\xbfhello"), "bad": {0xe2, 0x82, 0xff, 0xed, 0xa0, 0x80},
		"utf16le": {0xff, 0xfe, 0x61, 0, 0x3d, 0xd8, 0, 0xde}, "utf16be": {0xfe, 0xff, 0, 0x61, 0xd8, 0x3d, 0xde, 0}, "utf16be-odd": {0xfe, 0xff, 0, 0x61, 0x62}, "empty": {},
		"locked/file": []byte("locked"), "closed/file": []byte("closed"),
	} {
		if err := os.WriteFile(filepath.Join(root, name), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	for name, target := range map[string]string{"link": "file", "dangling": "missing", "directory-link": "directory"} {
		if err := os.Symlink(target, filepath.Join(root, name)); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Chmod(filepath.Join(root, "locked"), 0555); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(filepath.Join(root, "closed"), 0000); err != nil {
		t.Fatal(err)
	}
	how := inputRun{directory: root, arguments: []string{root}}
	if os.Geteuid() == 0 {
		how.credential = &syscall.Credential{Uid: 65534, Gid: 65534}
	}
	return how
}

// Like the input oracle's snapshot, also recording symlinks without following
// them. Every run has its own directory, including the leak-check run.
func fsFileSnapshot(t *testing.T, root string) map[string]string {
	t.Helper()
	result := map[string]string{}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			if entry != nil && entry.IsDir() {
				return filepath.SkipDir
			}
			return err
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if entry.Type()&os.ModeSymlink != 0 {
			target, err := os.Readlink(path)
			if err != nil {
				return err
			}
			result[relative] = "link " + target
			return nil
		}
		if entry.IsDir() {
			result[relative] = fmt.Sprintf("directory %v", info.Mode().Perm())
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		result[relative] = fmt.Sprintf("file %v %q", info.Mode().Perm(), data)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func TestNodeFSFileAgreesWithNode(t *testing.T) {
	t.Parallel()
	for _, fixture := range fsFileFixtures {
		t.Run(fixture, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/node_fs_file_"+fixture+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			shared := sharedDirectory(t)
			truthRun := fsFilePrepare(t, shared, "node")
			truth := onNodeWith(t, truthRun, path)
			if truth.exitCode != 0 || len(truth.stderr) != 0 {
				t.Fatalf("source failed on Node: exit %d stderr %s", truth.exitCode, truth.stderr)
			}
			backendRun := fsFilePrepare(t, shared, "javascript")
			backend := inputBackend(t, backendRun, program, shared)
			nativeRun := fsFilePrepare(t, shared, "native")
			got, binary := inputNatively(t, nativeRun, program, shared)
			for name, observation := range map[string]run{"native": got, "javascript": backend} {
				if difference := disagreement(truth, observation); difference != "" {
					t.Errorf("%s: %s\nNode: %q stderr %q exit %d\ngot: %q stderr %q exit %d", name, difference, truth.stdout, truth.stderr, truth.exitCode, observation.stdout, observation.stderr, observation.exitCode)
				}
			}
			wanted := fsFileSnapshot(t, truthRun.arguments[0])
			for name, how := range map[string]inputRun{"native": nativeRun, "javascript": backendRun} {
				if difference := filesDiffer(wanted, fsFileSnapshot(t, how.arguments[0])); difference != "" {
					t.Errorf("%s filesystem: %s", name, difference)
				}
			}
			if leaked := inputLeaks(t, fsFilePrepare(t, shared, "leaks"), program, binary); leaked != "" {
				t.Errorf("leaks: %s", leaked)
			}
			t.Logf("Node bytes, effects, ASan/UBSan and leak check passed")
		})
	}
}

// Each mutant compiles, exits 0, and passes the sanitizer and leak checks. Only
// the comparison with the source's stdout on Node is allowed to catch it.
func TestNodeFSFileMutants(t *testing.T) {
	t.Parallel()
	cases := []struct{ name, fixture, operation, helper string }{
		{"mkdtempSync suffix", "mkdtemp", "mkdtemp", `static adamic_string *fs_file_mutant(const adamic_string *prefix) {static adamic_string suffix=ADAMIC_STRING("!");adamic_string *changed=adamic_string_concat(2,(adamic_string *const[]){(adamic_string *)prefix,&suffix});adamic_string *result=adamic_fs_file_mkdtemp(changed);adamic_release(changed);return result;}`},
		{"rmSync drop force", "rm", "rm", `static double fs_file_mutant(const adamic_string *path,bool recursive,bool force) {(void)force;return adamic_fs_file_rm(path,recursive,false);}`},
		{"rmSync force", "rm", "rm", `static double fs_file_mutant(const adamic_string *path,bool recursive,bool force) {(void)force;return adamic_fs_file_rm(path,recursive,true);}`},

		{"writeFileSync Buffer bytes", "write_buffer", "write_buffer", `static double fs_file_mutant(const adamic_string *path,const adamic_array *data,const adamic_string *flag,double mode,bool flush) {static adamic_string truncate=ADAMIC_STRING("w");if(flag->length==1&&flag->bytes[0]=='a')flag=&truncate;return adamic_fs_file_write_buffer(path,data,flag,mode,flush);}`},
		{"writeFileSync Buffer fd", "write_buffer", "write_buffer_fd", `static double fs_file_mutant(double fd,const adamic_array *data,const adamic_string *flag,double mode,bool flush) {(void)fd;(void)data;(void)flag;(void)mode;(void)flush;return 0;}`},
		{"readSync byte count", "read_sync", "read_sync", `static double fs_file_mutant(double fd,adamic_array *buffer,double offset,double length,double position) {return adamic_fs_file_read_sync(fd,buffer,offset,length,position)+1;}`},
		{"readFileSync raw bytes", "buffer", "read_buffer", `static adamic_array *fs_file_mutant(const adamic_string *path,const adamic_string *flag) {adamic_array *value=adamic_fs_file_read_buffer(path,flag);if(value!=NULL && value->length>0)value->elements[0].number=fmod(value->elements[0].number+1,256);return value;}`},
		{"readFileSync raw fd", "buffer", "read_buffer_fd", `static adamic_array *fs_file_mutant(double fd,const adamic_string *flag) {adamic_array *value=adamic_fs_file_read_buffer_fd(fd,flag);if(value!=NULL && value->length>0)value->elements[0].number=fmod(value->elements[0].number+1,256);return value;}`},
		{"readFileSync UTF8", "read", "read_file", `static adamic_string *fs_file_mutant(const adamic_string *path,const adamic_string *flag) { adamic_string *value=adamic_fs_file_read_file(path,flag); if(value==NULL)return NULL; static adamic_string suffix=ADAMIC_STRING("!"); adamic_string *result=adamic_string_concat(2,(adamic_string *const[]){value,&suffix});adamic_release(value);return result; }`},
		{"readFileSync fd", "read", "read_fd", `static adamic_string *fs_file_mutant(double fd,const adamic_string *flag) { adamic_string *value=adamic_fs_file_read_fd(fd,flag); if(value==NULL)return NULL; static adamic_string suffix=ADAMIC_STRING("!"); adamic_string *result=adamic_string_concat(2,(adamic_string *const[]){value,&suffix});adamic_release(value);return result; }`},
		{"openSync truncation", "open", "open", `static double fs_file_mutant(const adamic_string *path,const adamic_string *flag,double mode) { static adamic_string append=ADAMIC_STRING("a"); if(flag->length==1 && flag->bytes[0]=='w')flag=&append;return adamic_fs_file_open(path,flag,mode); }`},
		{"writeSync byte count", "write", "write", `static double fs_file_mutant(double fd,const adamic_string *value,double position) { return adamic_fs_file_write(fd,value,position)+1; }`},
		{"closeSync swallowed failure", "close", "close", `static double fs_file_mutant(double fd) { double result=adamic_fs_file_close(fd);if(adamic_thrown!=NULL){adamic_release(adamic_thrown);adamic_thrown=NULL;}return result; }`},
		{"writeFileSync append flag", "write_file", "write_file", `static double fs_file_mutant(const adamic_string *path,const adamic_string *value,const adamic_string *flag,double mode,bool flush) { static adamic_string truncate=ADAMIC_STRING("w");if(flag->length==1 && flag->bytes[0]=='a')flag=&truncate;return adamic_fs_file_write_file(path,value,flag,mode,flush); }`},
		{"existsSync directories", "exists", "exists", `static bool fs_file_mutant(const adamic_string *path) { adamic_object *stat=adamic_fs_file_stat(path,false);if(adamic_thrown!=NULL){adamic_release(adamic_thrown);adamic_thrown=NULL;}if(stat==NULL)return false;bool result=adamic_fs_file_is_file(stat);adamic_release(stat);return result; }`},
		{"statSync missing option", "stat", "stat", `static adamic_object *fs_file_mutant(const adamic_string *path,bool throws) { (void)throws;return adamic_fs_file_stat(path,false); }`},
		{"Stats size", "stat", "stat", `static adamic_object *fs_file_mutant(const adamic_string *path,bool throws) { adamic_object *result=adamic_fs_file_stat(path,throws);if(result!=NULL)result->slots[0].number+=1;return result; }`},
		{"Stats mtimeMs", "stat", "stat", `static adamic_object *fs_file_mutant(const adamic_string *path,bool throws) { adamic_object *result=adamic_fs_file_stat(path,throws);if(result!=NULL)result->slots[1].number+=1;return result; }`},
		{"Stats atime", "utimes", "stat", `static adamic_object *fs_file_mutant(const adamic_string *path,bool throws) {adamic_object *result=adamic_fs_file_stat(path,throws);if(result!=NULL){adamic_object *date=result->slots[4].reference;date->slots[0].number+=1;}return result;}`},
		{"Stats mtime", "stat", "stat", `static adamic_object *fs_file_mutant(const adamic_string *path,bool throws) { adamic_object *result=adamic_fs_file_stat(path,throws);if(result!=NULL){adamic_object *date=result->slots[2].reference;date->slots[0].number+=1;}return result; }`},
		{"Stats isFile", "stat", "is_file", `static bool fs_file_mutant(const adamic_object *value) { return !adamic_fs_file_is_file(value); }`},
		{"Stats isDirectory", "stat", "is_directory", `static bool fs_file_mutant(const adamic_object *value) { return !adamic_fs_file_is_directory(value); }`},
		{"Stats isSymbolicLink follows", "stat", "is_symbolic_link", `static bool fs_file_mutant(const adamic_object *value) { (void)value;return true; }`},
		{"mkdirSync first directory", "mkdir", "mkdir", `static adamic_string *fs_file_mutant(const adamic_string *path,bool recursive,double mode) { adamic_string *result=adamic_fs_file_mkdir(path,recursive,mode);adamic_release(result);return NULL; }`},
		{"unlinkSync swallowed failure", "unlink", "unlink", `static double fs_file_mutant(const adamic_string *path) { double result=adamic_fs_file_unlink(path);if(adamic_thrown!=NULL){adamic_release(adamic_thrown);adamic_thrown=NULL;}return result; }`},
		{"utimesSync milliseconds", "utimes", "utimes", `static double fs_file_mutant(const adamic_string *path,double atime,double mtime) { return adamic_fs_file_utimes(path,atime,mtime+1); }`},
		{"Date timestamp", "date", "date_time", `static double fs_file_mutant(const adamic_object *date) { return adamic_fs_file_date_time(date)+1; }`},

		{"System fileExists", "system", "is_file", `static bool fs_file_mutant(const adamic_object *value) { (void)value;return false; }`},
		{"System getFileSize", "system", "stat", `static adamic_object *fs_file_mutant(const adamic_string *path,bool throws) { adamic_object *result=adamic_fs_file_stat(path,throws);if(result!=NULL)result->slots[0].number+=1;return result; }`},
		{"System getModifiedTime", "system", "stat", `static adamic_object *fs_file_mutant(const adamic_string *path,bool throws) { adamic_object *result=adamic_fs_file_stat(path,throws);if(result!=NULL){adamic_object *date=result->slots[2].reference;date->slots[0].number+=1;}return result; }`},
		{"System setModifiedTime", "system", "utimes_dates", `static double fs_file_mutant(const adamic_string *path,const adamic_object *atime,const adamic_object *mtime) {adamic_object *wrong=adamic_fs_file_date_new(adamic_fs_file_date_time(mtime)+1000);double result=adamic_fs_file_utimes_dates(path,atime,wrong);adamic_release(wrong);return result;}`},
		{"System deleteFile", "system", "unlink", `static double fs_file_mutant(const adamic_string *path) {(void)path;return 0;}`},
		{"System createDirectory", "system", "mkdir", `static adamic_string *fs_file_mutant(const adamic_string *path,bool recursive,double mode) {adamic_string *result=adamic_fs_file_mkdir(path,recursive,mode);if(adamic_thrown!=NULL){adamic_release(adamic_thrown);adamic_thrown=NULL;}return result;}`},
		{"System writeFile BOM", "system", "write", `static double fs_file_mutant(double fd,const adamic_string *value,double position) { if(value->length>=3 && (unsigned char)value->bytes[0]==0xef && (unsigned char)value->bytes[1]==0xbb && (unsigned char)value->bytes[2]==0xbf){adamic_string shortened=*value;shortened.bytes+=3;shortened.length-=3;return adamic_fs_file_write(fd,&shortened,position);}return adamic_fs_file_write(fd,value,position); }`},
	}
	for _, one := range cases {
		t.Run(one.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/node_fs_file_"+one.fixture+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			code := native.C(program)
			original := code
			code = strings.ReplaceAll(code, "adamic_fs_file_"+one.operation+"(", "fs_file_mutant(")
			if code == original {
				t.Fatal("mutant changed nothing")
			}
			code = strings.Replace(code, "#include \"adamic.h\"", "#include \"adamic.h\"\n"+one.helper, 1)
			shared := sharedDirectory(t)
			binary := filepath.Join(shared, "mutant")
			if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			how := fsFilePrepare(t, shared, "mutant-input")
			got := executeInput(t, how, []string{"ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary, how.arguments...)
			if got.exitCode != 0 || len(got.stderr) != 0 {
				t.Fatalf("mutant failed outside Node comparison: exit %d stderr %s", got.exitCode, got.stderr)
			}
			truth := onNodeWith(t, fsFilePrepare(t, shared, "node"), path)
			if truth.exitCode != 0 || len(truth.stderr) != 0 {
				t.Fatalf("Node source failed: %+v", truth)
			}
			if bytes.Equal(got.stdout, truth.stdout) {
				t.Fatal("Node stdout did not catch the mutant")
			}
			t.Log("caught only by Node stdout; exit 0, no sanitizer findings or leaks")
		})
	}
}
