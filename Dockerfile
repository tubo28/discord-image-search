FROM golang:1.13 as build
ADD . /workspace
WORKDIR /workspace
COPY go.mod .
COPY go.sum .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main

FROM alpine
COPY --from=build /workspace/main .
CMD ["./main"]