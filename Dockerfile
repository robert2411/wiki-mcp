FROM debian:12-slim AS git-layer
RUN apt-get update \
 && apt-get install -y --no-install-recommends git \
 && rm -rf /var/lib/apt/lists/* \
 && mkdir /git-libs \
 && find /usr/lib \( -name 'libz.so.1' -o -name 'libpcre2-8.so.0' \) -exec cp {} /git-libs/ \;

FROM gcr.io/distroless/base-debian12:nonroot
COPY --from=git-layer /usr/bin/git /usr/bin/git
COPY --from=git-layer /usr/lib/git-core /usr/lib/git-core
COPY --from=git-layer /usr/share/git-core /usr/share/git-core
COPY --from=git-layer /git-libs/ /usr/lib/

COPY wiki-mcp /wiki-mcp

EXPOSE 9000
VOLUME ["/wiki"]

ENTRYPOINT ["/wiki-mcp"]
CMD ["--serve-only", "--wiki-path=/wiki", "--bind=0.0.0.0", "--port=9000"]
