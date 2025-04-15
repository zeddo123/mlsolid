FROM golang:1.24 AS build
ADD . /src
WORKDIR /src

RUN go get
RUN go test --cover -v ./...
RUN CGO_ENABLED=0 go build -v


FROM alpine:latest
EXPOSE 8000
CMD [ "mlsolid" ]
COPY --from=build /src/mlsolid /usr/local/bin/mlsolid
RUN chmod +x /usr/local/bin/mlsolid
