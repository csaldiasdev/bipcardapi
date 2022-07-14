package main

import "bipcardapi/internal/server"

func main() {
	srv := server.NewHTTPServer(":8080")
	srv.ListenAndServe()
}
