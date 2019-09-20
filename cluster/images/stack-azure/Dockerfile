FROM BASEIMAGE
RUN apk --no-cache add ca-certificates bash

ARG ARCH
ARG TINI_VERSION

ADD stack-azure /usr/local/bin/

COPY stack-package /

EXPOSE 8080
ENTRYPOINT ["stack-azure"]
