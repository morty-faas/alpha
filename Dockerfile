FROM golang:1.19 as build
WORKDIR /build
COPY . .
RUN go build -o alpha -a -gcflags=all="-l -B -wb=false" -ldflags="-w -s" main.go

FROM scratch
COPY --from=build /build/alpha ./alpha
ENTRYPOINT ["alpha"]