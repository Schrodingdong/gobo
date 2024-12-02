# GoBo - Open source Rate limiter
GoBo (aka. Go One By One), is a rate limiter implementation written in go.

- It uses the `Token Bucket algorithm`

## Quickstart
- Clone and build the project
```bash
go mod tidy
go build .
```

## Usage
```bash
./gobo [--src ip:port] [--max-bucket-size n] --dest ip:port
```
- `--src`: source of the requests.
- `--max-bucket-size` : Number of possible requests in a specific time frame (default: 5)
- `--refil-delay` : Number of possible requests in a specific time frame (default: 15s)
- `--dest`: dest to proxy the request to.

## Examples
```bash
./gobo --dest :9999
./gobo --dest localhost:9999
./gobo --dest 127.0.0.1:9999
./gobo --dest http://hostname_here:9999
./gobo --dest https://hostname_here:9999
./gobo --dest :9999 --refil-delay 5
```
