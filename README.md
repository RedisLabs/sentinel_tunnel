# Sentinel Tunnel
Redis Sentinel provides high availability for Redis. In practical terms this means that using Sentinel you can create a Redis deployment that resists without human intervention to certain kind of failures. For more information about the Redis Sentinel refer to: https://redis.io/topics/sentinel.

Most of the redis clients has a special implementation for sentinel based connection. When one wishes to start using the HA capabilities of redis with sentinel, he must modified his code to use this spacific sentinel based connection. Moreover some clients do not even support sentinel based connection and so if one wishes to start using the sentinel he must change his entire client library.

Sentinel Tunnel (ST) is a tool that allows you using the Redis Sentinel capabilities, without any code modifications to your application. Sentinel Tunnel discovers the current Redis master via Sentinel, and creates a TCP tunnel between a local port on the client computer to the master. When the master fails, ST disconnects your client's connection. When the client reconnects, ST rediscovers the current master via Sentinel and provides the new address.

## Install
Make sure you have a working Go environment - [see the installation instructions here](http://golang.org/doc/install.html).

To install `sentinel_tunnel`, run:
```console
$ go get github.com/RedisLabs/sentinel_tunnel
```
Make sure your `PATH` includes the `$GOPATH/bin` directory so your commands can be easily used:

```
export PATH=$PATH:$GOPATH/bin
```

## Configure
The code contains an example configuration file named `sentinel_tunnel_configuration_example.json`. The configuration file is a json file that contains the following information:
* The Sentinels addresses list
* The list of databases and their corresponding local port

For example, the following config file contains two Sentinel addresses and two databases. When the client connects to the local port `12345` it actually connect to `db1`.
```
{
	"Sentinels_addresses_list":[
		"node1.local:8001",
		"node2.local:8001"
	],
	"Databases":[
		{
			"Name":"db1",
			"Local_port":"12345"
		},
		{
			"Name":"db2",
			"Local_port":"12346"	
		}
	]
}
```

## Run
In order to run the `sentinel_tunnel`:

```
./sentinel_tunnel <config_file_path> <log_file_path>
```
You can set the `log_file_path` to `/dev/null` if you do not want any log file.

## License

[2-Clause BSD](LICENSE)
