# Covid Commons


## Setup

0. Make sure to initialize the `googleapis` submodule if you are editing the .proto definitions:
    ```
    git submodule update --init --recursive
    ```

1. Compile server and proto files
    ```
    make proto # if necessary
    make commons-server
    ```

2. Build Docker files and deploy. **Note: you will need to rebuild the commons-server using `make` before running `docker-compose`; the Dockerfile uses the prebuilt binary**
    ```
    docker-compose build
    docker-compose up
    ```
