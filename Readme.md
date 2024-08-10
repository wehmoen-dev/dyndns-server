# DynDNS Server for Google Cloud DNS

This is a simple DynDNS server that updates a Google Cloud DNS record with the client's IP address. It is designed to be
run as a small docker container on a server that has a static IP address.

## Usage

1. Clone this repository
2. Build the docker image: `docker build -t dyndns:latest .`
3. Show the help: `docker run --rm dyndns:latest --help`

```shell
Usage of /dyndns.bin:
  -auth string
        Basic Auth username:password
  -auth-file string
        Path to file containing the Google Cloud credentials (default "google.json")
  -bind-address string
        Bind address for the server (default ":8080")
  -dns-zone-name string
        DNS zone name
  -domain-name string
        Domain name
  -project-id string
        Google Cloud project ID
```

4. Create a Google Cloud service account and download the JSON key file. It needs read/write permission for the DNS zone
   you want to update.

### Parameters

| Parameter     | Description                                                                    | Required                   |
|---------------|--------------------------------------------------------------------------------|----------------------------|
| auth          | Basic Auth username:password. Use format `username:password`                   | No                         |
| auth-file     | Path to the Google Cloud credentials file                                      | No - default `google.json` |
| bind-address  | Bind address for the server. Use format `ip:port`                              | No - default `:8080`       |
| dns-zone-name | DNS zone name from Cloud DNS                                                   | Yes                        |
| domain-name   | Domain name to update including the subdomain. For example `home.mydomain.tld` | Yes                        |
| project-id    | Google Cloud project ID                                                        | Yes                        |

*When using docker you should not change the `--bind-address` flag. The container will only expose port 8080.*

## Run the server

```shell
docker run -d --name dyndns \
  -p 8080:8080 \
  -v /path/to/google.json:/google.json \
  dyndns:latest \
  -auth username:password \
  -dns-zone-name my-dns-zone \
  -domain-name home.mydomain.tld \
  -project-id my-google-project
```

## Update the DNS record

This server supports updates of IPv4 and IPv6 addresses. The client can send a `GET` request to the server. 

```http
GET https://my-dyndns-server.lan/dyn?ip_address=IP_V4_ADDRESS&ipv6_address=IP_V6_ADDRESS
```

The server will validate the provided IP addresses and create/update the DNS records if they are valid.

### Example Response

```json
{
   "name": "home.mydomain.tld.",
   "v4": {
      "rr_type": "A",
      "success": true,
      "created": false,
      "updated": true,
      "value": "20.15.79.10"
   },
   "v6": {
      "rr_type": "AAAA",
      "success": true,
      "created": false,
      "updated": true,
      "value": "2a05:3100:0000:572:cafe:f00a:af39:beef"
   }
}

```

If validation fails or the record could not be updated, the response will contain an error message.

*You can only use public routable IP addresses. The server will not accept private or otherwise reserved IP addresses.*

---

## Run without docker

You can also run the server without docker. You need to have Go installed on your system.

1. Clone this repository
2. Build the server:

   - Linux/Mac: `go build -o dyndns.bin cmd/server/main.go`
   - Windows: `go build -o dyndns.exe cmd/server/main.go`

Run the server: `./dyndns.(bin|exe) --bind-address="localhost:3000" --auth username:password --dns-zone-name my-dns-zone --domain-name home.mydomain.tld --project-id my-google-project`

This will start the server on `localhost:3000` with basic auth and the provided Google Cloud credentials.