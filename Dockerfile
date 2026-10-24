FROM alpine:3.23@sha256:ff7f0c727dd796aa36787bfcd9a64bf35a6ce72a73451b3359fde849b6703d63 AS trust
RUN apk add --no-cache 'ca-certificates>=20261020-r0' && update-ca-certificates

FROM gcr.io/distroless/static:nonroot@sha256:18f4cc26c9ffa7f20a359298be7deadac5c607c8e03a02b8119f1a5fc2f82c99
COPY --from=trust /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
ARG TARGETOS
ARG TARGETARCH
COPY ./dist/spegel_${TARGETOS}_${TARGETARCH}/spegel /
USER root:root
ENTRYPOINT ["/spegel"]
