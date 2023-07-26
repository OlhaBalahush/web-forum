FROM golang:1.20
RUN mkdir /forum-app
ADD . /forum-app
WORKDIR /forum-app
RUN go mod download
RUN go build -o server .
RUN apt-get update && apt-get install -y sqlite3
EXPOSE 8080
CMD ["/forum-app/server", "--docker=true"]
