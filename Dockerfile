FROM golang:1.19.2-bullseye AS build

WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .
RUN GOOS=linux go build -ldflags="-s -w" -o /main .

FROM golang:1.19.2-bullseye AS dist
COPY --from=build /main /main
