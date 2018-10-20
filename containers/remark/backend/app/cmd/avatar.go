package cmd

import (
	"log"
	"path"
	"time"

	"github.com/coreos/bbolt"
	"github.com/go-pkgz/mongo"
	"github.com/pkg/errors"

	"github.com/umputun/remark/backend/app/store/avatar"
)

// AvatarCommand set of flags and command for avatar migration
// it converts all avatars from src.type to dst.type.
// Note: it is possible to run migration for the same types (src = dst) in order to resize all avatars.
type AvatarCommand struct {
	AvatarSrc AvatarGroup `group:"src" namespace:"src"`
	AvatarDst AvatarGroup `group:"dst" namespace:"dst"`
	Mongo     MongoGroup  `group:"mongo" namespace:"mongo" env-namespace:"MONGO"`

	migrator AvatarMigrator
	CommonOpts
}

// AvatarMigrator defines interface for migration
type AvatarMigrator interface {
	Migrate(avatar.Store, avatar.Store) (int, error)
}

type avatarMigrator struct{}

func (a avatarMigrator) Migrate(dst, src avatar.Store) (int, error) {
	return avatar.Migrate(dst, src)
}

// Execute runs  with AvatarCommand parameters, entry point for "avatar" command
func (ac *AvatarCommand) Execute(args []string) error {
	log.Printf("[INFO] migrate avatars from %s to %s", ac.AvatarSrc.Type, ac.AvatarDst.Type)

	src, err := ac.makeAvatarStore(ac.AvatarSrc)
	if err != nil {
		return errors.Wrapf(err, "can't make avatart store for %s", ac.AvatarSrc.Type)
	}

	dst, err := ac.makeAvatarStore(ac.AvatarDst)
	if err != nil {
		return errors.Wrapf(err, "can't make avatart store for %s", ac.AvatarDst.Type)
	}

	if ac.migrator == nil {
		ac.migrator = avatarMigrator{}
	}

	count, err := ac.migrator.Migrate(dst, src)
	if err != nil {
		return err
	}

	if err = dst.Close(); err != nil {
		log.Printf("[WARN] failed to close dst store %s", ac.AvatarDst.Type)
	}
	if err = src.Close(); err != nil {
		log.Printf("[WARN] failed to close src store %s", ac.AvatarSrc.Type)
	}

	log.Printf("[INFO] completed, migrated avatars = %d", count)
	return nil
}

func (ac *AvatarCommand) makeAvatarStore(gr AvatarGroup) (avatar.Store, error) {
	log.Printf("[DEBUG] make avatar store, type=%s", gr.Type)
	switch gr.Type {
	case "fs":
		if err := makeDirs(gr.FS.Path); err != nil {
			return nil, err
		}
		return avatar.NewLocalFS(gr.FS.Path, gr.RszLmt), nil
	case "mongo":
		mgServer, err := ac.makeMongo()
		if err != nil {
			return nil, errors.Wrap(err, "failed to create mongo server")
		}
		conn := mongo.NewConnection(mgServer, ac.Mongo.DB, "")
		return avatar.NewGridFS(conn, gr.RszLmt), nil
	case "bolt":
		if err := makeDirs(path.Dir(gr.Bolt.File)); err != nil {
			return nil, err
		}
		return avatar.NewBoltDB(gr.Bolt.File, bolt.Options{}, gr.RszLmt)
	}
	return nil, errors.Errorf("unsupported avatar store type %s", gr.Type)
}

func (ac *AvatarCommand) makeMongo() (result *mongo.Server, err error) {
	if ac.Mongo.URL == "" {
		return nil, errors.New("no mongo URL provided")
	}
	return mongo.NewServerWithURL(ac.Mongo.URL, 10*time.Second)
}
