FROM debian:12-slim AS git-layer
RUN apt-get update \
 && apt-get install -y --no-install-recommends git \
 && rm -rf /var/lib/apt/lists/*

FROM gcr.io/distroless/base-debian12:nonroot

# Copy git and its supporting files from the build layer so the optional
# git auto-commit feature works inside containers.
COPY --from=git-layer /usr/bin/git /usr/bin/git
COPY --from=git-layer /usr/lib/git-core /usr/lib/git-core
COPY --from=git-layer /usr/share/git-core /usr/share/git-core
COPY --from=git-layer /usr/lib/x86_64-linux-gnu/libz.so.1 /usr/lib/x86_64-linux-gnu/libz.so.1
COPY --from=git-layer /usr/lib/x86_64-linux-gnu/libpcre2-8.so.0 /usr/lib/x86_64-linux-gnu/libpcre2-8.so.0

COPY wiki-mcp /wiki-mcp

EXPOSE 9000
VOLUME ["/wiki"]

ENTRYPOINT ["/wiki-mcp"]
CMD ["--serve-only", "--wiki-path=/wiki", "--bind=0.0.0.0", "--port=9000"]
