package main

import (
	"database/sql"
	"encoding/xml"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	_ "github.com/mattn/go-sqlite3"
)

//Tvs xml array
type Tvs struct {
	XMLName xml.Name `xml:"tvs"`
	Tv      []struct {
		Text int64 `xml:",chardata"`
		ID   int64 `xml:"id,attr"`
	} `xml:"tv"`
}

func main() {
	db, err := sql.Open("sqlite3", "./db.db")

	if err != nil {
		log.Println(err)
	}

	router := mux.NewRouter()

	server := &Server{
		db:     db,
		router: router,
	}

	server.Route()

	hhtpSrv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	hash := make(chan [32]byte)
	fileRead := make(chan bool)

	go server.ReturnsChecker(hash, fileRead)

	go func() {
		hashSet := make(map[[32]byte]struct{})
		for {
			sha := <-hash
			if _, ok := hashSet[sha]; ok {
				fileRead <- true
			} else {
				hashSet[sha] = struct{}{}
				fileRead <- false
			}
		}
	}()

	log.Fatal(hhtpSrv.ListenAndServe())

}
