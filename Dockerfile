FROM golang:1.23 AS build

WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/ai-change-risk-gate ./cmd/server

FROM gcr.io/distroless/static-debian12

ENV ADDR=:8080

WORKDIR /
COPY --from=build /out/ai-change-risk-gate /ai-change-risk-gate

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/ai-change-risk-gate"]