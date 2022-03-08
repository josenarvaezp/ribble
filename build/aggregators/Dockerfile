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

# build aggregators
RUN env GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o /build/map_aggregator ./aggregators/map_aggregator/main/map_aggregator.go 
RUN env GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o /build/map_random_aggregator ./aggregators/map_random_aggregator/main/map_random_aggregator.go 
RUN env GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o /build/final_aggregator ./aggregators/final_aggregator/main/final_aggregator.go 

# compress
RUN upx --best --lzma /build/map_aggregator
RUN upx --best --lzma /build/map_random_aggregator
RUN upx --best --lzma /build/final_aggregator

# Build runtime for map aggregator
FROM alpine as map_aggregator

COPY --from=build /build/map_aggregator /map_aggregator

ENTRYPOINT [ "/map_aggregator" ]

# Build runtime for map_random_aggregator aggregator
FROM alpine as map_random_aggregator

COPY --from=build /build/map_random_aggregator /map_random_aggregator

ENTRYPOINT [ "/map_random_aggregator" ]

# Build runtime for final_aggregator aggregator
FROM alpine as final_aggregator

COPY --from=build /build/final_aggregator /final_aggregator

ENTRYPOINT [ "/final_aggregator" ]
