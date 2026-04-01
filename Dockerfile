FROM golang:1.25.0 AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 go build -o /out/short-link ./cmd/short-link

FROM scratch

COPY --from=build /out/short-link /short-link

EXPOSE 8081

ENTRYPOINT ["/short-link"]
