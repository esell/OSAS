FROM alpine:latest

RUN mkdir /bs
COPY conf.json /bs/
COPY scanner /bs/

CMD ["-c", "/bs/conf.json"]
ENTRYPOINT ["/bs/scanner"]
