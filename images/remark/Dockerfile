FROM docker.pkg.github.com/ngalaiko/server/base-alpine

ENV REMARK_VERSION=1.4.0 \
	REMARK_SHA256SUM=4f4d423689afba79e62ef5549804966b74258449f6f85b1c9072a7eeb00eab17

RUN apk add --no-cache --update \
    	curl \
	&& curl -OL "https://github.com/umputun/remark/releases/download/v${REMARK_VERSION}/remark42.linux-amd64.tar.gz" \
    && sha256sum remark42.linux-amd64.tar.gz \
    && echo "${REMARK_SHA256SUM}  remark42.linux-amd64.tar.gz" | sha256sum -c - \
    && tar -xvzf remark42.linux-amd64.tar.gz 

CMD ["./remark42.linux-amd64"]
