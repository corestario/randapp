FROM alpine:3.7

RUN apk update && \
    apk upgrade && \
    apk --no-cache add curl jq file bash

VOLUME [ /rd ]
WORKDIR /rd
EXPOSE 26656 26657
ENTRYPOINT ["/usr/bin/wrapper.sh"]
#CMD ["/bin/sh"]

#CMD ["rd", "collect-gentxs"]
CMD ["rd", "start"]

STOPSIGNAL SIGTERM

COPY wrapper.sh /usr/bin/wrapper.sh
#ENTRYPOINT ["/bin/sh"]

