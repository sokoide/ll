# 手書きLL(1)パーサー・スタックVM (Go)

小さいEBNFから1トークン先読みの再帰下降パーサーを自分で書き、ASTの構造をテストできるようになるための教材。Goの関数・struct・interface・errorとテストを前提とする。VMはパーサーを理解した後の応用として学ぶ。

## 作る言語のコード例

変数への代入、整数計算、文字列と数値の出力、whileによる繰返しを持つ小さい言語を作る。

```text
a=0;
while(a<3){
    put("a=",a," next=",a+1);
    a=a+1;
}
put("answer=",(1+2)*3);
```

実行結果:

```text
a=0 next=1
a=1 next=2
a=2 next=3
answer=9
```

`put`は引数を区切り文字なしで出力し、最後に改行する。各代入・put文には`;`が必要。whileの閉じ`}`には不要だが、後ろに`;`を書けば空文として受理する。コメントは言語に実装していない。

## 完成版のGrammar（EBNF）

EBNF（Extended Backus–Naur Form、拡張BNF）は、言語の構文を規則として書くための記法。BNFに反復や省略の表記を加え、文法を短く表せるようにしたもの。EBNFには表記の流儀があり、この教材では次の記号を使う。

`,`は連結、`|`は選択、`{ ... }`は0回以上の反復、`[ ... ]`は省略可能、引用符はソース中の記号やキーワードを表す。NUMBERは整数、IDENTは識別子、STRINGは文字列、EOFは入力終端。

Yaccも文法規則を書くが、パーサー生成ツールへの入力として使い、BNFに近い規則に型・優先順位・ASTを作る処理などを加える。たとえば文の反復は、この教材のEBNFでは`stmts = { stmt } ;`、Yaccでは`stmts : /* empty */ | stmts stmt ;`と再帰で書ける。EBNFの`{ ... }`は反復を表す一方、Yaccの`{ ... }`は規則に対応して実行するコード（意味アクション）を囲む。文法中の反復・省略と、手書きパーサーの`for`・`if`を対応させて読むと分かりやすい。

### なぜこのLL(1)教材ではEBNFを使うのか

LL(1)は「次の1トークンで生成規則を選べる」という文法の性質・解析方式であり、EBNFは文法の書き方。この2つは別で、LL(1)だからEBNFが必須になるわけではない。理論や予測解析表の説明では、`E → T R`、`R → '+' T R | ε`のようなBNF形式の生成規則も使う。

手書きの再帰下降では、同じ反復を`E = T, { "+", T } ;`と書くと、最初のTを読む関数呼出しと、その後の`for`が直接対応する。省略の`[ ... ]`も、先読みで進むか決める`if`へ対応させやすい。この教材ではその読みやすさを理由にEBNFを維持し、FIRST／FOLLOWを考える際は必要に応じてBNFへ展開する。EBNFと再帰下降コードの対応は[Cornell大学の講義資料（スライド14〜16）](https://www.cs.cornell.edu/courses/cs412/2001SP/lectures/lec06.pdf)でも説明されている。

ただし、EBNFで書くだけでLL(1)になるわけではない。選択肢の先頭や、反復・省略を終える条件が1トークンで区別できるかを別途確認する。たとえば`primary = IDENT | IDENT, "(", ")" ;`は両方がIDENTで始まるため、そのままでは選べず、`primary = IDENT, [ "(", ")" ] ;`のような左因子分解が必要になる。

```ebnf
program     = stmts, EOF ;
stmts       = { stmt } ;
stmt        = assignment | while_stmt | put_stmt | ";" ;
assignment  = IDENT, "=", expr, ";" ;
while_stmt  = "while", "(", cond, ")", "{", stmts, "}" ;
put_stmt    = "put", "(", put_list, ")", ";" ;
put_list    = put_item, { ",", put_item } ;
put_item    = STRING | expr ;
cond        = expr, [ compare_op, expr ] ;
compare_op  = "==" | "!=" | ">=" | ">" | "<=" | "<" ;
expr        = term, { ("+" | "-"), term } ;
term        = primary, { ("*" | "/" | "%"), primary } ;
primary     = NUMBER | IDENT | "(", expr, ")" ;
```

字句規則も含む正本は[grammar.ebnf](grammar.ebnf)。NUMBERは数字1個以上、IDENTはASCII英字または`_`で始まり以後は英数字と`_`、STRINGはシングル／ダブルクォートで囲む。`put`と`while`は予約語。トークン間の空白・タブ・改行は読み飛ばす。文字列エスケープはない。

完成版は[参照元goyacc4](https://github.com/sokoide/bison_flex_1/tree/main/goyacc4)を基にしている。学習のため文字列はputの直接引数だけに文法を絞った。参照元リポジトリを用意しなくても学べる。

## ビルドして観察する

```sh
make
./mini-compiler -dump tokens < examples/arithmetic.mini
./mini-compiler -dump ast < examples/arithmetic.mini
./mini-compiler -dump bytecode < examples/arithmetic.mini
./mini-compiler -engine tree < examples/count.mini
./mini-compiler -engine vm < examples/count.mini
make testall
```

`examples/arithmetic.mini`は`put(1+2*3);`で、結果は7。`-dump`は実行せず、指定した中間結果を表示する。

```text
Put
  Binary '+'
    Number 1
    Binary '*'
      Number 2
      Number 3
```

ASTから生成する命令列:

```text
0000 PUSH 1
0001 PUSH 2
0002 PUSH 3
0003 BINARY '*'
0004 BINARY '+'
0005 PRINT_INT
0006 NEWLINE
```

先に乗算の木を作り、それを加算の右の子にする。VMは木を直接たどらず、この命令列を実行する。

## 学習の進め方

最初に完成版を実行し、その後は別のGoモジュールで各ステップの小さい文法を実装する。各ページに作業用文法、読む関数、実装課題、合格条件を置いている。完成版CLIがステップごとの文法に切り替わるわけではない。

| ステップ | 作るもの | 学ぶ仕組み |
| --- | --- | --- |
| [1. 整数1個](docs/steps/01-number.md) | NUMBERとEOFを読む | lexer、トークン、advance / expect |
| [2. 加減算](docs/steps/02-addition.md) | 左結合のBinary AST | 左再帰除去、反復、不変条件 |
| [3. 優先順位と括弧](docs/steps/03-precedence.md) | expr / term / primary | 関数の階層、消費範囲、AST形状 |
| [4. 文と変数](docs/steps/04-statements.md) | 代入・put・文リスト | FIRST、文と式、解析と評価の分離 |
| [5. while](docs/steps/05-while.md) | 比較・ブロック・繰返し | FOLLOW、省略、左因子分解、スコープ |
| [6. スタックVM](docs/steps/06-vm.md) | AST→命令列→実行 | 後行順、PC、オペランドスタック、バックパッチ |

仕上げは[単項マイナス・if・呼出しの拡張課題](docs/exercises.md)。[レビューと改善理由](docs/review.md)には教材の設計判断と検証範囲をまとめた。

## ディレクトリ構成

```text
cmd/mini-compiler/main.go   CLIのエントリーポイント
pkg/minilang/
  minilang.go              公開API（Parse、Program.Execute）
  parser.go                手書きLL(1)パーサー
  lexer.go                 字句解析器
  ast.go                   文と式を区別したAST
  eval.go                  ASTを直接たどる評価器
  operations.go            整数演算の意味規則
  scope.go                 スコープ管理
  compiler.go              ASTから命令列への変換
  vm.go                    スタックVM・逆アセンブル
  inspect.go               トークン・ASTの表示
  parser_test.go           構文・評価・エラーのテスト
  minilang_test.go         公開APIのテスト
  vm_test.go               両エンジンの期待出力・エラー検証
  teaching_test.go         教材のAST・命令列・例題を検証
docs/steps/                小さい文法から進む学習手順
examples/                  実行例と期待出力
grammar.ebnf               構文と字句規則の文書
Makefile                   ビルド・テスト
```

Goが必須とする唯一のディレクトリ構成はないが、このプロジェクトでは`cmd`にCLI、`pkg`に他のGoコードからimportできるライブラリを配置する。[parser.go](pkg/minilang/parser.go)の各関数に対応する生成規則を記載。字句解析の行・列は1始まり、列はUnicode文字単位。

## ライブラリとして使う

```go
import (
    "os"
    "sokoide.com/ll/pkg/minilang"
)

func example() error {
    program, err := minilang.Parse(`a=40;put("ans=",a+2);`)
    if err != nil {
        return err
    }
    return program.Execute(os.Stdout)
}
```

`Parse`は入力全体を検証し、`Execute`は指定された出力先へ書き込む。変数スコープは実行ごとに作成するので、同じプログラムを再実行しても状態は持ち越さない。VMを使う場合は`program.Compile()`で得た`Bytecode`の`Execute(writer)`を呼ぶ。ライブラリはログを出さずエラーを返し、CLIがlogrusで表示する。

## MakefileとCLI

```sh
make generate  # 互換ターゲット。手書きなので生成不要と表示
make           # ./mini-compiler をビルド
make test      # Goテスト
make testcli   # CLIテスト（既定はtree。ENGINE=vmも指定可能）
make testall   # Goテスト・両エンジンのCLIテスト・go vet
make run       # 標準入力から実行
make clean     # バイナリのみ削除。ソースと文法は残す
```

既定のエンジンは`tree`。`-dump tokens`は字句解析だけなので構文の正しさまでは判定しない。`-dump ast`と`-dump bytecode`は構文解析も行う。`-logLevel TRACE|DEBUG|INFO|WARN|ERROR`も受け付ける（既定はdebug）。出力はstdout、エラーはstderr。構文・実行・入出力エラーでは非ゼロで終了する。

## 完成版の意味規則

これらはこの言語の選択で、LL(1)の必須条件ではない。

- 整数の`+ - * / %`を扱う。`* / %`は`+ -`より優先し、同じ優先順位では左結合。
- 比較はwhile条件に1個だけ書ける。比較の括弧による入れ子は不可。正の整数を真とし、0以下は偽。
- while全体でスコープを1個作り、全反復で共有し、終了時に破棄する。外側の既存変数への代入はその変数を更新する。
- 未定義変数を読むと現在のスコープに0で作る。
- STRINGはputの直接引数に限る。代入・演算・条件や括弧内の文字列は構文エラー。
- 直接出力した数値リテラルは先頭ゼロを維持する。`put(001);`は001、`put((001));`は数値評価を経由して1。
- 整数はGoのint。ゼロ除算・剰余ゼロは実行時エラー。
- 単項演算、独立したブロック、式だけの文、関数はない。単項マイナス等は拡張課題にしている。

未閉鎖文字列、不正文字、余分な末尾トークンは拒否する。CRLFも空白として受理し、文字列内の改行は許可する。パーサー生成ツールは不要。

## 左再帰とLL(1)・LALRの違い

参照先のYacc文法には、次のような左再帰がある。生成規則の右辺が、自分自身から始まる形を直接左再帰という。

```text
expr  → expr '+' expr | expr '*' expr | NUMBER | ...
stmts → stmts stmt | ε
```

この規則をそのまま再帰下降の関数にすると、`parseExpr()`が最初のトークンを消費する前に`parseExpr()`を呼び続け、無限再帰になる。先読みを増やすだけでは、この呼出し方の問題は解消しない。左再帰は間接的にも発生する（`A → B x`、`B → A y | z`）。

YaccのLALRパーサーは、トークンを読み、スタック上の記号を生成規則へ還元するボトムアップ方式なので、左再帰を扱える。ただし、上の`expr op expr`は演算子の優先順位・結合方向が曖昧であり、左再帰を許すことだけでは解決しない。参照先は`%left '+' '-'`と`%left '*' '/' '%'`で競合を解決している。

LL(1)は、左から入力を読み、左端導出を1トークン先読みで選ぶトップダウン方式。この実装では、次のように演算子の優先順位を文法の階層にし、左再帰を反復に変換する。

```ebnf
expr = term, { ("+" | "-"), term } ;
term = primary, { ("*" | "/" | "%"), primary } ;
stmts = { stmt } ;
```

一般に`A → A α | β`は`A → β A'`、`A' → α A' | ε`へ変換でき、EBNFでは`A = β, { α }`と書ける。ただし、言語として同じ列を受理することと、ASTの結合方向を保つことは別。`parser.go`では反復のたびに、それまでの左辺を新しいノードの左の子にする。

```text
20 - 5 - 3
→ BinOp(BinOp(20, '-', 5), '-', 3)
→ (20 - 5) - 3 = 12
```

残りの式を右へ再帰させて`20 - (5 - 3)`というASTを作ると結果は18になり、元の左結合と異なる。除算も同様なので、テストで両方を確認している。

また、左再帰を除去するだけで、すべての文法がLL(1)になるわけではない。選択肢の先頭トークンが重なる場合は左因子分解などが必要で、省略可能な規則ではFIRST集合（開始できるトークン）とFOLLOW集合（直後に来るトークン）の衝突も調べる。この文法では次のように1トークンで判断できる。

- `stmt`: IDENT、`while`、`put`、`;`で選択肢が分かれる。
- `primary`: NUMBER、IDENT、`(`で選択肢が分かれる。
- `expr`と`term`の反復: 対応する演算子なら継続し、それ以外なら呼出し元へ戻る。
- `cond`の省略部分: 比較演算子なら右辺を読み、`)`なら比較なしで終える。
- `stmts`: 文の開始トークンなら継続し、`}`またはEOFで終える。最上位の`}`やブロック内のEOFは、その後の呼出し元の必須トークン検査で拒否する。

LALRも先読みが1トークンであることと、LL(1)で書けることは同義ではない。ボトムアップで既に読んだ部分をスタックに保持して判断するLALRと、展開する生成規則を先に選ぶLL(1)では、扱える文法の範囲が異なる。


## 適用範囲

多倍長整数や算術オーバーフローの検出は実装しない。無限ループの命令数制限、深い括弧の再帰制限、複数エラーからの回復も扱わない。VMはコンパイラが生成したメモリ内命令列を実行する教材で、任意の外部bytecodeを実行するサービスではない。
