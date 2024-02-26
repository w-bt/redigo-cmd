# REDIS Command

Script for migrating redis key

### Requirements

- Golang 1.21
- Redis 6.0

# How it works

## Migration
- Change command/const.go
```
sourceUsername = ""
sourcePassword = ""
hostDest       = "localhost:6379"
destUsername   = ""
destPassword   = ""
hostSource     = "localhost:6380"
```
- Go build -a
- `./redigo-cmd migrate` for migrating redis keys

## Migration Specific Keys
- Change command/const.go
```
sourceUsername = ""
sourcePassword = ""
hostDest       = "localhost:6379"
destUsername   = ""
destPassword   = ""
hostSource     = "localhost:6380"

prefixKeys
```
- Go build -a
- `./redigo-cmd migrate_specific` for migrating redis keys

## Validation
- Change command/const.go
```
sourceUsername = ""
sourcePassword = ""
hostDest       = "localhost:6379"
destUsername   = ""
destPassword   = ""
hostSource     = "localhost:6380"
```
- Go build -a
- `./redigo-cmd validate 100 true` for validating redis keys, `100` will check the first 100 keys, `true` if we want to force update on the destination redis.

## Retry
- Change command/const.go
```
sourceUsername = ""
sourcePassword = ""
hostDest       = "localhost:6379"
destUsername   = ""
destPassword   = ""
hostSource     = "localhost:6380"
```
- Go build -a
- Create file `failed_keys.txt`, list down failed keys separated by new line
- `./redigo-cmd retry` for retrying failed keys