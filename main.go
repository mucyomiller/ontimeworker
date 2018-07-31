package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/facebookgo/grace/gracehttp"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	_ "github.com/joho/godotenv/autoload"
	"github.com/mucyomiller/ontimeworker/common"
	"github.com/mucyomiller/ontimeworker/enqueue"
	"github.com/mucyomiller/ontimeworker/processor"
	log "github.com/sirupsen/logrus"
	"github.com/unrolled/render"
)

// Q job Queue
type Q map[string]interface{}

// Initialize render
var rs = render.New()

func main() {

	// configure router to handler incomming job request
	r := chi.NewRouter()
	r.Use(middleware.DefaultCompress)
	r.Use(middleware.StripSlashes)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// routes
	r.Post("/jobs/create", handleJob)

	// getting server port
	serverPort := common.Getenv("SERVER_PORT", ":8080")
	// spinning up jobProcessor
	go processor.StartJobProcessor()
	// starting http server with graceful restart
	log.Info("Starting web server to accept submitted job")
	err := gracehttp.Serve(&http.Server{Addr: serverPort, Handler: r})
	if err != nil {
		log.Errorf("Server Error %v", err)
		return
	}

}

func handleJob(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := middleware.GetReqID(ctx)
	log.Infof("we got a request with ID: %v", reqID)

	// retreiving data from http payload
	transactionID := strings.TrimSpace(r.FormValue("transactionId"))

	// forming job map[string]interface{}
	tx := Q{
		"transactionId": transactionID,
	}

	// Enqueue job tx to worker
	ok, err := enqueue.Enqueue(tx)
	if err != nil {
		log.Errorf("Error occured while enqueuing job %v ", err)
		rs.JSON(w, http.StatusOK, map[string]string{"error": "job not queued"})
	}
	if ok == true {
		rs.JSON(w, http.StatusOK, map[string]string{"success": "job successful queued"})
	}
}
