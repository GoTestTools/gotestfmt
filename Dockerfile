FROM golang:1.16-alpine AS build

COPY . /source
WORKDIR /source
RUN cd /source && go build -o /gotestfmt cmd/gotestfmt/main.go
RUN chmod +x /gotestfmt

FROM alpine
COPY --from=build /gotestfmt /
ENTRYPOINT ["/gotestfmt"]

LABEL org.opencontainers.image.title="gotestfmt: go test output for humans"
LABEL org.opencontainers.image.documentation="https://github.com/haveyoudebuggedit/gotestfmt#readme"
LABEL org.opencontainers.image.source="https://github.com/haveyoudebuggedit/gotestfmt"
LABEL org.opencontainers.image.description="This image wraps gotestfmt, a go test output formatter. You can access it by running /gotestfmt in this image, or you can use this as a base image to extract the binary. The binary itself has no dependencies and can be safely copied to other AMD64/Linux container image."
LABEL org.opencontainers.image.licenses="Unlicense"