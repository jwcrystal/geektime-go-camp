package web

// type Middlewares []HandleFunc
// 函數式的責任鏈模式， or 洋蔥模式
type Middleware func(next HandleFunc) HandleFunc
