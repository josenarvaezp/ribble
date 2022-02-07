FROM golang:1.16-alpine as build

# create work directory
WORKDIR /build

# add dependancies
ADD go.mod go.sum ./
ADD ./vendor ./vendor

# add source files
ADD ./internal ./internal
ADD ./lambdamap ./lambdamap
ADD ./lambdareduce ./lambdareduce
ADD ./lambdacoordinator ./lambdacoordinator

# build lambdas
RUN go build -o /build/lambdas/ ./lambdamap/main/lambdamap.go 
RUN go build -o /build/lambdas/ ./lambdareduce/main/lambdareduce.go 
RUN go build -o /build/lambdas/ ./lambdacoordinator/main/lambdacoordinator.go 

# Build runtime for mapper
FROM alpine as lambdamap
COPY --from=build /build/lambdas/lambdamap /lambdas/lambdamap
ENTRYPOINT [ "/lambdas/lambdamap" ]

# Build runtime for reducer
FROM alpine as lambdareduce
COPY --from=build /build/lambdas/lambdareduce /lambdas/lambdareduce
ENTRYPOINT [ "/lambdas/lambdareduce" ]

# Build runtime for coordinator
FROM alpine as lambdacoordinator
COPY --from=build /build/lambdas/lambdacoordinator /lambdas/lambdacoordinator
ENTRYPOINT [ "/lambdas/lambdacoordinator" ]
