# Step 5: while・省略可能な規則・FOLLOW

Step 4に次の規則を加える。ここで[完成版のEBNF](../../grammar.ebnf)になる。

```ebnf
stmt = assignment | while_stmt | put_stmt | ";" ;
while_stmt = "while", "(", cond, ")", "{", stmts, "}" ;
cond = expr, [ compare_op, expr ] ;
compare_op = "==" | "!=" | ">=" | ">" | "<=" | "<" ;
```

exprを読んだ後に比較演算子なら右辺へ進み、`)`なら比較の省略を選ぶ。`while(a)`も`while(a<3)`も、共通のexprを一度だけ読む。バックトラックは要らない。

## FIRSTだけでは省略を判断できない

FOLLOWは「ある規則の直後に現れ得るトークンの集合」。省略・反復は、開始できないからといって任意のトークンで成功してよいわけではない。

完成版の主要な集合を示す。`C`は6個の比較演算子、`A`は`+ -`、`M`は`* / %`。

| 規則 | FIRST | FOLLOW |
| --- | --- | --- |
| stmt | IDENT, while, put, `;` | IDENT, while, put, `;`, `}`, EOF |
| stmts | FIRST(stmt), ε | `}`, EOF |
| put_item | STRING, NUMBER, IDENT, `(` | `,`, `)` |
| cond | NUMBER, IDENT, `(` | `)` |
| expr | NUMBER, IDENT, `(` | `;`, `,`, `)`, C |
| term | NUMBER, IDENT, `(` | A, FOLLOW(expr) |
| primary | NUMBER, IDENT, `(` | M, FOLLOW(term) |

εはトークンではなく、空の列。規則の選択肢ごとにFIRSTが重ならず、空にできる選択肢では他の選択肢のFIRSTとその規則のFOLLOWも重ならないことを確かめる。たとえばcondの省略部分ではCと`)`が重ならない。

完成版の`parseExpr`は対応する演算子以外で戻り、呼出し元が`;`等の必須トークンを検査する。`a=1 2;`ではexprが1を返しても、assignmentの`expect(';')`で失敗する。成功したexpr単体と、入力全体を受理するprogramは区別する。

`parseStmts`は`}`かEOFで戻る。最上位の余分な`}`はprogramのEOF検査、whileの閉じ忘れはwhileの`expect('}')`で拒否する。

## 左因子分解の練習

```text
cond → expr compare_op expr | expr
```

この形では両方のFIRSTが同じ。`cond → expr [compare_op expr]`へ共通部分をくくり出すと、exprを読んだ後で選択できる。左再帰除去と左因子分解は違う操作である。

## whileの評価課題

ASTには条件の木と本体の文リストを保存する。実行では同じ条件ASTを毎回評価し直す。パース時に条件値を計算して保存すると、変数が変わってもループが終わらない。

この言語では正の整数を真とする。while全体でスコープを1個作り、各反復で共有し、終了時に破棄する。外側にある変数への代入は外側を更新する。これはLL(1)の要件ではなく言語の意味規則である。

```sh
./mini-compiler -dump ast < examples/count.mini
./mini-compiler < examples/count.mini
go test ./pkg/minilang -run 'TestWhileAST|TestComparisons|TestEvaluation|TestSyntaxErrors'
```

合格条件: 条件が最初から偽、本体が空、whileの入れ子、閉じ忘れ、`while(a<3<4)`をそれぞれ説明・テストできる。

次は[スタックVM](06-vm.md)。
