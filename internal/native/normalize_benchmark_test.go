package native

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestNormalizeLongMeasurements is the integration performance probe, not an oracle fixture.
// Opt in with ADAMIC_NORMALIZE_BENCH_LOG=/absolute/path/results.jsonl. It measures normalization
// after constructing and flattening the input, serially, best of five against Node. The Python
// driver records wait4 peak RSS and load. Oversized K forms have a ten-second time limit, and
// native processes a twelve-GiB address-space limit. The ordinary gate skips this heavy program.
func TestNormalizeLongMeasurements(t *testing.T) {
	// Not parallel: timing and peak RSS need an otherwise idle worker.
	destination := os.Getenv("ADAMIC_NORMALIZE_BENCH_LOG")
	if destination == "" {
		// census: measurement ADAMIC_NORMALIZE_BENCH_LOG names a writable JSONL output for opt-in long-string timing and RSS observations.
		t.Skip("set ADAMIC_NORMALIZE_BENCH_LOG to run the long-string measurements")
	}
	destination, err := filepath.Abs(destination)
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	binary := filepath.Join(directory, "probe")
	if err := Build(normalizeLongProbe, binary, Options{}); err != nil {
		t.Fatal(err)
	}
	oracle := filepath.Join(directory, "oracle.js")
	driver := filepath.Join(directory, "measure.py")
	for path, content := range map[string]string{oracle: normalizeLongOracle, driver: normalizeLongDriver} {
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	command := exec.Command("python3", driver, binary, oracle, destination)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("measurements: %v\n%s", err, output)
	} else {
		t.Logf("%s", output)
	}
}

const normalizeLongProbe = `#define _POSIX_C_SOURCE 200809L
#include "adamic.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
static double now(void) { struct timespec t; clock_gettime(CLOCK_MONOTONIC, &t); return t.tv_sec + t.tv_nsec * 1e-9; }
int main(int argc, char **argv) {
 if (argc != 4) return 2;
 const char *patterns[] = {"a", "e\314\201a\314\243\314\201", "\352\260\201", "\341\204\200\341\205\241\341\206\250", "\357\267\272"};
 size_t counts[] = {1,5,1,3,1};
 int kind = atoi(argv[1]); size_t n = strtoull(argv[2], NULL, 10), p = strlen(patterns[kind]);
 size_t reps = n / counts[kind], extra = n % counts[kind];
 adamic_string *input = adamic_string_allocate(reps*p+extra);
 for (size_t i=0;i<reps;i++) memcpy((char *)input->bytes+i*p,patterns[kind],p);
 memset((char *)input->bytes+reps*p,'a',extra);
 adamic_string form = ADAMIC_STRING(""); form.bytes=argv[3]; form.length=strlen(argv[3]);
 double start=now(); adamic_string *out=adamic_string_normalize(input,&form); double elapsed=now()-start;
 printf("%.9f %zu %u %u\n",elapsed,out->length,(unsigned char)out->bytes[0],(unsigned char)out->bytes[out->length-1]);
 adamic_release(out); adamic_release(input); return 0;
}
`

const normalizeLongOracle = `const [kind,n,form]=process.argv.slice(2); const patterns=['a','e\u0301a\u0323\u0301','각','각','ﷺ']; const counts=[1,5,1,3,1];
let input=patterns[kind].repeat(Math.floor(Number(n)/counts[kind]))+'a'.repeat(Number(n)%counts[kind]);
// Force V8's repeat rope flat before timing normalization.
input.charCodeAt(input.length-1);
const start=process.hrtime.bigint(); const out=input.normalize(form); const elapsed=Number(process.hrtime.bigint()-start)/1e9;
console.log(elapsed,Buffer.byteLength(out),out.codePointAt(0),out.charCodeAt(out.length-1));
`

const normalizeLongDriver = `import subprocess,os,json,sys,time,resource,threading
binary, oracle, destination = sys.argv[1:]
out=open(destination,'w',buffering=1)
failed=False
def limit(): resource.setrlimit(resource.RLIMIT_AS,(12*1024**3,12*1024**3))
for n in [100000,10000000,100000000,29826160]:
 for kind in (range(5) if n!=29826160 else [4]):
  for form in (['NFC','NFD','NFKC','NFKD'] if n!=29826160 else ['NFKD']):
   for repeat in range(5):
    for impl in ['native','node']:
     command=[binary,str(kind),str(n),form] if impl=='native' else ['node',oracle,str(kind),str(n),form]
     load=os.getloadavg(); start=time.monotonic()
     r=subprocess.Popen(command,text=True,stdout=subprocess.PIPE,stderr=subprocess.PIPE,preexec_fn=limit if impl=='native' else None)
     bounded = n==100000000 and kind==4 and form in ['NFKC','NFKD']
     timer=threading.Timer(10,r.terminate) if bounded else None
     if timer:timer.start()
     _,status,usage=os.wait4(r.pid,0); r.returncode=os.waitstatus_to_exitcode(status)
     if timer:timer.cancel()
     stdout=r.stdout.read(); stderr=r.stderr.read()+'\nRSS '+str(usage.ru_maxrss)
     row=dict(n=n,kind=kind,form=form,repeat=repeat,impl=impl,load=load,wall=time.monotonic()-start,code=r.returncode,stdout=stdout.strip(),stderr=stderr.strip())
     out.write(json.dumps(row)+'\n')
     invalid = n==100000000 and kind==4 and form in ['NFKC','NFKD']
     failed = failed or (r.returncode==0 if invalid else r.returncode!=0)
print('610 measurements written to',destination)
sys.exit(1 if failed else 0)
`
