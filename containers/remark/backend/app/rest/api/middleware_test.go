package api

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/umputun/remark/backend/app/rest"
	"github.com/umputun/remark/backend/app/store"
)

func TestMiddleware_AppInfo(t *testing.T) {
	router := chi.NewRouter()
	router.With(AppInfo("remark42", "12345")).Get("/blah", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("blah blah"))
	})
	ts := httptest.NewServer(router)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/blah")
	require.Nil(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)

	assert.Equal(t, "blah blah", string(b))
	assert.Equal(t, "remark42", resp.Header.Get("App-Name"))
	assert.Equal(t, "12345", resp.Header.Get("App-Version"))
	assert.Equal(t, "Umputun", resp.Header.Get("Org"))
}

func TestMiddleware_GetBodyAndUser(t *testing.T) {
	req, err := http.NewRequest("GET", "http://example.com/request", strings.NewReader("body1\nbody2"))
	require.Nil(t, err)

	body, user := getBodyAndUser(req, []LoggerFlag{LogAll})
	assert.Equal(t, "body1 body2", body)
	assert.Equal(t, "", user, "no user")

	b, err := ioutil.ReadAll(req.Body)
	assert.NoError(t, err)
	assert.Equal(t, "body1\nbody2", string(b))

	req = rest.SetUserInfo(req, store.User{ID: "id1", Name: "user1"})
	_, user = getBodyAndUser(req, []LoggerFlag{LogAll})
	assert.Equal(t, ` - id1 "user1"`, user, "no user")

	body, user = getBodyAndUser(req, nil)
	assert.Equal(t, "", body)
	assert.Equal(t, "", user, "no user")

	body, user = getBodyAndUser(req, []LoggerFlag{LogNone})
	assert.Equal(t, "", body)
	assert.Equal(t, "", user, "no user")

	body, user = getBodyAndUser(req, []LoggerFlag{LogUser})
	assert.Equal(t, "", body)
	assert.Equal(t, ` - id1 "user1"`, user, "no user")
}

func TestMiddleware_sanitizeReqURL(t *testing.T) {
	tbl := []struct {
		in  string
		out string
	}{
		{"", ""},
		{"/aa/bb?xyz=123", "/aa/bb?xyz=123"},
		{"/aa/bb?xyz=123&secret=asdfghjk", "/aa/bb?xyz=123&secret=********"},
		{"/aa/bb?xyz=123&secret=asdfghjk&key=val", "/aa/bb?xyz=123&secret=********&key=val"},
		{"/aa/bb?xyz=123&secret=asdfghjk&key=val&password=1234", "/aa/bb?xyz=123&secret=********&key=val&password=****"},
	}
	for i, tt := range tbl {
		assert.Equal(t, tt.out, sanitizeQuery(tt.in), "check #%d, %s", i, tt.in)
	}
}
