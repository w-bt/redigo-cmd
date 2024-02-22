package command

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/zenthangplus/goccm"
	"log"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

var (
	count           = 500
	total     int64 = 0
	startTime time.Time
	duration  time.Duration
)

func Migrate() {
	startTime = time.Now()

	sourcePool := NewPool(hostSource, dbSource, sourceUsername, sourcePassword)
	destPool := NewPool(hostDest, dbDest, destUsername, destPassword)
	var wg sync.WaitGroup
	chanTotal := make(chan int64)
	chanBool := make(chan bool)

	iter := 0
	c1 := goccm.New(con1)

	go func() {
		for {
			select {
			case t := <-chanTotal:
				atomic.AddInt64(&total, t)
				log.Printf("total executed keys %d\n", total)
			case <-chanBool:
				break
			}
		}
	}()

	for {
		connSource := sourcePool.Get()

		arr, err := redis.Values(connSource.Do("SCAN", iter, "COUNT", count))
		if err != nil {
			log.Fatalln(fmt.Sprintf("error during execute command scan, err %+v", err))
		}

		iter, err = redis.Int(arr[0], nil)
		if err != nil {
			log.Fatalln(fmt.Sprintf("error during converting int, err %+v", err))
		}
		keys, err := redis.Strings(arr[1], nil)
		if err != nil {
			log.Fatalln(fmt.Sprintf("error during converting string array, err %+v", err))
		}

		c1.Wait()
		wg.Add(len(keys))
		go func(iter int, keys []string, sourcePool *redis.Pool, wgPointer *sync.WaitGroup, chanTotal chan int64) {
			c2 := goccm.New(con2)
			connSourceItem := sourcePool.Get()
			defer connSourceItem.Close()
			for _, key := range keys {
				connSourceItem.Send("PTTL", key)
				connSourceItem.Send("DUMP", key)
				connSourceItem.Flush()
				ttl, err := redis.Int(connSourceItem.Receive())
				if err != nil {
					log.Println(fmt.Sprintf("error during receiving ttl, err %+v", err))
				}
				if ttl < 0 {
					ttl = 0
				}
				hashVal, err := redis.String(connSourceItem.Receive())
				if err != nil {
					log.Println(fmt.Sprintf("error during receiving hashVal, err %+v", err))
					err = saveFailedKey(key, "failed_keys.txt")
					if err != nil {
						log.Println(fmt.Sprintf("error during save failed key key %s, err %+v", key, err))
					}
					wgPointer.Done()
					continue
				}
				c2.Wait()
				go func(key string, ttl int, hashVal string, destPool *redis.Pool, wgPointer2 *sync.WaitGroup, chanTotal chan int64) {
					connDest := destPool.Get()
					defer connDest.Close()
					_, err = redis.String(connDest.Do("RESTORE", key, ttl, hashVal, "REPLACE"))
					if err != nil {
						log.Println(fmt.Sprintf("error during execute command restore for key %s, err %+v", key, err))
						err = saveFailedKey(key, "failed_keys.txt")
						if err != nil {
							log.Println(fmt.Sprintf("error during save failed key key %s, err %+v", key, err))
						}
					}
					chanTotal <- 1
					c2.Done()
					wgPointer2.Done()
				}(key, ttl, hashVal, destPool, wgPointer, chanTotal)
			}
			c1.Done()
		}(iter, keys, sourcePool, &wg, chanTotal)

		//log.Println(fmt.Sprintf("next iteration: %d", iter))

		if iter == 0 {
			chanBool <- true
			connSource.Close()
			break
		}

		connSource.Close()
	}

	wg.Wait()
	duration = time.Since(startTime)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	signal.Notify(c, syscall.SIGKILL)
	go func() {
		<-c
		sourcePool.Close()
		destPool.Close()
		os.Exit(0)
	}()

	for {
		log.Printf("finish, total key %d, duration %f minutes, waiting for exit command...", total, duration.Minutes())
		time.Sleep(10 * time.Second)
	}
}

func saveFailedKey(key, fileName string) error {
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write([]byte(fmt.Sprintf("%s\n", key)))
	if err != nil {
		return err
	}

	return nil
}
