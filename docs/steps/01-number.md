# Step 1: 整数1個を最後まで読む

目標は「文字を読む処理」と「トークンを読む処理」を分け、未読入力を残したまま成功しないこと。Goの関数・struct・interface・errorを知っていれば始められる。

このステップの作業用文法は次の2規則だけ。完成版の`put`や変数はまだ実装しない。

```ebnf
program = NUMBER, EOF ;
NUMBER = digit, { digit } ;
```

`42`は成功、空入力・`42 17`・`42@`は失敗する。空白はトークン間で読み飛ばす。`EOF`は入力終端を示すトークンで、ソース中に`EOF`という文字を書くわけではない。

## 自分で作る順序

完成版を変更する前に、`/tmp/ll-step`など別のGoモジュールで小さい実装を作る。各ステップの文法は作業用の部分言語であり、完成版CLIがその文法へ切り替わるわけではない。

1. トークンを`kind`、`literal`、`position`で表す。整数の値への変換は後の評価処理へ任せ、ここでは字面を保つ。
2. `Scan()`を作る。`42`の2文字をまとめてNUMBERにし、不正文字はILLEGALにする。EOFと混ぜない。
3. `parser`に`current`を1個だけ持たせる。最初に`advance()`を呼ぶ。
4. `expect(NUMBER)`と`expect(EOF)`を順に呼ぶ。
5. 数値リテラルのASTノードを返す。まだ評価しない。

`advance`と`expect`の契約は以後ずっと同じ。

```go
// Pseudocode: Scan returns the next token.
func (p *parser) advance() { p.current = p.scanner.Scan() }

func (p *parser) expect(want Kind) error {
    if p.current.Kind != want {
        return fmt.Errorf("expected %v, got %v", want, p.current.Kind)
    }
    p.advance()
    return nil
}
```

`expect`が成功すれば1トークン消費する。失敗時に勝手に読み進めたり再試行したりしない。この教材では最初の構文エラーで停止する。

## 完成版で観察する

```sh
printf 'put(42);' | ./mini-compiler -dump tokens
```

```text
1:1 put      "put"
1:4 '('      "("
1:5 NUMBER   "42"
1:7 ')'      ")"
1:8 ';'      ";"
1:9 EOF      ""
```

文字の先読みとトークンの先読みは別。lexerが`<`の次の文字を見て`<=`を作っても、parserが見るのは比較トークン1個である。

読むコード: [lexer.go](../../pkg/minilang/lexer.go)の`Scan`、`scanNumber`、[parser.go](../../pkg/minilang/parser.go)の`advance`、`expect`。

## 合格条件

自作した小さいパーサーで、`42`と` 42 `が成功し、空入力・`42 17`・`42@`が位置付きエラーになるテストを書く。`42@`を受理したらEOF処理を見直す。

完成版の関連テスト:

```sh
go test ./pkg/minilang -run 'TestTokens|TestSyntaxErrors|TestErrorPosition'
```

次は[加減算](02-addition.md)。
