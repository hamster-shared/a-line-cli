FROM ubuntu:latest
RUN apt update &&\
    apt install -y curl &&\
    curl -fsSL https://get.docker.com | bash -s docker --mirror Aliyun

ADD a-line-cli /usr/local/bin/

EXPOSE 8080

CMD ["/usr/local/bin/a-line-cli","daemon"]
