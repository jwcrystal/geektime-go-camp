# ORM 之 Subquery 子查詢

## 目標

替現有的查詢語句，加上子查詢 `Subquery` 功能。

### 功能需求

支持下方三種情況：

- `WHERE`: 用來篩選特別的條件，例如平均、中位數等
- `FROM`: 重塑表格的格式、或預先進行計算或濾掉資料
- `JOIN`: 進行預先計算或過濾資料。

### 單元測試

- 用在 FROM 部分
  - 子查詢單獨使用
  - 子查詢和普通表 JOIN，子查詢在左邊或者在右邊
  - 子查詢和子查詢 JOIN
  - 子查詢和 JOIN 查詢進一步 JOIN，子查詢可以在左邊也可以在右邊

- 在 WHERE 部分
  - EXIST 和 NOT EXIST
  - ALL，SOME 和 ANY
  - IN 和 NOT IN

- 使用子查詢指定列：
  - 指定非法列
  - 指定列用於 WHERE，HAVING 或者 ON 部分
  - 指定列用於 SELECT 部分

## 思路

之前作業都是從自己撰寫的代碼中，再針對作業需求去寫，相對較為簡單。此次因為與原本自身代碼進度有不小的落差，加上此次作業相對複雜，因此直接採用大明老師提供的作業代碼去撰寫，所以需要先釐清原有的代碼思路，才會比較容易去寫作業部分，這部分也不少坑啊...

依序從較為單一的子查詢著手，循序漸進，跟據原有 SQL 的子查詢模式去使用且設計 API：

1）用在 `FROM` 部分：

```go
sub := NewSelector[OrderDetail](db).AsSubquery("sub")
NewSelector[Order](db).From(sub)
```

2）和`WHERE`查詢混用：

```go
sub := NewSelector[OrderDetail](db).Select(C("OrderId")).AsSubquery("sub")
				return NewSelector[Order](db).Where(Exist(sub))
```

3）和` JOIN` 查询混用：

```go
t1 := TableOf(&Order{})
sub := NewSelector[OrderDetail](db).AsSubquery("sub")
NewSelector[Order](db).Select(sub.C("ItemId")).From(t1.Join(sub).On(t1.C("Id").EQ(sub.C("OrderId"))))
```

## 問題點

- 如何 整合 subquery Build 構造 SQL 結果？

  - ```go
    // buildSubquery 構建子查詢 SQL，
    // useAlias 決定是否顯示別名，即使有別名
    func (b *builder) buildSubquery(sub Subquery, useAlias bool) error {
    	query, err := sub.q.Build()
    	if err != nil {
    		return err
    	}
    	b.sb.WriteByte('(')
    	// 拿掉最後 ';'
    	b.sb.WriteString(query.SQL[:len(query.SQL)-1])
      // 因為有 build() ，所以理應 args 也需要跟 SQL 一起處理
      if len(query.Args) > 0 {
    		b.addArgs(query.Args...)
    	}
    	b.sb.WriteByte(')')
    	if useAlias {
    		b.sb.WriteString(" AS ")
    		b.quote(sub.alias)
    	}
    	return nil
    }
    ```
  
- 構造 SQL 打印多次

  - 如下方單元測試 `some and any`，Build() 結果沒有覆蓋，而是 append 在 SQL 後面，導致重複兩次

    - Solve： 在 `Build()`中重置 `strings.builder` 的對象即可
  
  - ```go
    "SELECT * FROM `order` WHERE (`id` > SOME (SELECT `order_id` FROM `order_detail`)) AND (`id` < ANY (SELECT `order_id` FROM `order_detail`));"
    ==> 會變成下方樣式
    "SELECT * FROM `order` WHERE (`id` > SOME (SELECT `order_id` FROM `order_detail`)) AND (`id` < ANY (SELECT `order_id` FROM `order_detail;SELECT `order_id` FROM `order_detail`));"
    ```


  ```go
  			name: "some and any",
  			q: func() QueryBuilder {
  				sub := NewSelector[OrderDetail](db).Select(C("OrderId")).AsSubquery("sub")
  				return NewSelector[Order](db).Where(C("Id").GT(Some(sub)), C("Id").LT(Any(sub)))
  			}(),
  			wantQuery: &Query{
  				SQL: "SELECT * FROM `order` WHERE (`id` > SOME (SELECT `order_id` FROM `order_detail`)) AND (`id` < ANY (SELECT `order_id` FROM `order_detail`));",
  			},
  ```

- JOIN & Subquery 構造 SQL 表達式不支持

  - 因為 ColName() 沒有斷言支持 `Subquery` 型態 

  - ```go
    func (b *builder) colName(table TableReference, fd string) (string, error) {
    	switch tab := table.(type) {
    	...
    	case Subquery:
    		if len(tab.columns) > 0 {
    			for _, c := range tab.columns {
            // 如果 column 有別名與字段相等，直接採用字段
    				if c.selectedAlias() == fd {
    					return fd, nil
    				}
    
    				if c.fieldName() == fd {
    					return b.colName(c.target(), fd)
    				}
    			}
    			return "", errs.NewErrUnknownField(fd)
    		}
    		return b.colName(tab.entity, fd)
    		...
    	}
    }
    ```

  - ```go
    // 這邊 Column 的 Table 直接吃 TableReference 類型的結構體，需要在 buildColumn 斷言解析 Subquery 型態，不然會出現 不支持的表達式 Expression
    func (s Subquery) C(name string) Column {
    	return Column{
        table: s, // 此處如果採用 s.entity (為 TableReference)，就會印出下方輸出的結果
    		name:  name,
    	}
    }
    
    NewSelector[Order](db).Select(sub.C("ItemId")).From(t1.Join(sub).On(t1.C("Id").EQ(sub.C("OrderId")))).Where()
    ==> 
    輸出
    SQL: "SELECT item_id` FROM (`order` JOIN (SELECT * FROM `order_detail`) AS `sub` ON `id` = `order_id`);"
    正確：
    SQL: "SELECT `sub`.`item_id` FROM (`order` JOIN (SELECT * FROM `order_detail`) AS `sub` ON `id` = `sub`.`order_id`);"
    ```