package connection

import (
	"errors"
	"sync/atomic"
	"time"
)

type ThrowAwayConnection interface {
	Unusable() bool
}

// Type of function that opens connection
type ConnectFunction func() (interface{}, error)

type DynamicPool struct {
	// Maximum open connections
	max           int32

	// Timeout period for waiting for a connection when the maximum is exhausted
	timeout       time.Duration

	// Time interval between connect retries
	retryInterval time.Duration

	// Function that can open a new connection when needed
	connectFn     ConnectFunction

	// Implementation detail, buffered channel that keeps the available connections
	channel       chan interface{}

	// Implementation detail, the current number of connection in circulation
	size          int32

	// Implementation detail, active waiters
	waiters       int32
}

func (pool *DynamicPool) connect() {

	// Note that we would like to query a new connection even if there is one available
	// to start preparing it for the next query
	if len(pool.channel) > 1 {
		return
	}
	if pool.size >= pool.max {
		return
	}

	atomic.AddInt32(&pool.size, 1)

	for {
		connection, err := pool.connectFn()
		if err == nil {
			pool.channel <- connection
			return
		}
		// if there is no active waiter, abandon the request
		if pool.waiters == 0 {
			break
		}

		time.Sleep(pool.retryInterval)
	}

	atomic.AddInt32(&pool.size, -1)
}

// InitPool initializes the Dynamic pool.
// Makes a channel and triggers creation of the first connection.
func (pool *DynamicPool) InitPool(max int32, timeout time.Duration, retryInterval time.Duration, connectFn ConnectFunction) {

	pool.max = max
	pool.size = 0
	pool.waiters = 0
	pool.connectFn = connectFn
	pool.timeout = timeout
	pool.retryInterval = retryInterval
	pool.channel = make(chan interface{}, max)
	go pool.connect()
	return
}

func (pool *DynamicPool) GetConnection() (interface{}, error) {
	go pool.connect()

	atomic.AddInt32(&pool.waiters, 1)
	defer atomic.AddInt32(&pool.waiters, -1)

	select {
	case connection := <-pool.channel:
		return connection, nil
	case <-time.After(pool.timeout):
		return nil, errors.New("timeout waiting for connection")
	}
}

func (pool *DynamicPool) ReleaseConnection(connection interface{}) {
	throwAway := connection.(ThrowAwayConnection)
	if throwAway.Unusable() {
		atomic.AddInt32(&pool.size, -1)
		return
	}
	pool.channel <- connection
}

func (pool *DynamicPool) CirculationConnectionCount() int32 {
	return pool.size
}
