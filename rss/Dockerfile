FROM clue/ttrss:latest

RUN apt-get update \
	&& DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
		curl 

RUN curl -LO https://github.com/dasmurphy/tinytinyrss-fever-plugin/archive/1.4.7.tar.gz \
	&& tar xzf 1.4.7.tar.gz \
	&& mv tinytinyrss-fever-plugin-1.4.7/fever /var/www/plugins/
