FROM golang:1.25.7-alpine

RUN apk add --no-cache docker-cli
#RUN apt-get update && apt-get install -y docker.io
WORKDIR /app
COPY . .

#WORKDIR /app/worker
#RUN go mod tidy
#CMD ["go", "run", "main.go"]
CMD ["sh", "-c", "go mod tidy && go run main.go"]