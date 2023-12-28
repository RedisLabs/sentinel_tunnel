# Sentinel Tunnel

Sentinel Tunnel is a tool that allows you using the Redis Sentinel capabilities, without any code modifications to your application.

Redis Sentinel provides high availability (HA) for Redis. In practical terms this means that using Sentinel you can create a Redis deployment that tolerates certain kinds of failures without human intervention. For more information about Redis Sentinel refer to: <https://redis.io/topics/sentinel>.

## Overview

Connecting an application to a Sentinel-managed Redis deployment is usually done with a Sentinel-aware Redis client. While most Redis clients do support Sentinel, the application needs to call a specialized connection management interface of the client to use it. When one wishes to migrate to a Sentinel-enabled Redis deployment, she/he must modify the application to use Sentinel-based connection management. Moreover, when the application uses a Redis client that does not provide support for Sentinel, the migration becomes that much more complex because it also requires replacing the entire client library.

Sentinel Tunnel (ST) discovers the current Redis master via Sentinel, and creates a TCP tunnel between a local port on the client computer to the master. When the master fails, ST disconnects your client's connection. When the client reconnects, ST rediscovers the current master via Sentinel and provides the new address.
The following diagram illustrates that:

```                                                                                                          _
+----------------------------------------------------------+                                          _,-'*'-,_
| +---------------------------------------+                |                              _,-._      (_ o v # _)
| |                           +--------+  |  +----------+  |    +----------+          _,-'  *  `-._  (_'-,_,-'_)
| |Application code           | Redis  |  |  | Sentinel |  |    |  Redis   | +       (_  O     #  _) (_'|,_,|'_)
| |(uses regular connections) | client +<--->+  Tunnel  +<----->+ Sentinel +<--+---->(_`-._ ^ _,-'_)   '-,_,-'
| |                           +--------+  |  +----------+  |    +----------+ | |     (_`|._`|'_,|'_)
| +---------------------------------------+                |      +----------+ |     (_`|._`|'_,|'_)
| Application node                                         |        +----------+       `-._`|'_,-'
+----------------------------------------------------------+                               `-'
```

## Install

Make sure you have a working Go environment - [see the installation instructions here](http://golang.org/doc/install.html).

To install `sentinel_tunnel`, run:

```bash
go install github.com/USA-RedDragon/sentinel_tunnel@latest
```

Make sure your `PATH` includes the `$GOPATH/bin` directory so your commands can be easily used:

```bash
export PATH=$PATH:$GOPATH/bin
```

## Configure

The code contains an example configuration file named `configuration_example.json`. The configuration file is a json file that contains the following information:

* The Sentinels addresses list
* The list of databases and their corresponding local port

For example, the following config file contains two Sentinel addresses and two databases. When the client connects to the local port `12345` it actually connect to `db1`.

```json
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

### Manual

In order to run `sentinel_tunnel` manually:

```bash
./sentinel_tunnel <config_file_path>
```

### Docker

In order to run `sentinel_tunnel` using Docker:

```bash
docker run -v <config_file_path>:/config.json -p <local_port>:<port_in_docker> -d ghcr.io/usa-reddragon/sentinel_tunnel
```

## License

[2-Clause BSD](LICENSE)
