FROM golang:1.8

WORKDIR /openapi

ENV TERRAFORM_VERSION=0.12.0

RUN apt-get update && \
    apt-get install unzip openssl ca-certificates && \
    cd /tmp && \
    wget https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip && \
    unzip terraform_${TERRAFORM_VERSION}_linux_amd64.zip -d /usr/bin && \
    rm -rf /tmp/* && \
    rm -rf /var/cache/apk/* && \
    rm -rf /var/tmp/*

COPY version /

# provision openapi plugins
RUN export PROVIDER_NAME=goa && curl -fsSL https://raw.githubusercontent.com/dikhan/terraform-provider-openapi/master/scripts/install.sh | bash -s -- --provider-name $PROVIDER_NAME
RUN export PROVIDER_NAME=swaggercodegen && curl -fsSL https://raw.githubusercontent.com/dikhan/terraform-provider-openapi/master/scripts/install.sh | bash -s -- --provider-name $PROVIDER_NAME

# copy examples including terraform configurations
COPY examples/ .

# move plugin config file set up with openapi providers configuration to terraform plugins folder
RUN mv terraform-provider-openapi.yaml /root/.terraform.d/plugins/

CMD ["bash"]