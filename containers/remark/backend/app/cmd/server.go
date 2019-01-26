package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"

	bolt "github.com/coreos/bbolt"
	log "github.com/go-pkgz/lgr"
	auth_cache "github.com/patrickmn/go-cache"
	"github.com/pkg/errors"

	"github.com/go-pkgz/auth"
	"github.com/go-pkgz/auth/avatar"
	"github.com/go-pkgz/auth/provider"
	"github.com/go-pkgz/auth/token"
	"github.com/go-pkgz/mongo"
	"github.com/go-pkgz/rest/cache"

	"github.com/umputun/remark/backend/app/migrator"
	"github.com/umputun/remark/backend/app/notify"
	"github.com/umputun/remark/backend/app/rest/api"
	"github.com/umputun/remark/backend/app/rest/proxy"
	"github.com/umputun/remark/backend/app/store"
	"github.com/umputun/remark/backend/app/store/admin"
	"github.com/umputun/remark/backend/app/store/engine"
	"github.com/umputun/remark/backend/app/store/service"
)

// ServerCommand with command line flags and env
type ServerCommand struct {
	Store  StoreGroup  `group:"store" namespace:"store" env-namespace:"STORE"`
	Avatar AvatarGroup `group:"avatar" namespace:"avatar" env-namespace:"AVATAR"`
	Cache  CacheGroup  `group:"cache" namespace:"cache" env-namespace:"CACHE"`
	Mongo  MongoGroup  `group:"mongo" namespace:"mongo" env-namespace:"MONGO"`
	Admin  AdminGroup  `group:"admin" namespace:"admin" env-namespace:"ADMIN"`
	Notify NotifyGroup `group:"notify" namespace:"notify" env-namespace:"NOTIFY"`
	SSL    SSLGroup    `group:"ssl" namespace:"ssl" env-namespace:"SSL"`

	Sites           []string      `long:"site" env:"SITE" default:"remark" description:"site names" env-delim:","`
	AdminPasswd     string        `long:"admin-passwd" env:"ADMIN_PASSWD" default:"" description:"admin basic auth password"`
	BackupLocation  string        `long:"backup" env:"BACKUP_PATH" default:"./var/backup" description:"backups location"`
	MaxBackupFiles  int           `long:"max-back" env:"MAX_BACKUP_FILES" default:"10" description:"max backups to keep"`
	ImageProxy      bool          `long:"img-proxy" env:"IMG_PROXY" description:"enable image proxy"`
	MaxCommentSize  int           `long:"max-comment" env:"MAX_COMMENT_SIZE" default:"2048" description:"max comment size"`
	MaxVotes        int           `long:"max-votes" env:"MAX_VOTES" default:"-1" description:"maximum number of votes per comment"`
	LowScore        int           `long:"low-score" env:"LOW_SCORE" default:"-5" description:"low score threshold"`
	CriticalScore   int           `long:"critical-score" env:"CRITICAL_SCORE" default:"-10" description:"critical score threshold"`
	PositiveScore   bool          `long:"positive-score" env:"POSITIVE_SCORE" description:"enable positive score only"`
	ReadOnlyAge     int           `long:"read-age" env:"READONLY_AGE" default:"0" description:"read-only age of comments, days"`
	EditDuration    time.Duration `long:"edit-time" env:"EDIT_TIME" default:"5m" description:"edit window"`
	Port            int           `long:"port" env:"REMARK_PORT" default:"8080" description:"port"`
	WebRoot         string        `long:"web-root" env:"REMARK_WEB_ROOT" default:"./web" description:"web root directory"`
	UpdateLimit     float64       `long:"update-limit" env:"UPDATE_LIMIT" default:"0.5" description:"updates/sec limit"`
	RestrictedWords []string      `long:"restricted-words" env:"RESTRICTED_WORDS" default:"" description:"words prohibited to use in comments" env-delim:","`

	Auth struct {
		TTL struct {
			JWT    time.Duration `long:"jwt" env:"JWT" default:"5m" description:"jwt TTL"`
			Cookie time.Duration `long:"cookie" env:"COOKIE" default:"200h" description:"auth cookie TTL"`
		} `group:"ttl" namespace:"ttl" env-namespace:"TTL"`
		Google   AuthGroup `group:"google" namespace:"google" env-namespace:"GOOGLE" description:"Google OAuth"`
		Github   AuthGroup `group:"github" namespace:"github" env-namespace:"GITHUB" description:"Github OAuth"`
		Facebook AuthGroup `group:"facebook" namespace:"facebook" env-namespace:"FACEBOOK" description:"Facebook OAuth"`
		Yandex   AuthGroup `group:"yandex" namespace:"yandex" env-namespace:"YANDEX" description:"Yandex OAuth"`
		Dev      bool      `long:"dev" env:"DEV" description:"enable dev (local) oauth2"`
	} `group:"auth" namespace:"auth" env-namespace:"AUTH"`

	CommonOpts
}

// AuthGroup defines options group for auth params
type AuthGroup struct {
	CID  string `long:"cid" env:"CID" description:"OAuth client ID"`
	CSEC string `long:"csec" env:"CSEC" description:"OAuth client secret"`
}

// StoreGroup defines options group for store params
type StoreGroup struct {
	Type string `long:"type" env:"TYPE" description:"type of storage" choice:"bolt" choice:"mongo" default:"bolt"`
	Bolt struct {
		Path    string        `long:"path" env:"PATH" default:"./var" description:"parent dir for bolt files"`
		Timeout time.Duration `long:"timeout" env:"TIMEOUT" default:"30s" description:"bolt timeout"`
	} `group:"bolt" namespace:"bolt" env-namespace:"BOLT"`
}

// AvatarGroup defines options group for avatar params
type AvatarGroup struct {
	Type string `long:"type" env:"TYPE" description:"type of avatar storage" choice:"fs" choice:"bolt" choice:"mongo" default:"fs"`
	FS   struct {
		Path string `long:"path" env:"PATH" default:"./var/avatars" description:"avatars location"`
	} `group:"fs" namespace:"fs" env-namespace:"FS"`
	Bolt struct {
		File string `long:"file" env:"FILE" default:"./var/avatars.db" description:"avatars bolt file location"`
	} `group:"bolt" namespace:"bolt" env-namespace:"bolt"`
	RszLmt int `long:"rsz-lmt" env:"RESIZE" default:"0" description:"max image size for resizing avatars on save"`
}

// CacheGroup defines options group for cache params
type CacheGroup struct {
	Type string `long:"type" env:"TYPE" description:"type of cache" choice:"mem" choice:"mongo" choice:"none" default:"mem"`
	Max  struct {
		Items int   `long:"items" env:"ITEMS" default:"1000" description:"max cached items"`
		Value int   `long:"value" env:"VALUE" default:"65536" description:"max size of cached value"`
		Size  int64 `long:"size" env:"SIZE" default:"50000000" description:"max size of total cache"`
	} `group:"max" namespace:"max" env-namespace:"MAX"`
}

// MongoGroup holds all mongo params, used by store, avatar and cache
type MongoGroup struct {
	URL string `long:"url" env:"URL" description:"mongo url"`
	DB  string `long:"db" env:"DB" default:"remark42" description:"mongo database"`
}

// AdminGroup defines options group for admin params
type AdminGroup struct {
	Type   string `long:"type" env:"TYPE" description:"type of admin store" choice:"shared" choice:"mongo" default:"shared"`
	Shared struct {
		Admins []string `long:"id" env:"ID" description:"admin(s) ids" env-delim:","`
		Email  string   `long:"email" env:"EMAIL" default:"" description:"admin email"`
	} `group:"shared" namespace:"shared" env-namespace:"SHARED"`
}

// NotifyGroup defines options for notification
type NotifyGroup struct {
	Type      string `long:"type" env:"TYPE" description:"type of notification" choice:"none" choice:"telegram" default:"none"`
	QueueSize int    `long:"queue" env:"QUEUE" description:"size of notification queue" default:"100"`
	Telegram  struct {
		Token   string        `long:"token" env:"TOKEN" description:"telegram token"`
		Channel string        `long:"chan" env:"CHAN" description:"telegram channel"`
		Timeout time.Duration `long:"timeout" env:"TIMEOUT" default:"5s" description:"telegram timeout"`
		API     string        `long:"api" env:"API" default:"https://api.telegram.org/bot" description:"telegram api prefix"`
	} `group:"telegram" namespace:"telegram" env-namespace:"TELEGRAM"`
}

// SSLGroup defines options group for server ssl params
type SSLGroup struct {
	Type         string `long:"type" env:"TYPE" description:"ssl (auto)support" choice:"none" choice:"static" choice:"auto" default:"none"`
	Port         int    `long:"port" env:"PORT" description:"port number for https server" default:"8443"`
	Cert         string `long:"cert" env:"CERT" description:"path to cert.pem file"`
	Key          string `long:"key" env:"KEY" description:"path to key.pem file"`
	ACMELocation string `long:"acme-location" env:"ACME_LOCATION" description:"dir where certificates will be stored by autocert manager" default:"./var/acme"`
	ACMEEmail    string `long:"acme-email" env:"ACME_EMAIL" description:"admin email for certificate notifications"`
}

// serverApp holds all active objects
type serverApp struct {
	*ServerCommand
	restSrv       *api.Rest
	migratorSrv   *api.Migrator
	exporter      migrator.Exporter
	devAuth       *provider.DevAuthServer
	dataService   *service.DataStore
	avatarStore   avatar.Store
	notifyService *notify.Service
	terminated    chan struct{}
}

// Execute is the entry point for "server" command, called by flag parser
func (s *ServerCommand) Execute(args []string) error {
	log.Printf("[INFO] start server on port %d", s.Port)
	resetEnv("SECRET", "AUTH_GOOGLE_CSEC", "AUTH_GITHUB_CSEC", "AUTH_FACEBOOK_CSEC", "AUTH_YANDEX_CSEC", "ADMIN_PASSWD")

	ctx, cancel := context.WithCancel(context.Background())
	go func() { // catch signal and invoke graceful termination
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		log.Printf("[WARN] interrupt signal")
		cancel()
	}()

	app, err := s.newServerApp()
	if err != nil {
		log.Printf("[PANIC] failed to setup application, %+v", err)
	}
	if err = app.run(ctx); err != nil {
		log.Printf("[ERROR] remark terminated with error %+v", err)
		return err
	}
	log.Printf("[INFO] remark terminated")
	return nil
}

// newServerApp prepares application and return it with all active parts
// doesn't start anything
func (s *ServerCommand) newServerApp() (*serverApp, error) {

	if err := makeDirs(s.BackupLocation); err != nil {
		return nil, err
	}

	if !strings.HasPrefix(s.RemarkURL, "http://") && !strings.HasPrefix(s.RemarkURL, "https://") {
		return nil, errors.Errorf("invalid remark42 url %s", s.RemarkURL)
	}
	log.Printf("[INFO] root url=%s", s.RemarkURL)

	storeEngine, err := s.makeDataStore()
	if err != nil {
		return nil, errors.Wrap(err, "failed to make data store engine")
	}

	adminStore, err := s.makeAdminStore()
	if err != nil {
		return nil, errors.Wrap(err, "failed to make admin store")
	}

	dataService := &service.DataStore{
		Interface:              storeEngine,
		EditDuration:           s.EditDuration,
		AdminStore:             adminStore,
		MaxCommentSize:         s.MaxCommentSize,
		MaxVotes:               s.MaxVotes,
		PositiveScore:          s.PositiveScore,
		TitleExtractor:         service.NewTitleExtractor(http.Client{Timeout: time.Second * 5}),
		RestrictedWordsMatcher: service.NewRestrictedWordsMatcher(service.StaticRestrictedWordsLister{Words: s.RestrictedWords}),
	}

	loadingCache, err := s.makeCache()
	if err != nil {
		return nil, errors.Wrap(err, "failed to make cache")
	}

	avatarStore, err := s.makeAvatarStore()
	if err != nil {
		return nil, errors.Wrap(err, "failed to make avatar store")
	}
	authenticator := s.makeAuthenticator(dataService, avatarStore, adminStore)

	exporter := &migrator.Native{DataStore: dataService}

	migr := &api.Migrator{
		Cache:             loadingCache,
		NativeImporter:    &migrator.Native{DataStore: dataService},
		DisqusImporter:    &migrator.Disqus{DataStore: dataService},
		WordPressImporter: &migrator.WordPress{DataStore: dataService},
		NativeExporter:    &migrator.Native{DataStore: dataService},
		KeyStore:          adminStore,
	}

	notifyService, err := s.makeNotify(dataService)
	if err != nil {
		log.Printf("[WARN] failed to make notify service, %s", err)
		notifyService = notify.NopService // disable notifier
	}

	imgProxy := &proxy.Image{Enabled: s.ImageProxy, RoutePath: "/api/v1/img", RemarkURL: s.RemarkURL}
	commentFormatter := store.NewCommentFormatter(imgProxy)

	sslConfig, err := s.makeSSLConfig()
	if err != nil {
		return nil, errors.Wrap(err, "failed to make config of ssl server params")
	}

	srv := &api.Rest{
		Version:          s.Revision,
		DataService:      dataService,
		WebRoot:          s.WebRoot,
		RemarkURL:        s.RemarkURL,
		ImageProxy:       imgProxy,
		CommentFormatter: commentFormatter,
		Migrator:         migr,
		ReadOnlyAge:      s.ReadOnlyAge,
		SharedSecret:     s.SharedSecret,
		Authenticator:    authenticator,
		Cache:            loadingCache,
		NotifyService:    notifyService,
		SSLConfig:        sslConfig,
		UpdateLimiter:    s.UpdateLimit,
	}

	srv.ScoreThresholds.Low, srv.ScoreThresholds.Critical = s.LowScore, s.CriticalScore

	var devAuth *provider.DevAuthServer
	if s.Auth.Dev {
		da, err := authenticator.DevAuth()
		if err != nil {
			return nil, errors.Wrap(err, "can't make dev oauth2 server")
		}
		devAuth = da
	}

	return &serverApp{
		ServerCommand: s,
		restSrv:       srv,
		migratorSrv:   migr,
		exporter:      exporter,
		devAuth:       devAuth,
		dataService:   dataService,
		avatarStore:   avatarStore,
		notifyService: notifyService,
		terminated:    make(chan struct{}),
	}, nil
}

// Run all application objects
func (a *serverApp) run(ctx context.Context) error {
	if a.AdminPasswd != "" {
		log.Printf("[WARN] admin basic auth enabled")
	}

	go func() {
		// shutdown on context cancellation
		<-ctx.Done()
		log.Print("[INFO] shutdown initiated")
		a.restSrv.Shutdown()
		if a.devAuth != nil {
			a.devAuth.Shutdown()
		}
		if e := a.dataService.Close(); e != nil {
			log.Printf("[WARN] failed to close data store, %s", e)
		}
		if e := a.avatarStore.Close(); e != nil {
			log.Printf("[WARN] failed to close avatar store, %s", e)
		}
		a.notifyService.Close()
		log.Print("[INFO] shutdown completed")
	}()
	a.activateBackup(ctx) // runs in goroutine for each site
	if a.Auth.Dev {
		go a.devAuth.Run(context.Background()) // dev oauth2 server on :8084
	}
	a.restSrv.Run(a.Port)
	close(a.terminated)
	return nil
}

// Wait for application completion (termination)
func (a *serverApp) Wait() {
	<-a.terminated
}

// activateBackup runs background backups for each site
func (a *serverApp) activateBackup(ctx context.Context) {
	for _, siteID := range a.Sites {
		backup := migrator.AutoBackup{
			Exporter:       a.exporter,
			BackupLocation: a.BackupLocation,
			SiteID:         siteID,
			KeepMax:        a.MaxBackupFiles,
			Duration:       24 * time.Hour,
		}
		go backup.Do(ctx)
	}
}

// makeDataStore creates store for all sites
func (s *ServerCommand) makeDataStore() (result engine.Interface, err error) {
	log.Printf("[INFO] make data store, type=%s", s.Store.Type)

	switch s.Store.Type {
	case "bolt":
		if err = makeDirs(s.Store.Bolt.Path); err != nil {
			return nil, errors.Wrap(err, "failed to create bolt store")
		}
		sites := []engine.BoltSite{}
		for _, site := range s.Sites {
			sites = append(sites, engine.BoltSite{SiteID: site, FileName: fmt.Sprintf("%s/%s.db", s.Store.Bolt.Path, site)})
		}
		result, err = engine.NewBoltDB(bolt.Options{Timeout: s.Store.Bolt.Timeout}, sites...)
	case "mongo":
		mgServer, e := s.makeMongo()
		if e != nil {
			return result, errors.Wrap(e, "failed to create mongo server")
		}
		conn := mongo.NewConnection(mgServer, s.Mongo.DB, "")
		result, err = engine.NewMongo(conn, 500, 100*time.Millisecond)
	default:
		return nil, errors.Errorf("unsupported store type %s", s.Store.Type)
	}
	return result, errors.Wrap(err, "can't initialize data store")
}

func (s *ServerCommand) makeAvatarStore() (avatar.Store, error) {
	log.Printf("[INFO] make avatar store, type=%s", s.Avatar.Type)

	switch s.Avatar.Type {
	case "fs":
		if err := makeDirs(s.Avatar.FS.Path); err != nil {
			return nil, err
		}
		return avatar.NewLocalFS(s.Avatar.FS.Path), nil
	case "mongo":
		mgServer, err := s.makeMongo()
		if err != nil {
			return nil, errors.Wrap(err, "failed to create mongo server")
		}
		conn := mongo.NewConnection(mgServer, s.Mongo.DB, "")
		return avatar.NewGridFS(conn), nil
	case "bolt":
		if err := makeDirs(path.Dir(s.Avatar.Bolt.File)); err != nil {
			return nil, err
		}
		return avatar.NewBoltDB(s.Avatar.Bolt.File, bolt.Options{})
	}
	return nil, errors.Errorf("unsupported avatar store type %s", s.Avatar.Type)
}

func (s *ServerCommand) makeAdminStore() (admin.Store, error) {
	log.Printf("[INFO] make admin store, type=%s", s.Admin.Type)

	switch s.Admin.Type {
	case "shared":
		if s.Admin.Shared.Email == "" { // no admin email, use admin@domain
			if u, err := url.Parse(s.RemarkURL); err == nil {
				s.Admin.Shared.Email = "admin@" + u.Host
			}
		}
		return admin.NewStaticStore(s.SharedSecret, s.Admin.Shared.Admins, s.Admin.Shared.Email), nil
	case "mongo":
		mgServer, e := s.makeMongo()
		if e != nil {
			return nil, errors.Wrap(e, "failed to create mongo server")
		}
		conn := mongo.NewConnection(mgServer, s.Mongo.DB, "admin")
		return admin.NewMongoStore(conn, s.SharedSecret), nil
	default:
		return nil, errors.Errorf("unsupported admin store type %s", s.Admin.Type)
	}
}

func (s *ServerCommand) makeCache() (cache.LoadingCache, error) {
	log.Printf("[INFO] make cache, type=%s", s.Cache.Type)
	switch s.Cache.Type {
	case "mem":
		return cache.NewMemoryCache(cache.MaxCacheSize(s.Cache.Max.Size), cache.MaxValSize(s.Cache.Max.Value),
			cache.MaxKeys(s.Cache.Max.Items))
	// case "mongo":
	// 	mgServer, err := s.makeMongo()
	// 	if err != nil {
	// 		return nil, errors.Wrap(err, "failed to create mongo server")
	// 	}
	// 	conn := mongo.NewConnection(mgServer, s.Mongo.DB, "cache")
	// 	return cache.NewMongoCache(conn, cache.MaxCacheSize(s.Cache.Max.Size), cache.MaxValSize(s.Cache.Max.Value),
	// 		cache.MaxKeys(s.Cache.Max.Items))
	case "none":
		return &cache.Nop{}, nil
	}
	return nil, errors.Errorf("unsupported cache type %s", s.Cache.Type)
}

func (s *ServerCommand) makeMongo() (result *mongo.Server, err error) {
	if s.Mongo.URL == "" {
		return nil, errors.New("no mongo URL provided")
	}
	return mongo.NewServerWithURL(s.Mongo.URL, 10*time.Second)
}

func (s *ServerCommand) addAuthProviders(authenticator *auth.Service) {

	providers := 0
	if s.Auth.Google.CID != "" && s.Auth.Google.CSEC != "" {
		authenticator.AddProvider("google", s.Auth.Google.CID, s.Auth.Google.CSEC)
		providers++
	}
	if s.Auth.Github.CID != "" && s.Auth.Github.CSEC != "" {
		authenticator.AddProvider("github", s.Auth.Github.CID, s.Auth.Github.CSEC)
		providers++
	}
	if s.Auth.Facebook.CID != "" && s.Auth.Facebook.CSEC != "" {
		authenticator.AddProvider("facebook", s.Auth.Facebook.CID, s.Auth.Facebook.CSEC)
		providers++
	}
	if s.Auth.Yandex.CID != "" && s.Auth.Yandex.CSEC != "" {
		authenticator.AddProvider("yandex", s.Auth.Yandex.CID, s.Auth.Yandex.CSEC)
		providers++
	}
	if s.Auth.Dev {
		authenticator.AddProvider("dev", "", "")
		providers++
	}

	if providers == 0 {
		log.Printf("[WARN] no auth providers defined")
	}
}

func (s *ServerCommand) makeNotify(dataStore *service.DataStore) (*notify.Service, error) {
	log.Printf("[INFO] make notify, type=%s", s.Notify.Type)
	switch s.Notify.Type {
	case "telegram":
		tg, err := notify.NewTelegram(s.Notify.Telegram.Token, s.Notify.Telegram.Channel,
			s.Notify.Telegram.Timeout, s.Notify.Telegram.API)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create telegram notification destination")
		}
		return notify.NewService(dataStore, s.Notify.QueueSize, tg), nil
	case "none":
		return notify.NopService, nil
	}
	return nil, errors.Errorf("unsupported notification type %q", s.Notify.Type)
}

func (s *ServerCommand) makeSSLConfig() (config api.SSLConfig, err error) {
	switch s.SSL.Type {
	case "none":
		config.SSLMode = api.None
	case "static":
		if s.SSL.Cert == "" {
			return config, errors.New("path to cert.pem is required")
		}
		if s.SSL.Key == "" {
			return config, errors.New("path to key.pem is required")
		}
		config.SSLMode = api.Static
		config.Port = s.SSL.Port
		config.Cert = s.SSL.Cert
		config.Key = s.SSL.Key
	case "auto":
		config.SSLMode = api.Auto
		config.Port = s.SSL.Port
		config.ACMELocation = s.SSL.ACMELocation
		if s.SSL.ACMEEmail != "" {
			config.ACMEEmail = s.SSL.ACMEEmail
		} else if s.Admin.Type == "shared" && s.Admin.Shared.Email != "" {
			config.ACMEEmail = s.Admin.Shared.Email
		} else if u, e := url.Parse(s.RemarkURL); e == nil {
			config.ACMEEmail = "admin@" + u.Hostname()
		}
	}
	return config, err
}

func (s *ServerCommand) makeAuthenticator(ds *service.DataStore, avas avatar.Store, admns admin.Store) *auth.Service {
	authenticator := auth.NewService(auth.Opts{
		URL:            strings.TrimSuffix(s.RemarkURL, "/"),
		Issuer:         "remark42",
		TokenDuration:  s.Auth.TTL.JWT,
		CookieDuration: s.Auth.TTL.Cookie,
		SecureCookies:  strings.HasPrefix(s.RemarkURL, "https://"),
		SecretReader: token.SecretFunc(func() (string, error) { // get secret per site
			return admns.Key()
		}),
		ClaimsUpd: token.ClaimsUpdFunc(func(c token.Claims) token.Claims { // set attributes, on new token or refresh
			if c.User == nil {
				return c
			}
			c.User.SetAdmin(ds.IsAdmin(c.Audience, c.User.ID))
			c.User.SetBoolAttr("blocked", ds.IsBlocked(c.Audience, c.User.ID))
			return c
		}),
		AdminPasswd: s.AdminPasswd,
		Validator: token.ValidatorFunc(func(token string, claims token.Claims) bool { // check on each auth call (in middleware)
			if claims.User == nil {
				return false
			}
			return !claims.User.BoolAttr("blocked")
		}),
		AvatarStore:       avas,
		AvatarResizeLimit: s.Avatar.RszLmt,
		AvatarRoutePath:   "/api/v1/avatar",
		Logger:            log.Default(),
		RefreshCache:      newAuthRefreshCache(),
	})
	s.addAuthProviders(authenticator)
	return authenticator
}

// authRefreshCache used by authenticator to minimize repeatable token refreshes
type authRefreshCache struct {
	*auth_cache.Cache
}

func newAuthRefreshCache() *authRefreshCache {
	return &authRefreshCache{Cache: auth_cache.New(5*time.Minute, 10*time.Minute)}
}

func (c *authRefreshCache) Get(key interface{}) (interface{}, bool) {
	return c.Cache.Get(key.(string))
}

func (c *authRefreshCache) Set(key, value interface{}) {
	c.Cache.Set(key.(string), value, auth_cache.DefaultExpiration)
}
