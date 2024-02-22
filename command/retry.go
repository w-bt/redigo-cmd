package command

import (
	"bufio"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"log"
	"os"
	"time"
)

func Retry() {
	startTime = time.Now()

	sourcePool := NewPool(hostSource, dbSource, sourceUsername, sourcePassword)
	destPool := NewPool(hostDest, dbDest, destUsername, destPassword)

	connSource := sourcePool.Get()
	defer connSource.Close()

	connDest := destPool.Get()
	defer connDest.Close()

	keys := getFailedKeys()

	for _, key := range keys {
		if key == "" {
			continue
		}

		connSource.Send("PTTL", key)
		connSource.Send("DUMP", key)
		connSource.Flush()
		ttl, err := redis.Int(connSource.Receive())
		if err != nil {
			log.Println(fmt.Sprintf("error during receiving ttl, err %+v", err))
		}
		if ttl < 0 {
			ttl = 0
		}
		hashVal, err := redis.String(connSource.Receive())
		if err != nil {
			log.Println(fmt.Sprintf("error during receiving hashVal, err %+v", err))
			err = saveFailedKey(key, "failed_retry.txt")
			if err != nil {
				log.Println(fmt.Sprintf("error during save failed key key %s, err %+v", key, err))
			}
			continue
		}

		_, err = redis.String(connDest.Do("RESTORE", key, ttl, hashVal, "REPLACE"))
		if err != nil {
			log.Println(fmt.Sprintf("error during execute command restore for key %s, err %+v", key, err))
			err = saveFailedKey(key, "failed_retry.txt")
			if err != nil {
				log.Println(fmt.Sprintf("error during save failed key key %s, err %+v", key, err))
			}
		}
	}
}

func getFailedKeys() (keys []string) {
	readFile, err := os.Open("failed_keys.txt")
	if err != nil {
		log.Fatalf("unable to read file: %v", err)
	}

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)

	for fileScanner.Scan() {
		keys = append(keys, fileScanner.Text())
	}

	readFile.Close()

	return keys
}
