# Step 3: 乗除算・優先順位・括弧

```ebnf
program = expr, EOF ;
expr = term, { ("+" | "-"), term } ;
term = primary, { ("*" | "/" | "%"), primary } ;
primary = NUMBER | "(", expr, ")" ;
```

優先順位は演算子の数値を比較する仕組みではなく、呼出しの階層で作る。exprの左右はtermが完成するまで返らないため、掛け算が先にまとまる。

`1+2*3`の関数呼出しを手で追う。各parse関数は、自分の規則が消費する分だけ読み、次のトークンをcurrentに残して戻る。

| 場所 | 動作 | 戻る時のcurrent |
| --- | --- | --- |
| 最初のprimary | Number(1)を作る | `+` |
| 最初のterm | `+`は乗除算でないので戻る | `+` |
| expr | `+`を消費し、右辺のtermへ | `2` |
| 右辺のprimary | Number(2)を作る | `*` |
| 右辺のterm | `*`とNumber(3)を読み、Binary(2,*,3)を作る | EOF |
| expr | Binary(1,+,Binary(2,*,3))を作る | EOF |

完成版で`put(1+2*3);`を読むときは、表のEOFが`)`になる。exprは`)`を消費しない。`)`を必要としている`parsePut`が消費する。この役割分担が括弧の入れ子でも成り立つ。

## 実装課題

1. Step 2の`parseExpr`が読む要素をNUMBERからtermへ置き換える。
2. `parseTerm`に同じ反復構造を作る。
3. `parsePrimary`でNUMBERか`(`を選び、`(`ならexprを読んだ後に必ず`)`を検査する。
4. エラーを呼出し元まで返す。

`primary → '(' expr ')'`からexprへの再帰は、先に`(`を消費しているため左再帰ではない。

```sh
printf 'put(1+2*3);' | ./mini-compiler -dump ast
printf 'put((1+2)*3);' | ./mini-compiler -dump ast
go test ./pkg/minilang -run 'TestASTShape|TestEvaluation'
```

## 合格条件

自作版で`1+2*3=7`、`(1+2)*3=9`、`24/4/2=3`を確認する。`(1+2`、`1**2`を拒否する。ASTも比較し、加算の右の子が乗算である理由を説明できるようにする。

次は[文と変数](04-statements.md)。
