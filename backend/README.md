## Docker commands

### Build docker image

docker build -t vim-zombies -f vim-zombies-docker .

### Run docker image

docker run -d -p 8080:8080 --name run vim-zombies

Runs in detached mode so you can stop it from same terminal, remove -d for debugging logs streamed to terminal. To see logs in detached mode run docker logs -f run.


### Stopping and removing docker container

docker stop run; docker rm run


### Doing everything at once

docker stop run; docker rm run; docker build -t run -f docker-prod .; docker run -p 8080:8080 --name run run

