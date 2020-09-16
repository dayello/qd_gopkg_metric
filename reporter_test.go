package metric

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"testing"
	"time"
)

func TestReadMemStats(t *testing.T) {

	go func() {
		ticket := time.NewTicker(time.Second * 3)

		for {
			<-ticket.C
			mem := runtime.MemStats{}
			runtime.ReadMemStats(&mem)
			fmt.Printf("all: %d byte, inuse: %d\n", mem.HeapAlloc, mem.HeapInuse)
			fmt.Printf("all object: %d\n", mem.HeapObjects)
			fmt.Printf("GC: %d\n", mem.NumGC)
		}
	}()

	g := make([][]byte, 0)
	// 模拟内存分配
	ticker := time.NewTicker(time.Millisecond * 100)
	for {
		<-ticker.C
		round := rand.Intn(1 << 20)
		byts := make([]byte, round)
		rand.Read(byts)
		if string(byts) == "---" {
			return
		}
		if round%2 == 0 {
			g = append(g, byts)
		}
	}
}

func TestRunning(t *testing.T) {
	fmt.Println("goroutine: ", runtime.NumGoroutine())
}

func TestHttp(t *testing.T) {
	Init(AppVer("test", "test"), EnableProcess(), EnableRuntime())
	var h = Handler()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		AddReqCount(request)
		CollectReqCostTime(request, rand.Int63n(3000))
		h.ServeHTTP(writer, request)
	})

	g := make([][]byte, 0)
	go func() {
		// 模拟内存分配
		ticker := time.NewTicker(time.Millisecond * 100)
		for {
			<-ticker.C
			round := rand.Intn(1 << 20)
			byts := make([]byte, round)
			rand.Read(byts)
			if string(byts) == "---" {
				return
			}
			if round%2 == 0 {
				g = append(g, byts)
			}
		}
	}()

	server := http.Server{
		Handler: mux,
		Addr:    "0.0.0.0:8888",
	}
	go func() {
		log.Fatal(server.ListenAndServe())
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	server.Close()
}
