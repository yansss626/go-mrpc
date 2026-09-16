package mrpc

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/silenceper/pool"
)

const (
	defaultInitialCap  = 0
	defaultMaxIdle     = 5
	defaultMaxCap      = 20
	defaultIdleTimeout = 10 * time.Minute
)

type ClientPoolConfig struct {
	Host        string
	Port        int
	InitialCap  int
	MaxIdle     int
	MaxCap      int
	IdleTimeout time.Duration
}

type ClientPool struct {
	pool pool.Pool
}

func NewClientPool(cnf *ClientPoolConfig) (*ClientPool, error) {

	config := *cnf
	if cnf.IdleTimeout == 0 {
		config.IdleTimeout = defaultIdleTimeout
	}
	if cnf.InitialCap == 0 {
		config.InitialCap = defaultInitialCap
	}
	if cnf.MaxCap == 0 {
		config.MaxCap = defaultMaxCap
	}
	if cnf.MaxIdle == 0 {
		config.MaxIdle = defaultMaxIdle
	}

	pool, err := newPool(&config)
	if err != nil {
		return nil, err
	}

	return &ClientPool{
		pool: pool,
	}, nil
}

func (c *ClientPool) Get() (*Client, error) {
	for {
		obj, err := c.pool.Get()
		if err != nil {
			return nil, err
		}
		client, ok := obj.(*Client)
		if !ok {
			return nil, fmt.Errorf("invalid connection type")
		}

		if client.closed.Load() {
			c.Discard(client)
			continue
		}
		return client, nil
	}

}

// 放回一个连接
func (c *ClientPool) Put(client *Client) error {
	return c.pool.Put(client)
}

// 关闭一个连接
func (c *ClientPool) Discard(client *Client) error {
	return c.pool.Close(client)
}

// 释放连接池
func (c *ClientPool) Release() {
	c.pool.Release()
}

func newPool(cnf *ClientPoolConfig) (pool.Pool, error) {

	factory := func() (interface{}, error) {
		addr := net.JoinHostPort(cnf.Host, strconv.Itoa(cnf.Port))
		client, err := Dial(addr)
		if err != nil {
			return nil, err
		}
		return client, nil
	}

	close := func(v interface{}) error {

		client, ok := v.(*Client)
		if !ok {
			return fmt.Errorf("invalid connection type")
		}
		return client.Close()
	}

	poolConfig := &pool.Config{
		InitialCap:  cnf.InitialCap,
		MaxIdle:     cnf.MaxIdle,
		MaxCap:      cnf.MaxCap,
		Factory:     factory,
		Close:       close,
		IdleTimeout: cnf.IdleTimeout,
	}
	pool, err := pool.NewChannelPool(poolConfig)
	if err != nil {
		return nil, err
	}

	return pool, nil
}
