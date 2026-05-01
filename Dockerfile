FROM golang:1.26.2-trixie AS build

WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .
RUN GOOS=linux go build -ldflags="-s -w" -o /main .

FROM golang:1.26.2-trixie AS dist
COPY --from=build /main /main
