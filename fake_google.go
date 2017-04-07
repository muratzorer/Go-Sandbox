package main

import (
	"fmt"
	"math/rand"
	"time"
	_ "expvar"
	"net/http"
	"github.com/stackimpact/stackimpact-go"
	"encoding/json"
)

type Result string
type Search func(query string) Result

var (
	Web1 = fakeSearch("web1")
	Web2 = fakeSearch("web2")
	Image1 = fakeSearch("image1")
	Image2 = fakeSearch("image2")
	Video1 = fakeSearch("video1")
	Video2 = fakeSearch("video2")
)

func Google(query string) (results []Result) {
	c := make(chan Result)
	go func() { c <- First(query, Web1, Web2) } ()
	go func() { c <- First(query, Image1, Image2) } ()
	go func() { c <- First(query, Video1, Video2) } ()
	timeout := time.After(80 * time.Millisecond)
	for i := 0; i < 3; i++ {
		select {
		case result := <-c:
			results = append(results, result)
		case <-timeout:
			fmt.Println("timed out")
			return
		}
	}
	return
}

func First(query string, replicas ...Search) Result {
	c := make(chan Result)
	searchReplica := func(i int) {
		c <- replicas[i](query)
	}
	for i := range replicas {
		go searchReplica(i)
	}
	return <-c
}

func fakeSearch(kind string) Search {
	return func(query string) Result {
		time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
		return Result(fmt.Sprintf("%s result for %q\n", kind, query))
	}
}

func fakeGoogleHandler(w http.ResponseWriter, r *http.Request) {
	rand.Seed(time.Now().UnixNano())
	//start := time.Now()
	results := Google("golang")
	//elapsed := time.Since(start)
	//fmt.Println(results)
	//fmt.Println(elapsed)

	b, _ := json.Marshal(results)
	p := &Page{Title: "Google Search Results", Body: b}

	fmt.Fprintf(w, "<h1>%s</h1><div>%s</div>", p.Title, p.Body)
}

type Page struct {
	Title string
	Body  []byte
}

func main() {
	// for StackImpact Monitor
	agent := stackimpact.NewAgent()
	agent.Start(stackimpact.Options{
		AgentKey: "9593b1747aa6ef5466e2340b08a5bcbe820292ab",
		AppName: "MyGoApp",
	})

	http.HandleFunc("/google/", fakeGoogleHandler)
	http.ListenAndServe(":8080", nil) //expvarmon will poll port 8080

	rand.Seed(time.Now().UnixNano())
	start := time.Now()
	results := Google("golang")
	elapsed := time.Since(start)
	fmt.Println(results)
	fmt.Println(elapsed)
}