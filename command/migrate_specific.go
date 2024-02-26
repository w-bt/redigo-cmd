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

func MigrateSpecific() {
	sourcePool := NewPool(hostSource, dbSource, sourceUsername, sourcePassword)
	destPool := NewPool(hostDest, dbDest, destUsername, destPassword)
	var wg sync.WaitGroup
	var wgP sync.WaitGroup
	chanTotal := make(chan int64)
	chanBool := make(chan bool)

	prefix := prefixKeys

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

	t1 := time.Now()
	for _, keyP := range prefix {
		wgP.Add(1)
		iter := 0
		go func(wgP *sync.WaitGroup, keyP string, chanTotal chan int64, iter int) {
			log.Println("Starting to execute", keyP)
			defer wgP.Done()
			for {
				connSource := sourcePool.Get()
				arr, err := redis.Values(connSource.Do("SCAN", iter, "MATCH", keyP, "COUNT", count))
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
				go func(iter int, keys []string, wgPointer *sync.WaitGroup, chanTotal chan int64) {
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
						hashVal, err := connSourceItem.Receive()
						if err != nil {
							log.Println(fmt.Sprintf("error during receiving hashVal, err %+v", err))
							wgPointer.Done()
							continue
						}
						c2.Wait()
						go func(key string, ttl int, hashVal interface{}, wgPointer2 *sync.WaitGroup, chanTotal chan int64) {
							connDest := destPool.Get()
							_, err = redis.String(connDest.Do("RESTORE", key, ttl, hashVal, "REPLACE"))
							if err != nil {
								log.Println(fmt.Sprintf("error during execute command restore for key %s, err %+v", key, err))
							}
							chanTotal <- 1
							c2.Done()
							wgPointer2.Done()
							connDest.Close()
						}(key, ttl, hashVal, wgPointer, chanTotal)
					}
					c1.Done()
				}(iter, keys, &wg, chanTotal)

				//log.Println(fmt.Sprintf("next iteration: %d", iter))

				if iter == 0 {
					connSource.Close()
					break
				}

				connSource.Close()
			}
		}(&wgP, keyP, chanTotal, iter)
	}

	wg.Wait()
	wgP.Wait()
	chanBool <- true
	log.Println("finish")
	diff := time.Since(t1)
	log.Println(fmt.Sprintf("Total time: %+q", diff))

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
}
