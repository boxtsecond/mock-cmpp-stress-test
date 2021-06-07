package cache

import "errors"

type Cache struct {
	Size uint32
	Data map[string]string
}

const MaxSize = 1e6

var RetryQueue = make(chan []string, 1e4)

func (c *Cache) New(size uint32) *Cache {
	c.Data = make(map[string]string, size)
	c.Size = 0
	return c
}

func (c *Cache) Get(key string) string {
	value, ok := c.Data[key]
	if !ok {
		return ""
	}
	return value
}

func (c *Cache) Set(key, value string) error {
	if c.Size >= MaxSize {
		return errors.New("Cache capacity is zero! ")
	}
	c.Size += 1
	c.Data[key] = value
	return nil
}

func (c *Cache) Delete(key string) {
	c.Size -= 1
	delete(c.Data, key)
}

func (c *Cache) Clear() {
	c.Data = nil
}

func (c *Cache) RetrySet(key, value string) {
	RetryQueue <- []string{key, value}
}

func (c *Cache) StartRetry() {
	for {
		select {
		case retry := <-RetryQueue:
			c.Set(retry[0], retry[1])
		default:
		}
	}
}
