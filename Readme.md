# go-api

Program to save all Odds of the upcoming matches with set delay from Odds-Api.com

## Installation

Use the package manager  to install all dependencies.

```bash
go get ./...
```

## Usage

Set the enviornment variables in config.yaml file

```yaml
API_KEY : "YOUR_API_KEY"

DB:
  type: "YOUR_DB_TYPE"
  host: "YOUR_DB_HOST"
  port: "YOUR_DB_PORT"
  user: "YOUR_DB_USER"
  dbname: "YOUR_DB_NAME"
  sslmode: "disable"
  password: "YOUR_DB_PASSWORD"

DELAY : DELAY_FOR_ALL_MATCHES
DELAY_UPCOMING: DELAY_FOR_IN_PLAY_MATCHES
```

#Run the main.go file

```bash
go run main.go
```