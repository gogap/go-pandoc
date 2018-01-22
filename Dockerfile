FROM idocking/pandoc:alpine-2.1.1

# Install golang and Install go-pandoc
RUN apk update \
    && apk add --no-cache --virtual .fetch-deps go git musl-dev\
    && go get -v github.com/gogap/go-pandoc \
    && mkdir app \
    && cd app \
    && go build github.com/gogap/go-pandoc \
    && cp -r $(go env GOPATH)/src/github.com/gogap/go-pandoc/templates . \
    && cp $(go env GOPATH)/src/github.com/gogap/go-pandoc/app.conf . \
    && rm -rf $(go env GOPATH) \
    && apk del .fetch-deps

WORKDIR /app

VOLUME /app

CMD ["./go-pandoc"]