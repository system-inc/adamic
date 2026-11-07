cd "$(git rev-parse --show-toplevel)"
go build -o /tmp/a ./cmd/adamic || exit 1
for p in p1 p2 p3 p4; do
  echo "===== $p"
  cat probe/$p.a
  /tmp/a js probe/$p.a > /tmp/$p.mjs; echo "js exit=$?"
  echo "--- emitted js program body"
  sed -n '/^let local_0_handle/,$p' /tmp/$p.mjs
  grep -n "const adamicCallee = " /tmp/$p.mjs | cut -c1-300
  echo "--- node source (pipe)"; node --disable-warning=ExperimentalWarning oracle/node.mjs probe/$p.a | cat; echo "exit=${PIPESTATUS[0]}"
  echo "--- js backend (pipe)"; node --disable-warning=ExperimentalWarning oracle/node.mjs /tmp/$p.mjs | cat; echo "exit=${PIPESTATUS[0]}"
  /tmp/a c probe/$p.a > /tmp/$p.c; echo "c exit=$?"
  echo "--- emitted c initializer"
  sed -n '/^void adamic_initialize_p/,/^}/p' /tmp/$p.c
  /tmp/a build probe/$p.a -o /tmp/$p.bin; echo "build exit=$?"
  echo "--- native (pipe)"; /tmp/$p.bin | cat; echo "exit=${PIPESTATUS[0]}"
done
