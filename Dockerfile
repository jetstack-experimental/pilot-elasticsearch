FROM docker.elastic.co/elasticsearch/elasticsearch:5.2.0

ADD sidecar/sidecar_linux_amd64 /usr/bin/sidecar

ADD elasticsearch.yml /usr/share/elasticsearch/config/elasticsearch.yml

ENTRYPOINT ["/usr/bin/sidecar"]
