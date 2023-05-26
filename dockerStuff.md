### perform stop on all docker containers
```shell 
$ docker stop $(docker ps -q)
```

### output docker build to log file and stdout (screen)
```shell
docker build --no-cache --progress=plain -t containerName . 2>&1 | tee build.log
```

### manual watchtower run to update all docker containers
```shell
for image in $(docker images --format "{{.Repository}}:{{.Tag}}" | grep -v '<none>'); do docker run --rm --name watchtower -v /var/run/docker.sock:/var/run/docker.sock containrrr/watchtower --run-once --cleanup $image; done;
```

### AWS ECS Task Definition
- make sure to set the root file system in all containers to read only via ```"readonlyRootFilesystem": true,```
  - this is set in JSON immediately above ```"logConfiguration": {```
  - note the trailing comma, make sure to include that for the JSON to valiate.
