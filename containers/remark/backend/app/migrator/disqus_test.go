package migrator

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/coreos/bbolt"
	"github.com/stretchr/testify/require"
	"github.com/umputun/remark/backend/app/store"
	"github.com/umputun/remark/backend/app/store/admin"
	"github.com/umputun/remark/backend/app/store/engine"
	"github.com/umputun/remark/backend/app/store/service"

	"github.com/stretchr/testify/assert"
)

func TestDisqus_Import(t *testing.T) {
	defer os.Remove("/tmp/remark-test.db")
	b, err := engine.NewBoltDB(bolt.Options{}, engine.BoltSite{FileName: "/tmp/remark-test.db", SiteID: "test"})
	require.Nil(t, err, "create store")
	dataStore := service.DataStore{Interface: b, AdminStore: admin.NewStaticStore("12345", []string{}, "")}
	d := Disqus{DataStore: &dataStore}
	size, err := d.Import(strings.NewReader(xmlTestDisqus), "test")
	assert.Nil(t, err)
	assert.Equal(t, 3, size)

	last, err := dataStore.Last("test", 10)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(last), "3 comments imported")

	c := last[0]
	assert.True(t, strings.HasPrefix(c.Text, "<p>Google App Engine"))
	assert.Equal(t, "299986072", c.ID)
	assert.Equal(t, "", c.ParentID)
	assert.Equal(t, store.Locator{SiteID: "test", URL: "http://radio-t.umputun.com/2011/03/229_8880.html"}, c.Locator)
	assert.Equal(t, "Dmitry Noname", c.User.Name)
	assert.Equal(t, "disqus_8799342cdf328253e03313958ffc6a433659d7ff", c.User.ID)
	assert.Equal(t, "7001968ea3f6c9013a9f0a3650f200c10c927638", c.User.IP)

	posts, err := dataStore.List("test", 0, 0)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(posts), "2 posts")

	count, err := dataStore.Count(store.Locator{SiteID: "test", URL: "https://radio-t.com/p/2011/03/05/podcast-229/"})
	assert.Nil(t, err)
	assert.Equal(t, 2, count)
}

func TestDisqus_Convert(t *testing.T) {
	d := Disqus{}
	ch := d.convert(strings.NewReader(xmlTestDisqus), "test")

	res := []store.Comment{}
	for comment := range ch {
		res = append(res, comment)
	}
	assert.Equal(t, 3, len(res), "3 comments total, 1 spam excluded")

	exp0 := store.Comment{
		ID: "299619020",
		Locator: store.Locator{
			SiteID: "test",
			URL:    "https://radio-t.com/p/2011/03/05/podcast-229/",
		},
		Text: `<p>The quick brown fox jumps over the lazy dog.</p><p><a href="https://https://radio-t.com" rel="nofollow noopener" title="radio-t">some link</a></p>`,
		User: store.User{
			Name: "Alexander Blah",
			ID:   "disqus_328c8b68974aef73785f6b38c3d3fedfdf941434",
			IP:   "178.178.178.178",
		},
	}
	exp0.Timestamp, _ = time.Parse("2006-01-02T15:04:05Z", "2011-08-31T15:16:29Z")
	assert.Equal(t, exp0, res[0])
}

var xmlTestDisqus = `<?xml version="1.0" encoding="utf-8"?>
<disqus xmlns="http://disqus.com" xmlns:dsq="http://disqus.com/disqus-internals" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://disqus.com/api/schemas/1.0/disqus.xsd http://disqus.com/api/schemas/1.0/disqus-internals.xsd">

	<category dsq:id="707279">
		<forum>radiot</forum>
		<title>General</title>
		<isDefault>true</isDefault>
	</category>

	<thread dsq:id="247918464">
		<id/>
		<forum>radiot</forum>
		<category dsq:id="707279"/>
		<link>http://radio-t.umputun.com/2011/03/229_8880.html</link>
		<title>Радио-Т: Радио-Т 229</title>
		<message/>
		<createdAt>2011-03-07T20:46:25Z</createdAt>
		<author>
			<email>umputun@gmail.com</email>
			<name>Umputun</name>
			<isAnonymous>false</isAnonymous>
			<username>umputun</username>
		</author>
		<ipAddress>98.212.28.115</ipAddress>
		<isClosed>false</isClosed>
		<isDeleted>false</isDeleted>
	</thread>
	<thread dsq:id="247937687">
		<id>http://www.radio-t.com/p/2011/03/05/podcast-229/</id>
		<forum>radiot</forum>
		<category dsq:id="707279"/>
		<link>https://radio-t.com/p/2011/03/05/podcast-229/</link>
		<title>Радио-Т: Радио-Т 229</title>
		<message/>
		<createdAt>2011-03-07T21:17:17Z</createdAt>
		<author>
			<email>umputun@gmail.com</email>
			<name>Umputun</name>
			<isAnonymous>false</isAnonymous>
			<username>umputun</username>
		</author>
		<ipAddress>80.250.214.235</ipAddress>
		<isClosed>true</isClosed>
		<isDeleted>false</isDeleted>
	</thread>

	<post dsq:id="299619020">
		<id>3565798471341011339</id>
		<message>
			<![CDATA[<p>The quick brown fox jumps over the lazy dog.</p><p><a href="https://https://radio-t.com" rel="nofollow noopener" title="radio-t">some link</a></p>]]>
		</message>
		<createdAt>2011-08-31T15:16:29Z</createdAt>
		<isDeleted>false</isDeleted>
		<isSpam>false</isSpam>
		<author>
			<email/>
			<name>Alexander Blah</name>
			<isAnonymous>false</isAnonymous>
			<username>facebook-1787732238</username>
		</author>
		<ipAddress>178.178.178.178</ipAddress>
		<thread dsq:id="247937687"/>
	</post>

	<post dsq:id="299744309">
		<id>3029154520436241933</id>
		<message>
			<![CDATA[<p>Microsoft показал проводник Windows 8 с ленточным интерфейсом.</p><p><a href="http://blogs.msdn.com/b/b8/archive/2011/08/29/improvements-in-windows-explorer.aspx" rel="nofollow noopener" title="http://blogs.msdn.com/b/b8/archive/2011/08/29/improvements-in-windows-explorer.aspx">http://blogs.msdn.com/b/b8/...</a> </p>]]>
		</message>
		<createdAt>2011-08-31T17:44:22Z</createdAt>
		<isDeleted>false</isDeleted>
		<isSpam>false</isSpam>
		<author>
			<email>mihail.noname@gmail.com</email>
			<name>mikhail</name>
			<isAnonymous>false</isAnonymous>
			<username>mikhail-noname</username>
		</author>
		<ipAddress>195.195.195.139</ipAddress>
		<thread dsq:id="247937687"/>
	</post>

	<post dsq:id="299986072">
		<id>6580890074280459209</id>
		<message>
			<![CDATA[<p>Google App Engine скоро выходит из превью статуса.</p><p>Сейчас письмо пришло от гугла.</p><p>Для платных приложений использущих High Replication Datastore (HRD) будет 99.95% uptime SLA.<br>Будут Премьер аккаунты за 500 баксов/месяц с оперативной поддержкой и любым количеством приложений на аккаунте (+ плата за потребленные ресурсы).<br>В связи с переходом на новую систему оплаты, обещают снизить бесплатные квоты.<br>Всем кто включит биллинг до 31 октября, обещают 50 баксов :)</p>]]>
		</message>
		<createdAt>2011-08-31T22:48:43Z</createdAt>
		<isDeleted>false</isDeleted>
		<isSpam>false</isSpam>
		<author>
			<email>dmitri.noname@gmail.com</email>
			<name>Dmitry Noname</name>
			<isAnonymous>false</isAnonymous>
			<username>google-74b9e7568ef6860e93862c5d77590123</username>
		</author>
		<ipAddress>89.89.89.139</ipAddress>
		<thread dsq:id="247918464"/>
	</post>

	<post dsq:id="299986073">
		<id>6580890074280459219</id>
		<message>some ugly spam</message>
		<createdAt>2011-09-30T22:48:43Z</createdAt>
		<isDeleted>false</isDeleted>
		<isSpam>true</isSpam>
		<author>
			<email>spam.noname@gmail.com</email>
			<name>Spam Noname</name>
			<isAnonymous>false</isAnonymous>
			<username>google-2c5d77590123</username>
		</author>
		<ipAddress>189.89.89.139</ipAddress>
		<thread dsq:id="247937687"/>
	</post>
</disqus>
`
