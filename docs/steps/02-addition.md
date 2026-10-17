# Step 2: 加減算と左結合

作業用文法を整数1個から加減算へ広げる。

```ebnf
program = expr, EOF ;
expr = NUMBER, { ("+" | "-"), NUMBER } ;
```

`20-5-3`を`(20-5)-3`として表すASTを作る。評価値12だけでなく木の形を確認する。

## 左再帰をそのまま関数にしない

`expr → expr '-' NUMBER | NUMBER`を関数に直すと、入力を消費する前に`parseExpr`が自身を呼び続ける。そこで、最初のNUMBERを1回読み、残りをforで反復する。

```go
// Pseudocode: error handling is shown in the completed parseExpr implementation.
lhs := parseNumber()
for current == '+' || current == '-' {
    op := current
    advance()
    rhs := parseNumber()
    lhs = Binary{Left: lhs, Op: op, Right: rhs}
}
return lhs
```

| 消費済みの入力 | lhsが表す木 | 未読の先頭 |
| --- | --- | --- |
| `20` | `20` | `-` |
| `20-5` | `Binary(20,-,5)` | `-` |
| `20-5-3` | `Binary(Binary(20,-,5),-,3)` | EOF |

ループの先頭で「lhsが消費済みの式全体を表す」という不変条件を保つ。右辺を同じ`parseExpr`で読むと右結合になり得る。ここではNUMBERだけを読む。

## 実装課題

Step 1のASTにBinaryノードを加え、上の文法を実装する。NUMBERの次が`+`か`-`なら反復し、そうでなければexprから戻る。呼出し元のprogramがEOFを検査するので、`1 2`は成功しない。

`1+`では、演算子を消費した後にNUMBERがないため失敗する。右辺の失敗を無視してlhsだけ返してはいけない。

完成版ではNUMBERの代わりにtermを読む。[parseExpr](../../pkg/minilang/parser.go)のfor構造を比較する。

```sh
printf 'put(20-5-3);' | ./mini-compiler -dump ast
```

```text
Put
  Binary '-'
    Binary '-'
      Number 20
      Number 5
    Number 3
```

## 合格条件

自作版で`1+2`、`20-5-3`を受理し、`1+`、`+1`、`1 2`を拒否する。ASTが左結合であることをテストする。

```sh
go test ./pkg/minilang -run TestASTShape
```

次は[優先順位と括弧](03-precedence.md)。
