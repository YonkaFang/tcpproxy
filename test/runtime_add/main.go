package main

import (
	"log"
	"tcpproxy"
	"time"
	"net/http"
	"fmt"
	"net"
)

const (
	baiduIP   = "220.181.57.216"
	neteaseIP = "123.58.180.7"
	sinaIP    = "66.102.251.33"
)

func main() {
	p := tcpproxy.NewProxy()
	var err error
	var port int

	p.AddRoute(":8124", tcpproxy.To(neteaseIP+":80")) // fallback
	go func() {
		log.Fatal(p.Run())
	}()

	time.Sleep(1 * time.Second) // wait for server-startup

	testHost(neteaseIP, 8124)

	var r *tcpproxy.Credential

	log.Printf("AddRuntimeRoute(%d), -> %s:80", 8123, baiduIP)
	if r, err = p.AddRuntimeRoute(":8123", tcpproxy.To(baiduIP+":80")); err != nil {
		log.Printf("AddRuntimeRoute: met err (%v)", err)
	}
	// time.Sleep(1 * time.Second) // wait for server-startup
	log.Printf("after add %s -> %d", baiduIP, 8123)
	testHost(baiduIP, 8123)

	log.Printf("remove(%d), -> %s:80", 8123, baiduIP)
	if err = r.Remove(); err != nil {
		log.Printf("r.Remove: met err (%v)", err)
	}

	log.Printf("after remove %s -> %d", baiduIP, 8123)
	testHost(baiduIP, 8123)

	log.Printf("AddRuntimeRoute(%d), -> %s:80", 0, sinaIP)
	if r, err = p.AddRuntimeRoute("", tcpproxy.To(sinaIP+":80")); err != nil {
		log.Printf("AddRuntimeRoute: met err (%v)", err)
	} else {
		port = r.Addr().(*net.TCPAddr).Port
		log.Printf("AddRuntimeRoute get addr: (%s)", r.Addr().String())
	}

	testHost(sinaIP, port)

	log.Printf("remove(%d), -> %s:80", port, sinaIP)
	if err = r.Remove(); err != nil {
		log.Printf("r.Remove: met err (%v)", err)
	}
	testHost(sinaIP, port)

	p.Close()
}

func testHost(host string, port int) {
	client := &http.Client{}
	if resp, err := client.Do(newReq(fmt.Sprintf("http://127.0.0.1:%d", port), host)); err == nil {
		log.Printf("%d: (%d)", port, resp.StatusCode)
	} else {
		log.Printf("%d: met err (%v)", port, err)
	}
}

func newReq(url, host string) *http.Request {
	req, _ := http.NewRequest("GET", url, nil)
	//req.Header.Set("Host", baiduIP)  // https://github.com/golang/go/issues/7682
	req.Host = host
	return req
}
