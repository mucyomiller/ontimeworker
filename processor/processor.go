package processor

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gocraft/work"
	"github.com/gomodule/redigo/redis"
	"github.com/mucyomiller/ontimeworker/common"
	log "github.com/sirupsen/logrus"
)

// Havanao response status codes
// See: https://havanao.com/docs/1.0/sales-request#purchase-status
const (
	StatusApproved  = "APPROVED"
	StatusDeclined  = "DECLINED"
	StatusRequested = "REQUESTED"
	StatusUnknown   = "UNKNOWN"
)

// Initialise Env variables
var redisPort = common.Getenv("REDIS_PORT", ":6379")
var paymentURL = common.Getenv("PAYMENT_URL", " ")
var momoToken = common.Getenv("HAVANAO_KEY", " ")
var webhookURL = common.Getenv("WEBHOOK", " ")
var ourServiceKey = common.Getenv("OUR_SERVICE_KEY", " ")

// http client configs
var netTransport = &http.Transport{
	Dial: (&net.Dialer{
		Timeout: 10 * time.Second,
	}).Dial,
	TLSHandshakeTimeout: 10 * time.Second,
}

var httpClient = &http.Client{
	Timeout:   time.Second * 30,
	Transport: netTransport,
}

// Make a redis pool
var redisPool = &redis.Pool{
	MaxActive: 5,
	MaxIdle:   5,
	Wait:      true,
	Dial: func() (redis.Conn, error) {
		return redis.Dial("tcp", redisPort)
	},
}

// Context custom context
type Context struct{}

// Havanao API response struct
type Havanao struct {
	Code              int64  `json:"code"`
	Status            string `json:"status"`
	Transactionid     string `json:"transactionid"`
	TransactionStatus string `json:"transactionStatus"`
	Description       string `json:"description"`
}

// StartJobProcessor process submit tasks
func StartJobProcessor() {
	// Make a new pool. Arguments:
	// Context{} is a struct that will be the context for the request.
	// 10 is the max concurrency
	// "ontimepaymentservice" is the Redis namespace
	// redisPool is a Redis pool
	pool := work.NewWorkerPool(Context{}, 10, "ontimepaymentservice", redisPool)

	// Add middleware that will be executed for each job
	pool.Middleware((*Context).Log)

	// Map the name of jobs to handler functions
	pool.Job("validate_transaction", (*Context).CheckTransaction)

	// Start processing jobs
	pool.Start()
	log.Info("Starting Job Processor to process submitted tasks")
	// Wait for a signal to quit:
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, os.Kill)
	<-signalChan

	// Stop the pool
	pool.Stop()
}

// Log middleware to log each started job
func (c *Context) Log(job *work.Job, next work.NextMiddlewareFunc) error {
	log.Infof("Starting job: ", job.Name)
	return next()
}

// CheckTransaction pull work in queue and checks it against
// momo payment processor server
func (c *Context) CheckTransaction(job *work.Job) error {
	// Extract arguments:
	tx := job.ArgString("transactionId")
	if err := job.ArgError(); err != nil {
		return err
	}

	// checking transaction id against momo processor[havanao]
	resp, err := httpClient.Get(paymentURL + "?transactionId=" + tx + "&api_token=" + momoToken)
	if err != nil {
		return err
	}

	havanao := &Havanao{}
	defer resp.Body.Close()
	// Decode resp body directly into Havanao struct
	json.NewDecoder(resp.Body).Decode(havanao)

	// b, err := json.Marshal(havanao)
	// if err != nil {
	// 	return err
	// }
	// log.Infof("completed transaction Id %v : result %v:", tx, string(b))
	if havanao.TransactionStatus == StatusApproved {
		// WebHook and notify transaction successful completed
		payload := []byte(fmt.Sprintf(`{"transaction_id":%q,"status":"success"}`, tx))
		req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(payload))
		if err != nil {
			// return error so that worker will retry again
			log.Error("malformed payload or webhook url")
			return err
		}
		req.Header.Set("X-Custom-Header", ourServiceKey)
		req.Header.Set("Content-Type", "application/json")
		rsp, err := httpClient.Do(req)
		if err != nil {
			// return error so that worker will retry again
			log.Error("submitting worker result to ontimservice webhook failed")
			return err
		}
		defer rsp.Body.Close()
		if rsp.StatusCode == http.StatusOK {
			return nil
		}
		return errors.New("error occured while trying to submit result to webhook")

	} else if havanao.TransactionStatus == StatusDeclined {
		log.Warnf("user canceled transaction with ID %v", tx)
		return nil
	} else if havanao.TransactionStatus == StatusUnknown {
		log.Errorf("unknown error halt job & skip to next")
		return nil
	} else {
		return errors.New("User doesn't yet confirmed transaction")
	}
}
