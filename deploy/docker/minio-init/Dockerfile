FROM minio/mc:latest

COPY deploy/docker/minio-init/init-buckets.sh /usr/local/bin/init-buckets.sh
RUN chmod +x /usr/local/bin/init-buckets.sh

ENTRYPOINT ["init-buckets.sh"]
