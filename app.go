package main

import (
	"encoding/json"
	"fmt"

	// "io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"gopkg.in/validator.v2"
)

// App eccapsulates Env, Router and middlewares
type App struct {
	Router      *mux.Router
	Middlewares *Middleware
	Config      *Env
}

type shortlinkReq struct {
	URL        string `json:"url" validate:"nonzero"`
	Expiration int64  `json:"expiration" validate:"min=0"`
}

type shortlinkResp struct {
	Shortlink string `json:"shortlink"`
}

// Initialize is initializtion of app
func (a *App) Initialize(e *Env) {
	// set log formatter
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	a.Config = e
	a.Router = mux.NewRouter()
	a.Middlewares = &Middleware{}
	a.InitializeRouter()
}

func (a *App) InitializeRouter() {
	m := alice.New(a.Middlewares.LoggingHandler, a.Middlewares.RecoverHandler)
	// a.Router.HandleFunc("/api/shorten", a.CreateShortLink).Methods("POST")
	// a.Router.HandleFunc("/api/info", a.GetShortLinkInfo).Methods("GET")
	// a.Router.HandleFunc("/{shortlink:[a-zA-Z0-9]{1,11}}", a.Redirect).Methods("GET")
	a.Router.Handle("/api/shorten", m.ThenFunc(a.CreateShortLink)).Methods("POST")
	a.Router.Handle("/api/info", m.ThenFunc(a.GetShortLinkInfo)).Methods("GET")
	a.Router.Handle("/{shortlink:[a-zA-Z0-9]{1,11}}", m.ThenFunc(a.Redirect)).Methods("GET")
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

	s, err := a.Config.S.Shorten(req.URL, req.Expiration)
	if err != nil {
		respondWithError(w, err)
	} else {
		respondWithJSON(w, http.StatusCreated, shortlinkResp{Shortlink: s})
	}
}

func (a *App) GetShortLinkInfo(w http.ResponseWriter, r *http.Request) {
	vals := r.URL.Query()
	url := vals.Get("shortlink")

	d, err := a.Config.S.ShortlinkInfo(url)
	if err != nil {
		respondWithError(w, err)
	} else {
		respondWithJSON(w, http.StatusOK, d)
	}
}

func (a *App) Redirect(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	u, err := a.Config.S.Unshorten(vars["shortlink"])
	if err != nil {
		respondWithError(w, err)
	} else {
		http.Redirect(w, r, u, http.StatusTemporaryRedirect)
	}
}

// Run starts listen and server
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
