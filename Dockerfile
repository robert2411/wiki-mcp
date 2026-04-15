FROM gcr.io/distroless/static-debian12:nonroot

COPY wiki-mcp /wiki-mcp

EXPOSE 9000
VOLUME ["/wiki"]

ENTRYPOINT ["/wiki-mcp"]
CMD ["--serve-only", "--wiki-path=/wiki", "--bind=0.0.0.0", "--port=9000"]
