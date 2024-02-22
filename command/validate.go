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
)

var (
	totalEq   int64 = 0
	totalNeq  int64 = 0
	totalSkip int64 = 0
)

func Validate(threshold int, forceRewrite bool) {
	sourcePool := NewPool(hostSource, dbSource, sourceUsername, sourcePassword)
	destPool := NewPool(hostDest, dbDest, destUsername, destPassword)
	var wg sync.WaitGroup
	chanTotalEq := make(chan int64)
	chanTotalNeq := make(chan int64)
	chanTotalSkip := make(chan int64)
	chanBool := make(chan bool)

	iter := 0
	c1 := goccm.New(con1)

	go func() {
		for {
			select {
			case t := <-chanTotalEq:
				atomic.AddInt64(&totalEq, t)
				//log.Printf("total equal keys %d, total not equal keys %d, total skipped keys %d\n", totalEq, totalNeq, totalSkip)
			case t := <-chanTotalNeq:
				atomic.AddInt64(&totalNeq, t)
				//log.Printf("total equal keys %d, total not equal keys %d, total skipped keys %d\n", totalEq, totalNeq, totalSkip)
			case t := <-chanTotalSkip:
				atomic.AddInt64(&totalSkip, t)
				//log.Printf("total equal keys %d, total not equal keys %d, total skipped keys %d\n", totalEq, totalNeq, totalSkip)
			case <-chanBool:
				break
			}
		}
	}()

	counter := 0
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
		go func(iter int, keys []string, sourcePool *redis.Pool, destPool *redis.Pool, wgPointer *sync.WaitGroup, chanTotalEq, chanTotalNeq chan int64, forceRewrite bool) {
			connSourceItem, err := sourcePool.Dial()
			if err != nil {
				log.Fatalln(fmt.Sprintf("error during dialing redis source, err %+v", err))
			}
			defer connSourceItem.Close()

			connDestItem, err := destPool.Dial()
			if err != nil {
				log.Fatalln(fmt.Sprintf("error during dialing redis destination, err %+v", err))
			}
			defer connDestItem.Close()
			for _, key := range keys {
				hashSource, ttlSource, keyType, err := GetHashAndTTL(connSourceItem, key)
				if err != nil {
					log.Println(fmt.Sprintf("error during get hash and ttl source, err %+v", err))
					wgPointer.Done()
					continue
				}

				hashDest, ttlDest, _, err := GetHashAndTTL(connDestItem, key)
				if err != nil {
					log.Println(fmt.Sprintf("error during get hash and ttl dest, err %+v", err))
					chanTotalNeq <- 1
					wgPointer.Done()
					continue
				}

				eqHash := hashSource == hashDest
				eqTTL := false
				if (ttlSource == 0 && ttlDest == 0) || (ttlSource > 0 && ttlDest > 0) {
					eqTTL = true
				}

				if keyType == "hash" {
					chanTotalSkip <- 1
				} else if eqHash && eqTTL {
					chanTotalEq <- 1
				} else {
					chanTotalNeq <- 1
					log.Println("invalid key", key)

					if forceRewrite {
						connDest := destPool.Get()
						defer connDest.Close()
						_, err = redis.String(connDest.Do("RESTORE", key, ttlSource, hashSource, "REPLACE"))
						if err != nil {
							log.Println(fmt.Sprintf("error during execute command restore for key %s, err %+v", key, err))
						}
					}
				}

				wgPointer.Done()
			}
			c1.Done()
		}(iter, keys, sourcePool, destPool, &wg, chanTotalEq, chanTotalNeq, forceRewrite)
		counter += len(keys)
		//log.Println(counter, len(keys))
		if counter >= threshold {
			connSource.Close()
			break
		}

		if iter == 0 {
			connSource.Close()
			break
		}

		connSource.Close()
	}

	wg.Wait()
	chanBool <- true
	log.Println("finish")
	log.Printf("total equal keys %d, total not equal keys %d, total skipped keys %d\n", totalEq, totalNeq, totalSkip)

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

func GetHashAndTTL(conn redis.Conn, key string) (string, int, string, error) {
	ttl, err := redis.Int(conn.Do("PTTL", key))
	if err != nil {
		log.Println(fmt.Sprintf("error during receiving ttl, err %+v", err))
	}
	if ttl < 0 {
		ttl = 0
	}
	hashVal, err := redis.String(conn.Do("DUMP", key))
	if err != nil {
		log.Println(fmt.Sprintf("error during receiving hashVal, err %+v", err))
		return "", 0, "", err
	}

	keyType, err := redis.String(conn.Do("TYPE", key))
	if err != nil {
		log.Println(fmt.Sprintf("error during receiving keyType, err %+v", err))
		return "", 0, "", err
	}
	return hashVal, ttl, keyType, nil
}
