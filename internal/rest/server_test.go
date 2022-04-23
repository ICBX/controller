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

func (suite *TestSuite) TestVideoCycle() {
	var res *http.Response
	/// create video
	// assert no videos exists (yet)
	assert.Equal(suite.T(), 0, len(suite.utilFindVideos()))

	// create video
	res = suite.jsonReq("POST", "/video/add", newVideoPayload{VideoID: "hello"})
	suite.assert(res, fiber.StatusCreated)
	assert.Equal(suite.T(), 1, len(suite.utilFindVideos()))

	/// disable video
	res = suite.req("DELETE", "/media/videos/hello")
	suite.assert(res, fiber.StatusCreated)
	assert.Equal(suite.T(), 0, len(suite.utilFindVideos()))

	/// re-enable video
	res = suite.req("DELETE", "/media/videos/hello?state=enable")
	suite.assert(res, fiber.StatusCreated)
	assert.Equal(suite.T(), 1, len(suite.utilFindVideos()))

	/// completely disable video
	res = suite.req("DELETE", "/media/videos/hello?perm=yes")
	suite.assert(res, fiber.StatusCreated)
	assert.Equal(suite.T(), 0, len(suite.utilFindVideos()))

	/// re-enable video
	res = suite.req("DELETE", "/media/videos/hello?state=enable")
	suite.assert(res, fiber.StatusNotFound)
	assert.Equal(suite.T(), 0, len(suite.utilFindVideos()))

	/// invalid state
	res = suite.req("DELETE", "/media/videos/hello?state=braun")
	suite.assert(res, fiber.StatusBadRequest)
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
