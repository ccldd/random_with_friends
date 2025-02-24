package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

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
	suite.Len(suite.App.rooms, 1, "should have created a room")

	loc, err := resp.Request.Response.Location()
	suite.Require().NoError(err)
	suite.Require().NotEmpty(loc)

	roomId := loc.Query().Get("roomId")
	suite.Contains(suite.App.rooms, roomId)

	// hosts connect via websockets
	suite.mustConnectWSNo(roomId, "foo")
	suite.NotNil(suite.App.rooms[roomId].host, "should have a host")
}

func (suite *ServerTestSuite) TestJoinRoomNoName() {
	resp, err := suite.Server.Client().Get(suite.Server.URL + "/room/join")
	suite.NoError(err)
	suite.NotNil(resp)
	suite.Equal(http.StatusBadRequest, resp.StatusCode)
}

func (suite *ServerTestSuite) TestJoinNonExistentRoom() {
	resp, err := suite.Server.Client().Get(suite.Server.URL + "/room/join?name=foo&roomId=123")
	suite.NoError(err)
	suite.NotNil(resp)
	suite.Equal(http.StatusNotFound, resp.StatusCode)
}

func (suite *ServerTestSuite) TestJoinRoomOk() {
	// create the room
	resp, _ := suite.Server.Client().Post(suite.Server.URL+"/create?name=foo", "application/json", nil)
	loc, _ := resp.Request.Response.Location()
	roomId := loc.Query().Get("roomId")
	suite.mustConnectWSNo(roomId, "foo")

	// join via websockets
	suite.mustConnectWSNo(roomId, "bar")
	suite.Eventually(func() bool { 
		return suite.Len(suite.App.rooms[roomId].members, 1, "should have joined the room")
	}, 1 * time.Second, 100 * time.Millisecond)
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
