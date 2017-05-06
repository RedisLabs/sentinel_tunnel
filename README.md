# Sentinel Tunnel
Sentinel Tunnel is a tool that allows you using the redis sentinel capabilities without any code modification on you application 
side. For more information about the redis sentinel: https://redis.io/topics/sentinel.

The sentinel tunnel creates a tcp tunnelling between a local port on the client computer to the current running master 
(it retrieves the current running muster from the sentinel).

When the master fails, the sentinel tunnel reflects the error to you client. When the client reconnects, the sentinel tunnel 
re-ask the sentinel for the new master address and connects to that address.

## Install
Make sure you have a working Go environment. [See the install instructions](http://golang.org/doc/install.html).

To install `sentinel_tunnel`, simply run:
```console
$ go get github.com/RedisLabs/sentinel_tunnel
```
Make sure your `PATH` includes the `$GOPATH/bin` directory so your
commands can be easily used:

```
export PATH=$PATH:$GOPATH/bin
```

## Configure
The code contains an example configuration file called `sentinel_tunnel_configuration_example.json`.
The configuration file is simply a json file that contains the following information:
* The sentinels addresses list
* The list of databases and their corresponding local port

For example, the following config file contains two sentinel addresses and two databases. When the client connects to
the local port `12345` it actually connect to `db1`.
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
In order to run the `sentinel_tunnel` get into the `bin` directory (created by the `go get` command) and perform:

```
./sentinel_tunnel <config_file_path> <log_file_path>
```
You can set the `log_file_path` to `/dev/null` if you do not want any log file.


