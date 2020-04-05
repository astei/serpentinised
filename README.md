# serpentinised

`serpentinised` acts as a proxy that allows clients that are unaware of
Redis Sentinel to connect to a Redis Sentinel cluster.

## Motivation

At Mineteria, we have a few applications using Django that are not aware
of Redis Sentinel. Of course, there are also plenty of applications that
are unaware of Redis Sentinel. `serpentinised` can help them cope with
the introduction or use of Redis Sentinel.

We previously used [redis-ellison](https://github.com/metal3d/redis-ellison).
While this solution worked, it had numerous CPU consumption issues, including
spawning a `redis-cli` instance every second. `serpentinised` connects to the
Redis Sentinel server directly and listens for any failover changes.

## Usage

```
Usage of serpentinised:
  -bind string
    	the address to bind to proxy connections to the active Sentinel (default "127.0.0.1:26380")
  -connect-timeout int
    	seconds before a connection to a master times out (default 1)
  -sentinel-address string
    	the address of the Sentinel master
  -sentinel-master string
    	the name of the Sentinel master (default "mymaster")
```