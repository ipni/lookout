FROM golang:1.19-bullseye as build

WORKDIR /go/src/lookout

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /go/bin/lookout ./cmd/lookout

FROM gcr.io/distroless/static-debian11
COPY --from=build /go/bin/lookout /usr/bin/

ENTRYPOINT ["/usr/bin/lookout"]
