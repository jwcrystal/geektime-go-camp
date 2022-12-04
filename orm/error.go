package orm

import "geektime-go/orm/internal/errs"

// 透過下面形式將內部錯誤，暴露再外部
var ErrNoRows = errs.ErrNoRows
