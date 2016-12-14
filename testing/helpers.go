package testing

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"

	mgo "gopkg.in/mgo.v2"
	redsync "gopkg.in/redsync.v1"

	"github.com/garyburd/redigo/redis"
	"github.com/labstack/echo/engine/standard"
	"github.com/onsi/gomega"
	"github.com/topfreegames/donations/api"
	"github.com/uber-go/zap"
)

//GetTestRedis returns a configured redsync connection
func GetTestRedis() redis.Conn {
	conn, err := redis.DialURL("redis://localhost:9998/1")
	if err != nil {
		panic("Failed to connect to redis")
	}
	return conn
}

//GetTestRedsync returns a configured redsync connection
func GetTestRedsync() *redsync.Redsync {
	return redsync.New([]redsync.Pool{
		&redis.Pool{
			MaxIdle:     3,
			IdleTimeout: 240 * time.Second,
			Dial: func() (redis.Conn, error) {
				conn, err := redis.DialURL("redis://localhost:9998/1")
				if err != nil {
					return nil, err
				}
				return conn, nil
			},
			TestOnBorrow: func(c redis.Conn, t time.Time) error {
				_, err := c.Do("PING")
				if err != nil {
					return err
				}
				return nil
			},
		},
	})
}

//GetTestMutex returns a mutex for the name specified
func GetTestMutex(name string, rs *redsync.Redsync) *redsync.Mutex {
	return rs.NewMutex(name)
}

//GetTestMongoDB returns a test connection to mongo db
func GetTestMongoDB() (*mgo.Session, *mgo.Database) {
	return getMongoDB("localhost", "9999", "donations-test")
}

//GetPerfMongoDB returns a test connection to PERF mongo db
func GetPerfMongoDB() (*mgo.Session, *mgo.Database) {
	return getMongoDB("localhost", "9999", "donations-perf")
}
func getMongoDB(host, port, db string) (*mgo.Session, *mgo.Database) {
	hostEnv := os.Getenv("DONATIONS_MONGO_HOST")
	if hostEnv != "" {
		host = hostEnv
	}
	portEnv := os.Getenv("DONATIONS_MONGO_PORT")
	if portEnv != "" {
		port = portEnv
	}

	mongoURL := fmt.Sprintf("mongodb://%s:%s", host, port)
	session, err := mgo.Dial(mongoURL)
	if err != nil {
		panic(fmt.Sprintf("Could not connect to mongodb: %s.", err.Error()))
	}
	session.SetMode(mgo.Monotonic, true)
	database := session.DB(db)
	return session, database
}

//GetConfPath returns teh configuration path
func GetConfPath() string {
	conf := "../config/test.yaml"
	return conf
}

//GetDefaultTestApp returns a new podium API Application bound to 0.0.0.0:8890 for test
func GetDefaultTestApp(logger zap.Logger) *api.App {
	app, err := api.GetApp("0.0.0.0", 8890, GetConfPath(), false, logger, false, false)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	app.Configure()
	return app
}

//Get from server
func Get(app *api.App, url string) (int, string) {
	return doRequest(app, "GET", url, "")
}

//Post to server
func Post(app *api.App, url, body string) (int, string) {
	return doRequest(app, "POST", url, body)
}

//Put to server
func Put(app *api.App, url, body string) (int, string) {
	return doRequest(app, "PUT", url, body)
}

//Delete from server
func Delete(app *api.App, url, body string) (int, string) {
	return doRequest(app, "DELETE", url, body)
}

var client *http.Client
var transport *http.Transport

func initClient() {
	if client == nil {
		transport = &http.Transport{DisableKeepAlives: true}
		client = &http.Client{Transport: transport}
	}
}

//InitializeTestServer for tests
func InitializeTestServer(app *api.App) *httptest.Server {
	initClient()
	app.Engine.SetHandler(app.App)
	return httptest.NewServer(app.Engine.(*standard.Server))
}

//GetRequest builds a new request for tests
func GetRequest(app *api.App, ts *httptest.Server, method, url string, bodyBuff io.Reader) *http.Request {
	req, err := http.NewRequest(method, fmt.Sprintf("%s%s", ts.URL, url), bodyBuff)
	req.Header.Set("Connection", "close")
	req.Close = true
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	return req
}

//PerformRequest against specified test server
func PerformRequest(ts *httptest.Server, req *http.Request) *http.Response {
	res, err := client.Do(req)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return res
}

//ReadBody from response
func ReadBody(res *http.Response) []byte {
	//Wait for port of httptest to be reclaimed by OS
	b, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return b
}

func doRequest(app *api.App, method, url, body string) (int, string) {
	ts := InitializeTestServer(app)
	defer transport.CloseIdleConnections()
	defer ts.Close()

	var bodyBuff io.Reader
	if body != "" {
		bodyBuff = bytes.NewBuffer([]byte(body))
	}

	req := GetRequest(app, ts, method, url, bodyBuff)
	res := PerformRequest(ts, req)
	bodyRes := ReadBody(res)
	return res.StatusCode, string(bodyRes)
}

//ResetStdout back to os.Stdout
var ResetStdout func()

//ReadStdout value
var ReadStdout func() string

//MockStdout to read it's value later
func MockStdout() {
	stdout := os.Stdout
	r, w, err := os.Pipe()
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	os.Stdout = w

	ReadStdout = func() string {
		var buf bytes.Buffer
		_, err := io.Copy(&buf, r)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		r.Close()
		return buf.String()
	}

	ResetStdout = func() {
		w.Close()
		os.Stdout = stdout
	}
}

//TestBuffer is a mock buffer
type TestBuffer struct {
	bytes.Buffer
}

//Sync does nothing
func (b *TestBuffer) Sync() error {
	return nil
}

//Lines returns all lines of log
func (b *TestBuffer) Lines() []string {
	output := strings.Split(b.String(), "\n")
	return output[:len(output)-1]
}

//Stripped removes new lines
func (b *TestBuffer) Stripped() string {
	return strings.TrimRight(b.String(), "\n")
}
