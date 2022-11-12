package main

import (
	"geektime-go/web"
	"log"
)

func main() {

	s := &web.HttpServer{}

	s.Get("/", func(ctx *web.Context) {
		_, err := ctx.Res.Write([]byte("hello, world"))
		if err != nil {
			log.Fatal(err)
			return
		}
	})

	err := s.Start(":8081")
	if err != nil {
		log.Fatal(err)
	}
	//helloHandler := func(w http.ResponseWriter, req *http.Request) {
	//	io.WriteString(w, "Hello, world!\n")
	//}
	//http.HandleFunc("/hello", helloHandler)
	//log.Fatal(http.ListenAndServe(":8080", nil))
}
