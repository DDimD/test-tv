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
			http.Error(w, err.Error(), 400)
			return
		}

		tv, err := GetTV(srv.db, id)

		if err != nil {
			fmt.Println(err, "nothing get")
		}
		if tv == nil {
			w.WriteHeader(http.StatusNotFound) //todo, ответов в переборе
			http.Error(w, http.StatusText(404), 404)
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
			http.Error(w, err.Error(), 400)
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
			http.Error(w, http.StatusText(404), 404)

			return
		}

		if err != nil {
			fmt.Println(err)
		}

		w.WriteHeader(http.StatusNoContent)

	case "PUT":

		id, err := srv.checkID(r)

		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			http.Error(w, err.Error(), 400)
			return
		}
		inTv := InPutTv{}
		err = json.NewDecoder(r.Body).Decode(&inTv)

		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			http.Error(w, http.StatusText(400), 400)
			return
		}

		err = checkData(inTv)

		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			http.Error(w, err.Error(), 400)
			return
		}

		updtID, err := UpdateTV(srv.db, id, &inTv)

		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			http.Error(w, http.StatusText(500), 500)
			return
		}

		if updtID < 0 {
			w.WriteHeader(http.StatusNotFound)
			http.Error(w, http.StatusText(404), 404)

			return
		}

		if err != nil {
			fmt.Println(err)
		}

		w.WriteHeader(http.StatusNoContent)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		http.Error(w, http.StatusText(405), 405)
		return
	}
}

//PostHandler handles
func (srv *Server) PostHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		http.Error(w, http.StatusText(405), 405)
		return
	}
	tv := TV{}
	err := json.NewDecoder(r.Body).Decode(&tv)

	if err != nil {
		fmt.Println(err)
		http.Error(w, http.StatusText(400), 400)
		return
	}

	err = validID(tv.ID)

	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), 400)
		return
	}
	inTv := InPutTv{
		tv.Brand,
		tv.Manufacturer,
		tv.Model,
		tv.Year,
	}

	err = checkData(inTv)

	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), 400)
		return
	}

	addID, err := AddTV(srv.db, tv.ID, &inTv)

	if err != nil {
		fmt.Println(err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	if addID < 0 {
		fmt.Println("index not found")
		http.Error(w, http.StatusText(404), 404)
		return
	}

	if err != nil {
		fmt.Println(err)
	}

	mapID := make(map[string]int64)
	mapID["id"] = addID
	jsonID, err := json.Marshal(mapID)

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonID)
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

	err = validID(int64(id))

	if err != nil {
		return 0, err
	}
	return int64(id), nil
}

func validID(id int64) error {
	if id <= 0 {
		return errors.New("invalid id")
	}

	return nil
}

func checkData(inTv InPutTv) error {
	if len(inTv.Manufacturer) < 3 {
		return errors.New("minimum string length 3 characters")
	} else if len(inTv.Model) < 2 {
		return errors.New("minimum string length 2 characters")
	} else if inTv.Year < 2010 {
		return errors.New("	year must be at least 2010")
	}

	return nil
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
