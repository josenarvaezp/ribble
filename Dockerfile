FROM golang:1.16-alpine as build

# create work directory
WORKDIR /build

# add dependancies
ADD go.mod go.sum ./
ADD ./vendor ./vendor

# add source files
ADD ./internal ./internal
ADD ./mapper ./mapper

# build
RUN go build -o /build/lambdas/ ./mapper/main/mapper.go 

# copy artifacts to a clean image
FROM alpine
COPY --from=build /build/lambdas/mapper /lambdas/mapper
ENTRYPOINT [ "/lambdas/mapper" ]  