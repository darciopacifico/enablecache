FROM docker-registry.qa.vmcommerce.intra/docker/golang:1.4-godep

ENV http_proxy=http://proxy.qa.vmcommerce.intra:3128 \
    https_proxy=http://proxy.qa.vmcommerce.intra:3128 \
    no_proxy=localhost,vmcommerce.intra

ENV APP_PACKAGE=gitlab.wmxp.com.br/bis/biro
ENV APP_SRC_DIR=/go/src/${APP_PACKAGE}

COPY biro.docker.conf /etc/biro/biro.conf
COPY . $APP_SRC_DIR
WORKDIR $APP_SRC_DIR

RUN godep go install -v "$APP_PACKAGE"

ENTRYPOINT ["/go/bin/biro"]
CMD ["-c", "biro.conf"]
EXPOSE 8080
