# 拡張課題の設計ヒント

実装する前に自分の文法を書いてから読む。

## 単項マイナス

```ebnf
term = unary, { ("*" | "/" | "%"), unary } ;
unary = "-", unary | primary ;
```

この形なら`--2`も受理する。再帰する前に`-`を消費するので左再帰ではない。FIRST(unary)とFIRST(expr)に`-`を加える。Unary ASTを加え、VMでは子を評価してNEGするか、`PUSH 0`→子の命令→`BINARY '-'`へ変換できる。

## ifとelse

```ebnf
if_stmt = "if", "(", cond, ")", "{", stmts, "}",
          [ "else", "{", stmts, "}" ] ;
```

elseがあればthen末尾からelse後へ飛ぶJUMPも必要。スコープを導入するなら、どちらの経路もENTER/LEAVEが対応する配置にする。現在のcondはFOLLOWが`)`なので、そのまま再利用できる。

## 識別子と関数呼出し

```ebnf
primary = NUMBER | IDENT, [ "(", args, ")" ] | "(", expr, ")" ;
```

IDENTを先に読み、`(`なら呼出し、そうでなければ変数。argsを空にできる設計では、`)`で空を選択し、引数のFIRSTと重ならないか確認する。これは左因子分解であり、左再帰除去ではない。

## エラー予測

| 入力 | 停止する関数 | current | 期待 |
| --- | --- | --- | --- |
| `put(1+);` | parsePrimary | `)` | NUMBER / IDENT / `(` |
| `while(a<){}` | parsePrimary | `)` | NUMBER / IDENT / `(` |
| `a=1 2;` | parseAssignment内のexpect | NUMBER(2) | `;` |
| `while(1){` | parseWhile内のexpect | EOF | `}` |

課題の拡張は完成版に実装していない。ここに示すものは設計例で、追加後のテストは自分で用意する。
