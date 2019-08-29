FROM lambci/lambda:build-go1.x

WORKDIR /opt

ENV NPM_CONFIG_USERCONFIG /opt/nodejs/.npmrc

RUN curl -sL https://rpm.nodesource.com/setup_12.x | bash - && \
    yum install -y nodejs && \
    mkdir bin nodejs && \
    cp /usr/bin/node bin/node && \
    cp -r /usr/lib/node_modules ./nodejs/node_modules && \
    npm config set cache /tmp/.npm && \
    npm config set init-module /tmp/.npm-init.js && \ 
    npm config set update-notifier false && \
    chmod a+r $NPM_CONFIG_USERCONFIG && \
    echo -e "#!/bin/sh\n/opt/nodejs/node_modules/npm/bin/npm-cli.js \$@" > ./bin/npm && \
    echo -e "#!/bin/sh\n/opt/nodejs/node_modules/npm/bin/npx-cli.js \$@" > ./bin/npx && \
    chmod a+x ./bin/npm && \
    chmod a+x ./bin/npx && \
    cp /usr/bin/docker ./bin/docker && \
    zip -yr /tmp/layer.zip ./*