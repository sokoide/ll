# Step 4: 文・変数・出力

ここから入力全体は式ではなく、文の列になる。Step 3の算術規則を引き継ぎ、primaryにIDENTを加える。

```ebnf
program = stmts, EOF ;
stmts = { stmt } ;
stmt = assignment | put_stmt | ";" ;
assignment = IDENT, "=", expr, ";" ;
put_stmt = "put", "(", put_list, ")", ";" ;
put_list = put_item, { ",", put_item } ;
put_item = STRING | expr ;
primary = NUMBER | IDENT | "(", expr, ")" ;
```

Step 3で必要だったprogramのEOF検査は残す。空プログラムはstmtsが0回なので成功する。`put()`はput_listに最初のput_itemが必須なので失敗する。

## FIRST集合からswitchを書く

FIRSTは「その規則から始められる先頭トークンの集合」。文の選択肢は次のように重ならない。

| 選択肢 | FIRST | 実装 |
| --- | --- | --- |
| assignment | IDENT | parseAssignment |
| put_stmt | put | parsePut |
| 空文 | `;` | 1トークン消費してEmptyを返す |

IDENTを見て代入と決められるのは、この文法にIDENTで始まる他の文がないため。一般的なGo/Cの式文まで加えるなら改めて設計する。

`put_item`もSTRINGとFIRST(expr)が重ならないので1トークンで選べる。文字列をprimaryへ入れないことで、`a="x";`や`put("x"+1);`は構文エラーになる。

## ASTと評価を分ける

[ast.go](../../pkg/minilang/ast.go)では`statement`と`expression`を分ける。代入はstatement、右辺の計算はexpression。構文解析は木を作るだけで、変数への代入や出力を行わない。

[evaluateStmt / evaluateExpr](../../pkg/minilang/eval.go)は、構文解析完了後に呼ぶ。この分離により、`put("bad");@`は何も出力せず構文エラーになる。

## 実装課題と合格条件

Step 3に`parseStmts`と`parseStmt`を追加する。まず代入だけ、その後putとカンマ区切りを追加する。変数はmapで管理し、未定義の読取りは0とする（この言語固有の決定）。

- `a=40;put("ans=",a+2);`の出力が`ans=42`と改行になる。
- `put(1,);`と`a=1`を拒否する。
- `put("bad");@`で評価を一度も始めない。
- `expr`関数が`;`や`,`を消費しない理由を説明する。

```sh
printf 'a=40;put("ans=",a+2);' | ./mini-compiler -dump ast
make testcli
```

次は[whileとLL(1)判定](05-while.md)。
