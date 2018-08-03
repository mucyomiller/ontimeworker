package enqueue

import (
	"github.com/gocraft/work"
	"github.com/gomodule/redigo/redis"
	"github.com/mucyomiller/ontimeworker/common"
)

var redisPort = common.Getenv("REDIS_PORT", ":6379")

// Make a redis pool.
var redisPool = &redis.Pool{
	MaxActive: 5,
	MaxIdle:   5,
	Wait:      true,
	Dial: func() (redis.Conn, error) {
		return redis.Dial("tcp", redisPort)
	},
}

// Make an enqueuer with a particular namespace.
var enqueuer = work.NewEnqueuer("ontimepaymentservice", redisPool)

// Enqueue method used to created job in worker pool.
func Enqueue(tx map[string]interface{}) (bool, error) {
	// Enqueue a job named "validate_transaction" with the specified parameters.
	_, err := enqueuer.Enqueue("validate_transaction", tx)
	if err != nil {
		return false, err
	}
	return true, nil
}
