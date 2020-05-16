FROM golang:1.14.1 AS builder

# Build the DO controller
WORKDIR $GOPATH/src/github.com/astei/serpentinised
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix nocgo -ldflags="-s -w" -o /serpentinised .

FROM gcr.io/distroless/static-debian10:22bd467b41e5e656e31db347265fae118db166d9
USER nobody:nobody
COPY --from=builder serpentinised /
EXPOSE 6379
CMD ["/serpentinised", "-bind=0.0.0.0:6379"]
