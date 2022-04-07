FROM BASEIMAGE
RUN apk --no-cache add ca-certificates bash

ARG ARCH
ARG TERRAFORM_VERSION
ARG TERRAFORM_PROVIDER_SOURCE
ARG TERRAFORM_PROVIDER_VERSION
ARG TINI_VERSION
ENV USER_ID=1001
ENV TERRAFORM_PROVIDER_VERSION=$TERRAFORM_PROVIDER_VERSION
ENV TERRAFORM_PROVIDER_SOURCE=$TERRAFORM_PROVIDER_SOURCE

# Setup Terraform environment
ENV TERRAFORM_VERSION=$TERRAFORM_VERSION
ENV PLUGIN_DIR /terraform/provider-mirror/registry.terraform.io/hashicorp/azurerm/${TERRAFORM_PROVIDER_VERSION}/linux_${ARCH}
ENV TF_CLI_CONFIG_FILE /terraform/.terraformrc
ENV TF_FORK 0

RUN mkdir -p ${PLUGIN_DIR}

ADD https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_${ARCH}.zip /tmp
ADD https://releases.hashicorp.com/terraform-provider-azurerm/${TERRAFORM_PROVIDER_VERSION}/terraform-provider-azurerm_${TERRAFORM_PROVIDER_VERSION}_linux_${ARCH}.zip /tmp
ADD terraformrc.hcl ${TF_CLI_CONFIG_FILE}

RUN unzip /tmp/terraform_${TERRAFORM_VERSION}_linux_${ARCH}.zip -d /usr/local/bin \
  && chmod +x /usr/local/bin/terraform \
  && rm /tmp/terraform_${TERRAFORM_VERSION}_linux_${ARCH}.zip \
  && unzip /tmp/terraform-provider-azurerm_${TERRAFORM_PROVIDER_VERSION}_linux_${ARCH}.zip -d ${PLUGIN_DIR} \
  && chmod +x ${PLUGIN_DIR}/* \
  && rm /tmp/terraform-provider-azurerm_${TERRAFORM_PROVIDER_VERSION}_linux_${ARCH}.zip \
  && chown -R ${USER_ID}:${USER_ID} /terraform
# End of - Setup Terraform environment

ADD provider /usr/local/bin/crossplane-jet-azure-provider

USER ${USER_ID}
EXPOSE 8080
ENTRYPOINT ["crossplane-jet-azure-provider"]
