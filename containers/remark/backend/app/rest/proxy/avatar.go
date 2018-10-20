package proxy

import (
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/pkg/errors"

	"github.com/umputun/remark/backend/app/rest"
	"github.com/umputun/remark/backend/app/store"
	"github.com/umputun/remark/backend/app/store/avatar"
)

// Avatar provides http handler for avatars from avatar.Store
// On user login auth will call Put and it will retrieve and save picture locally.
type Avatar struct {
	Store     avatar.Store
	RoutePath string
	RemarkURL string
}

// Put stores retrieved avatar to avatar.Store. Gets image from user info. Returns proxied url
func (p *Avatar) Put(u store.User) (avatarURL string, err error) {

	// no picture for user, try default avatar
	if u.Picture == "" {
		return "", errors.Errorf("no picture for %s", u.ID)
	}

	// load avatar from remote location
	client := http.Client{Timeout: 10 * time.Second}
	var resp *http.Response
	err = retry(5, time.Second, func() error {
		var e error
		resp, e = client.Get(u.Picture)
		return e
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to fetch avatar from the orig")
	}

	defer func() {
		if e := resp.Body.Close(); e != nil {
			log.Printf("[WARN] can't close response body, %s", e)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", errors.Errorf("failed to get avatar from the orig, status %s", resp.Status)
	}

	avatarID, err := p.Store.Put(u.ID, resp.Body) // put returns avatar base name, like 123456.image
	if err != nil {
		return "", err
	}

	log.Printf("[DEBUG] saved avatar from %s to %s, user %q", u.Picture, avatarID, u.Name)
	return p.RemarkURL + p.RoutePath + "/" + avatarID, nil
}

// Routes returns auth routes for given provider
func (p *Avatar) Routes(middlewares ...func(http.Handler) http.Handler) (string, chi.Router) {
	router := chi.NewRouter()
	router.Use(middlewares...)

	// GET /123456789.image
	router.Get("/{avatar}", func(w http.ResponseWriter, r *http.Request) {

		avatarID := chi.URLParam(r, "avatar")

		// enforce client-side caching
		etag := `"` + p.Store.ID(avatarID) + `"`
		w.Header().Set("Etag", etag)
		w.Header().Set("Cache-Control", "max-age=604800") // 7 days
		if match := r.Header.Get("If-None-Match"); match != "" {
			if strings.Contains(match, etag) {
				w.WriteHeader(http.StatusNotModified)
				return
			}
		}

		avReader, size, err := p.Store.Get(avatarID)
		if err != nil {
			rest.SendErrorJSON(w, r, http.StatusBadRequest, err, "can't load avatar")
			return
		}

		defer func() {
			if e := avReader.Close(); e != nil {
				log.Printf("[WARN] can't close avatar reader for %s, %s", avatarID, e)
			}
		}()

		w.Header().Set("Content-Type", "image/*")
		w.Header().Set("Content-Length", strconv.Itoa(size))
		w.WriteHeader(http.StatusOK)
		if _, err = io.Copy(w, avReader); err != nil {
			log.Printf("[WARN] can't send response to %s, %s", r.RemoteAddr, err)
		}
	})

	return p.RoutePath, router
}

func retry(retries int, delay time.Duration, fn func() error) (err error) {
	for i := 0; i < retries; i++ {
		if err = fn(); err == nil {
			return nil
		}
		time.Sleep(delay)
	}
	return errors.Wrap(err, "retry failed")
}
