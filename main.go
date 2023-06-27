package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

// Server is an interface that will be implemented by the simpleServer struct
// it will help us to create a slice of servers that we want to load balance
// and then we will call the Serve function on the server that we want to serve the request to
type Server interface {
	Address() string

	IsAlive() bool

	Serve(rw http.ResponseWriter, req *http.Request)
}

// simpleServer is a struct that will implement the Server interface
type simpleServer struct {
	addr  string
	proxy *httputil.ReverseProxy
}

func (s *simpleServer) Address() string { return s.addr }

func (s *simpleServer) IsAlive() bool { return true }

func (s *simpleServer) Serve(rw http.ResponseWriter, req *http.Request) {
	// it will start the server on the address that we have passed in the newSimpleServer function
	// s.proxy is getting the proxy that we have created in the newSimpleServer function
	// then we are calling the ServeHTTP function on the proxy that we have created in the newSimpleServer function
	// ServeHTTP  will serve the request to the server that we have passed in the newSimpleServer function
	s.proxy.ServeHTTP(rw, req)
}

func newSimpleServer(addr string) *simpleServer {
	serverUrl, err := url.Parse(addr)
	handleErr(err)

	return &simpleServer{
		addr:  addr,
		proxy: httputil.NewSingleHostReverseProxy(serverUrl),
	}
}

type LoadBalancer struct {
	port            string
	roundRobinCount int
	servers         []Server
}

func NewLoadBalancer(port string, servers []Server) *LoadBalancer {
	return &LoadBalancer{
		port:            port,
		roundRobinCount: 0,
		servers:         servers,
	}
}

func handleErr(err error) {
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
}

func (lb *LoadBalancer) getNextAvailableServer() Server {
	// this for loop will loop through the servers slice and will return the server that is alive
	// if the server is not alive then it will increment the roundRobinCount and will return the next server
	server := lb.servers[lb.roundRobinCount%len(lb.servers)]
	for !server.IsAlive() {
		lb.roundRobinCount++
		server = lb.servers[lb.roundRobinCount%len(lb.servers)]
	}
	// increment the roundRobinCount so that the next time we call the getNextAvailableServer function
	lb.roundRobinCount++

	// return the server that is alive
	return server
}

// this function is called when a request is made to the load balancer
func (lb *LoadBalancer) serveProxy(rw http.ResponseWriter, req *http.Request) {
	// get the next available server
	targetServer := lb.getNextAvailableServer()

	fmt.Printf("forwarding request to address %q\n", targetServer.Address())

	// serve the request to the target server how ? targetserver saves the returned server from the getNextAvailableServer function
	// and then we are calling the Serve function on the targetServer which is the server that we want to serve the request to
	targetServer.Serve(rw, req)
}

func main() {

	// this servers var will hold the servers that we want to load balance
	servers := []Server{

		// newSimpleServer is a helper function that returns a simpleServer struct that is saved in the servers slice
		newSimpleServer("https://www.facebook.com"),
		newSimpleServer("https://www.bing.com"),
		newSimpleServer("https://www.duckduckgo.com"),
		newSimpleServer("https://www.google.com"),
		newSimpleServer("https://www.yahoo.com"),
	}

	// lb will cointain the LoadBalancer struct as we are defined it to return
	lb := NewLoadBalancer("8000", servers)

	// handelRedirect is a function that will be called when a request is made to the load balancer
	// it is a blank function that will always call .

	// then it is calling the function lb.serveProxy , as we have defined lb which contains the LoadBalancer struct
	// then we are calling like call lb.serveProxy , how its working is lb is the receiver of the serveProxy function
	// so lb is the receiver and serveProxy is the function that is being called on the receiver lb
	// its like we are saving the function serveProxy in the LoadBalancer struct and then calling it when a request is made

	handleRedirect := func(rw http.ResponseWriter, req *http.Request) {
		lb.serveProxy(rw, req)
	}

	http.HandleFunc("/", handleRedirect)

	fmt.Printf("serving requests at 'localhost:%s'\n", lb.port)
	http.ListenAndServe(":"+lb.port, nil)
}
