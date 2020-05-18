# Covid Commons


## Setup

0. Make sure to initialize the `googleapis` submodule if you are editing the .proto definitions
    ```
    git submodule update --init --recursive
    ```

1. Install dependencies:
    ```
    sudo apt install libprotobuf-dev  # or equivalent on other systems
    go install \
        github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway \
        github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger \
        github.com/golang/protobuf/protoc-gen-go
    ```

2. Compile server and proto files
    ```
    make proto # if necessary
    make commons-server
    ```

3. Build Docker files and deploy. **Note: you will need to rebuild the commons-server using `make` before running `docker-compose`; the Dockerfile uses the prebuilt binary**
    ```
    docker-compose build
    docker-compose up
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
