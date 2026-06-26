# scenario-test-framework
#---------------------------------------------------------------------
FROM debian:stretch
ENV LANG=ja_JP.UTF-8
ENV TZ=Asia/Tokyo

RUN apt update

# based https://github.com/Ragnaroek/kcov_docker/blob/master/Dockerfile
# kcov build & install
RUN apt install -y binutils-dev libcurl4-openssl-dev \
                   zlib1g-dev libdw-dev libiberty-dev \
                   g++ pkg-config git cmake python-dev

# unstable latest version
ENV SRC_DIR=/home/kcov-src \
    URL_GIT_KCOV=https://github.com/SimonKagstrom/kcov.git
RUN git clone $URL_GIT_KCOV $SRC_DIR; \
    cd $SRC_DIR; \
    # default to latest tagged version
    DEFAULT_KCOV_GIT_REF=$(git tag --list | grep "^v[0-9]\+$" | sort -V | tail --lines 1); \
    KCOV_GIT_REF=${KCOV_GIT_REF:-$DEFAULT_KCOV_GIT_REF}; \
    git reset --hard $KCOV_GIT_REF;

RUN cd $SRC_DIR && \
    mkdir build && \
    cd build    && \
    cmake ..    && \
    make        && \
    make install

# shunit2
RUN git clone https://github.com/kward/shunit2.git

# for stfw
RUN apt install -y curl python-pip openjdk-8-jdk \
                   libyaml-dev 
RUN pip install pyaml docopt
RUN apt install -y ruby
RUN gem install serverspec

RUN adduser stfwuser
USER stfwuser

ENTRYPOINT ["kcov"]
CMD ["--help"]
