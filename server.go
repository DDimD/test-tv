package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
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
}

//IndexHandler handler of index page
func (srv *Server) IndexHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("HELLO!"))
}

//Handler Handling Insert, Delete, and Modify Requests
func (srv *Server) Handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		id, err := srv.checkID(r)

		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), 400)
			return
		}

		tv, err := GetTV(srv.db, id)

		if err != nil {
			log.Println(err, "nothing get")
		}
		if tv == nil {
			http.Error(w, http.StatusText(404), 404)
			return
		}

		b, err := json.Marshal(tv)

		if err != nil {
			log.Println(err)
		}

		log.Println(r.Host, r.Method, " ", tv)

		w.Header().Set("Content-Type", "application/json")
		w.Write(b)

	case "DELETE":
		id, err := srv.checkID(r)

		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), 400)
			return
		}

		delID, err := DelTV(srv.db, id)

		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(500), 500)
			return
		}

		if delID < 0 {
			log.Println("id not found")
			http.Error(w, http.StatusText(404), 404)

			return
		}

		log.Println(r.Host, r.Method, id)

		w.WriteHeader(http.StatusNoContent)

	case "PUT":

		id, err := srv.checkID(r)

		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), 400)
			return
		}
		inTv := InPutTv{}
		err = json.NewDecoder(r.Body).Decode(&inTv)

		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(400), 400)
			return
		}

		err = checkData(inTv)

		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), 400)
			return
		}

		updtID, err := UpdateTV(srv.db, id, &inTv)

		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(500), 500)
			return
		}

		if updtID < 0 {
			http.Error(w, http.StatusText(404), 404)

			return
		}

		log.Println(r.Host, r.Method, id, inTv)

		w.WriteHeader(http.StatusNoContent)

	default:
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
		log.Println(err)
		http.Error(w, http.StatusText(400), 400)
		return
	}

	err = validID(tv.ID)

	if err != nil {
		log.Println(err)
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
		log.Println(err)
		http.Error(w, err.Error(), 400)
		return
	}

	addID, err := AddTV(srv.db, tv.ID, &inTv)

	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	if addID < 0 {
		log.Println("index not found")
		http.Error(w, http.StatusText(404), 404)
		return
	}

	if err != nil {
		log.Println(err)
	}

	mapID := make(map[string]int64)
	mapID["id"] = addID
	jsonID, err := json.Marshal(mapID)

	log.Println(r.Host, r.Method, addID)

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonID)
}

func notFound(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(404), 404)
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

// ReturnsChecker handles returns
func (srv *Server) ReturnsChecker(hash chan [32]byte, fileRead chan bool) {

	ticker := time.NewTicker(time.Minute)
	var wg sync.WaitGroup

	for {

		wg.Wait()

		select {
		case <-ticker.C:

			wg.Add(1)

			go func() {
				defer wg.Done()

				xmlFile, err := os.Open("returns.xml")

				if err != nil {
					log.Println(err)
					return
				}

				log.Println("Successfully Opened users.xml")
				defer xmlFile.Close()

				byteVal, _ := ioutil.ReadAll(xmlFile)

				hash <- sha256.Sum256(byteVal)

				if <-fileRead {
					log.Println("file already read")
					return
				}

				var tvs Tvs

				err = xml.Unmarshal(byteVal, &tvs)

				if err != nil {
					log.Println("xml unmarshal error: ", err)
				}

				for _, row := range tvs.Tv {
					returns := int64(row.Text)
					id := int64(row.ID)

					availible, err := UpdtateReturns(srv.db, id, returns)

					if err != nil {
						log.Println(err)
						continue
					}

					log.Println("tv id ", id, "returns ", returns, "availible ", availible)
				}
			}()

		}
	}
}
