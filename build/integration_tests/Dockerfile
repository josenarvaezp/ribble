FROM golang as build

ARG CGO_ENABLED=0

# install tools
RUN apt-get update && apt-get install -y unzip && \
    apt-get install -y make

# create work directory
WORKDIR /build

# add dependancies
ADD go.mod go.sum ./
RUN go mod download

ADD ./Makefile ./Makefile
ADD ./pkg ./pkg
ADD ./internal ./internal
ADD ./cmd ./cmd
ADD ./build/build_dockerfiles.sh ./build/
ADD ./build/generate_lambda_files.sh ./build/
ADD ./build/upload_images_locally.sh ./build/
ADD ./build/integration_tests/tests ./build/tests/

# Build the ribble CLI
RUN env GOOS=linux GOARCH=arm64 go build -o ./ribble ./cmd/driver/main.go

FROM golang

WORKDIR /build

COPY --from=build /build/ribble /build/ribble

ADD ./build/build_dockerfiles.sh ./build/
ADD ./build/generate_lambda_files.sh ./build/
ADD ./build/upload_images_locally.sh ./build/
ADD ./build/integration_tests/tests ./build/tests/

ADD $HOME/.ssh/id_ed25519  /root/.ssh/id_ed25519 
RUN chmod 600 /root/.ssh/id_ed25519 

RUN echo "[url \"git@github.com:\"]\n\tinsteadOf = https://github.com/" >> /root/.gitconfig
RUN export GOPRIVATE=github.com/josenarvaezp/displ

RUN go mod init integration_tests
RUN go mod tidy
# RUN go get github.com/josenarvaezp/displ/build/integration_tests/tests/query1

# FROM alpine

# # COPY --from=build /build/lambdas/coordinator /lambdas/coordinator
# RUN apk update && apk add unzip && \
#     apk add curl && \
#     apk add sudo && \
#     apk add libc6-compat

# RUN curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
# RUN unzip ./awscliv2.zip
# RUN export PATH=/usr/local/bin:$PATH
# RUN sudo ./aws/install


# # Build job
# RUN ./ribble build --job ./build/tests/query1/job/query1_job.go