# 子查詢

## 需求分析

在實際使用過程中，用戶經常會直接在查詢的 `Expression` 表達式中再採用查詢語句，這也被稱為子查詢 `Subquery`。

- 子查詢基本語法
```sql
SELECT 欄位名稱1,欄位名稱2,...,欄位名稱n 
FROM 資料表名稱1
WHERE 欄位名稱 = 
(SELECT 欄位名稱 FROM 資料表名稱2 WHERE 條件)
```

### 場景分析

每一個子查詢都是一個Select指令，必須用小括號 `()` 包起來，能夠針對不同資料表進行查詢。

如果SQL查詢指令內有子查詢，首先處理的是子查詢條件，然後再依子查詢取得的條件值來處理主查詢，然後就去得最後的查詢結果。

下方為使用子查詢的情況：

- `WHERE`: 用來篩選特別的條件，例如平均、中位數等
    - 希望藉由**某些條件來過濾資料**的情況，我們就會把子查詢放在 `WHERE` 之後
    - e.g. 查找比平均身高高的學生
    ```sql
    SELECT id, height
    FROM Students
    WHERE height > (SELECT AVG(height) FROM Students);
    ```
- `FROM`: 重塑表格的格式、或預先進行計算或濾掉資料
    -  **預先過濾篩選表格**的情況下，就會把子查詢放在 `FROM` 之後
    -  e.g. 查找比平均身高高，且體重超過80公斤的學生
    ```sql
    SELECT id, height, weight
    FROM (SELECT id, height, weight 
    FROM Students WHERE weight > 80)
    WHERE height > (SELECT AVG(height) FROM Students);
    ```
- `JOIN`: 進行預先計算或過濾資料
    - **預先與其他表格交互過濾篩選表格**的情況，**但不想讓整個表格過於複雜**，就會把子查詢與 `JOIN` 一起使用
    - 因為顯示欄位過多情況下，如果只用 `FROM` 會讓子表格變得過於長
    ```sql
    SELECT id, height, weight
    FROM Students AS s 
    INNER JOIN (SELECT id, height, weight 
    FROM Students WHERE weight > 80) AS sub
    ON s.id = sub.id
    WHERE height > (SELECT AVG(height) FROM Students);
    ```
- `SELECT`: 快速計算單一數值，特別是多表格存取時
    - 針對特定數值的情況下，會採用 `SELECT` 子查詢
    - 單一子查詢，需要留意下列兩種情況：
        - `SELECT` 子查詢僅能用於單一數值，否則就會跳出 `Error`
        - 若 `SELECT` 子查詢有特定條件，則在子查詢和總表都要記得加上條件。
    ```sql
    # 示意，例子不好
    # 或許換成 weight > 80 換成 BMI > 28 
    SELECT id, height, (SELECT weight FROM Students) AS weight
    FROM Students
    WHERE height > (SELECT AVG(height) FROM Students)
    AND weight > 80；
    ```

#### 功能需求
支持下方三種情況：
- `WHERE`: 用來篩選特別的條件，例如平均、中位數等
- `FROM`: 重塑表格的格式、或預先進行計算或濾掉資料
- `JOIN`: 進行預先計算或過濾資料。

## 行業方案

- GORM
    - 支持兩種情況下使用子查詢：`WHERE`、`FROM`
    - `HAVING` 同 `WHERE` 表達式
        - subquery in `WHERE` clause with the method Table
            ```go
            db.Where("amount > (?)", db.Table("orders").Select("AVG(amount)")).Find(&orders)
            // SELECT * FROM "orders" WHERE amount > (SELECT AVG(amount) FROM "orders");
            
            subQuery := db.Select("AVG(age)").Where("name LIKE ?", "name%").Table("users")
            db.Select("AVG(age) as avgage").Group("name").Having("AVG(age) > (?)", subQuery).Find(&results)
            // SELECT AVG(age) as avgage FROM `users` GROUP BY `name` HAVING AVG(age) > (SELECT AVG(age) FROM `users` WHERE name LIKE "name%")
            ```
        - subquery in `FROM` clause with the method Table
            ```go
            db.Table("(?) as u", db.Model(&User{}).Select("name", "age")).Where("age = ?", 18).Find(&User{})
            // SELECT * FROM (SELECT `name`,`age` FROM `users`) as u WHERE `age` = 18
            
            subQuery1 := db.Model(&User{}).Select("name")
            subQuery2 := db.Model(&Pet{}).Select("name")
            db.Table("(?) as u, (?) as p", subQuery1, subQuery2).Find(&User{})
            // SELECT * FROM (SELECT `name` FROM `users`) as u, (SELECT `name` FROM `pets`) as p
            ```
    - 沒有明確的子查詢方法，透過 `Expression` 去解析支持
    - the method Table on GORM
        - 採用 `Raw Expression` 解析:
    ```go
    TableExpr =  &clause.Expr{SQL: name, Vars: args}
    ```
    ```go
    // Table specify the table you would like to run db operations
    //
    //	// Get a user
    //	db.Table("users").take(&result)
    func (db *DB) Table(name string, args ...interface{}) (tx *DB) {
    	tx = db.getInstance()
    	if strings.Contains(name, " ") || strings.Contains(name, "`") || len(args) > 0 {
    	   // 採用 Raw Expression 
    		tx.Statement.TableExpr = &clause.Expr{SQL: name, Vars: args}
    		if results := tableRegexp.FindStringSubmatch(name); len(results) == 2 {
    			tx.Statement.Table = results[1]
    		}
    	} else if tables := strings.Split(name, "."); len(tables) == 2 {
    	
    		tx.Statement.TableExpr = &clause.Expr{SQL: tx.Statement.Quote(name)}
    		tx.Statement.Table = tables[1]
    	} else if name != "" {
    		tx.Statement.TableExpr = &clause.Expr{SQL: tx.Statement.Quote(name)}
    		tx.Statement.Table = name
    	} else {
    		tx.Statement.TableExpr = nil
    		tx.Statement.Table = ""
    	}
    	return
    }
    
    // Expr raw expression
    type Expr struct {
    	SQL                string
    	Vars               []interface{}
    	WithoutParentheses bool
    }
    ```
- BEEGO
    - 設計在 `QueryBuilder` 接口
    - 特別的是，有一個明確的子查詢方法為 `Subquery`
    - 風格都是用戶自己拼接
    ``` go
    // QueryBuilder is the Query builder interface
    type QueryBuilder interface {
    	Select(fields ...string) QueryBuilder
    	ForUpdate() QueryBuilder
    	From(tables ...string) QueryBuilder
    	InnerJoin(table string) QueryBuilder
    	LeftJoin(table string) QueryBuilder
    	RightJoin(table string) QueryBuilder
    	On(cond string) QueryBuilder
    	Where(cond string) QueryBuilder
    	And(cond string) QueryBuilder
    	Or(cond string) QueryBuilder
    	In(vals ...string) QueryBuilder
    	OrderBy(fields ...string) QueryBuilder
    	Asc() QueryBuilder
    	Desc() QueryBuilder
    	Limit(limit int) QueryBuilder
    	Offset(offset int) QueryBuilder
    	GroupBy(fields ...string) QueryBuilder
    	Having(cond string) QueryBuilder
    	Update(tables ...string) QueryBuilder
    	Set(kv ...string) QueryBuilder
    	Delete(tables ...string) QueryBuilder
    	InsertInto(table string, fields ...string) QueryBuilder
    	Values(vals ...string) QueryBuilder
    	Subquery(sub string, alias string) string // here
    	String() string
    }
    // 實現
    // Subquery join the sub as alias
    func (qb *MySQLQueryBuilder) Subquery(sub string, alias string) string {
    	return (*orm.MySQLQueryBuilder)(qb).Subquery(sub, alias)
    }
    // Subquery join the sub as alias
    func (qb *MySQLQueryBuilder) Subquery(sub string, alias string) string {
    	return fmt.Sprintf("(%s) AS %s", sub, alias)
    }
    ```

- ENT
    - 子查詢包含在 `Select Expression` 裡面
    - `SelectExpr` 使用自定義表達式列表更改 `SELECT` 語句的列選擇
    ```go
    type Selector struct {
    	Builder
    	// ctx stores contextual data typically from
    	// generated code such as alternate table schemas.
    	ctx       context.Context
    	as        string
    	selection []any //here
    	from      []TableView
    	joins     []join
    	where     *Predicate
    	or        bool
    	not       bool
    	order     []any
    	group     []string
    	having    *Predicate
    	limit     *int
    	offset    *int
    	distinct  bool
    	setOps    []setOp
    	prefix    Queries
    	lock      *LockOptions
    }
    ```
    ```go
    // example
    client.User.
		Query().
		Modify(func(s *sql.Selector) {
			subQuery := sql.SelectExpr(sql.Raw("1")).As("s")
			s.Select("*").From(subQuery)
		}).
		Int(ctx)
	
	// SelectExpr is like Select, but supports passing arbitrary
    // expressions for SELECT clause.
    func SelectExpr(exprs ...Querier) *Selector {
    	return (&Selector{}).SelectExpr(exprs...)
    }
    // SelectExpr changes the columns selection of the SELECT statement
    // with custom list of expressions.
    func (s *Selector) SelectExpr(exprs ...Querier) *Selector {
    	s.selection = make([]any, len(exprs))
    	for i := range exprs {
    		s.selection[i] = exprs[i]
    	}
    	return s
    }
    ```


## 設計

可以復用 `JOIN` 查詢裡面的 `TableReference` 抽象，提供一個代表子查詢的抽象
```go
type Subquery struct {
	// 使用 QueryBuilder 僅是為了讓 Subquery 可以是非泛型
	q QueryBuilder
	alias string
}
```

特殊支持在於， `Subquery` 理論上應該使用 `Selector` 的，但是因為本身 `Selector` 是泛型的，而我們並不希望在 `Subquery` 中引入類型參數，所以直接使用 `QueryBuilder`。

Subquery 會實現下方接口：
- `TableReference`: 保證了子查詢可以用在 `FROM` 部分
- `Expression`: 保證了子查詢可以用在 `WHERE` 部分

從 `Selector` 中構建出來的
```go
func (s *Selector[T]) AsSubquery(alias string) Subquery {
   panic("implement me")
}
```
### 詳細設計

#### 指定列

類似於 `JOIN` 查詢，我們可以在 `Subquery` 結構體上定義一個指定列的方法：
```go
func (s Subquery) C(name string) Column {
	return Column{
		table: s,
		name:  name,
	}
}
```
那麼這個返回的 `Column` 就可以被用在 `SELECT` 部分，或者 `WHERE` 部分。

如果 `Column` 直接傳入 `TableReference`，需要提供 `buildColumn()` 斷言解析，不然構造 SQL 無法辨別，會導致出現不支持表示式的問題

```go
func (b *builder) colName(table TableReference, fd string) (string, error) {
	switch tab := table.(type) {
	...
	case Subquery:
    panic("implement me")
		return "", nil
		...
	}
}
```

#### JOIN 查詢

類似於 Table 結構體，我們在 Subquery 上定義 `Join`, `LeftJoin` 和 `RightJoin` 方法：
```go
func (s Subquery) Join(target TableReference) *JoinBuilder {
    panic("implement me")
}

func (s Subquery) LeftJoin(target TableReference) *JoinBuilder {
	panic("implememnt me")
}

func (s Subquery) RightJoin(target TableReference) *JoinBuilder {
	panic("implement me")
}
```

定義了之後，用戶就可以通過 JoinBuilder 將子查詢和普通的表、JOIN 查詢關聯在一起。

#### Any，All 和 Some

我們只需要直接定義三個方法，它們都接收子查詢來作為輸入：
```go
type SubqueryExpr struct {
	s Subquery
	// 謂詞，ALL，ANY 或者 SOME
	pred string
}

func Any(sub Subquery) SubqueryExpr {
	return SubqueryExpr{
		s: sub,
		pred: "ANY",
	}
}

func All(sub Subquery) SubqueryExpr {
	return SubqueryExpr{
		s: sub,
		pred: "ALL",
	}
}

func Some(sub Subquery) SubqueryExpr {
	return SubqueryExpr{
		s: sub,
		pred: "SOME",
	}
}
```
注意，這裡的設計我們並沒有把謂詞直接定義在 `Subquery` 上，而是額外定義了一個結構體。是因為我們希望用戶不能將這個 `SubqueryExpr` 用在 `FROM` 部分，避免類似於 `Aggregate` 別名處理的尷尬問題。

`SubqueryExpr` 本身會實現 `Expression` 接口，這也確保了 `SubqueryExpr` 可以被用在 `WHERE` 部分，以構建複雜的查詢條件。

#### EXIST 和 NOT EXIST

類似於 `EQ` 之類的方法，我們可以定義新的方法：
```go
func Exist(sub Subquery) Predicate {
	return Predicate{
		op: opExist,
		right: sub,
	}
}
```

和 `Eq` 之類的方法比起來，不同之處就是 `Exist` 是 `left` 是沒有取值的，並且 `op` 定義了一個新的 `opExist`。
顯然，`Not Exist` 可以復用已有的 `Not` 方法，而不需要我們額外定義一個新的方法。


## 例子

用在 FROM 部分：
```go
sub := NewSelector[OrderDetail](db).AsSubquery("sub")
NewSelector[Order](db).From(sub)
```

和 JOIN 查询混用：
```go
t1 := TableOf(&Order{})
sub := NewSelector[OrderDetail](db).AsSubquery("sub")
NewSelector[Order](db).Select(sub.C("ItemId")).From(t1.Join(sub).On(t1.C("Id").EQ(sub.C("OrderId"))))
```

這裡我們可以看到使用子查詢來指定列名，並且用在 `ON` 部分
在 WHERE 中使用子查询的例子：

```go
// NOT EXIST
sub := NewSelector[OrderDetail](db).Select(C("OrderId")).AsSubquery("sub")
NewSelector[Order](db).Where(Not(Exist(sub)))

// SOME 和 ANY
sub := NewSelector[OrderDetail](db).Select(C("OrderId")).AsSubquery("sub")
NewSelector[Order](db).Where(C("Id").GT(Some(sub)), C("Id").LT(Any(sub)))
```

## 測試
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

### 集成測試

- 考慮在單元測試中出現的測試用例，但是並不需要測試非法用例，例如指定非法列。