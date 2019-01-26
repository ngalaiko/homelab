package api

import (
	"errors"
	"net/http"
	"path"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/go-pkgz/auth"
	log "github.com/go-pkgz/lgr"
	R "github.com/go-pkgz/rest"
	"github.com/go-pkgz/rest/cache"

	"github.com/umputun/remark/backend/app/rest"
	"github.com/umputun/remark/backend/app/store"
	"github.com/umputun/remark/backend/app/store/service"
)

// admin provides router for all requests available for admin users only
type admin struct {
	dataService   *service.DataStore
	cache         cache.LoadingCache
	authenticator *auth.Service
	readOnlyAge   int
	migrator      *Migrator
}

func (a *admin) routes(middlewares ...func(http.Handler) http.Handler) chi.Router {
	router := chi.NewRouter()
	router.Use(middlewares...)
	router.Delete("/comment/{id}", a.deleteCommentCtrl)
	router.Put("/user/{userid}", a.setBlockCtrl)
	router.Delete("/user/{userid}", a.deleteUserCtrl)
	router.Get("/user/{userid}", a.getUserInfoCtrl)
	router.Get("/deleteme", a.deleteMeRequestCtrl)
	router.Put("/verify/{userid}", a.setVerifyCtrl)
	router.Put("/pin/{id}", a.setPinCtrl)
	router.Get("/blocked", a.blockedUsersCtrl)
	router.Put("/readonly", a.setReadOnlyCtrl)
	router.Put("/title/{id}", a.setTitleCtrl)

	a.migrator.withRoutes(router) // set migrator routes, i.e. /export and /import

	return router
}

// DELETE /comment/{id}?site=siteID&url=post-url - removes comment
func (a *admin) deleteCommentCtrl(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	locator := store.Locator{SiteID: r.URL.Query().Get("site"), URL: r.URL.Query().Get("url")}
	log.Printf("[INFO] delete comment %s", id)

	err := a.dataService.Delete(locator, id, store.SoftDelete)
	if err != nil {
		rest.SendErrorJSON(w, r, http.StatusInternalServerError, err, "can't delete comment")
		return
	}
	a.cache.Flush(cache.Flusher(locator.SiteID).Scopes(locator.SiteID, locator.URL, lastCommentsScope))
	render.Status(r, http.StatusOK)
	render.JSON(w, r, R.JSON{"id": id, "locator": locator})
}

// DELETE /user/{userid}?site=side-id - delete all user comments for requested userid
func (a *admin) deleteUserCtrl(w http.ResponseWriter, r *http.Request) {

	userID := chi.URLParam(r, "userid")
	siteID := r.URL.Query().Get("site")
	log.Printf("[INFO] delete all user comments for %s, site %s", userID, siteID)

	if err := a.dataService.DeleteUser(siteID, userID); err != nil {
		rest.SendErrorJSON(w, r, http.StatusInternalServerError, err, "can't delete user")
		return
	}
	a.cache.Flush(cache.Flusher(siteID).Scopes(userID, siteID, lastCommentsScope))
	render.Status(r, http.StatusOK)
	render.JSON(w, r, R.JSON{"user_id": userID, "site_id": siteID})
}

// GET /user/{userid}?site=side-id - get user info for requested userid
func (a *admin) getUserInfoCtrl(w http.ResponseWriter, r *http.Request) {

	userID := chi.URLParam(r, "userid")
	siteID := r.URL.Query().Get("site")
	log.Printf("[INFO] get user info for %s, site %s", userID, siteID)

	ucomments, err := a.dataService.User(siteID, userID, 1, 0)
	if err != nil || len(ucomments) == 0 {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, err, "can't get user info")
		return
	}
	render.Status(r, http.StatusOK)
	render.JSON(w, r, ucomments[0].User)
}

// GET /deleteme?token=jwt - delete all user comments by user's request. Gets info about deleted used from provided token
// request made GET to allow direct click from the email sent by user
func (a *admin) deleteMeRequestCtrl(w http.ResponseWriter, r *http.Request) {

	token := r.URL.Query().Get("token")

	claims, err := a.authenticator.TokenService().Parse(token)
	if err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, err, "can't process token")
		return
	}

	log.Printf("[INFO] delete all user comments by request for %s, site %s", claims.User.ID, claims.Audience)

	// deleteme set by deleteMeCtrl, this check just to make sure we not trying to delete with leaked token
	if !claims.User.BoolAttr("delete_me") {
		rest.SendErrorJSON(w, r, http.StatusForbidden, errors.New("forbidden"), "can't use provided token")
		return
	}

	if err := a.dataService.DeleteUser(claims.Audience, claims.User.ID); err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, err, "can't delete user")
		return
	}

	if claims.User.Picture != "" && a.authenticator.AvatarProxy() != nil {
		avatartStore := a.authenticator.AvatarProxy().Store
		if err := avatartStore.Remove(path.Base(claims.User.Picture)); err != nil {
			rest.SendErrorJSON(w, r, http.StatusBadRequest, err, "can't delete user's avatar")
			return
		}
	}

	a.cache.Flush(cache.Flusher(claims.Audience).Scopes(claims.Audience, claims.User.ID, lastCommentsScope))
	render.Status(r, http.StatusOK)
	render.JSON(w, r, R.JSON{"user_id": claims.User.ID, "site_id": claims.Audience})
}

// PUT /user/{userid}?site=side-id&block=1&ttl=7d - block or unblock user
func (a *admin) setBlockCtrl(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userid")
	siteID := r.URL.Query().Get("site")
	blockStatus := r.URL.Query().Get("block") == "1"

	ttl := time.Duration(0) // unlimited duration by default
	if ttlParam := r.URL.Query().Get("ttl"); ttlParam != "" {
		if d, err := time.ParseDuration(ttlParam); err == nil {
			ttl = d
		}
	}

	if err := a.dataService.SetBlock(siteID, userID, blockStatus, ttl); err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, err, "can't set blocking status")
		return
	}
	a.cache.Flush(cache.Flusher(siteID).Scopes(userID, siteID, lastCommentsScope))
	render.JSON(w, r, R.JSON{"user_id": userID, "site_id": siteID, "block": blockStatus})
}

// GET /blocked?site=siteID - list blocked users
func (a *admin) blockedUsersCtrl(w http.ResponseWriter, r *http.Request) {
	siteID := r.URL.Query().Get("site")
	users, err := a.dataService.Blocked(siteID)
	if err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, err, "can't get blocked users")
		return
	}
	render.JSON(w, r, users)
}

// PUT /readonly?site=siteID&url=post-url&ro=1 - set or reset read-only status for the post
func (a *admin) setReadOnlyCtrl(w http.ResponseWriter, r *http.Request) {
	locator := store.Locator{SiteID: r.URL.Query().Get("site"), URL: r.URL.Query().Get("url")}
	roStatus := r.URL.Query().Get("ro") == "1"

	isRoByAge := func(info store.PostInfo) bool {
		return a.readOnlyAge > 0 && !info.FirstTS.IsZero() &&
			info.FirstTS.AddDate(0, 0, a.readOnlyAge).Before(time.Now())
	}

	// don't allow to reset ro for posts turned to ro by ReadOnlyAge
	if !roStatus {
		if info, e := a.dataService.Info(locator, a.readOnlyAge); e == nil && isRoByAge(info) {
			rest.SendErrorJSON(w, r, http.StatusForbidden, errors.New("rejected"), "read-only due the age")
			return
		}
	}

	if err := a.dataService.SetReadOnly(locator, roStatus); err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, err, "can't set readonly status")
		return
	}
	a.cache.Flush(cache.Flusher(locator.SiteID).Scopes(locator.URL, locator.SiteID))
	render.JSON(w, r, R.JSON{"locator": locator, "read-only": roStatus})
}

// PUT /title/{id}?site=siteID&url=post-url - set comment PostTitle to page's title
func (a *admin) setTitleCtrl(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	locator := store.Locator{SiteID: r.URL.Query().Get("site"), URL: r.URL.Query().Get("url")}

	c, err := a.dataService.SetTitle(locator, id)
	if err != nil {
		rest.SendErrorJSON(w, r, http.StatusInternalServerError, err, "can't set title")
		return
	}
	log.Printf("[INFO] set comment's title %s to %q", id, c.PostTitle)

	a.cache.Flush(cache.Flusher(locator.SiteID).Scopes(locator.URL, lastCommentsScope))
	render.Status(r, http.StatusOK)
	render.JSON(w, r, R.JSON{"id": id, "locator": locator})
}

// PUT /verify?site=siteID&url=post-url&ro=1 - set or reset read-only status for the post
func (a *admin) setVerifyCtrl(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userid")
	siteID := r.URL.Query().Get("site")
	verifyStatus := r.URL.Query().Get("verified") == "1"

	if err := a.dataService.SetVerified(siteID, userID, verifyStatus); err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, err, "can't set verify status")
		return
	}
	a.cache.Flush(cache.Flusher(siteID).Scopes(siteID, userID))
	render.JSON(w, r, R.JSON{"user": userID, "verified": verifyStatus})
}

// PUT /pin/{id}?site=siteID&url=post-url&pin=1
// mark/unmark comment as a special
func (a *admin) setPinCtrl(w http.ResponseWriter, r *http.Request) {
	commentID := chi.URLParam(r, "id")
	locator := store.Locator{SiteID: r.URL.Query().Get("site"), URL: r.URL.Query().Get("url")}
	pinStatus := r.URL.Query().Get("pin") == "1"

	if err := a.dataService.SetPin(locator, commentID, pinStatus); err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, err, "can't set pin status")
		return
	}
	a.cache.Flush(cache.Flusher(locator.SiteID).Scopes(locator.URL))
	render.JSON(w, r, R.JSON{"id": commentID, "locator": locator, "pin": pinStatus})
}

func (a *admin) checkBlocked(siteID string, user store.User) bool {
	return a.dataService.IsBlocked(siteID, user.ID)
}

// post-processes comments, hides text of all comments for blocked users,
// resets score and votes too. Also hides sensitive info for non-admin users
func (a *admin) alterComments(comments []store.Comment, r *http.Request) (res []store.Comment) {
	res = make([]store.Comment, len(comments))

	user, err := rest.GetUserInfo(r)
	isAdmin := err == nil && user.Admin

	for i, c := range comments {

		blocked := a.dataService.IsBlocked(c.Locator.SiteID, c.User.ID)
		// process blocked users
		if blocked {
			if !isAdmin { // reset comment to deleted for non-admins
				c.SetDeleted(store.SoftDelete)
			}
			c.User.Blocked = true
			c.Deleted = true
		}

		// set verified status retroactively
		if !blocked {
			c.User.Verified = a.dataService.IsVerified(c.Locator.SiteID, c.User.ID)
		}

		// hide info from non-admins
		if !isAdmin {
			c.User.IP = ""
		}

		res[i] = c
	}
	return res
}
