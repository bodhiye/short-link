package main

import (
	"encoding/json"
	"fmt"

	// "io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"gopkg.in/validator.v2"
)

type App struct {
	Router *mux.Router
}

type shortlinkReq struct {
	URL        string `json:"url" validate:"nonzero"`
	ExpireDate int64  `json:"expireDate" validate:"min=0"`
}

type shortlinkResp struct {
	Shortlink string `json:"shortlink"`
}

func (a *App) Initialize() {
	// set log formatter
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	a.Router = mux.NewRouter()
	a.InitializeRouter()
}

func (a *App) InitializeRouter() {
	a.Router.HandleFunc("/api/shorten", a.CreateShortLink).Methods("POST")
	a.Router.HandleFunc("/api/info", a.GetShortLinkInfo).Methods("GET")
	a.Router.HandleFunc("/{shortlink:[a-zA-Z0-9]{1,11}}", a.Redirect).Methods("GET")
}

func (a *App) CreateShortLink(w http.ResponseWriter, r *http.Request) {
	var req shortlinkReq

	// bs, err := ioutil.ReadAll(r.Body)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// if err := json.Unmarshal(bs, &req); err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, StatusError{http.StatusBadRequest,
			fmt.Errorf("parse parameters failed %v", r.Body)})
		return
	}
	defer r.Body.Close()

	if err := validator.Validate(req); err != nil {
		respondWithError(w, StatusError{http.StatusBadRequest,
			fmt.Errorf("validate parameters failed %v", req)})
		return
	}

	fmt.Printf("%v\n", req)
}

func (a *App) GetShortLinkInfo(w http.ResponseWriter, r *http.Request) {
	vals := r.URL.Query()
	url := vals.Get("shortlink")

	fmt.Printf("%s\n", url)
}

func (a *App) Redirect(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	fmt.Printf("%s\n", vars["shortlink"])
}

func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

func respondWithError(w http.ResponseWriter, err error) {
	switch e := err.(type) {
	case Error:
		log.Printf("HTTP %d - %s", e.Status(), e)
		respondWithJSON(w, e.Status(), e.Error())
	default:
		respondWithJSON(w, http.StatusInternalServerError,
			http.StatusText(http.StatusInternalServerError))
	}
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	resp, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(resp)
}
