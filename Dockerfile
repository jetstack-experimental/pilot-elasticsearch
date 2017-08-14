FROM docker.elastic.co/elasticsearch/elasticsearch:5.2.2

ADD pilot_linux_amd64 /usr/bin/pilot

ADD elasticsearch.yml /usr/share/elasticsearch/config/elasticsearch.yml
ADD log4j2.properties /usr/share/elasticsearch/config/log4j2.properties

ENTRYPOINT ["/usr/bin/pilot"]
