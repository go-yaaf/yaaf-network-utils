FROM ubuntu:22.04

RUN apt-get update
RUN apt-get -yq install apt-utils
RUN apt-get -yq install ca-certificates && update-ca-certificates

RUN mkdir yaaf && cd yaaf

ADD yaaf-network-utils /yaaf/.

RUN chmod +x /yaaf/yaaf-network-utils

WORKDIR "/yaaf"

CMD ["./yaaf-network-utils"]
