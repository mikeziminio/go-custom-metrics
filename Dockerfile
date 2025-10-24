FROM golang:1.24.7-bookworm

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

RUN  git clone https://github.com/Yandex-Practicum/go-autotests.git ./tmp-go-autotests
RUN (cd ./tmp-go-autotests && go test -c -o ../bin/metricstest ./cmd/metricstest)
RUN (cd ./tmp-go-autotests && go build -o ../bin/statictest ./cmd/statictest)
RUN (cd ./tmp-go-autotests && go build -o ../bin/random ./cmd/random)
RUN rm -r -f ./tmp-go-autotests
