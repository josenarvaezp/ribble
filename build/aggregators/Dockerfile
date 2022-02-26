FROM golang:1.16 as build

ARG CGO_ENABLED=0

# create work directory
WORKDIR /build

# install tools
RUN apt-get update && apt-get install -y upx

# add dependancies
ADD go.mod go.sum ./
RUN go mod download

# add source files
ADD ./pkg ./pkg
ADD ./internal ./internal
ADD ./aggregators ./aggregators

# build lambdas
RUN env GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o /build/map_sum ./aggregators/map_sum/main/map_sum.go 

# compress
RUN upx --best --lzma /build/map_sum

# Build runtime for map sum aggregator
FROM alpine as map_sum

COPY --from=build /build/map_sum /map_sum

ENTRYPOINT [ "/map_sum" ]
