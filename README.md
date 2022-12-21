# Go Proxy Autostart - Autostart container on the fly

Go Proxy Autorstart is a go program that act like a reverse proxy, and start a given container when the proxy receive a connection. This project was initially designed to work with [itzg/minecraft-server](https://hub.docker.com/r/itzg/minecraft-server), in order to automatically start the server on incoming connection, while the server shutdown itself after a timeout.

This project should work with any other container accepting TCP connections, the proxy will start it, however, it will not stop it.

**Warning** The proxy start the container on the first TCP connction it receive on the exposed port! Exposing this on Internet may be useless if not behind a firewall, because bots scanning the web may start the server for you :/

# Usage

A `docker-compose.yml` should look like this :

```yml
version: "2.2"

services:
  minecraft:
    container_name: minecraft
    image: itzg/minecraft-server:latest
    volumes:
      - "/path/to/data:/data"
    env_file: stack.env # from portainer, you should put your env here
  go-proxy:
    container_name: go-proxy-minecraft
    image: tolfsh/go-proxy-autostart
    volumes:
      - "/var/run/docker.sock:/var/run/docker.sock:ro"
    environment:
      - "CONTAINER_NAME=minecraft" # this is the name of the minecraft, used to start the container, and initialize the TCP connection
      - "LISTEN_IP=0.0.0.0" # OPTIONAL : the IP address to listen on, default is 0.0.0.0
      - "LISTEN_PORT=25565" # OPTIONAL : the port to listen on, default is 25565
      - "SERVICE_PORT=25565" # OPTIONAL : the port your Minecraft container is listening on, default is 25565
      - "TZ=Europe/Paris"
    ports:
      - 25565:25565 
```

# Disclaimer

This project was made because the implementation of [Lazytainer](https://github.com/vmorganp/Lazytainer) did not really satisfied my needs. I wanted the process to be flowless for the enduser, just looking a the "Pinging..." prompt from its Minecraft client util the server is up. Also, it was an excuse to learn Go, thus, the code may not be the cleanest, but it works!