FROM golang:1.25 as build

COPY . /src

WORKDIR /src

RUN CGO_ENABLED=0 GOOS=linux go build -o dkvs

FROM scratch

COPY --from=build /src/dkvs .

COPY --from=build /src/*.pem .

EXPOSE 8000

CMD ["/dkvs"]
