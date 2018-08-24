# base image
FROM centos

MAINTAINER meission@aliyun.com 

RUN mkdir -p /data/app/sander
WORKDIR /data/app/sander
COPY ./bin ./bin
COPY ./config ./config
COPY ./data ./data 
COPY ./static ./static
COPY ./template ./template

CMD ["./bin/main"]

EXPOSE 8088

# docker run -it sander:latest /bin/bash
# docker run --name sander1 -p 127.0.0.1:8088:8088 -d sander
# docker exec -it sander1 /bin/bash