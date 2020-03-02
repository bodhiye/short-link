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
		fmt.Println(err)
		return
	}
	defer r.Body.Close()

	if err := validator.Validate(req); err != nil {
		fmt.Println(err)
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
