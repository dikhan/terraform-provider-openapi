FROM golang:latest

ENV WORK_DIR "$GOPATH/src/github.com/dikhan/terraform-provider-api/service_provider_example"

RUN mkdir -p $WORK_DIR
ADD . $WORK_DIR
WORKDIR $WORK_DIR

RUN go get
RUN go build -o cdn-service-provider-api .

EXPOSE 80

CMD ["./cdn-service-provider-api"]