FROM golang:1.16-alpine as build

# create work directory
WORKDIR /build

# add dependancies
ADD go.mod go.sum ./
ADD ./vendor ./vendor

# add source files
ADD ./pkg ./pkg
ADD ./internal ./internal
ADD ./aggregators ./aggregators

# build lambdas
RUN go build -o /build/map_sum ./aggregators/map_sum/main/map_sum.go 

# Build runtime for map sum aggregator
FROM alpine as map_sum
COPY --from=build /build/map_sum /map_sum
ENTRYPOINT [ "/map_sum" ]
