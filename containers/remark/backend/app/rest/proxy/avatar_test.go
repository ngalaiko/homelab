package proxy

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/umputun/remark/backend/app/store"
	"github.com/umputun/remark/backend/app/store/avatar"
)

func TestAvatar_Put(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/pic.png" {
			w.Header().Set("Content-Type", "image/*")
			fmt.Fprint(w, "some picture bin data")
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer ts.Close()

	p := Avatar{RoutePath: "/avatar", RemarkURL: "http://localhost:8080", Store: avatar.NewLocalFS("/tmp/avatars.test", 300)}
	os.MkdirAll("/tmp/avatars.test", 0700)
	defer os.RemoveAll("/tmp/avatars.test")

	u := store.User{ID: "user1", Name: "user1 name", Picture: ts.URL + "/pic.png"}
	res, err := p.Put(u)
	assert.NoError(t, err)
	assert.Equal(t, "http://localhost:8080/avatar/b3daa77b4c04a9551b8781d03191fe098f325e67.image", res)
	fi, err := os.Stat("/tmp/avatars.test/30/b3daa77b4c04a9551b8781d03191fe098f325e67.image")
	assert.NoError(t, err)
	assert.Equal(t, int64(21), fi.Size())

	u.ID = "user2"
	res, err = p.Put(u)
	assert.NoError(t, err)
	assert.Equal(t, "http://localhost:8080/avatar/a1881c06eec96db9901c7bbfe41c42a3f08e9cb4.image", res)
	fi, err = os.Stat("/tmp/avatars.test/84/a1881c06eec96db9901c7bbfe41c42a3f08e9cb4.image")
	assert.NoError(t, err)
	assert.Equal(t, int64(21), fi.Size())
}

func TestAvatar_PutFailed(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Print("request: ", r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	p := Avatar{RoutePath: "/avatar", Store: avatar.NewLocalFS("/tmp/avatars.test", 300)}

	u := store.User{ID: "user1", Name: "user1 name"}
	_, err := p.Put(u)
	assert.EqualError(t, err, "no picture for user1")

	u = store.User{ID: "user1", Name: "user1 name", Picture: "http://127.0.0.1:12345/avater/pic"}
	_, err = p.Put(u)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "connect: connection refused")

	u = store.User{ID: "user1", Name: "user1 name", Picture: ts.URL + "/avatar/pic"}
	_, err = p.Put(u)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get avatar from the orig")
}

func TestAvatar_Routes(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/pic.png" {
			w.Header().Set("Content-Type", "image/*")
			w.Header().Set("Custom-Header", "xyz")
			fmt.Fprint(w, "some picture bin data")
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer ts.Close()

	p := Avatar{RoutePath: "/avatar", Store: avatar.NewLocalFS("/tmp/avatars.test", 300)}
	os.MkdirAll("/tmp/avatars.test", 0700)
	defer os.RemoveAll("/tmp/avatars.test")

	u := store.User{ID: "user1", Name: "user1 name", Picture: ts.URL + "/pic.png"}
	_, err := p.Put(u)
	assert.NoError(t, err)

	// status 400
	req, err := http.NewRequest("GET", "/some_random_name.image", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	_, routes := p.Routes()
	handler := http.Handler(routes)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	// status 200
	req, err = http.NewRequest("GET", "/b3daa77b4c04a9551b8781d03191fe098f325e67.image", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr = httptest.NewRecorder()
	_, routes = p.Routes()
	handler = http.Handler(routes)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	assert.Equal(t, []string{"image/*"}, rr.HeaderMap["Content-Type"])
	assert.Equal(t, []string{"21"}, rr.HeaderMap["Content-Length"])
	assert.Equal(t, []string(nil), rr.HeaderMap["Custom-Header"], "strip all custom headers")
	assert.NotNil(t, rr.HeaderMap["Etag"])

	bb := bytes.Buffer{}
	sz, err := io.Copy(&bb, rr.Body)
	assert.NoError(t, err)
	assert.Equal(t, int64(21), sz)
	assert.Equal(t, "some picture bin data", bb.String())

	// status 304
	req, err = http.NewRequest("GET", "/some_random_name.image", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add("If-None-Match", `"a008de0a2ccb3308b5d99ffff66436e15538f701"`) // hash of `some_random_name.image` since the file doesn't exist

	rr = httptest.NewRecorder()
	_, routes = p.Routes()
	handler = http.Handler(routes)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotModified, rr.Code)
	assert.Equal(t, []string{`"a008de0a2ccb3308b5d99ffff66436e15538f701"`}, rr.HeaderMap["Etag"])
}

func TestAvatar_Retry(t *testing.T) {
	i := 0
	err := retry(5, time.Millisecond, func() error {
		if i == 3 {
			return nil
		}
		i++
		return errors.New("err")
	})
	assert.Nil(t, err)
	assert.Equal(t, 3, i)

	st := time.Now()
	err = retry(5, time.Millisecond, func() error {
		return errors.New("err")
	})
	assert.NotNil(t, err)
	assert.True(t, time.Since(st) >= time.Microsecond*5)
}
