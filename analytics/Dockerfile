FROM matomo:3.5.1-fpm-alpine as matomo

FROM nginx:1.13.3-alpine

COPY --from=matomo /var/www/html /var/www/html

ADD nginx.conf /etc/nginx/nginx.conf

ENTRYPOINT ["nginx", "-c", "/etc/nginx/nginx.conf", "-g", "daemon off;"]
