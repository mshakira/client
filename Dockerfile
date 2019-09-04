FROM scratch
ADD bin/client /
CMD ["/client", "https://host.docker.internal/api/v1/list/incidents"]
