package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"text/tabwriter"
	"time"

	"git.containerum.net/ch/solutions/pkg/db"
	"git.containerum.net/ch/solutions/pkg/router"
	"git.containerum.net/ch/solutions/pkg/server"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"git.containerum.net/ch/solutions/pkg/clients"
	"github.com/urfave/cli"
)

func getService(service interface{}, err error) interface{} {
	exitOnErr(err)
	return service
}

func initServer(c *cli.Context) error {
	if c.Bool(debugFlag) {
		gin.SetMode(gin.DebugMode)
		log.SetLevel(log.DebugLevel)
	} else {
		gin.SetMode(gin.ReleaseMode)
		log.SetLevel(log.InfoLevel)
	}

	if c.Bool(textlogFlag) {
		log.SetFormatter(&log.TextFormatter{})
	} else {
		log.SetFormatter(&log.JSONFormatter{})
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.TabIndent|tabwriter.Debug)
	for _, f := range c.GlobalFlagNames() {
		fmt.Fprintf(w, "Flag: %s\t Value: %s\n", f, c.String(f))
	}
	w.Flush()

	solutionssrv, err := getSolutionsSrv(c, server.Services{
		DB:              getService(getDB(c)).(db.DB),
		DownloadClient:  clients.NewHTTPDownloadClient(c.String(csvURLFlag)),
		ResourceClient:  clients.NewHTTPResourceClient(c.String(resourceURLFlag)),
		KubeAPIClient:   clients.NewHTTPKubeAPIClient(c.String(kubeURLFlag)),
		ConverterClient: clients.NewHTTPConverterClient(c.String(converterURLFlag)),
	})
	exitOnErr(err)

	app := router.CreateRouter(&solutionssrv, c.Bool(corsFlag))

	// for graceful shutdown
	srv := &http.Server{
		Addr:    ":" + c.String(portFlag),
		Handler: app,
	}

	go exitOnErr(srv.ListenAndServe())

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt) // subscribe on interrupt event
	<-quit                            // wait for event
	log.Infoln("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return srv.Shutdown(ctx)
}

func exitOnErr(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
