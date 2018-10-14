package proxy

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPicture_Extract(t *testing.T) {

	tbl := []struct {
		inp string
		res []string
	}{
		{
			`<p> blah <img src="http://radio-t.com/img.png"/> test</p>`,
			[]string{"http://radio-t.com/img.png"},
		},
		{
			`<p> blah <img src="https://radio-t.com/img.png"/> test</p>`,
			[]string{},
		},
		{
			`<img src="http://radio-t.com/img2.png"/>`,
			[]string{"http://radio-t.com/img2.png"},
		},
		{
			`<img src="http://radio-t.com/img3.png"/> <div>xyz <img src="http://images.pexels.com/67636/img4.jpeg"> </div>`,
			[]string{"http://radio-t.com/img3.png", "http://images.pexels.com/67636/img4.jpeg"},
		},
		{
			`<img src="https://radio-t.com/img3.png"/> <div>xyz <img src="http://images.pexels.com/67636/img4.jpeg"> </div>`,
			[]string{"http://images.pexels.com/67636/img4.jpeg"},
		},
		{
			`abcd <b>blah</b> <h1>xxx</h1>`,
			[]string{},
		},
	}
	img := Image{Enabled: true}

	for i, tt := range tbl {
		res, err := img.extract(tt.inp)
		assert.Nil(t, err, "err in #%d", i)
		assert.Equal(t, tt.res, res, "mismatch in #%d", i)
	}
}

func TestPicture_Replace(t *testing.T) {
	img := Image{Enabled: true, RoutePath: "/img"}
	r := img.replace(`<img src="http://radio-t.com/img3.png"/> xyz <img src="http://images.pexels.com/67636/img4.jpeg">`,
		[]string{"http://radio-t.com/img3.png", "http://images.pexels.com/67636/img4.jpeg"})
	assert.Equal(t, `<img src="/img?src=aHR0cDovL3JhZGlvLXQuY29tL2ltZzMucG5n"/> xyz <img src="/img?src=aHR0cDovL2ltYWdlcy5wZXhlbHMuY29tLzY3NjM2L2ltZzQuanBlZw==">`, r)
}

func TestImage_Routes(t *testing.T) {
	img := Image{Enabled: true, RemarkURL: "https://demo.remark42.com", RoutePath: "/api/v1/proxy"}
	router := img.Routes()

	httpSrv := imgHTTPServer(t)
	defer httpSrv.Close()
	ts := httptest.NewServer(router)
	defer ts.Close()

	encodedImgURL := base64.URLEncoding.EncodeToString([]byte(httpSrv.URL + "/image/img1.png"))

	resp, err := http.Get(ts.URL + "/?src=" + encodedImgURL)
	require.Nil(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	t.Logf("%+v", resp.Header)
	assert.Equal(t, "123", resp.Header["Content-Length"][0])
	assert.Equal(t, "image/png", resp.Header["Content-Type"][0])

	encodedImgURL = base64.URLEncoding.EncodeToString([]byte(httpSrv.URL + "/image/no-such-image.png"))
	resp, err = http.Get(ts.URL + "/?src=" + encodedImgURL)
	require.Nil(t, err)
	assert.Equal(t, 404, resp.StatusCode)

	encodedImgURL = base64.URLEncoding.EncodeToString([]byte(httpSrv.URL + "bad encoding"))
	resp, err = http.Get(ts.URL + "/?src=" + encodedImgURL)
	require.Nil(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestPicture_Convert(t *testing.T) {
	img := Image{Enabled: true, RoutePath: "/img"}
	r := img.Convert(`<img src="http://radio-t.com/img3.png"/> xyz <img src="http://images.pexels.com/67636/img4.jpeg">`)
	assert.Equal(t, `<img src="/img?src=aHR0cDovL3JhZGlvLXQuY29tL2ltZzMucG5n"/> xyz <img src="/img?src=aHR0cDovL2ltYWdlcy5wZXhlbHMuY29tLzY3NjM2L2ltZzQuanBlZw==">`, r)

	r = img.Convert(`<img src="https://radio-t.com/img3.png"/> xyz <img src="http://images.pexels.com/67636/img4.jpeg">`)
	assert.Equal(t, `<img src="https://radio-t.com/img3.png"/> xyz <img src="/img?src=aHR0cDovL2ltYWdlcy5wZXhlbHMuY29tLzY3NjM2L2ltZzQuanBlZw==">`, r)

	img = Image{Enabled: true, RoutePath: "/img", RemarkURL: "http://example.com"}
	r = img.Convert(`<img src="http://radio-t.com/img3.png"/> xyz`)
	assert.Equal(t, `<img src="http://radio-t.com/img3.png"/> xyz`, r, "http:// remark url, no proxy")

	img = Image{Enabled: false, RoutePath: "/img"}
	r = img.Convert(`<img src="http://radio-t.com/img3.png"/> xyz`)
	assert.Equal(t, `<img src="http://radio-t.com/img3.png"/> xyz`, r, "disabled, no proxy")
}

func imgHTTPServer(t *testing.T) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/image/img1.png" {
			t.Log("http img request", r.URL)
			w.Header().Add("Content-Length", "123")
			w.Header().Add("Content-Type", "image/png")
			w.Write([]byte(fmt.Sprintf("%123s", "X")))
			return
		}
		t.Log("http img request - not found", r.URL)
		w.WriteHeader(404)
	}))

	return ts
}
