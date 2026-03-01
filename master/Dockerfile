FROM golang:1.25.7-alpine
WORKDIR /app
# Copy ทั้งโปรเจกต์เพื่อให้เห็น ../proto
COPY . .
#WORKDIR /app/master
#RUN go mod tidy
#CMD ["go", "run", "main.go"]
CMD ["sh", "-c", "go mod tidy && go build -o main . && ./main"]