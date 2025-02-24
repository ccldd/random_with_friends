package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/suite"
)

type ServerTestSuite struct {
	suite.Suite
	App    *App
	Server *httptest.Server
}

func (suite *ServerTestSuite) SetupTest() {
	suite.App = NewApp()
	suite.Server = httptest.NewServer(routes(suite.App))
}

func (suite *ServerTestSuite) TearDownTest() {
	suite.Server.Close()
}

func (suite *ServerTestSuite) connectWS(roomId RoomId, name string) (*websocket.Conn, *http.Response, error) {
	u, err := url.Parse(suite.Server.URL)
	suite.Require().NoError(err)
	u.Scheme = "ws"
	u.Path = "/ws/" + roomId
	q := u.Query()
	q.Set("name", name)
	u.RawQuery = q.Encode()

	conn, resp, err := websocket.DefaultDialer.Dial(u.String(), nil)
	return conn, resp, err
}

func (suite *ServerTestSuite) mustConnectWSNo(roomId RoomId, name string) (*websocket.Conn, *http.Response) {
	conn, resp, err := suite.connectWS(roomId, name)
	suite.Require().NotNil(conn)
	suite.Require().NotNil(resp)
	suite.Require().NoError(err)
	return conn, resp
}

func (suite *ServerTestSuite) TestCreateRoomNoName() {
	resp, err := suite.Server.Client().Post(suite.Server.URL+"/create", "application/json", nil)
	suite.NoError(err)
	suite.NotNil(resp)
	suite.Equal(http.StatusBadRequest, resp.StatusCode)
}

func (suite *ServerTestSuite) TestCreateRoomOk() {
	resp, err := suite.Server.Client().Post(suite.Server.URL+"/create?name=foo", "application/json", nil)
	suite.NoError(err)
	suite.NotNil(resp)
	suite.Equal(http.StatusOK, resp.StatusCode)
}

func (suite *ServerTestSuite) TestWSNoRoomId() {
	_, resp, err := suite.connectWS("", "")
	suite.Error(err)
	suite.Equal(http.StatusNotFound, resp.StatusCode, "should be 404 because it room is missing and doesn't match any routes")
}

func (suite *ServerTestSuite) TestWSNoName() {
	_, resp, err := suite.connectWS("a", "")
	suite.Error(err)
	suite.Equal(http.StatusBadRequest, resp.StatusCode, "should be 404 because it room is missing and doesn't match any routes")
}

func (suite *ServerTestSuite) TestApp() {
	suite.NotNil(suite.App)
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(ServerTestSuite))
}
