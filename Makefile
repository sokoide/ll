TARGET=./mini-compiler
SRCS=$(wildcard cmd/mini-compiler/*.go pkg/minilang/*.go) go.mod go.sum
LL=ERROR

.PHONY: all clean generate run test testcli

all: $(TARGET)

$(TARGET): $(SRCS)
	go build -o $(TARGET) ./cmd/mini-compiler

generate:
	@echo 'Handwritten LL parser: no code generation required (see grammar.ebnf).'

clean:
	rm -f $(TARGET)

run: $(TARGET)
	$(TARGET)

test: $(TARGET)
	go test -v ./...

testcli: $(TARGET)
	@set -e; out=$$(printf 'put("ans=",1*2);\n' | $(TARGET) -logLevel $(LL)); \
	if [ ! "$$out" = "ans=2" ]; then echo 'FAIL: put("ans=",1*2); ->'" got '$$out'"; exit 1; fi; \
	echo 'ok: put("ans=",1*2); -> ans=2'
	@set -e; out=$$(printf 'put("ans=",(1*2)+10/3);\n' | $(TARGET) -logLevel $(LL)); \
	if [ ! "$$out" = "ans=5" ]; then echo 'FAIL: put("ans=",(1*2)+10/3); ->'" got '$$out'"; exit 1; fi; \
	echo 'ok: put("ans=",(1*2)+10/3); -> ans=5'
	@set -e; out=$$(printf 'put("ans=",(40+1)+2/2);\n' | $(TARGET) -logLevel $(LL)); \
	if [ ! "$$out" = "ans=42" ]; then echo 'FAIL: put("ans=",(40+1)+2/2); ->'" got '$$out'"; exit 1; fi; \
	echo 'ok: put("ans=",(40+1)+2/2); -> ans=42'
	@set -e; out=$$(printf 'a=40;b=2;c=a+b;put("c=",c);\n' | $(TARGET) -logLevel $(LL)); \
	if [ ! "$$out" = "c=42" ]; then echo 'FAIL: a=40;b=2;c=a+b;put("c=",c); ->'" got '$$out'"; exit 1; fi; \
	echo 'ok: a=40;b=2;c=a+b;put("c=",c); -> c=42'
	@set -e; out=$$(printf 'a=0;while(a<5){a=a+1;};put(a);\n' | $(TARGET) -logLevel $(LL)); \
	if [ ! "$$out" = "5" ]; then echo 'FAIL: a=0;while(a<5){a=a+1;};put(a); ->'" got '$$out'"; exit 1; fi; \
	echo 'ok: a=0;while(a<5){a=a+1;};put(a); -> 5'
	@set -e; out=$$(printf 'a=0;while(a<5){a=a+1;b=b+1;};put(a+b);\n' | $(TARGET) -logLevel $(LL)); \
	if [ ! "$$out" = "5" ]; then echo 'FAIL: a=0;while(a<5){a=a+1;b=b+1;};put(a+b); ->'" got '$$out'"; exit 1; fi; \
	echo 'ok: a=0;while(a<5){a=a+1;b=b+1;};put(a+b); -> 5 (undefined b reads as 0)'
	@if printf 'put("a"+"b");\n' | $(TARGET) -logLevel $(LL) >/dev/null 2>&1; \
	then echo 'FAIL: put("a"+"b"); should exit non-zero (strings are not supported in arithmetic)'; exit 1; \
	else echo 'ok: put("a"+"b"); exits with an evaluation error'; fi
	@set -e; out=$$(printf 'put("", '\''日本語'\'',001);\n' | $(TARGET) -logLevel $(LL)); \
	if [ ! "$$out" = "日本語001" ]; then echo "FAIL: string/number output: $$out"; exit 1; fi; \
	echo 'ok: quoted strings and literal leading zeros'
	@set -e; for src in 'put(-1);' 'put(1);@put(2);' 'put("unterminated);' 'put(1/0);' 'put(1%0);'; do \
	if printf '%s\n' "$$src" | $(TARGET) -logLevel $(LL) >/dev/null 2>&1; \
	then echo "FAIL: $$src should exit non-zero"; exit 1; \
	else echo "ok: $$src exits non-zero"; fi; done
	@set -e; out=$$(printf 'put("should not execute");@' | $(TARGET) -logLevel $(LL) 2>/dev/null) && exit 1; \
	if [ -n "$$out" ]; then echo 'FAIL: invalid program executed partially'; exit 1; fi; \
	echo 'ok: parse failure prevents execution'
