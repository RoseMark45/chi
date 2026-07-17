package main

import (
	"fmt"
	"net/http"
)

func main() {
	r := NewRouter()
	r.Get("/users/{userID}/profile", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "user=%s", URLParam(req, "userID"))
	})
	fmt.Println("chi router ready")
	_ = http.ListenAndServe(":3333", r)
}
