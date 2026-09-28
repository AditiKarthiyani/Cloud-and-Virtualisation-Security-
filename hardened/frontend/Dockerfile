FROM php:8.5-apache

RUN apt-get update \
    && apt-get upgrade -y  \
    && apt-get install -y --no-install-recommends \
        curl \
    && rm -rf /var/lib/apt/lists/*

COPY apache.conf /etc/apache2/sites-available/000-default.conf
COPY index.html style.css  /var/www/html/

RUN a2ensite 000-default.conf \
    && a2enmod rewrite

RUN chmod -R 544 /var/www/html/* \
 && find /var/www/html -type d -exec chmod 755 {} \; \   
 && chown -R www-data:www-data /var/www/html

EXPOSE 80
USER www-data

HEALTHCHECK --interval=50s --timeout=5s --retries=5 \
  CMD curl -f http://localhost/ || exit 1

CMD ["apache2ctl", "-D", "FOREGROUND"]


