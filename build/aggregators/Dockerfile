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
RUN env GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o /build/map_sum ./aggregators/map_sum/main/map_sum.go 
RUN env GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o /build/map_max ./aggregators/map_max/main/map_max.go 
RUN env GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o /build/map_min ./aggregators/map_min/main/map_min.go 
RUN env GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o /build/sum ./aggregators/sum/main/sum.go 
RUN env GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o /build/sum_final ./aggregators/sum_final/main/sum_final.go 

# compress
RUN upx --best --lzma /build/map_sum
RUN upx --best --lzma /build/map_max
RUN upx --best --lzma /build/map_min
RUN upx --best --lzma /build/sum
RUN upx --best --lzma /build/sum_final

# Build runtime for map sum aggregator
FROM alpine as map_sum

COPY --from=build /build/map_sum /map_sum

ENTRYPOINT [ "/map_sum" ]

# Build runtime for map max aggregator
FROM alpine as map_max

COPY --from=build /build/map_max /map_max

ENTRYPOINT [ "/map_max" ]

# Build runtime for map min aggregator
FROM alpine as map_min

COPY --from=build /build/map_min /map_min

ENTRYPOINT [ "/map_min" ]

# Build runtime for sum aggregator
FROM alpine as sum

COPY --from=build /build/sum /sum

ENTRYPOINT [ "/sum" ]

# Build runtime for final sum aggregator
FROM alpine as sum_final

COPY --from=build /build/sum_final /sum_final

ENTRYPOINT [ "/sum_final" ]
