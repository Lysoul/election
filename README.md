## Setup local development

### Install tools

- [Golang](https://go.dev/)

- [PostgresSQL](https://www.postgresql.org/)

- [Angular-Cli](https://angular.io/cli)

- [Migrate](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate)

    ```bash
    scoop install migrate
    ```

- [Sqlc](https://github.com/kyleconroy/sqlc#installation)

    ```bash
    go install github.com/kyleconroy/sqlc/cmd/sqlc@latest
    ```

- [Gomock](https://github.com/golang/mock)

    ``` bash
    go install github.com/golang/mock/mockgen@v1.6.0
    ```

### Setup infrastructure (Windows)

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

- Install UI dependencies:

    ```bash
    make ui-package
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
    


