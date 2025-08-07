FROM golang:1.24.6-bullseye AS build

WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .
RUN GOOS=linux go build -ldflags="-s -w" -o /main .

FROM golang:1.24.6-bullseye AS dist
COPY --from=build /main /main
