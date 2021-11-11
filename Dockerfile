FROM golang:1.16-alpine AS build

COPY . /source
WORKDIR /source
RUN cd /source && go build -o /gotestfmt cmd/gotestfmt/main.go
RUN chmod +x /gotestfmt

FROM alpine
COPY --from=build /gotestfmt /
ENTRYPOINT ["/gotestfmt"]
