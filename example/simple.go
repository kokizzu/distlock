// code is from https://gist.github.com/fiorix/816117cfc7573319b72d#file-groupcache-go-L42
// Simple groupsync example: https://github.com/golang/groupcache
// Running 3 instances:
// go run simple.go -addr=:8080 -pool=http://127.0.0.1:8080,http://127.0.0.1:8081,http://127.0.0.1:8082
// go run simple.go -addr=:8081 -pool=http://127.0.0.1:8081,http://127.0.0.1:8080,http://127.0.0.1:8082
// go run simple.go -addr=:8082 -pool=http://127.0.0.1:8082,http://127.0.0.1:8080,http://127.0.0.1:8081
// Testing:
// curl localhost:8080/color?name=red
package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"strings"

	"github.com/kokizzu/groupsync"
)

var Store = map[string][]byte{
	"red":   []byte("#FF0000"),
	"green": []byte("#00FF00"),
	"blue":  []byte("#0000FF"),
}

var Group = groupsync.NewGroup("foobar", 64<<20, groupsync.GetterFunc(
	func(ctx groupsync.Context, key string, dest groupsync.Sink) error {
		log.Println("looking up", key)
		v, ok := Store[key]
		if !ok {
			return errors.New("color not found")
		}
		return dest.SetBytes(v)
	},
))

func main() {
	addr := flag.String("addr", ":8080", "server address")
	peers := flag.String("pool", "http://localhost:8080", "server pool list")
	flag.Parse()
	http.HandleFunc("/color", func(w http.ResponseWriter, r *http.Request) {
		color := r.FormValue("name")
		log.Println("/color uri hit", color)
		var b []byte
		err := Group.Get(context.Background(), color, groupsync.AllocatingByteSliceSink(&b))
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		_, _ = w.Write(b)
		_, _ = w.Write([]byte{'\n'})
	})
	p := strings.Split(*peers, ",")
	pool := groupsync.NewHTTPPool(p[0])
	pool.Set(p...)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
