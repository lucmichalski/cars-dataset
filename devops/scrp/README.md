# What
A fully resumable horizontally (infinitely) scalable webscraper in Go. Think 1000's of machines scraping sites in a distributed way. Based on Docker swarm, Cassandra, Traefik, colly, gRPC, and my other [boilerplate](https://github.com/dioptre/gtrpc).

# Note
You could probably just use colly... Especially if you don't care about scalability... or use a shell script ([Example](https://github.com/dioptre/scrp/blob/master/simple.sh))

# Why
I built this to distribute scraping across multiple servers, so as to go undetected. I could have used proxies, but wanted to reuse the code for other distributed apps.

# Instructions
Run
```
docker-compose up
```
Then in the scrp container (```docker exec -it 045 bash```) run gcli to issue the command to service:

```
/app/scrp/gcli https://en.wikipedia.org/wiki/List_of_HTTP_status_codes
```

# Dependencies
## gRPC SSL Certificate (https://docs.traefik.io/v2.0/user-guides/grpc/)
```

In order to secure the gRPC server, we generate a self-signed certificate for service url:

openssl req -new -x509 -sha256 -newkey rsa:2048 -nodes -keyout backend.key -days 365 -out backend.cert -subj '/CN=backend.local'

openssl req -new -x509 -sha256 -newkey rsa:2048 -nodes -keyout frontend.key -days 365 -out frontend.cert -subj '/CN=frontend.local'


```

That will prompt for information, the important answer is:

Common Name (e.g. server FQDN or YOUR name) []: backend.local / frontend.local

# Thanks
Cheers to the engineers of Cassandra, colly, gRPC,Consul, Traefik & protobuf to name a few.

