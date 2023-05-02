### perform stop on all docker containers
```shell 
$ docker stop $(docker ps -q)
```

### output docker build to log file and stdout (screen)
```shell
docker build --no-cache --progress=plain -t containerName . 2>&1 | tee build.log
```
