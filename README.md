# Go Redis

A simple implementation of Redis server in Go programming lanugage.

## Motivation

I built this Redis server in Go as a personal project and a learning experience. It helped me better understand in-memory databases and how to handle network connections in Go efficiently.

## 🚀 Quick Start

Ensure you have either [Docker](https://www.docker.com/get-started) or [Go](https://golang.org/doc/install) environments set up. Additionally, make sure [Redis](https://redis.io/download) is installed on your local machine.

### Clone the project:

```bash
git clone https://github.com/danilovict2/go-redis.git
cd go-redis
```

### Run with Docker:

```bash
docker build .
docker container run -p "6379:6379" <IMAGE ID>
```

### Run locally:

```bash
./your_program.sh
```

### Interact with Go Redis:

Open a new terminal and run the following command to start the Redis client:
```bash
redis-cli
```

## 📖 Usage

### Available commands

* `PING`
* `ECHO`
* `SET`
* `GET`
* `CONFIG GET`
* `KEYS`
* `INFO REPLICATION`
* `REPLCONF GETACK`
* `WAIT`
* `TYPE`
* `XADD`
* `XRANGE`
* `XREAD`
* `INCR`
* `RPUSH`
* `LPUSH`
* `LRANGE`
* `LLEN`
* `LPOP`
* `BLPOP`
* `ZADD`
* `ZRANK`
* `ZRANGE`
* `ZCARD`
* `ZSCORE`
* `ZREM`
* `GEOADD`
* `GEOPOS`
* `GEODIST`
* `GEOSEARCH`
* `SUBSCRIBE`
* `UNSUBSCRIBE`
* `PUBLISH`
* `MULTI`
* `EXEC`
* `DISCARD`
* `WATCH`
* `UNWATCH`
* `AUTH`
* `ACL`

## Examples

### Creating a replica
A replica is a clone of a Redis instance that listens to all write commands and executes them.

### Creating a replica with Docker

1. **Create a network:**

```bash
docker network create my-network
```

2. **Run the main program:**
    
```bash
docker container run -p "6379:6379" --network my-network --name leader <IMAGE ID>
```

3. **Start the replica in a new terminal window:**

```bash
docker container run -p "<PORT>:<PORT>" --network my-network --name replica <IMAGE ID> --port <PORT> --replicaof "leader 6379"
```

4. **Connect to the replica:**
    
```bash
redis-cli -p <PORT>
```

### Creating a replica locally

1. **Run the main program:**

```bash
./your_program.sh
```

2. **Start the replica in a new terminal window:**

```bash
./your_program.sh --port <PORT> --replicaof "localhost 6379"
```

3. **Connect to the replica:**

```bash
redis-cli -p <PORT>
```

### Reading Data from an RDB File

To load data from an RDB file into the program, you can either:

1. Place the file in the project's root directory and name it `dump.rdb`.
2. Or, run the following command to specify the file's location and name:

```bash
./your_program --dir "/path/to/your/file" --dbfilename "file.rdb"
```

For example, if you have `file.rdb` in the home directory, run:
```bash
./your_program.sh --dir "/home" --dbfilename "file.rdb"
```

### Streams

#### Creating a Stream

To create a stream and add an entry to it, use the following command:

```bash
redis-cli XADD some_key 1526985054069-0 temperature 36 humidity 95
```

The output will be the ID of the first added entry:

```bash
"1526985054069-0"
```

#### Querying Entries from a Stream

Once you've added a few entries to the stream, you can query them using the `XRANGE` command.

1. **Querying a Specific Range:**

```bash
$ redis-cli XADD some_key 1526985054069-0 temperature 36 humidity 95
"1526985054069-0"
$ redis-cli XADD some_key 1526985054079-0 temperature 37 humidity 94
"1526985054079-0"
$ redis-cli XRANGE some_key 1526985054069 1526985054079
1) 1) 1526985054069-0
   2) 1) temperature
      2) 36
      3) humidity
      4) 95
2) 1) 1526985054079-0
   2) 1) temperature
      2) 37
      3) humidity
      4) 94
```

2. **Querying Until the End of the Stream:**

Use the `+` symbol to query entries up to the end of the stream:

```bash
$ redis-cli XRANGE some_key 1526985054069 +
1) 1) 1526985054069-0
   2) 1) temperature
      2) 36
      3) humidity
      4) 95
2) 1) 1526985054079-0
   2) 1) temperature
      2) 37
      3) humidity
      4) 94
```

3. **Querying from the Beginning of the Stream:**

Use the `-` symbol to query entries from the beginning of the stream:

```bash
$ redis-cli XRANGE some_key - 1526985054079
1) 1) 1526985054069-0
   2) 1) temperature
      2) 36
      3) humidity
      4) 95
2) 1) 1526985054079-0
   2) 1) temperature
      2) 37
      3) humidity
      4) 94
```

#### Reading Data from Streams with XREAD

The `XREAD` command is used to read data from one or more streams, starting from a specified entry ID:

```bash
$ redis-cli XADD some_key 1526985054069-0 temperature 36 humidity 95
"1526985054069-0"
$ redis-cli XADD some_key 1526985054079-0 temperature 37 humidity 94
"1526985054079-0"
$ redis-cli XREAD streams some_key 1526985054069-0
1) 1) "some_key"
   2) 1) 1) 1526985054079-0
         2) 1) temperature
            2) 37
            3) humidity
            4) 94
```

#### Blocking Reads Using $

You can block reads until a new entry is added to the stream or the specified timeout is reached. Use the `$` symbol for this:

```bash
$ redis-cli XADD some_key 1526985054069-0 temperature 36 humidity 95
"1526985054069-0"
$ redis-cli XREAD block 1000 streams some_key $
```

This command will block reads until either 1000ms has passed or another Redis client adds a new entry to the stream:

```bash
$ redis-cli -p <PORT> XADD some_key 1526985054079-0 temperature 37 humidity 94
"1526985054079-0"
```

### Lists

#### Creating a list and appending elements

```bash
> RPUSH list "bar" "baz"
> LPUSH list_key "a" "b" "c"
```

#### List elements

```bash
> RPUSH list_key "a" "b" "c" "d" "e"
(integer) 5

# List first 2 items 
> LRANGE list_key 0 1
1) "a"
2) "b"

# List last 2 items 
> LRANGE list_key -2 -1
1) "d"
2) "e"
```

#### Query list length

```bash
> RPUSH list_key "a" "b" "c" "d"
(integer) 4

> LLEN list_key
(integer) 4
```

#### Remove elements

```bash
> RPUSH list_key "a" "b" "c" "d"
(integer) 4
> LPOP list_key 2
1) "a"
2) "b"
> LRANGE list_key 0 -1
1) "c"
2) "d"
```

#### Blocking retrieval

```bash
> BLPOP list_key 0

# ... this blocks until an element is added to the list

# As soon as an element is added, it responds with a resp array:
1) "list_key"
2) "foobar"

$ redis-cli BLPOP list_key 0.1
# (Blocks for 0.1 seconds)
```

### Sorted Sets

#### Adding members

```bash
> ZADD scores 10 alice
(integer) 1
> ZADD scores 20 bob
(integer) 1
> ZADD scores 15 carol
(integer) 1
```

#### Querying members

```bash
# Number of members in the set
> ZCARD scores
(integer) 3

# Rank of a member (lowest score first, 0-based)
> ZRANK scores bob
(integer) 2

# Score of a member
> ZSCORE scores carol
"15"

# Members in a rank range, ordered by score
> ZRANGE scores 0 -1
1) "alice"
2) "carol"
3) "bob"
```

#### Removing members

```bash
> ZREM scores bob
(integer) 1
```

### Geospatial

The geospatial commands store coordinates as members of a sorted set, so positions are encoded by their longitude and latitude.

#### Adding locations

```bash
> GEOADD places 13.361389 38.115556 "Palermo"
(integer) 1
> GEOADD places 15.087269 37.502669 "Catania"
(integer) 1
```

#### Querying positions and distance

```bash
# Position of one or more members
> GEOPOS places Palermo
1) 1) "13.361389"
   2) "38.115556"

# Distance between two members (in meters by default)
> GEODIST places Palermo Catania
"166274.1516"
```

#### Searching by radius

Search for members within a given radius of a longitude/latitude. The radius unit can be `m`, `km` or `mi`:

```bash
> GEOSEARCH places FROMLONLAT 15 37 BYRADIUS 200 km
1) "Palermo"
2) "Catania"
```

### Pub/Sub

#### Subscribing to a channel

```bash
> SUBSCRIBE news
1) "subscribe"
2) "news"
3) (integer) 1

# ... this connection now blocks, waiting for messages
```

#### Publishing a message

In another terminal, publish a message to the channel:

```bash
$ redis-cli PUBLISH news "hello"
```

The subscriber will then receive:

```bash
1) "message"
2) "news"
3) "hello"
```

#### Unsubscribing

```bash
> UNSUBSCRIBE news
```

### Transactions

A transaction lets you queue multiple commands and run them together with `EXEC`.

#### Running a transaction

```bash
> MULTI
OK
> SET foo 1
QUEUED
> INCR foo
QUEUED
> EXEC
1) OK
2) (integer) 2
```

#### Aborting a transaction

Use `DISCARD` to throw away a queued transaction without running it:

```bash
> MULTI
OK
> SET foo 1
QUEUED
> DISCARD
OK
```

#### Watching keys

`WATCH` marks keys for optimistic locking. If a watched key is modified before `EXEC`, the transaction is aborted. `UNWATCH` clears all watched keys, and `DISCARD` also clears them automatically:

```bash
> WATCH foo
OK
> MULTI
OK
> INCR foo
QUEUED
> EXEC
# Aborts if foo was changed by another client after WATCH
```

### Authentication

The server starts with a `default` user that requires no password. You can add a password to a user with `ACL SETUSER` and then authenticate with `AUTH`.

#### Inspecting users

```bash
# The currently authenticated user
> ACL WHOAMI
"default"

# The flags and passwords configured for a user
> ACL GETUSER default
1) "flags"
2) 1) "nopass"
3) "passwords"
4) (empty array)
```

#### Setting a password and authenticating

```bash
> ACL SETUSER default >mypassword
OK
> AUTH default mypassword
OK
```


## 🤝 Contributing

### Build the project

```bash
go build -o redis app/*.go
```

### Run the project

```bash
./redis
```

### Submit a pull request

If you'd like to contribute, please fork the repository and open a pull request to the `master` branch.

