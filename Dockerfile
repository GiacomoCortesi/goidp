FROM docker.io/golang:1.20.6-alpine3.17 AS builder

COPY . /app
WORKDIR /app/cmd

# Set build args
ARG APP_BUILD
ARG APP_NAME
ARG APP_VERSION
ARG CHART_VERSION
ARG API_VERSION

RUN CGO_ENABLED=0 go build \
    -ldflags="-X idp/controllers.AppVersion=${APP_VERSION} \
    -X idp/controllers.AppName=${APP_NAME} \
    -X idp/controllers.AppBuild=${APP_BUILD} \
    -X idp/controllers.ChartVersion=${CHART_VERSION} \
    -X idp/controllers.ApiVersion=${API_VERSION} " \
    -o /app/idp

# Test image
FROM builder AS test
COPY idp /app
WORKDIR /app
RUN CGO_ENABLED=0 go test ./...

FROM docker.io/alpine:3.12

COPY --from=builder /app/idp /app/

WORKDIR /app

# Set build args
ARG APP_BUILD
ARG APP_NAME
ARG APP_VERSION
ARG CHART_VERSION
ARG API_VERSION

# Set env from args
ENV APP_BUILD=$APP_BUILD
ENV APP_NAME=$APP_NAME
ENV APP_VERSION=$APP_VERSION
ENV CHART_VERSION=$CHART_VERSION
ENV API_VERSION=$API_VERSION

# Set labels
LABEL idp.name=${APP_NAME}
LABEL idp.chartversion=${CHART_VERSION}
LABEL idp.apiversion=${API_VERSION}
LABEL idp.version=${APP_VERSION}
LABEL idp.build=${APP_BUILD}

CMD ["./idp"]
