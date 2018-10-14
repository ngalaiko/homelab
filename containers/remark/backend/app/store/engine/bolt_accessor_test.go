package engine

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/coreos/bbolt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/umputun/remark/backend/app/store"
)

var testDb = "test-remark.db"

func TestBoltDB_CreateAndFind(t *testing.T) {
	defer os.Remove(testDb)
	var b = prep(t)

	res, err := b.Find(store.Locator{URL: "https://radio-t.com", SiteID: "radio-t"}, "time")
	assert.Nil(t, err)
	assert.Equal(t, 2, len(res))
	assert.Equal(t, `some text, <a href="http://radio-t.com">link</a>`, res[0].Text)
	assert.Equal(t, "user1", res[0].User.ID)
	t.Log(res[0].ID)

	_, err = b.Create(store.Comment{ID: res[0].ID, Locator: store.Locator{URL: "https://radio-t.com", SiteID: "radio-t"}})
	assert.NotNil(t, err)
	assert.Equal(t, "key id-1 already in store", err.Error())

	_, err = b.Find(store.Locator{URL: "https://radio-t.com", SiteID: "radio-t-bad"}, "time")
	assert.EqualError(t, err, `site "radio-t-bad" not found`)

	assert.NoError(t, b.Close())
}

func TestBoltDB_CreateReadOnly(t *testing.T) {
	defer os.Remove(testDb)
	var b = prep(t)

	comment := store.Comment{
		ID:        "id-ro",
		Text:      `some text, <a href="http://radio-t.com">link</a>`,
		Timestamp: time.Date(2017, 12, 20, 15, 18, 22, 0, time.Local),
		Locator:   store.Locator{URL: "https://radio-t.com/ro", SiteID: "radio-t"},
		User:      store.User{ID: "user1", Name: "user name"},
	}
	err := b.SetReadOnly(comment.Locator, true)
	require.Nil(t, err)

	_, err = b.Create(comment)
	assert.NotNil(t, err)
	assert.Equal(t, "post https://radio-t.com/ro is read-only", err.Error())

	err = b.SetReadOnly(comment.Locator, false)
	require.Nil(t, err)
	_, err = b.Create(comment)
	assert.Nil(t, err)
}

func TestBoltDB_Get(t *testing.T) {
	defer os.Remove(testDb)
	b := prep(t)

	res, err := b.Find(store.Locator{URL: "https://radio-t.com", SiteID: "radio-t"}, "time")
	assert.Nil(t, err)
	assert.Equal(t, 2, len(res))

	comment, err := b.Get(store.Locator{URL: "https://radio-t.com", SiteID: "radio-t"}, res[1].ID)
	assert.Nil(t, err)
	assert.Equal(t, "some text2", comment.Text)

	comment, err = b.Get(store.Locator{URL: "https://radio-t.com", SiteID: "radio-t"}, "1234567")
	assert.NotNil(t, err)

	_, err = b.Get(store.Locator{URL: "https://radio-t.com", SiteID: "bad"}, res[1].ID)
	assert.EqualError(t, err, `site "bad" not found`)
}

func TestBoltDB_Put(t *testing.T) {
	defer os.Remove(testDb)
	b := prep(t)
	loc := store.Locator{URL: "https://radio-t.com", SiteID: "radio-t"}
	res, err := b.Find(loc, "time")
	assert.Nil(t, err)
	assert.Equal(t, 2, len(res))

	comment := res[0]
	comment.Text = "abc 123"
	comment.Score = 100
	err = b.Put(loc, comment)
	assert.Nil(t, err)

	comment, err = b.Get(loc, res[0].ID)
	assert.Nil(t, err)
	assert.Equal(t, "abc 123", comment.Text)
	assert.Equal(t, res[0].ID, comment.ID)
	assert.Equal(t, 100, comment.Score)

	err = b.Put(store.Locator{URL: "https://radio-t.com", SiteID: "bad"}, comment)
	assert.EqualError(t, err, `site "bad" not found`)

	err = b.Put(store.Locator{URL: "https://radio-t.com-bad", SiteID: "radio-t"}, comment)
	assert.EqualError(t, err, `no bucket https://radio-t.com-bad in store`)
}

func TestBoltDB_Last(t *testing.T) {
	defer os.Remove(testDb)
	b := prep(t)

	res, err := b.Last("radio-t", 0)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(res))
	assert.Equal(t, "some text2", res[0].Text)

	res, err = b.Last("radio-t", 1)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(res))
	assert.Equal(t, "some text2", res[0].Text)

	_, err = b.Last("bad", 0)
	assert.EqualError(t, err, `site "bad" not found`)
}

func TestBoltDB_Count(t *testing.T) {
	defer os.Remove(testDb)
	b := prep(t)

	c, err := b.Count(store.Locator{URL: "https://radio-t.com", SiteID: "radio-t"})
	assert.Nil(t, err)
	assert.Equal(t, 2, c)

	c, err = b.Count(store.Locator{URL: "https://radio-t.com-xxx", SiteID: "radio-t"})
	assert.Nil(t, err)
	assert.Equal(t, 0, c)

	_, err = b.Count(store.Locator{URL: "https://radio-t.com", SiteID: "bad"})
	assert.EqualError(t, err, `site "bad" not found`)
}

func TestBoltDB_List(t *testing.T) {
	defer os.Remove(testDb)
	b := prep(t) // two comments for https://radio-t.com

	// add one more for https://radio-t.com/2
	comment := store.Comment{
		ID:        "12345",
		Text:      `some text, <a href="http://radio-t.com">link</a>`,
		Timestamp: time.Date(2017, 12, 20, 15, 18, 22, 0, time.Local),
		Locator:   store.Locator{URL: "https://radio-t.com/2", SiteID: "radio-t"},
		User:      store.User{ID: "user1", Name: "user name"},
	}
	_, err := b.Create(comment)
	assert.Nil(t, err)

	ts := func(sec int) time.Time { return time.Date(2017, 12, 20, 15, 18, sec, 0, time.Local) }

	res, err := b.List("radio-t", 0, 0)
	assert.Nil(t, err)
	assert.Equal(t, []store.PostInfo{{URL: "https://radio-t.com/2", Count: 1, FirstTS: ts(22), LastTS: ts(22)},
		{URL: "https://radio-t.com", Count: 2, FirstTS: ts(22), LastTS: ts(23)}},
		res)

	res, err = b.List("radio-t", -1, -1)
	assert.Nil(t, err)
	assert.Equal(t, []store.PostInfo{{URL: "https://radio-t.com/2", Count: 1, FirstTS: ts(22), LastTS: ts(22)},
		{URL: "https://radio-t.com", Count: 2, FirstTS: ts(22), LastTS: ts(23)}}, res)

	res, err = b.List("radio-t", 1, 0)
	assert.Nil(t, err)
	assert.Equal(t, []store.PostInfo{{URL: "https://radio-t.com/2", Count: 1, FirstTS: ts(22), LastTS: ts(22)}}, res)

	res, err = b.List("radio-t", 1, 1)
	assert.Nil(t, err)
	assert.Equal(t, []store.PostInfo{{URL: "https://radio-t.com", Count: 2, FirstTS: ts(22), LastTS: ts(23)}}, res)

	_, err = b.List("bad", 1, 1)
	assert.EqualError(t, err, `site "bad" not found`)
}

func TestBoltDB_Info(t *testing.T) {
	defer os.Remove(testDb)
	b := prep(t) // two comments for https://radio-t.com

	ts := func(min int) time.Time { return time.Date(2017, 12, 20, 15, 18, min, 0, time.Local) }

	// add one more for https://radio-t.com/2
	comment := store.Comment{
		ID:        "12345",
		Text:      `some text, <a href="http://radio-t.com">link</a>`,
		Timestamp: time.Date(2017, 12, 20, 15, 18, 24, 0, time.Local),
		Locator:   store.Locator{URL: "https://radio-t.com/2", SiteID: "radio-t"},
		User:      store.User{ID: "user1", Name: "user name"},
	}
	_, err := b.Create(comment)
	assert.Nil(t, err)

	r, err := b.Info(store.Locator{URL: "https://radio-t.com/2", SiteID: "radio-t"}, 0)
	require.Nil(t, err)
	assert.Equal(t, store.PostInfo{URL: "https://radio-t.com/2", Count: 1, FirstTS: ts(24), LastTS: ts(24)}, r)

	r, err = b.Info(store.Locator{URL: "https://radio-t.com/2", SiteID: "radio-t"}, 10)
	require.Nil(t, err)
	assert.Equal(t, store.PostInfo{URL: "https://radio-t.com/2", Count: 1, FirstTS: ts(24), LastTS: ts(24), ReadOnly: true}, r)

	r, err = b.Info(store.Locator{URL: "https://radio-t.com", SiteID: "radio-t"}, 0)
	require.Nil(t, err)
	assert.Equal(t, store.PostInfo{URL: "https://radio-t.com", Count: 2, FirstTS: ts(22), LastTS: ts(23)}, r)

	_, err = b.Info(store.Locator{URL: "https://radio-t.com/error", SiteID: "radio-t"}, 0)
	require.NotNil(t, err)

	_, err = b.Info(store.Locator{URL: "https://radio-t.com", SiteID: "radio-t-error"}, 0)
	require.NotNil(t, err)

	err = b.SetReadOnly(store.Locator{URL: "https://radio-t.com/2", SiteID: "radio-t"}, true)
	require.Nil(t, err)
	r, err = b.Info(store.Locator{URL: "https://radio-t.com/2", SiteID: "radio-t"}, 0)
	require.Nil(t, err)
	assert.Equal(t, store.PostInfo{URL: "https://radio-t.com/2", Count: 1, FirstTS: ts(24), LastTS: ts(24), ReadOnly: true}, r)

}

func TestBoltDB_GetForUser(t *testing.T) {
	defer os.Remove(testDb)
	b := prep(t)

	res, err := b.User("radio-t", "user1", 5, 0)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(res))
	assert.Equal(t, "some text2", res[0].Text, "sorted by -time")

	res, err = b.User("radio-t", "user1", 1, 0)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(res), "allow 1 comment")
	assert.Equal(t, "some text2", res[0].Text, "sorted by -time")

	res, err = b.User("radio-t", "user1", 1, 1)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(res), "allow 1 comment")
	assert.Equal(t, `some text, <a href="http://radio-t.com">link</a>`, res[0].Text, "second comment")

	_, err = b.User("bad", "user1", 1, 0)
	assert.EqualError(t, err, `site "bad" not found`)

	_, err = b.User("radio-t", "userZ", 1, 0)
	assert.EqualError(t, err, `no comments for user userZ in store`)
}

func TestBoltDB_GetForUserPagination(t *testing.T) {
	os.Remove(testDb)
	b, err := NewBoltDB(bolt.Options{}, BoltSite{FileName: testDb, SiteID: "radio-t"})
	require.Nil(t, err)
	defer os.Remove(testDb)

	c := store.Comment{
		Locator: store.Locator{URL: "https://radio-t.com", SiteID: "radio-t"},
		User:    store.User{ID: "user1", Name: "user name"},
	}

	// write 200 comments
	for i := 0; i < 200; i++ {
		c.ID = fmt.Sprintf("id-%d", i)
		c.Text = fmt.Sprintf("text #%d", i)
		c.Timestamp = time.Date(2017, 12, 20, 15, 18, i, 0, time.Local)
		_, err = b.Create(c)
		require.Nil(t, err)
	}

	// get all comments
	res, err := b.User("radio-t", "user1", 0, 0)
	assert.Nil(t, err)
	assert.Equal(t, 200, len(res))
	assert.Equal(t, "id-199", res[0].ID)

	// seek 0, 5 comments
	res, err = b.User("radio-t", "user1", 5, 0)
	assert.Nil(t, err)
	assert.Equal(t, 5, len(res))
	assert.Equal(t, "id-199", res[0].ID)
	assert.Equal(t, "id-195", res[4].ID)

	// seek 10, 3 comments
	res, err = b.User("radio-t", "user1", 3, 10)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(res))
	assert.Equal(t, "id-189", res[0].ID)
	assert.Equal(t, "id-187", res[2].ID)

	// seek 195, ask 10 comments
	res, err = b.User("radio-t", "user1", 10, 195)
	assert.Nil(t, err)
	assert.Equal(t, 5, len(res))
	assert.Equal(t, "id-4", res[0].ID)
	assert.Equal(t, "id-0", res[4].ID)

	// seek 255, ask 10 comments
	res, err = b.User("radio-t", "user1", 10, 255)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(res))
}

func TestBoltDB_GetForUserCounter(t *testing.T) {
	defer os.Remove(testDb)
	b := prep(t)
	count, err := b.UserCount("radio-t", "user1")
	assert.Nil(t, err)
	assert.Equal(t, 2, count)

	_, err = b.UserCount("bad", "user1")
	assert.EqualError(t, err, `site "bad" not found`)

	_, err = b.UserCount("radio-t", "userZ")
	assert.EqualError(t, err, `no comments for user userZ in store`)
}

func TestBoltDB_Ref(t *testing.T) {
	b := BoltDB{}
	comment := store.Comment{
		ID:        "12345",
		Text:      `some text, <a href="http://radio-t.com">link</a>`,
		Timestamp: time.Date(2017, 12, 20, 15, 18, 22, 0, time.Local),
		Locator:   store.Locator{URL: "https://radio-t.com/2", SiteID: "radio-t"},
		User:      store.User{ID: "user1", Name: "user name"},
	}
	res := b.makeRef(comment)
	assert.Equal(t, "https://radio-t.com/2!!12345", string(res))

	url, id, err := b.parseRef([]byte("https://radio-t.com/2!!12345"))
	assert.Nil(t, err)
	assert.Equal(t, "https://radio-t.com/2", url)
	assert.Equal(t, "12345", id)

	_, _, err = b.parseRef([]byte("https://radio-t.com/2"))
	assert.NotNil(t, err)
}

func TestBoltDB_New(t *testing.T) {
	_, err := NewBoltDB(bolt.Options{}, BoltSite{FileName: "/tmp/no-such-place/tmp.db", SiteID: "radio-t"})
	assert.EqualError(t, err, "failed to make boltdb for /tmp/no-such-place/tmp.db: open /tmp/no-such-place/tmp.db: no such file or directory")
}

// makes new boltdb, put two records
func prep(t *testing.T) *BoltDB {
	os.Remove(testDb)

	boltStore, err := NewBoltDB(bolt.Options{}, BoltSite{FileName: testDb, SiteID: "radio-t"})
	assert.Nil(t, err)
	b := boltStore

	comment := store.Comment{
		ID:        "id-1",
		Text:      `some text, <a href="http://radio-t.com">link</a>`,
		Timestamp: time.Date(2017, 12, 20, 15, 18, 22, 0, time.Local),
		Locator:   store.Locator{URL: "https://radio-t.com", SiteID: "radio-t"},
		User:      store.User{ID: "user1", Name: "user name"},
	}
	_, err = b.Create(comment)
	assert.Nil(t, err)

	comment = store.Comment{
		ID:        "id-2",
		Text:      "some text2",
		Timestamp: time.Date(2017, 12, 20, 15, 18, 23, 0, time.Local),
		Locator:   store.Locator{URL: "https://radio-t.com", SiteID: "radio-t"},
		User:      store.User{ID: "user1", Name: "user name"},
	}
	_, err = b.Create(comment)
	assert.Nil(t, err)

	return b
}
