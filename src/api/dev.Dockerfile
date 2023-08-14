ARG BASE_IMAGE

RUN apt-get update -y && \
    apt-get upgrade -y && \
    apt-get install -y tabix && \
    rm -rf /var/lib/apt/lists/*

RUN npm install -g nodemon

WORKDIR /gohan_api

COPY run.dev.bash .
COPY nodemon.json .

CMD ["bash", "./run.dev.bash"]
