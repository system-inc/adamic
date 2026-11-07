set -x
cd "$(git rev-parse --show-toplevel)"
go build -o /tmp/a ./cmd/adamic || exit 1
for p in p1 p2 p3 p4; do
  echo "===== $p"
  cat probe/$p.a
  /tmp/a js probe/$p.a > /tmp/$p.mjs; echo "js exit=$?"
  grep -n "local_\|setBlocking\|adamicCallee = \|_handle" /tmp/$p.mjs | grep -v "^.*const adamic[A-Z][a-zA-Z]* = (" | head -40
  grep -n "const adamicCallee = " /tmp/$p.mjs | cut -c1-400
  echo "--- node source (pipe)"; node --disable-warning=ExperimentalWarning oracle/node.mjs probe/$p.a | cat; echo "exit=${PIPESTATUS[0]}"
  echo "--- js backend (pipe)"; node --disable-warning=ExperimentalWarning oracle/node.mjs /tmp/$p.mjs | cat; echo "exit=${PIPESTATUS[0]}"
  /tmp/a c probe/$p.a > /tmp/$p.c; echo "c exit=$?"
  grep -n "setBlocking\|stdout_handle\|->code(" /tmp/$p.c | head -30
  sed -n '/^int main\|adamic_main\|function_0_main/,+60p' /tmp/$p.c | head -90
  /tmp/a build probe/$p.a -o /tmp/$p.bin; echo "build exit=$?"
  echo "--- native (pipe)"; /tmp/$p.bin | cat; echo "exit=${PIPESTATUS[0]}"
done
