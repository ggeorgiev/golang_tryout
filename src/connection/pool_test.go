package connection_test

import (
	"connection"
	"testing"
	"time"
)

type DummyConnection struct {
	id int
}

func (connection DummyConnection) Unusable() bool {
	return connection.id == 10
}

func OpenDummyConnection() (interface{}, error) {
	return &DummyConnection{}, nil
}

func TestDynamicPool_ReuseConnection(t *testing.T) {
	pool := &connection.DynamicPool{}
	pool.InitPool(10, time.Second*10, OpenDummyConnection)

	func() {
		connectionInterface, _ := pool.GetConnection()
		connection := connectionInterface.(*DummyConnection)
		defer pool.ReleaseConnection(connection)
		connection.id = 1
	}()

	// check that the modified dummy connection will be reused
	for {
		currentInterface, err := pool.GetConnection()
		current := currentInterface.(*DummyConnection)
		if err == nil && current.id == 1 {
			break
		}
	}
}

func TestDynamicPool_DoNotMakeNewConnections(t *testing.T) {
	pool := &connection.DynamicPool{}
	pool.InitPool(10, time.Millisecond*10, OpenDummyConnection)

	// check with arbitrary number of attempts if we create new connection
	for i := 0; i < 10; i++ {
		func() {
			connectionInterface, _ := pool.GetConnection()
			connection := connectionInterface.(*DummyConnection)
			defer pool.ReleaseConnection(connection)
			count := pool.CirculationConnectionCount()

			// We expect to get to 3 connections, because when we query for
			// 1 we start preparing second, and the creation of connection is
			// async, it might not be reflected when the next yet.
			if count > 3 {
				t.Error("the pool created too many connections:", count)
			}
		}()
	}
}

func TestDynamicPool_UseAllConnections(t *testing.T) {
	pool := &connection.DynamicPool{}
	pool.InitPool(10, time.Millisecond*10, OpenDummyConnection)

	var connections [10]*DummyConnection

	for i := 0; i < 10; i++ {
		connectionInterface, err := pool.GetConnection()
		if err != nil {
			t.Error("pool get connection unexpectedly returned error")
		}
		connections[i] = connectionInterface.(*DummyConnection)
	}

	// this should timeout
	_, err := pool.GetConnection()
	if err == nil {
		t.Error("the pool should return error.")
	}
}

func TestDynamicPool_Unusable(t *testing.T) {
	pool := &connection.DynamicPool{}
	pool.InitPool(10, time.Millisecond*10, OpenDummyConnection)

	func() {
		connectionInterface, _ := pool.GetConnection()
		connection := connectionInterface.(*DummyConnection)
		defer pool.ReleaseConnection(connection)
		connection.id = 10
	}()

	// check if we get back the unusable connection
	for {
		currentInterface, err := pool.GetConnection()
		if err != nil {
			break // we got all connections, no unusable one
		}
		current := currentInterface.(*DummyConnection)
		if current.Unusable() {
			t.Error("we reseaved unusable connection.")
		}
	}
}
