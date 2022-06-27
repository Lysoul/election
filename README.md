## Setup local development

### Install tools

- [Migrate](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate)

    ```bash
    brew install golang-migrate
    ```

- [Sqlc](https://github.com/kyleconroy/sqlc#installation)

    ```bash
    brew install sqlc
    ```

- [Gomock](https://github.com/golang/mock)

    ``` bash
    go install github.com/golang/mock/mockgen@v1.6.0
    ```

### Setup infrastructure

- Create election database:

    ```bash
    make createdb
    ```
    
- Run db migration up all versions:

    ```bash
    make migrateup
    ```

- Run db migration down all versions:

    ```bash
    make migratedown
    ```


### How to generate code

- Generate SQL CRUD with sqlc:

    ```bash
    make sqlc
    ```

- Generate DB mock with gomock:

    ```bash
    make mock
    ```

- Create a new db migration:

    ```bash
    migrate create -ext sql -dir db/migration -seq <migration_name>
    ```

### How to run

- Run server:

    ```bash
    make server
    ```
- Run UI:

    ```bash
    make ui
    ```

- Run test:

    ```bash
    make test
    ```
    


