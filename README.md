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

