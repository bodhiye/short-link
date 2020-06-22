package main_test

import (
	// 引入当前包
	// "."
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify.v2/mock"
)

const (
	expTime       = 60
	longURL       = "https://www.example.com"
	shortLink     = "YQZgxf"
	shortLinkInfo = `{"url": "https://www.example.com", "created_at": "2020-06-21 22:32:29", "expiration_in_minutes": 60`
)

type storageMock struct {
	mock.Mock
}

var app main.App
var mockR *storageMock

func (s *storageMock) Shorten(url string, exp int64) (string, error) {
	args := s.Called(url, exp)
	return args.String(0), args.Error(1)
}

func (s *storageMock) Unshorten(eid string) (string, error) {
	args := s.Called(eid)
	return args.String(0), args.Error(1)
}

func (s *storageMock) ShortLinkInfo(eid string) (interface{}, error) {
	args := s.Called(eid)
	return args.String(0), args.Error(1)
}

func init() {
	app = main.App{}
	mockR = new(storageMock)
	app.Initialize(&main.Env{S: mockR})
}

func TestCreateShortLink(t *testing.T) {
	var jsonStr = []byte(`{
		"url": "https://www.example.com",
		"expiration_in_minutes": 60}`)
	req, err := http.NewRequest("POST", "/api/shorten",
		bytes.NewBuffer(jsonStr))
	if err != nil {
		t.Fatal("Shoule be able to create a request.", err)
	}
	req.Header.Set("Content-Type", "application/json")

	mockR.On("Shorten", longURL, int64(expTime)).Return(shortLink, nil).Once()
	rw := httptest.NewRecorder()
	app.Router.ServeHTTP(rw, req)

	if rw.Code != http.StatusCreated {
		t.Fatal("Excepted reveive %d. Got %d", http.StatusCreated, rw.Code)
	}

	resp := struct {
		ShortLink string `json:"short_link"`
	}{}
	if err := json.NewDecoder(rw.Body).Decode(&resp); err != nil {
		t.Fatal("Shoule decode the reponse")
	}

	if resp.ShortLink != shortLink {
		t.Fatal("Excepted receive %s, Got %s", shortLink, resp.ShortLink)
	}
}

func TestRedirect(t *testing.T) {
	r := fmt.Sprintf("/%s", shortLink)
	req, err := http.NewRequest("GET", r, nil)
	if err != nil {
		t.Fatal("Should be able to create a request.", err)
	}

	mockR.On("Unshorten", shortLink).Return(longURL, nil).Once()
	rw := httptest.NewRecorder()
	app.Router.ServeHTTP(rw, req)

	if rw.Code != http.StatusTemporaryRedirect {
		t.Fatal("Excepted reveive %d. Got %d", http.StatusTemporaryRedirect, rw.Code)
	}
}
