# ORM 之 Select 多功能支持 

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

- 支持`SELECT`多功能

  - GROUP BY , HAVING 
  - ORDER BY
  - [ ] OFFSET  x LIMIT y
  - [ ] 支持 `LIMIT` 後，原本 `GET` 方法設計為 `LIMIT 1`
  - [ ] HAVING，可提供 2 種寫法

  ```sql
  # 基本會支持
  1) SELECT * FROM xx  GROUP BY aa HAVING(AVG(column_b)) < ?
  # 附加支持
  2) SELECT AVG(column_b) AS avg_b FROM xx GROUP BY aa HAVING avg_b <?
  ```

- Unit Test 需要考慮下面幾點
  - GROUP BY：
    - 單個列
    - 多個列
    - 非法列
  - HAVING：
    - 單個查詢條件
    - 多個查詢條件
    - 復合查詢條件（AND，OR）
    - 使用聚合函數的查詢條件
  - ORDER BY：
    - 單個列
    - 多個列，並且指定不同的升降序
    - 非法列
  - OFFSET 和 LIMIT
    - 單獨使用 OFFSET
    - 單獨使用 LIMIT
    - 混合使用
    - 使用負數，我們並不做校驗

## 思路

藉由此次練習及複習之前學過的內容，以及**學習重構思路**。

可選部分儘量跟上，因為時間問題，不確定是否能加上，因為需要重構代碼。

