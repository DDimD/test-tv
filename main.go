package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	_ "github.com/mattn/go-sqlite3"
)

// Server struct
type Server struct {
	db     *sql.DB
	router *mux.Router
}

// Route create handlers
func (srv *Server) Route() {
	srv.router.HandleFunc("/", srv.IndexHandler)
	srv.router.NotFoundHandler = http.HandlerFunc(notFound)
	srv.router.HandleFunc("/tv/{id:[0-9]+}", srv.Handler).Methods("GET", "PUT", "DELETE")
	srv.router.HandleFunc("/tv/", srv.PostHandler).Methods("POST")
	srv.router.HandleFunc("/tv/", srv.GetAllHandler).Methods("GET")
}

//IndexHandler handler of index page
func (srv *Server) IndexHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("HELLO!"))
}

//Handler Handling Insert, Delete, and Modify Requests
func (srv *Server) Handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case "GET": //todo вынести в проверку id

		id, err := srv.checkID(r)

		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			http.Error(w, http.StatusText(400), 400)
			return
		}

		tv, err := GetTV(srv.db, id)

		if err != nil {
			fmt.Println(err, "nothing get")
		}
		if tv == nil {
			w.WriteHeader(http.StatusNotFound)
			http.Error(w, http.StatusText(404), 404)
			w.Write([]byte("Cant find item with such params"))
			return
		}
		b, err := json.Marshal(tv)

		if err != nil {
			fmt.Println(err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)

	case "DELETE":
		id, err := srv.checkID(r)

		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			http.Error(w, http.StatusText(400), 400)
			return
		}

		delID, err := DelTV(srv.db, id)

		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			http.Error(w, http.StatusText(500), 500)
			return
		}

		if delID < 0 {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Cant find item with such params"))
			http.Error(w, http.StatusText(404), 404)

			return
		}

		w.WriteHeader(http.StatusNoContent)
		b, err := json.Marshal(tv)

		if err != nil {
			fmt.Println(err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)

		http.Error(w, http.StatusText(204), 204)

	case "PUT":

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		http.Error(w, http.StatusText(405), 405)
		return
	}
}

func (srv *Server) PostHandler(w http.ResponseWriter, r *http.Request) {

}

func (srv *Server) GetAllHandler(w http.ResponseWriter, r *http.Request) {

}

func notFound(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Custom 404 Page not found")
}

func (srv *Server) checkID(r *http.Request) (int64, error) {

	vars := mux.Vars(r)
	stringID := vars["id"]

	id, err := strconv.Atoi(stringID)

	if err != nil {
		return 0, err
	}

	if id <= 0 {
		return 0, errors.New("invalid id")
	}
	return 0, nil

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

	log.Fatal(hhtpSrv.ListenAndServe())
}
