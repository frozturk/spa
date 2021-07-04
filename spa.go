package spa

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os/exec"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/koding/websocketproxy"
)

type spa struct {
	config             Config
	isDevServerStarted bool
	url                *url.URL
}

func (s *spa) startDevServer() {
	cmd := exec.Command("npm", "run", s.config.NPMScript)
	cmd.Dir = s.config.SPADirectory
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	cmd.Start()

	r := regexp.MustCompile(`open your browser on (?P<url>.*\/)`)
	outScanner := bufio.NewScanner(stdout)
	outScanner.Split(bufio.ScanLines)
	go func() {
		for outScanner.Scan() {
			m := outScanner.Text()
			fmt.Println(m)
			if !s.isDevServerStarted {
				match := r.FindStringSubmatch(m)
				if len(match) > 0 {
					s.devServerStarted(match[1])
				}
			}
		}
	}()

	errScanner := bufio.NewScanner(stderr)
	errScanner.Split(bufio.ScanLines)
	go func() {
		for errScanner.Scan() {
			m := errScanner.Text()
			fmt.Println(m)
		}
	}()

	cmd.Wait()
}

func (s *spa) devServerStarted(urlString string) {
	url, _ := url.Parse(urlString)
	s.url = url
	s.isDevServerStarted = true
}

func (s *spa) proxyWebSocket(c *gin.Context) {
	urlstr := fmt.Sprintf("ws://%s%s", s.url.Host, c.Request.URL.Path)
	u, _ := url.Parse(urlstr)
	ws := websocketproxy.NewProxy(u)
	ws.ServeHTTP(c.Writer, c.Request)
}

func (s *spa) proxyHttp(c *gin.Context) {
	urlstr := fmt.Sprintf("%s%s", s.url.String(), c.Request.URL.Path)
	request, _ := http.NewRequest("GET", urlstr, nil)
	request.Header.Add("Accept", "text/html, application/xhtml+xml, application/xml;q=0.9, image/webp, */*;q=0.8")
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	for name, values := range resp.Header {
		c.Header(name, values[0])
	}
	c.Data(200, resp.Header.Get("Content-Type"), body)
}

func (s *spa) serveSPA(c *gin.Context) {
	if s.isDevServerStarted {
		if strings.Contains(c.Request.URL.Path, "sockjs-node") && !strings.Contains(c.Request.URL.Path, "info") {
			s.proxyWebSocket(c)
		} else {
			s.proxyHttp(c)
		}
	}
	c.Next()
}

func newSpa(config *Config) *spa {
	return &spa{config: *config}
}

func UseSpaStaticFiles(config Config) gin.HandlerFunc {
	fileserver := http.FileServer(gin.Dir(config.SPADirectory, false))
	return func(c *gin.Context) {
		if strings.Contains(c.Request.Header.Get("Accept"), "text/html") {
			c.File(fmt.Sprintf("%s/%s", config.SPADirectory, "index.html"))
		} else {
			fileserver.ServeHTTP(c.Writer, c.Request)
			c.Abort()
		}
	}
}

func UseAngularCliServer(config Config) gin.HandlerFunc {
	spaConfig, err := newConfig(&config)
	if err != nil {
		panic(err.Error())
	}
	spa := newSpa(spaConfig)
	go spa.startDevServer()
	return func(c *gin.Context) {
		spa.serveSPA(c)
	}
}
