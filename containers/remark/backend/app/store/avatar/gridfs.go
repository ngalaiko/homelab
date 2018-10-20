package avatar

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"time"

	"github.com/globalsign/mgo"
	"github.com/go-pkgz/mongo"
	"github.com/pkg/errors"

	"github.com/umputun/remark/backend/app/store"
)

// NewGridFS makes gridfs (mongo) avatar store
func NewGridFS(conn *mongo.Connection, resizeLimit int) *GridFS {
	return &GridFS{Connection: conn, resizeLimit: resizeLimit}
}

// GridFS implements Store for GridFS
type GridFS struct {
	Connection  *mongo.Connection
	resizeLimit int
}

// Put avatar to gridfs object, try to resize
func (gf *GridFS) Put(userID string, reader io.Reader) (avatar string, err error) {
	id := encodeID(userID)
	err = gf.Connection.WithDB(func(dbase *mgo.Database) error {
		fh, e := dbase.GridFS("fs").Create(id + imgSfx)
		if e != nil {
			return e
		}
		defer func() {
			if err = fh.Close(); err != nil {
				log.Printf("[WARN] can't close avatar file %v, %s", fh, err)
			}
		}()

		// Trying to resize avatar.
		if reader = resize(reader, gf.resizeLimit); reader == nil {
			return errors.New("avatar resize reader is nil")
		}
		_, e = io.Copy(fh, reader)
		return e
	})
	return id + imgSfx, err
}

// Get avatar reader for avatar id.image
func (gf *GridFS) Get(avatar string) (reader io.ReadCloser, size int, err error) {
	buf := &bytes.Buffer{}
	err = gf.Connection.WithDB(func(dbase *mgo.Database) error {
		fh, e := dbase.GridFS("fs").Open(avatar)
		if e != nil {
			return errors.Wrapf(e, "can't load avatar %s", avatar)
		}
		if _, e = io.Copy(buf, fh); e != nil {
			return errors.Wrapf(e, "can't copy avatar %s", avatar)
		}
		size = int(fh.Size())
		return fh.Close()
	})
	return ioutil.NopCloser(buf), size, err
}

// ID returns a fingerprint of the avatar content. Uses MD5 because gridfs provides it directly
func (gf *GridFS) ID(avatar string) (id string) {
	err := gf.Connection.WithDB(func(dbase *mgo.Database) error {
		fh, e := dbase.GridFS("fs").Open(avatar)
		if e != nil {
			return errors.Wrapf(e, "can't open avatar %s", avatar)
		}
		id = fh.MD5()
		return errors.Wrapf(fh.Close(), "can't close avatar")
	})
	if err != nil {
		log.Printf("[DEBUG] can't get file info '%s', %s", avatar, err)
		return store.EncodeID(avatar)
	}
	return id
}

// Remove avatar from gridfs
func (gf *GridFS) Remove(avatar string) error {
	return gf.Connection.WithDB(func(dbase *mgo.Database) error {
		fh, e := dbase.GridFS("fs").Open(avatar)
		if e != nil {
			return errors.Wrapf(e, "can't get avatar %s", avatar)
		}
		if e = fh.Close(); e != nil {
			log.Printf("[WARN] can't close avatar %s, %s", avatar, e)
		}
		return dbase.GridFS("fs").Remove(avatar)
	})
}

// List all avatars (ids) on gfs
// note: id includes .image suffix
func (gf *GridFS) List() (ids []string, err error) {

	type gfsFile struct {
		UploadDate time.Time `bson:"uploadDate"`
		Length     int64     `bson:",minsize"`
		MD5        string
		Filename   string `bson:",omitempty"`
	}

	files := []gfsFile{}
	err = gf.Connection.WithDB(func(dbase *mgo.Database) error {
		return dbase.GridFS("fs").Find(nil).All(&files)
	})

	for _, f := range files {
		ids = append(ids, f.Filename)
	}
	return ids, errors.Wrap(err, "can't list avatars")
}

// Close gridfs does nothing but satisfies interface
func (gf *GridFS) Close() error {
	return nil
}
