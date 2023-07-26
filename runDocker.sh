docker image build -f dockerfile -t forum-app .
docker volume create --name dbvolume --opt type=none --opt device=$(pwd)/db/src/ --opt o=bind
docker container run -p 8080:8080 -v dbvolume:/data/db --detach --name forumcontainer forum-app
