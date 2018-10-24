/**
 * geoip.c -- implementation of GeoIP
 *    ______      ___
 *   / ____/___  /   | _____________  __________
 *  / / __/ __ \/ /| |/ ___/ ___/ _ \/ ___/ ___/
 * / /_/ / /_/ / ___ / /__/ /__/  __(__  |__  )
 * \____/\____/_/  |_\___/\___/\___/____/____/
 *
 * The MIT License (MIT)
 * Copyright (c) 2009-2016 Gerardo Orellana <hello @ goaccess.io>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

#if HAVE_CONFIG_H
#include <config.h>
#endif

#ifdef HAVE_LIBGEOIP
#include <GeoIP.h>
#include <GeoIPCity.h>
#endif

#include "geoip1.h"

#include "error.h"
#include "util.h"

static GeoIP *geo_location_data;

/* Determine if we have a valid geoip resource.
 *
 * If the geoip resource is NULL, 0 is returned.
 * If the geoip resource is valid and malloc'd, 1 is returned. */
int
is_geoip_resource (void)
{
  return geo_location_data != NULL ? 1 : 0;
}

/* Free up GeoIP resources */
void
geoip_free (void)
{
  if (!is_geoip_resource ())
    return;

  GeoIP_delete (geo_location_data);
  GeoIP_cleanup ();
}

/* Open the given GeoLocation database and set its charset.
 *
 * On error, it aborts.
 * On success, a new geolocation structure is returned. */
static GeoIP *
geoip_open_db (const char *db)
{
  GeoIP *geoip;
  geoip = GeoIP_open (db, GEOIP_MEMORY_CACHE);

  if (geoip == NULL)
    FATAL ("Unable to open GeoIP database: %s\n", db);

  GeoIP_set_charset (geoip, GEOIP_CHARSET_UTF8);
  LOG_DEBUG (("Opened GeoIP City database: %s\n", db));

  return geoip;
}

/* Set up and open GeoIP database */
void
init_geoip (void)
{
  /* open custom city GeoIP database */
  if (conf.geoip_database != NULL)
    geo_location_data = geoip_open_db (conf.geoip_database);
  /* fall back to legacy GeoIP database */
  else
    geo_location_data = GeoIP_new (conf.geo_db);
}

/* Get continent name concatenated with code.
 *
 * If continent not found, "Unknown" is returned.
 * On success, the continent code & name is returned . */
static const char *
get_continent_name_and_code (const char *continentid)
{
  if (memcmp (continentid, "NA", 2) == 0)
    return "NA North America";
  else if (memcmp (continentid, "OC", 2) == 0)
    return "OC Oceania";
  else if (memcmp (continentid, "EU", 2) == 0)
    return "EU Europe";
  else if (memcmp (continentid, "SA", 2) == 0)
    return "SA South America";
  else if (memcmp (continentid, "AF", 2) == 0)
    return "AF Africa";
  else if (memcmp (continentid, "AN", 2) == 0)
    return "AN Antarctica";
  else if (memcmp (continentid, "AS", 2) == 0)
    return "AS Asia";
  else
    return "-- Unknown";
}

/* Compose a string with the country name and code and store it in the
 * given buffer. */
static void
geoip_set_country (const char *country, const char *code, char *loc)
{
  if (country && code)
    snprintf (loc, COUNTRY_LEN, "%s %s", code, country);
  else
    snprintf (loc, COUNTRY_LEN, "%s", "Unknown");
}

/* Compose a string with the city name and state/region and store it
 * in the given buffer. */
static void
geoip_set_city (const char *city, const char *region, char *loc)
{
  snprintf (loc, CITY_LEN, "%s, %s", city ? city : "N/A City",
            region ? region : "N/A Region");
}

/* Compose a string with the continent name and store it in the given
 * buffer. */
static void
geoip_set_continent (const char *continent, char *loc)
{
  if (continent)
    snprintf (loc, CONTINENT_LEN, "%s",
              get_continent_name_and_code (continent));
  else
    snprintf (loc, CONTINENT_LEN, "%s", "Unknown");
}

/* Get detailed information found in the GeoIP Database about the
 * given IPv4 or IPv6.
 *
 * On error, NULL is returned
 * On success, GeoIPRecord structure is returned */
static GeoIPRecord *
get_geoip_record (const char *addr, GTypeIP type_ip)
{
  GeoIPRecord *rec = NULL;

  if (TYPE_IPV4 == type_ip)
    rec = GeoIP_record_by_name (geo_location_data, addr);
  else if (TYPE_IPV6 == type_ip)
    rec = GeoIP_record_by_name_v6 (geo_location_data, addr);

  return rec;
}

/* Set country data by record into the given `location` buffer based
 * on the IP version. */
static void
geoip_set_country_by_record (const char *ip, char *location, GTypeIP type_ip)
{
  GeoIPRecord *rec = NULL;
  const char *country = NULL, *code = NULL, *addr = ip;

  if (conf.geoip_database == NULL || geo_location_data == NULL)
    return;

  /* Custom GeoIP database */
  if ((rec = get_geoip_record (addr, type_ip))) {
    country = rec->country_name;
    code = rec->country_code;
  }

  geoip_set_country (country, code, location);
  if (rec != NULL) {
    GeoIPRecord_delete (rec);
  }
}

/* Get the GeoIP location id by name.
 *
 * On error, 0 is returned
 * On success, the GeoIP location id is returned */
static int
geoip_get_geoid (const char *addr, GTypeIP type_ip)
{
  int geoid = 0;

  if (TYPE_IPV4 == type_ip)
    geoid = GeoIP_id_by_name (geo_location_data, addr);
  else if (TYPE_IPV6 == type_ip)
    geoid = GeoIP_id_by_name_v6 (geo_location_data, addr);

  return geoid;
}

/* Get the country name by GeoIP location id.
 *
 * On error, NULL is returned
 * On success, the country name is returned */
static const char *
geoip_get_country_by_geoid (const char *addr, GTypeIP type_ip)
{
  const char *country = NULL;

  if (TYPE_IPV4 == type_ip)
    country = GeoIP_country_name_by_name (geo_location_data, addr);
  else if (TYPE_IPV6 == type_ip)
    country = GeoIP_country_name_by_name_v6 (geo_location_data, addr);

  return country;
}

/* Set country data by geoid into the given `location` buffer based on
 * the IP version. */
static void
geoip_set_country_by_geoid (const char *ip, char *location, GTypeIP type_ip)
{
  const char *country = NULL, *code = NULL, *addr = ip;
  int geoid = 0;

  if (!is_geoip_resource ())
    return;

  if (!(country = geoip_get_country_by_geoid (addr, type_ip)))
    goto out;

  /* return two letter country code */
  if (!(geoid = geoip_get_geoid (addr, type_ip)))
    goto out;
  code = GeoIP_code_by_id (geoid);

out:
  geoip_set_country (country, code, location);
}

/* Set country data by geoid or record into the given `location` buffer
 * based on the IP version and currently used database edition.  */
void
geoip_get_country (const char *ip, char *location, GTypeIP type_ip)
{
  unsigned char rec = GeoIP_database_edition (geo_location_data);

  switch (rec) {
  case GEOIP_COUNTRY_EDITION:
    if (TYPE_IPV4 == type_ip)
      geoip_set_country_by_geoid (ip, location, TYPE_IPV4);
    else
      geoip_set_country (NULL, NULL, location);
    break;
  case GEOIP_COUNTRY_EDITION_V6:
    if (TYPE_IPV6 == type_ip)
      geoip_set_country_by_geoid (ip, location, TYPE_IPV6);
    else
      geoip_set_country (NULL, NULL, location);
    break;
  case GEOIP_CITY_EDITION_REV0:
  case GEOIP_CITY_EDITION_REV1:
    if (TYPE_IPV4 == type_ip)
      geoip_set_country_by_record (ip, location, TYPE_IPV4);
    else
      geoip_set_country (NULL, NULL, location);
    break;
  case GEOIP_CITY_EDITION_REV0_V6:
  case GEOIP_CITY_EDITION_REV1_V6:
    if (TYPE_IPV6 == type_ip)
      geoip_set_country_by_record (ip, location, TYPE_IPV6);
    else
      geoip_set_country (NULL, NULL, location);
    break;
  }
}

/* Set continent data by record into the given `location` buffer based
 * on the IP version. */
static void
geoip_set_continent_by_record (const char *ip, char *location, GTypeIP type_ip)
{
  GeoIPRecord *rec = NULL;
  const char *continent = NULL, *addr = ip;

  if (conf.geoip_database == NULL || geo_location_data == NULL)
    return;

  /* Custom GeoIP database */
  if ((rec = get_geoip_record (addr, type_ip)))
    continent = rec->continent_code;

  geoip_set_continent (continent, location);
  if (rec != NULL) {
    GeoIPRecord_delete (rec);
  }
}

/* Set continent data by geoid into the given `location` buffer based
 * on the IP version. */
static void
geoip_set_continent_by_geoid (const char *ip, char *location, GTypeIP type_ip)
{
  const char *continent = NULL, *addr = ip;
  int geoid = 0;

  if (!is_geoip_resource ())
    return;

  if (!(geoid = geoip_get_geoid (addr, type_ip)))
    goto out;
  continent = GeoIP_continent_by_id (geoid);

out:
  geoip_set_continent (continent, location);
}

/* Set continent data by geoid or record into the given `location` buffer
 * based on the IP version and currently used database edition.  */
void
geoip_get_continent (const char *ip, char *location, GTypeIP type_ip)
{
  unsigned char rec = GeoIP_database_edition (geo_location_data);

  switch (rec) {
  case GEOIP_COUNTRY_EDITION:
    if (TYPE_IPV4 == type_ip)
      geoip_set_continent_by_geoid (ip, location, TYPE_IPV4);
    else
      geoip_set_continent (NULL, location);
    break;
  case GEOIP_COUNTRY_EDITION_V6:
    if (TYPE_IPV6 == type_ip)
      geoip_set_continent_by_geoid (ip, location, TYPE_IPV6);
    else
      geoip_set_continent (NULL, location);
    break;
  case GEOIP_CITY_EDITION_REV0:
  case GEOIP_CITY_EDITION_REV1:
    if (TYPE_IPV4 == type_ip)
      geoip_set_continent_by_record (ip, location, TYPE_IPV4);
    else
      geoip_set_continent (NULL, location);
    break;
  case GEOIP_CITY_EDITION_REV0_V6:
  case GEOIP_CITY_EDITION_REV1_V6:
    if (TYPE_IPV6 == type_ip)
      geoip_set_continent_by_record (ip, location, TYPE_IPV6);
    else
      geoip_set_continent (NULL, location);
    break;
  }
}

/* Set city data by record into the given `location` buffer based on
 * the IP version.  */
static void
geoip_set_city_by_record (const char *ip, char *location, GTypeIP type_ip)
{
  GeoIPRecord *rec = NULL;
  const char *city = NULL, *region = NULL, *addr = ip;

  /* Custom GeoIP database */
  if ((rec = get_geoip_record (addr, type_ip))) {
    city = rec->city;
    region = rec->region;
  }

  geoip_set_city (city, region, location);
  if (rec != NULL) {
    GeoIPRecord_delete (rec);
  }
}

/* Set city data by geoid or record into the given `location` buffer
 * based on the IP version and currently used database edition.
 * It uses the custom GeoIP database - i.e., GeoLiteCity.dat */
void
geoip_get_city (const char *ip, char *location, GTypeIP type_ip)
{
  unsigned char rec = GeoIP_database_edition (geo_location_data);

  if (conf.geoip_database == NULL || geo_location_data == NULL)
    return;

  switch (rec) {
  case GEOIP_CITY_EDITION_REV0:
  case GEOIP_CITY_EDITION_REV1:
    if (TYPE_IPV4 == type_ip)
      geoip_set_city_by_record (ip, location, TYPE_IPV4);
    else
      geoip_set_city (NULL, NULL, location);
    break;
  case GEOIP_CITY_EDITION_REV0_V6:
  case GEOIP_CITY_EDITION_REV1_V6:
    if (TYPE_IPV6 == type_ip)
      geoip_set_city_by_record (ip, location, TYPE_IPV6);
    else
      geoip_set_city (NULL, NULL, location);
    break;
  }
}

/* Entry point to set GeoIP location into the corresponding buffers,
 * (continent, country, city).
 *
 * On error, 1 is returned
 * On success, buffers are set and 0 is returned */
int
set_geolocation (char *host, char *continent, char *country, char *city)
{
  int type_ip = 0;

  if (!is_geoip_resource ())
    return 1;

  if (invalid_ipaddr (host, &type_ip))
    return 1;

  geoip_get_country (host, country, type_ip);
  geoip_get_continent (host, continent, type_ip);
  if (conf.geoip_database)
    geoip_get_city (host, city, type_ip);

  return 0;
}
