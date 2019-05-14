package api

import (
	"db2rest/conf"
	"db2rest/db"
	"os"
	"fmt"
	"log"
	"context"
	"time"
	"syscall"
	"os/signal"
	"net/http"
	"github.com/gorilla/mux"
)

type Server struct {
	conf 	*conf.Conf
	db		*db.Client
	server 	*http.Server
}

func NewServer(conf *conf.Conf) *Server {
	return &Server{conf: conf}
}

func (svr *Server) Start() error {
	db, err := db.New(svr.conf)
	if err != nil {
		return err
	}
	svr.db = db

	var router = mux.NewRouter()
	it, err := svr.conf.Iterator("api")
	if err != nil {
		return err
	}
	for it.HasNext() {
		api, err := it.Next()
		if err != nil {
			return err
		}
		e, err := NewEndpoint(api, db)
		if err != nil {
			return err
		}
		log.Printf("deployed: %s %s\n", e.method, e.url)
		router.HandleFunc(e.url, e.Handle).Methods(e.method)
	}

	addr := fmt.Sprintf("%s:%d", svr.conf.GetString("host", ""), svr.conf.GetInt("port", 3424))
	svr.server = &http.Server{
        Addr:         	addr,
        // WriteTimeout: 	time.Second * 15,
        // ReadTimeout:  	time.Second * 15,
        // IdleTimeout:  	time.Second * 15,
        Handler: 		router,
    }

    go func() {
        if err := svr.server.ListenAndServe(); err != nil {
			log.Println(err)
        }
    }()

	log.Printf("server listening at %s\n", addr)
	return nil
}

func (svr *Server) Run(signals ...os.Signal) error {
	if err  := svr.Start(); err != nil {
		return err
	}

	if len(signals) == 0 {
		return nil
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	sig := <- c
	log.Printf("server shutting down by signal: %v\n", sig)

	wait := time.Second * time.Duration(svr.conf.GetInt("graceful_timeout_seconds", 10))
    ctx, cancel := context.WithTimeout(context.Background(), wait)
    defer cancel()

	svr.Shutdown(ctx)
	return nil
}

func (svr *Server) Shutdown(ctx context.Context) {
	svr.server.Shutdown(ctx)
	svr.db.Close()
}

