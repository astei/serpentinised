FROM golang:1.14.1 AS builder

# Add dep
ADD https://github.com/golang/dep/releases/download/v0.5.4/dep-linux-amd64 /usr/bin/dep
RUN chmod +x /usr/bin/dep

# Build the DO controller
WORKDIR $GOPATH/src/github.com/astei/serpentinised
COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure --vendor-only
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix nocgo -ldflags="-s -w" -o /serpentinised .

FROM gcr.io/distroless/static-debian10:22bd467b41e5e656e31db347265fae118db166d9
USER nobody:nobody
COPY --from=builder serpentinised /
CMD ["/serpentinised"]