package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/umputun/remark/backend/app/store"
)

var testJwtValid = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjI3ODkxOTE4MjIsImp0aSI6InJhbmRvbSBpZCI" +
	"sImlzcyI6InJlbWFyazQyIiwibmJmIjoxNTI2ODg0MjIyLCJ1c2VyIjp7Im5hbWUiOiJuYW1lMSIsImlkIjoiaWQxIiwicGljdHVyZS" +
	"I6IiIsImFkbWluIjpmYWxzZX0sInN0YXRlIjoiMTIzNDU2IiwiZnJvbSI6ImZyb20ifQ._loFgh3g45gr9TtGqvM3N584I_6EHEOJnYb6Py84stQ"

var testJwtValidSess = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjI3ODkxOTE4MjIsImp0aSI6InJhbmRvbSBpZCIsImlzcyI6In" +
	"JlbWFyazQyIiwibmJmIjoxNTI2ODg0MjIyLCJ1c2VyIjp7Im5hbWUiOiJuYW1lMSIsImlkIjoiaWQxIiwicGljdHVyZSI6IiIsIm" +
	"FkbWluIjpmYWxzZX0sInN0YXRlIjoiMTIzNDU2IiwiZnJvbSI6ImZyb20iLCJzZXNzX29ubHkiOnRydWV9.p6w0sM_NYaRuyhyA9jqfWlB5cx1vZPGhXGC5geSX7nA"

var testJwtExpired = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1MjY4ODc4MjIsImp0aSI6InJhbmRvbSBpZCIs" +
	"ImlzcyI6InJlbWFyazQyIiwibmJmIjoxNTI2ODg0MjIyLCJ1c2VyIjp7Im5hbWUiOiJuYW1lMSIsImlkIjoiaWQxIiwicGljdHVyZSI6IiI" +
	"sImFkbWluIjpmYWxzZX0sInN0YXRlIjoiMTIzNDU2IiwiZnJvbSI6ImZyb20ifQ.4_dCrY9ihyfZIedz-kZwBTxmxU1a52V7IqeJrOqTzE4"

var testJwtBadSign = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjI3ODkxOTE4MjIsImp0aSI6InJhbmRvbSBpZCI" +
	"sImlzcyI6InJlbWFyazQyIiwibmJmIjoxNTI2ODg0MjIyLCJ1c2VyIjp7Im5hbWUiOiJuYW1lMSIsImlkIjoiaWQxIiwicGljdHVyZS" +
	"I6IiIsImFkbWluIjpmYWxzZX0sInN0YXRlIjoiMTIzNDU2IiwiZnJvbSI6ImZyb20ifQ._loFgh3g45gr9TtGqvM3N584I_6EHEOJnYb6Py84st"

var days31 = time.Hour * 24 * 31

func TestJWT_Token(t *testing.T) {
	j := NewJWT("xyz 12345", false, time.Hour, days31)

	claims := &CustomClaims{
		State: "123456",
		From:  "from",
		User: &store.User{
			ID:   "id1",
			Name: "name1",
		},
		StandardClaims: jwt.StandardClaims{
			Id:        "random id",
			Issuer:    "remark42",
			ExpiresAt: time.Date(2058, 5, 21, 1, 30, 22, 0, time.Local).Unix(),
			NotBefore: time.Date(2018, 5, 21, 1, 30, 22, 0, time.Local).Unix(),
		},
	}

	res, err := j.Token(claims)
	assert.Nil(t, err)
	assert.Equal(t, testJwtValid, res)
}

func TestJWT_Parse(t *testing.T) {
	j := NewJWT("xyz 12345", false, time.Hour, days31)
	claims, err := j.Parse(testJwtValid)
	assert.NoError(t, err)
	assert.False(t, j.IsExpired(claims))
	assert.Equal(t, &store.User{Name: "name1", ID: "id1"}, claims.User)

	claims, err = j.Parse(testJwtExpired)
	assert.NoError(t, err)
	assert.True(t, j.IsExpired(claims))

	_, err = j.Parse("bad")
	assert.NotNil(t, err, "bad token")

	_, err = j.Parse(testJwtBadSign)
	assert.EqualError(t, err, "can't parse jwt: signature is invalid")
}

func TestJWT_Set(t *testing.T) {
	j := NewJWT("xyz 12345", false, time.Hour, days31)

	claims := &CustomClaims{
		State: "123456",
		From:  "from",
		User: &store.User{
			ID:   "id1",
			Name: "name1",
		},
		StandardClaims: jwt.StandardClaims{
			Id:        "random id",
			Issuer:    "remark42",
			ExpiresAt: time.Date(2058, 5, 21, 1, 30, 22, 0, time.Local).Unix(),
			NotBefore: time.Date(2018, 5, 21, 1, 30, 22, 0, time.Local).Unix(),
		},
	}

	claims.SessionOnly = false
	rr := httptest.NewRecorder()
	err := j.Set(rr, claims, claims.SessionOnly)
	assert.Nil(t, err)
	cookies := rr.Result().Cookies()
	t.Log(cookies)
	require.Equal(t, 2, len(cookies))
	assert.Equal(t, "JWT", cookies[0].Name)
	assert.Equal(t, testJwtValid, cookies[0].Value)
	assert.Equal(t, 31*24*3600, cookies[0].MaxAge)
	assert.Equal(t, "XSRF-TOKEN", cookies[1].Name)
	assert.Equal(t, "random id", cookies[1].Value)

	claims.SessionOnly = true
	rr = httptest.NewRecorder()
	err = j.Set(rr, claims, claims.SessionOnly)
	assert.Nil(t, err)
	cookies = rr.Result().Cookies()
	t.Log(cookies)
	require.Equal(t, 2, len(cookies))
	assert.Equal(t, "JWT", cookies[0].Name)
	assert.Equal(t, testJwtValidSess, cookies[0].Value)
	assert.Equal(t, 0, cookies[0].MaxAge)
	assert.Equal(t, "XSRF-TOKEN", cookies[1].Name)
	assert.Equal(t, "random id", cookies[1].Value)
}

func TestJWT_GetFromHeader(t *testing.T) {
	j := NewJWT("xyz 12345", false, time.Hour, days31)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Add(jwtHeaderKey, testJwtValid)
	claims, err := j.Get(req)
	assert.Nil(t, err)
	assert.False(t, j.IsExpired(claims))
	assert.Equal(t, &store.User{Name: "name1", ID: "id1", Picture: "", Admin: false, Blocked: false, IP: ""}, claims.User)
	assert.Equal(t, "remark42", claims.Issuer)

	req = httptest.NewRequest("GET", "/", nil)
	req.Header.Add(jwtHeaderKey, testJwtExpired)
	claims, err = j.Get(req)
	assert.Nil(t, err)
	assert.True(t, j.IsExpired(claims))

	req = httptest.NewRequest("GET", "/", nil)
	req.Header.Add(jwtHeaderKey, "bad bad token")
	_, err = j.Get(req)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "can't parse jwt: token contains an invalid number of segments"), err.Error())

}

func TestJWT_SetAndGetWithCookies(t *testing.T) {
	j := NewJWT("xyz 12345", false, time.Hour, days31)

	claims := &CustomClaims{
		State:       "123456",
		From:        "from",
		SessionOnly: true,
		User: &store.User{
			ID:   "id1",
			Name: "name1",
		},
		StandardClaims: jwt.StandardClaims{
			Id:        "random id",
			Issuer:    "remark42",
			ExpiresAt: time.Date(2058, 5, 21, 1, 30, 22, 0, time.Local).Unix(),
			NotBefore: time.Date(2018, 5, 21, 1, 30, 22, 0, time.Local).Unix(),
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/valid" {
			assert.Nil(t, j.Set(w, claims, true))
			w.WriteHeader(200)
		}
	}))
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/valid")
	require.Nil(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	req := httptest.NewRequest("GET", "/valid", nil)
	req.AddCookie(resp.Cookies()[0])
	req.Header.Add(xsrfHeaderKey, "random id")
	claims, err = j.Get(req)
	assert.Nil(t, err)
	assert.Equal(t, &store.User{Name: "name1", ID: "id1", Picture: "", Admin: false, Blocked: false, IP: ""}, claims.User)
	assert.Equal(t, "remark42", claims.Issuer)
	assert.Equal(t, true, claims.SessionOnly)
	t.Log(resp.Cookies())
}

func TestJWT_SetAndGetWithXsrfMismatch(t *testing.T) {
	j := NewJWT("xyz 12345", false, time.Hour, days31)

	claims := &CustomClaims{
		State: "123456",
		From:  "from",
		User: &store.User{
			ID:   "id1",
			Name: "name1",
		},
		StandardClaims: jwt.StandardClaims{
			Id:        "random id",
			Issuer:    "remark42",
			ExpiresAt: time.Date(2058, 5, 21, 1, 30, 22, 0, time.Local).Unix(),
			NotBefore: time.Date(2018, 5, 21, 1, 30, 22, 0, time.Local).Unix(),
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/valid" {
			assert.Nil(t, j.Set(w, claims, true))
			w.WriteHeader(200)
		}
	}))
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/valid")
	require.Nil(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	req := httptest.NewRequest("GET", "/valid", nil)
	req.AddCookie(resp.Cookies()[0])
	req.Header.Add(xsrfHeaderKey, "random id wrong")
	claims, err = j.Get(req)
	assert.EqualError(t, err, "xsrf mismatch")
}

func TestJWT_SetAndGetWithCookiesExpired(t *testing.T) {
	j := NewJWT("xyz 12345", false, time.Hour, days31)

	claims := &CustomClaims{
		State: "123456",
		From:  "from",
		User: &store.User{
			ID:   "id1",
			Name: "name1",
		},
		StandardClaims: jwt.StandardClaims{
			Id:        "random id",
			Issuer:    "remark42",
			ExpiresAt: time.Date(2018, 5, 21, 1, 35, 22, 0, time.Local).Unix(),
			NotBefore: time.Date(2018, 5, 21, 1, 30, 22, 0, time.Local).Unix(),
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/expired" {
			assert.Nil(t, j.Set(w, claims, true))
			w.WriteHeader(200)
		}
	}))
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/expired")
	require.Nil(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	req := httptest.NewRequest("GET", "/expired", nil)
	req.AddCookie(resp.Cookies()[0])
	req.Header.Add(xsrfHeaderKey, "random id")
	claims, err = j.Get(req)
	assert.Nil(t, err)
	assert.True(t, j.IsExpired(claims))
}
