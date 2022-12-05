# ORM 之 Delete 刪除功能支持 

> SQL語句，複雜度為`SELECT` > `INSERT` > `UPDATE` > `DELETE`
> - Select: 有很多種變化，以及`join`等複雜實現
> - Insert: 有`InsertOrUpdate`等用法
> - Update: 有指定`Column`等用法
> - Delete: 相對前面幾個，基本只有`Where`語法較常用到
>   - `LIMIT` 和 `OFFSET`
>   - `RETURNING`
>   - 子查詢 
>   - 上面三個特性，很少使用，甚至涉及到各語法的支援（老師所說的方言）


## 目標

- 支持`DELETE`功能
- Unit Test 需要考慮下面幾點
  - 是否使用`From`指定 table name
  - 是否使用`Where`,及是否有搭配`And`、`Or`、`Not`

## 思路

大體跟老師說明的`Seletor`大同小異，藉由此次練習及複習`builder`模式，和`BuildExpression`的思路


## Benchmark

- TBD，待結果集後測試