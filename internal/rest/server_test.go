package rest

import (
	"bytes"
	"encoding/json"
	"github.com/ICBX/penguin/pkg/common"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type TestSuite struct {
	db *gorm.DB
	s  *Server
	suite.Suite
}

func TestTestSuite(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"))
	if err != nil {
		t.Error(err)
		return
	}
	if err = db.AutoMigrate(common.TableModels...); err != nil {
		t.Error(err)
		return
	}
	s := New(db)
	suite.Run(t, &TestSuite{
		db: db,
		s:  s,
	})
}

func (suite *TestSuite) TestURL() {
	assert.Equal(suite.T(), "/hello/world", suite.url("/:a/:b", "a", "hello", "b", "world"))
	assert.Equal(suite.T(), "/media/video/hello", suite.url(RouteDeleteVideo, VideoIDKey, "hello"))
}

func (suite *TestSuite) TestVideoCycle() {
	var res *http.Response
	/// create video
	// assert no videos exists (yet)
	assert.Equal(suite.T(), 0, len(suite.utilFindVideos()))

	// create video
	res = suite.jsonReq("POST", RouteAddVideo, newVideoPayload{VideoID: "hello"})
	suite.assert(res, fiber.StatusCreated)
	assert.Equal(suite.T(), 1, len(suite.utilFindVideos()))

	/// disable video
	res = suite.req("DELETE", suite.url(RouteDeleteVideo, VideoIDKey, "hello"))
	suite.assert(res, fiber.StatusCreated)
	assert.Equal(suite.T(), 0, len(suite.utilFindVideos()))

	/// re-enable video
	res = suite.req("DELETE", suite.url(RouteDeleteVideo, VideoIDKey, "hello")+"?state=enable")
	suite.assert(res, fiber.StatusCreated)
	assert.Equal(suite.T(), 1, len(suite.utilFindVideos()))

	/// completely disable video
	res = suite.req("DELETE", suite.url(RouteDeleteVideo, VideoIDKey, "hello")+"?perm=yes")
	suite.assert(res, fiber.StatusCreated)
	assert.Equal(suite.T(), 0, len(suite.utilFindVideos()))

	/// re-enable video
	res = suite.req("DELETE", suite.url(RouteDeleteVideo, VideoIDKey, "hello")+"?state=enable")
	suite.assert(res, fiber.StatusNotFound)
	assert.Equal(suite.T(), 0, len(suite.utilFindVideos()))

	/// invalid state
	res = suite.req("DELETE", suite.url(RouteDeleteVideo, VideoIDKey, "hello")+"?state=braun")
	suite.assert(res, fiber.StatusBadRequest)

	// create video
	res = suite.jsonReq("POST", suite.url(RouteAddVideo, VideoIDKey, "hello"), newVideoPayload{VideoID: "hello"})
	suite.assert(res, fiber.StatusCreated)
	assert.Equal(suite.T(), 1, len(suite.utilFindVideos()))

	/// create blobber
	// assert that none exist yet
	assert.Equal(suite.T(), 0, len(suite.utilFindBlobber()))

	// TODO: assert that missing secret and name throws error

	// create blobber
	res = suite.jsonReq("POST", suite.url(RouteAddBlobber), newBlobberPayload{Name: "blobby", Secret: "blobby"})
	suite.assert(res, fiber.StatusCreated)
	assert.Equal(suite.T(), 1, len(suite.utilFindBlobber()))

	/// add blobber to video
	res = suite.jsonReq("POST", suite.url(RouteAddBlobberToVideo, VideoIDKey, "hello"), newVideoBlobberPayload{BlobberID: 1})
	suite.assert(res, fiber.StatusCreated)
	assert.Equal(suite.T(), 1, len(suite.utilFindQueue()))

	/// remove blobber from video
	res = suite.req("DELETE", "/media/videos/hello/blobber/1")
	res = suite.req("DELETE", suite.url(RouteRemoveBlobberFromVideo, VideoIDKey, "hello", BlobberIDKey, "1"))
	suite.assert(res, fiber.StatusCreated)
	assert.Equal(suite.T(), 1, len(suite.utilFindQueue()))

}

func (suite *TestSuite) assert(res *http.Response, status int) {
	if res.StatusCode != status {
		d, _ := io.ReadAll(res.Body)
		assert.Equal(suite.T(), status, res.StatusCode, string(d))
	} else {
		assert.Equal(suite.T(), status, res.StatusCode)
	}
}

func (suite *TestSuite) utilFindVideos() []*common.Video {
	var videos []*common.Video
	err := suite.db.Model(&common.Video{}).Find(&videos).Error
	assert.NoError(suite.T(), err, "finding videos")
	return videos
}

func (suite *TestSuite) utilFindBlobber() []*common.BlobDownloader {
	var blobber []*common.BlobDownloader
	err := suite.db.Model(&common.BlobDownloader{}).Find(&blobber).Error
	assert.NoError(suite.T(), err, "finding blobber")
	return blobber
}

func (suite *TestSuite) utilFindQueue() []*common.Queue {
	var queues []*common.Queue
	err := suite.db.Model(&common.Queue{}).Find(&queues).Error
	assert.NoError(suite.T(), err, "finding queues")
	return queues
}

func (suite *TestSuite) jsonReq(typ, route string, val interface{}) *http.Response {
	// marshal json data
	data, err := json.Marshal(val)
	assert.NoError(suite.T(), err, "marshal data")
	// create request
	return suite.reqAdv(typ, route, http.Header{
		"Content-Type": []string{fiber.MIMEApplicationJSON},
	}, bytes.NewReader(data))
}

func (suite *TestSuite) req(typ, route string) *http.Response {
	return suite.reqAdv(typ, route, nil, nil)
}

func (suite *TestSuite) reqAdv(typ, route string, h http.Header, body io.Reader) *http.Response {
	req := httptest.NewRequest(typ, route, body)
	if h != nil {
		for k, v := range h {
			req.Header[k] = v
		}
	}
	resp, err := suite.s.app.Test(req, -1)
	assert.NoError(suite.T(), err, "fiber test")
	return resp
}

func (suite *TestSuite) url(orig string, replacements ...string) (res string) {
	res = orig
	if len(replacements)&1 == 0 {
		for i := 0; i < len(replacements); i += 2 {
			res = strings.Replace(res, ":"+replacements[i], replacements[i+1], -1)
		}
	}
	return
}
