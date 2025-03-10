package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func gethandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "ok")
}

func main() {

	router := mux.NewRouter()
	router.Methods(http.MethodGet).Path("/get").HandlerFunc(gethandler)

	router.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./"))))
	http.Handle("/", router)
	fmt.Println("Starting Server")
	//err := http.ListenAndServe(":8080", nil)

	var server *http.Server
	server = &http.Server{
		Addr:    ":8081",
		Handler: router,
	}
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
