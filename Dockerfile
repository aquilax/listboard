############################
# STEP 1 build executable binary
############################
FROM golang:alpine AS builder

# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git

WORKDIR $GOPATH/src/mypackage/myapp/

COPY . .

# Fetch dependencies.
# Using go get.
RUN go get -d -v

# Build the binary.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /go/bin/listboard

############################
# STEP 2 build a small image
############################

FROM scratch

ENV GO_ENV=production
ENV LB_CONFIG_FILE=config/listboard.json

WORKDIR /app

# Copy our static executable.
COPY --from=builder /go/bin/listboard /app/listboard

# Copy assets and config
COPY config /app/config
COPY public_html /app/public_html
COPY templates /app/templates
COPY translations /app/translations
COPY extra/* /app/

# Run the hello binary.
ENTRYPOINT ["/app/listboard"]
