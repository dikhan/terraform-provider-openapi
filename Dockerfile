FROM alpine:3.13

WORKDIR /openapi

ENV TERRAFORM_VERSION=0.13.7

RUN apk update && \
    apk add curl jq python3 bash ca-certificates git openssl unzip wget && \
    cd /tmp && \
    wget https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip && \
    unzip terraform_${TERRAFORM_VERSION}_linux_amd64.zip -d /usr/bin && \
    rm -rf /tmp/* && \
    rm -rf /var/cache/apk/* && \
    rm -rf /var/tmp/*

# provision openapi plugins
ENV PROVIDER_SOURCE_ADDRESS="terraform.example.com/examplecorp"
RUN export PROVIDER_NAME=goa && curl -fsSL https://raw.githubusercontent.com/dikhan/terraform-provider-openapi/master/scripts/install.sh | bash -s -- --provider-name $PROVIDER_NAME --provider-source-address ${PROVIDER_SOURCE_ADDRESS}
RUN export PROVIDER_NAME=swaggercodegen && curl -fsSL https://raw.githubusercontent.com/dikhan/terraform-provider-openapi/master/scripts/install.sh | bash -s -- --provider-name $PROVIDER_NAME --provider-source-address ${PROVIDER_SOURCE_ADDRESS}

# copy examples including terraform configurations
COPY examples/ .

# move plugin config file set up with openapi providers configuration to terraform plugins folder
RUN mv terraform-provider-openapi.yaml /root/.terraform.d/plugins/

CMD ["bash"]