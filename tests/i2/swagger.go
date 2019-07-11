package i2

const cdnSwaggerYAMLTemplate = `swagger: "2.0"

host: %s 
schemes:
- "http"

paths:
  ######################
  #### CDN Resource ####
  ######################

  /v1/cdns:
    post:
      x-terraform-resource-name: "cdn"
      summary: "Create cdn"
      operationId: "ContentDeliveryNetworkCreateV1"
      parameters:
      - in: "body"
        name: "body"
        description: "Created CDN"
        required: true
        schema:
          $ref: "#/definitions/ContentDeliveryNetworkV1"
      responses:
        201:
          description: "successful operation"
          schema:
            $ref: "#/definitions/ContentDeliveryNetworkV1"

  /v1/cdns/{id}:
    get:
      summary: "Get cdn by id"
      description: ""
      operationId: "ContentDeliveryNetworkGetV1"
      parameters:
      - name: "id"
        in: "path"
        description: "The cdn id that needs to be fetched."
        required: true
        type: "string"
      responses:
        200:
          description: "successful operation"
          schema:
            $ref: "#/definitions/ContentDeliveryNetworkV1"

    put:
      summary: "Updated cdn"
      operationId: "ContentDeliveryNetworkUpdateV1"
      parameters:
      - name: "id"
        in: "path"
        description: "cdn that needs to be updated"
        required: true
        type: "string"
      - in: "body"
        name: "body"
        description: "Updated cdn object"
        required: true
        schema:
          $ref: "#/definitions/ContentDeliveryNetworkV1"
      responses:
        200:
          description: "successful operation"
          schema:
            $ref: "#/definitions/ContentDeliveryNetworkV1"
    delete:
      summary: "Delete cdn"
      operationId: "ContentDeliveryNetworkDeleteV1"
      parameters:
      - name: "id"
        in: "path"
        description: "The cdn that needs to be deleted"
        required: true
        type: "string"
      responses:
        204:
          description: "successful operation, no content is returned"

  ## CDN sub-resource

  /v1/cdns/{parent_id}/v1/firewalls:
    post:
      summary: "Create cdn firewall"
      operationId: "ContentDeliveryNetworkFirewallCreateV1"
      parameters:
      - name: "parent_id"
        in: "path"
        description: "The cdn id that contains the firewall to be fetched."
        required: true
        type: "string"
      - in: "body"
        name: "body"
        description: "Created CDN firewall"
        required: true
        schema:
          $ref: "#/definitions/ContentDeliveryNetworkFirewallV1"
      responses:
        201:
          description: "successful operation"
          schema:
            $ref: "#/definitions/ContentDeliveryNetworkFirewallV1"

  /v1/cdns/{parent_id}/v1/firewalls/{id}:
    get:
      summary: "Get cdn firewall by id"
      description: ""
      operationId: "ContentDeliveryNetworkFirewallGetV1"
      parameters:
      - name: "parent_id"
        in: "path"
        description: "The cdn id that contains the firewall to be fetched."
        required: true
        type: "string"
      - name: "id"
        in: "path"
        description: "The cdn firewall id that needs to be fetched."
        required: true
        type: "string"
      responses:
        200:
          description: "successful operation"
          schema:
            $ref: "#/definitions/ContentDeliveryNetworkFirewallV1"
    delete: 
      operationId: ContentDeliveryNetworkFirewallDeleteV1
      parameters: 
        - description: "The cdn id that contains the firewall to be fetched."
          in: path
          name: parent_id
          required: true
          type: string
        - description: "The cdn firewall id that needs to be fetched."
          in: path
          name: id
          required: true
          type: string
      responses: 
        204: 
          description: "successful operation, no content is returned"
      summary: "Delete firewall"


definitions:
  ContentDeliveryNetworkFirewallV1:
    type: "object"
    properties:
      id:
        type: "string"
        readOnly: true
      label:
        type: "string"
  ContentDeliveryNetworkV1:
    type: "object"
    required:
      - label
    properties:
      id:
        type: "string"
        readOnly: true
      label:
        type: "string"
`
