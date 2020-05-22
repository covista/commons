# Covid Commons


## Setup
```
docker-compose build
docker-compose up
```

### Building `commons-server` From Source
Docker will take care of building the binary, but you can also do it yourself.

0. Make sure to initialize the `googleapis` submodule if you are editing the .proto definitions
    ```
    git submodule update --init --recursive
    ```

1. Install dependencies:
    ```
    sudo apt install libprotobuf-dev  # or equivalent on other systems
    sudo apt install protobuf-compiler
    pip3 install grpcio-tools
    go install \
        github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway \
        github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger \
        github.com/golang/protobuf/protoc-gen-go
    ```

2. Compile server and proto files
    ```
    make
    ```

## Simulation

We have a simple simulation demonstarting the workflow involved implemented in `simulation`. `N` entities randomly interact and query the database for a configurable number of days. To run:

```bash
$ cd simulate
$ time python simulation.py --entities 100 --days 7
2020-05-02 00:00:00+00:00
2020-05-03 00:00:00+00:00
entity-75 was exposed at 2020-05-02 00:00:00
2020-05-04 00:00:00+00:00
entity-73 was exposed at 2020-05-03 00:00:00
entity-75 was exposed at 2020-05-02 00:00:00
entity-79 was exposed at 2020-05-03 00:00:00
2020-05-05 00:00:00+00:00
entity-35 was exposed at 2020-05-04 00:00:00
entity-67 was exposed at 2020-05-04 00:00:00
... etc
```
