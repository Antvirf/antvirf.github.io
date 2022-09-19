+++ 
date = 2022-09-21
title = "Resolving port overlaps when using Docker compose"
description = "When resolving containers by their service names is the primary objective, Docker compose can run into issues with overlapping ports on localhost. This can be avoided using expose."
author = "Antti Viitala"
tags = [
    "devops",
    "docker"
]
+++

## Scenario

* Front-end container ```front```, serving a webapp on port 9001
* Back-end service container ```back```, serving an API on port 80
* **Primary Objective**: Reach and query ```back``` container from the ```front```, using its service name: ```http://back/...```
* Secondary Objective: Have a way to reach both ```front``` and ```back``` from the host machine for troubleshooting

## Problem

My first attempt was to use port mapping for both - the front must be available on 80 to be able to use it 'nicely' in a browser, and the back must be on port 80 so that it can be queried 'nicely' in code without thinking about the port. Example below:

```yaml
services:
  front:
    image: front-end-container:latest
    ports:
      - 80:9001

  back:
    image: back-end-container:latest
    ports:
      - 80:80
```

This will fail with ```port is already allocated``` as ```localhost:80``` can only host one service at at time.

## Solution: Connect containers with ```expose```

Use Docker compose configuration option [Expose](https://docs.docker.com/compose/compose-file/compose-file-v3/#expose):
> *Expose ports without publishing them to the host machine - they’ll only be accessible to linked services. Only the internal port can be specified.*

Expose "drills a hole in to the container", and nothing else. We can have a 100 containers, all with the same ports opened with this option.  There is no port mapping - the service running inside within *must* be configured to use the same port to be visible via that "hole".

In the example below, instead of both containers being served over port 80 on ```localhost```, we serve the front-end via 80 using port mapping as normal.

For the back-end we use ```expose``` to open the important port 80 for the front-end to query, and use a normal ```port``` mapping to serve the back-end on  ```localhost:8080``` instead to enable troubleshooting and avoiding the overlap on ```localhost:80```.

```yaml
services:
  front:
    image: front-end-container:latest
    ports:
      - 80:9001

  back:
    image: back-end-container:latest
    ports:
      - 8080:80
    expose:
      - 80 # the key bit - port used by other containers to access this container
```

As a result, with this configuration we can happily:

* Query ```back``` from the ```front``` with ```http://back/api/v2/...``` - just like we would inside kubernetes
* Access the ```front``` from ```http://localhost:80``` or ```http://localhost```
* Access the ```back``` from ```http://localhost:8080```
