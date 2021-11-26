FROM golang:1.16-alpine

WORKDIR /app

COPY go.mod go.sum ./
COPY cmd/ ./cmd
COPY models/ ./models

COPY .env ./

RUN go mod download

COPY *.go ./

WORKDIR /app/cmd/api

RUN go build -o /docker-gs-ping

EXPOSE 4000

CMD [ "/docker-gs-ping" ]

# docker image rm cred-gorm_web -f
# docker build -t inspirit_golang .
# docker compose up
# docker.for.mac.localhost:3306
# docker run -it 376e508a2137 sh
