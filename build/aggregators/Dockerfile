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
# RUN env GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o /build/sum ./aggregators/sum/main/sum.go 
# RUN env GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o /build/sum_final ./aggregators/sum_final/main/sum_final.go 

# compress
RUN upx --best --lzma /build/map_aggregator
# RUN upx --best --lzma /build/sum
# RUN upx --best --lzma /build/sum_final

# Build runtime for map aggregator
FROM alpine as map_aggregator

COPY --from=build /build/map_aggregator /map_aggregator

ENTRYPOINT [ "/map_aggregator" ]

# # Build runtime for sum aggregator
# FROM alpine as sum

# COPY --from=build /build/sum /sum

# ENTRYPOINT [ "/sum" ]

# # Build runtime for final sum aggregator
# FROM alpine as sum_final

# COPY --from=build /build/sum_final /sum_final

# ENTRYPOINT [ "/sum_final" ]
